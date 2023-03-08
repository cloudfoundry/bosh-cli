package stereoscope

import (
	"context"
	"fmt"

	"github.com/anchore/go-logger"

	"github.com/anchore/stereoscope/internal/bus"
	dockerClient "github.com/anchore/stereoscope/internal/docker"
	"github.com/anchore/stereoscope/internal/log"
	"github.com/anchore/stereoscope/internal/podman"
	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/anchore/stereoscope/pkg/image/docker"
	"github.com/anchore/stereoscope/pkg/image/oci"
	"github.com/anchore/stereoscope/pkg/image/sif"
	"github.com/wagoodman/go-partybus"
)

var rootTempDirGenerator = file.NewTempDirGenerator("stereoscope")

func WithRegistryOptions(options image.RegistryOptions) Option {
	return func(c *config) error {
		c.Registry = options
		return nil
	}
}

func WithInsecureSkipTLSVerify() Option {
	return func(c *config) error {
		c.Registry.InsecureSkipTLSVerify = true
		return nil
	}
}

func WithInsecureAllowHTTP() Option {
	return func(c *config) error {
		c.Registry.InsecureUseHTTP = true
		return nil
	}
}

func WithCredentials(credentials ...image.RegistryCredentials) Option {
	return func(c *config) error {
		c.Registry.Credentials = append(c.Registry.Credentials, credentials...)
		return nil
	}
}

func WithAdditionalMetadata(metadata ...image.AdditionalMetadata) Option {
	return func(c *config) error {
		c.AdditionalMetadata = append(c.AdditionalMetadata, metadata...)
		return nil
	}
}

func WithPlatform(platform string) Option {
	return func(c *config) error {
		p, err := image.NewPlatform(platform)
		if err != nil {
			return err
		}
		c.Platform = p
		return nil
	}
}

// GetImageFromSource returns an image from the explicitly provided source.
func GetImageFromSource(ctx context.Context, imgStr string, source image.Source, options ...Option) (*image.Image, error) {
	log.Debugf("image: source=%+v location=%+v", source, imgStr)

	var cfg config
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(&cfg); err != nil {
			return nil, fmt.Errorf("unable to parse option: %w", err)
		}
	}

	provider, err := selectImageProvider(imgStr, source, cfg)
	if err != nil {
		return nil, err
	}

	img, err := provider.Provide(ctx, cfg.AdditionalMetadata...)
	if err != nil {
		return nil, fmt.Errorf("unable to use %s source: %w", source, err)
	}

	err = img.Read()
	if err != nil {
		return nil, fmt.Errorf("could not read image: %+v", err)
	}

	return img, nil
}

func selectImageProvider(imgStr string, source image.Source, cfg config) (image.Provider, error) {
	var provider image.Provider
	tempDirGenerator := rootTempDirGenerator.NewGenerator()
	platformSelectionUnsupported := fmt.Errorf("specified platform=%q however image source=%q does not support selecting platform", cfg.Platform.String(), source.String())

	switch source {
	case image.DockerTarballSource:
		if cfg.Platform != nil {
			return nil, platformSelectionUnsupported
		}
		// note: the imgStr is the path on disk to the tar file
		provider = docker.NewProviderFromTarball(imgStr, tempDirGenerator)
	case image.DockerDaemonSource:
		c, err := dockerClient.GetClient()
		if err != nil {
			return nil, err
		}
		provider, err = docker.NewProviderFromDaemon(imgStr, tempDirGenerator, c, cfg.Platform)
		if err != nil {
			return nil, err
		}
	case image.PodmanDaemonSource:
		c, err := podman.GetClient()
		if err != nil {
			return nil, err
		}
		provider, err = docker.NewProviderFromDaemon(imgStr, tempDirGenerator, c, cfg.Platform)
		if err != nil {
			return nil, err
		}
	case image.OciDirectorySource:
		if cfg.Platform != nil {
			return nil, platformSelectionUnsupported
		}
		provider = oci.NewProviderFromPath(imgStr, tempDirGenerator)
	case image.OciTarballSource:
		if cfg.Platform != nil {
			return nil, platformSelectionUnsupported
		}
		provider = oci.NewProviderFromTarball(imgStr, tempDirGenerator)
	case image.OciRegistrySource:
		provider = oci.NewProviderFromRegistry(imgStr, tempDirGenerator, cfg.Registry, cfg.Platform)
	case image.SingularitySource:
		if cfg.Platform != nil {
			return nil, platformSelectionUnsupported
		}
		provider = sif.NewProviderFromPath(imgStr, tempDirGenerator)
	default:
		return nil, fmt.Errorf("unable determine image source")
	}
	return provider, nil
}

// GetImage parses the user provided image string and provides an image object;
// note: the source where the image should be referenced from is automatically inferred.
func GetImage(ctx context.Context, userStr string, options ...Option) (*image.Image, error) {
	source, imgStr, err := image.DetectSource(userStr)
	if err != nil {
		return nil, err
	}
	return GetImageFromSource(ctx, imgStr, source, options...)
}

func SetLogger(logger logger.Logger) {
	log.Log = logger
}

func SetBus(b *partybus.Bus) {
	bus.SetPublisher(b)
}

// Cleanup deletes all directories created by stereoscope calls. Note: please use image.Image.Cleanup() over this
// function when possible.
func Cleanup() {
	if err := rootTempDirGenerator.Cleanup(); err != nil {
		log.Errorf("failed to cleanup tempdir root: %w", err)
	}
}
