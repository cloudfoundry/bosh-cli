package docker

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"strings"
	"time"

	"github.com/anchore/stereoscope/internal/bus"
	"github.com/anchore/stereoscope/internal/log"
	"github.com/anchore/stereoscope/pkg/event"
	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/docker/cli/cli/config"
	configTypes "github.com/docker/cli/cli/config/types"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"
)

// DaemonImageProvider is a image.Provider capable of fetching and representing a docker image from the docker daemon API.
type DaemonImageProvider struct {
	imageStr  string
	tmpDirGen *file.TempDirGenerator
	client    client.APIClient
	platform  *image.Platform
}

// NewProviderFromDaemon creates a new provider instance for a specific image that will later be cached to the given directory.
func NewProviderFromDaemon(imgStr string, tmpDirGen *file.TempDirGenerator, c client.APIClient, platform *image.Platform) (*DaemonImageProvider, error) {
	ref, err := name.ParseReference(imgStr, name.WithDefaultRegistry(""))
	if err != nil {
		return nil, err
	}
	tag, ok := ref.(name.Tag)
	if ok {
		imgStr = tag.Name()
	}
	return &DaemonImageProvider{
		imageStr:  imgStr,
		tmpDirGen: tmpDirGen,
		client:    c,
		platform:  platform,
	}, nil
}

type daemonProvideProgress struct {
	SaveProgress *progress.TimedProgress
	CopyProgress *progress.Writer
	Stage        *progress.Stage
}

func (p *DaemonImageProvider) trackSaveProgress() (*daemonProvideProgress, error) {
	// fetch the expected image size to estimate and measure progress
	inspect, _, err := p.client.ImageInspectWithRaw(context.Background(), p.imageStr)
	if err != nil {
		return nil, fmt.Errorf("unable to inspect image: %w", err)
	}

	// docker image save clocks in at ~125MB/sec on my laptop... mileage may vary, of course :shrug:
	mb := math.Pow(2, 20)
	// "virtual size" is the total amount of disk-space used for the read-only image
	// data used by the container and the writable layer.
	// "size" (also provider by the inspect result) shows the amount of data (on disk)
	// that is used for the writable layer of each container.
	sec := float64(inspect.VirtualSize) / (mb * 125)
	approxSaveTime := time.Duration(sec*1000) * time.Millisecond

	estimateSaveProgress := progress.NewTimedProgress(approxSaveTime)
	copyProgress := progress.NewSizedWriter(inspect.VirtualSize)
	aggregateProgress := progress.NewAggregator(progress.NormalizeStrategy, estimateSaveProgress, copyProgress)

	// let consumers know of a monitorable event (image save + copy stages)
	stage := &progress.Stage{}

	bus.Publish(partybus.Event{
		Type:   event.FetchImage,
		Source: p.imageStr,
		Value: progress.StagedProgressable(&struct {
			progress.Stager
			*progress.Aggregator
		}{
			Stager:     progress.Stager(stage),
			Aggregator: aggregateProgress,
		}),
	})

	return &daemonProvideProgress{
		SaveProgress: estimateSaveProgress,
		CopyProgress: copyProgress,
		Stage:        stage,
	}, nil
}

// pull a docker image
func (p *DaemonImageProvider) pull(ctx context.Context) error {
	log.Debugf("pulling docker image=%q", p.imageStr)

	var status = newPullStatus()
	defer func() {
		status.complete = true
	}()

	// publish a pull event on the bus, allowing for read-only consumption of status
	bus.Publish(partybus.Event{
		Type:   event.PullDockerImage,
		Source: p.imageStr,
		Value:  status,
	})

	options, err := p.pullOptions()
	if err != nil {
		return err
	}

	resp, err := p.client.ImagePull(ctx, p.imageStr, options)
	if err != nil {
		return fmt.Errorf("pull failed: %w", err)
	}

	var thePullEvent *pullEvent
	decoder := json.NewDecoder(resp)
	for {
		if err := decoder.Decode(&thePullEvent); err != nil {
			if err == io.EOF {
				break
			}

			return fmt.Errorf("failed to pull image: %w", err)
		}

		// check for the last two events indicating the pull is complete
		if strings.HasPrefix(thePullEvent.Status, "Digest:") || strings.HasPrefix(thePullEvent.Status, "Status:") {
			continue
		}

		status.onEvent(thePullEvent)
	}

	return nil
}

