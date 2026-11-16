package cmd

import (
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	boshreldir "github.com/cloudfoundry/bosh-cli/v7/releasedir"
)

type InitReleaseCmd struct {
	releaseDir boshreldir.ReleaseDir
}

func NewInitReleaseCmd(releaseDir boshreldir.ReleaseDir) InitReleaseCmd {
	return InitReleaseCmd{releaseDir: releaseDir}
}

// initReleaseDir resolves where a release should be initialized. When --dir is
// omitted (empty Path) and a NAME is given, the release is created in ./NAME.
// An explicit --dir always names the release root itself.
func initReleaseDir(dir DirOrCWDArg, name string) (DirOrCWDArg, error) {
	explicit := len(dir.Path) > 0

	if !explicit {
		cwd, err := os.Getwd()
		if err != nil {
			return dir, bosherr.WrapError(err, "Getting working directory")
		}

		dir.Path = cwd
	}

	if !explicit && len(name) > 0 {
		dir.Path = filepath.Join(dir.Path, name)
	}

	return dir, nil
}

func (c InitReleaseCmd) Run(opts InitReleaseOpts) error {
	if !opts.Sample && len(opts.Blobstore) == 0 && (len(opts.BlobstoreOptions) > 0 || len(opts.BlobstorePrivateOptions) > 0) {
		return bosherr.Error("Cannot configure blobstore options without --blobstore")
	}

	blobstoreOptions, err := parseKeyValueOptions(opts.BlobstoreOptions)
	if err != nil {
		return err
	}

	privateBlobstoreOptions, err := parseKeyValueOptions(opts.BlobstorePrivateOptions)
	if err != nil {
		return err
	}

	return c.releaseDir.Init(boshreldir.InitOptions{
		Name:   opts.Args.Name,
		Git:    opts.Git,
		Sample: opts.Sample,
		Force:  opts.Force,

		BlobstoreProvider:       opts.Blobstore,
		BlobstoreOptions:        blobstoreOptions,
		BlobstorePrivateOptions: privateBlobstoreOptions,
	})
}