func (p *DaemonImageProvider) pullOptions() (types.ImagePullOptions, error) {
	var options = types.ImagePullOptions{
		Platform: p.platform.String(),
	}

	// note: this will search the default config dir and allow for a DOCKER_CONFIG override
	cfg, err := config.Load("")
	if err != nil {
		return options, fmt.Errorf("failed to load docker config: %w", err)
	}
	log.Debugf("using docker config=%q", cfg.Filename)

	// get a URL that works with docker credential helpers
	url, err := authURL(p.imageStr, true)
	if err != nil {
		log.Warnf("failed to determine auth url from image=%q: %+v", p.imageStr, err)
		return options, nil
	}

	authConfig, err := cfg.GetAuthConfig(url)
	if err != nil {
		log.Warnf("failed to fetch registry auth (url=%s): %+v", url, err)
		return options, nil
	}

	empty := configTypes.AuthConfig{}
	if authConfig == empty {
		// we didn't find any entries in any auth sources. This might be because the workaround needed for the
		// docker credential helper was unnecessary (since the user isn't using a credential helper). For this reason
		// lets try this auth config lookup again, but this time for a url that doesn't consider the dockerhub
		// workaround for the credential helper.
		url, err = authURL(p.imageStr, false)
		if err != nil {
			log.Warnf("failed to determine auth url from image=%q: %+v", p.imageStr, err)
			return options, nil
		}

		authConfig, err = cfg.GetAuthConfig(url)
		if err != nil {
			log.Warnf("failed to fetch registry auth (url=%s): %+v", url, err)
			return options, nil
		}
	}

	log.Debugf("using docker credentials for %q", url)

	options.RegistryAuth, err = encodeCredentials(authConfig)
	if err != nil {
		log.Warnf("failed to encode registry auth (url=%s): %+v", url, err)
	}

	return options, nil
}

func authURL(imageStr string, dockerhubWorkaround bool) (string, error) {
	ref, err := name.ParseReference(imageStr)
	if err != nil {
		return "", err
	}

	url := ref.Context().RegistryStr()
	if dockerhubWorkaround && url == "index.docker.io" {
		// why do this? There is an upstream issue here: https://github.com/docker/docker-credential-helpers/blob/e595cd69465c6b0f7af2d49582b82fdeddecbf75/wincred/wincred_windows.go#L113-L127
		// where the hostname used for the auth config lookup requires this or else even pulling public images
		// will fail with auth related problems (bad username/password, bad personal access token, etc).
		// The above note only applies to the credential helper, not to auth entries directly written to the docker config.
		// For this reason callers need to try getting the authconfig for both v1 and non-v1 routes.
		url += "/v1/"
	}
	return url, nil
}

// Provide an image object that represents the cached docker image tar fetched from a docker daemon.
func (p *DaemonImageProvider) Provide(ctx context.Context, userMetadata ...image.AdditionalMetadata) (*image.Image, error) {
	if err := p.pullImageIfMissing(ctx); err != nil {
		return nil, err
	}

	// inspect the image that might have been pulled
	inspectResult, _, err := p.client.ImageInspectWithRaw(ctx, p.imageStr)
	if err != nil {
		return nil, fmt.Errorf("unable to inspect existing image: %w", err)
	}

	// by this point the platform info should match based off of user input, so we should error out if this is not the case
	if err := p.validatePlatform(inspectResult); err != nil {
		return nil, err
	}

	tarFileName, err := p.saveImage(ctx)
	if err != nil {
		return nil, err
	}

	// use the existing tarball provider to process what was pulled from the docker daemon
	return NewProviderFromTarball(tarFileName, p.tmpDirGen).Provide(ctx, withInspectMetadata(inspectResult, userMetadata)...)
}

func (p *DaemonImageProvider) saveImage(ctx context.Context) (string, error) {
	// save the image from the docker daemon to a tar file
	providerProgress, err := p.trackSaveProgress()
	if err != nil {
		return "", fmt.Errorf("unable to trace image save progress: %w", err)
	}
	defer func() {
		// NOTE: progress trackers should complete at the end of this function
		// whether the function errors or succeeds.
		providerProgress.SaveProgress.SetCompleted()
		providerProgress.CopyProgress.SetComplete()
	}()

	imageTempDir, err := p.tmpDirGen.NewDirectory("docker-daemon-image")
	if err != nil {
		return "", err
	}

	// create a file within the temp dir
	tempTarFile, err := os.Create(path.Join(imageTempDir, "image.tar"))
	if err != nil {
		return "", fmt.Errorf("unable to create temp file for image: %w", err)
	}
	defer func() {
		err := tempTarFile.Close()
		if err != nil {
			log.Errorf("unable to close temp file (%s): %w", tempTarFile.Name(), err)
		}
	}()

	providerProgress.Stage.Current = "requesting image from docker"
	readCloser, err := p.client.ImageSave(ctx, []string{p.imageStr})
	if err != nil {
		return "", fmt.Errorf("unable to save image tar: %w", err)
	}
	defer func() {
		err := readCloser.Close()
		if err != nil {
			log.Errorf("unable to close temp file (%s): %w", tempTarFile.Name(), err)
		}
	}()

	// NOTE: The image save progress is only a guess (a timer counting up to a particular time where
	// the overall progress would be considered at 50%). It's logical to adjust the first image save timer
	// to complete when the image save operation returns. The defer statement is a fallback in case the numbers
	// from the docker daemon don't line up (as we saw when metadata and actual size differ)
	// or there is a problem that causes us to return early with an error.
	providerProgress.SaveProgress.SetCompleted()

	// save the image contents to the temp file
	// note: this is the same image that will be used to querying image content during analysis
	providerProgress.Stage.Current = "saving image to disk"
	nBytes, err := io.Copy(io.MultiWriter(tempTarFile, providerProgress.CopyProgress), readCloser)
	if err != nil {
		return "", fmt.Errorf("unable to save image to tar: %w", err)
	}
	if nBytes == 0 {
		return "", errors.New("cannot provide an empty image")
	}
	return tempTarFile.Name(), nil
}

func (p *DaemonImageProvider) pullImageIfMissing(ctx context.Context) error {
	// check if the image exists locally
	inspectResult, _, err := p.client.ImageInspectWithRaw(ctx, p.imageStr)
	if err != nil {
		if client.IsErrNotFound(err) {
			if err = p.pull(ctx); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unable to inspect existing image: %w", err)
		}
	} else {
		// looks like the image exists, but if the platform doesn't match what the user specified, we may need to
		// pull the image again with the correct platofmr specifier, which will override the local tag.
		if err := p.validatePlatform(inspectResult); err != nil {
			if err = p.pull(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *DaemonImageProvider) validatePlatform(i types.ImageInspect) error {
	if p.platform == nil {
		// the user did not specify a platform
		return nil
	}

	if i.Os != p.platform.OS {
		return fmt.Errorf("image has unexpected OS %q, which differs from the user specified PS %q", i.Os, p.platform.OS)
	}

	if i.Architecture != p.platform.Architecture {
		return fmt.Errorf("image has unexpected architecture %q, which differs from the user specified architecture %q", i.Architecture, p.platform.Architecture)
	}

	// note: there is no architecture variant captured in inspect responses

	return nil
}

func withInspectMetadata(i types.ImageInspect, userMetadata []image.AdditionalMetadata) (metadata []image.AdditionalMetadata) {
	metadata = append(metadata,
		image.WithTags(i.RepoTags...),
		image.WithRepoDigests(i.RepoDigests...),
		image.WithArchitecture(i.Architecture, ""), // since we don't have variant info from the image directly, we don't report it
		image.WithOS(i.Os),
	)

	// apply user-supplied metadata last to override any default behavior
	metadata = append(metadata, userMetadata...)
	return metadata
}

func encodeCredentials(authConfig configTypes.AuthConfig) (string, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	// note: the contents may contain characters that should not be escaped (such as password contents)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(authConfig); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(buffer.Bytes()), nil
}
