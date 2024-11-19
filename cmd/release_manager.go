package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/cloudfoundry/bosh-utils/work"
	"github.com/cppforlife/go-patch/patch"
	semver "github.com/cppforlife/go-semi-semantic/version"

	cmdopts "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
)

type ReleaseManager struct {
	createReleaseCmd ReleaseCreatingCmd
	uploadReleaseCmd ReleaseUploadingCmd
	parallelThreads  int
	uploadWithFix    bool
}

type ReleaseUploadingCmd interface {
	Run(cmdopts.UploadReleaseOpts) error
}

type ReleaseCreatingCmd interface {
	Run(cmdopts.CreateReleaseOpts) (boshrel.Release, error)
}

func NewReleaseManager(
	createReleaseCmd ReleaseCreatingCmd,
	uploadReleaseCmd ReleaseUploadingCmd,
	parallelThreads int,
) ReleaseManager {
	return ReleaseManager{createReleaseCmd, uploadReleaseCmd, parallelThreads, false}
}

func (m ReleaseManager) UploadReleasesWithFix(bytes []byte) ([]byte, error) {
	m.uploadWithFix = true
	return m.UploadReleases(bytes)
}

func (m ReleaseManager) UploadReleases(bytes []byte) ([]byte, error) {
	manifest, err := boshdir.NewManifestFromBytes(bytes)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Parsing manifest")
	}

	opss, err := m.parallelCreateAndUpload(manifest)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Creating and uploading releases")
	}

	tpl := boshtpl.NewTemplate(bytes)

	bytes, err = tpl.Evaluate(boshtpl.StaticVariables{}, opss, boshtpl.EvaluateOpts{})
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Updating manifest with created release versions")
	}

	return bytes, nil
}

func (m ReleaseManager) parallelCreateAndUpload(manifest boshdir.Manifest) (patch.Ops, error) {
	pool := work.Pool{
		Count: m.parallelThreads,
	}

	patchOpsChan := make(chan patch.Ops, len(manifest.Releases))
	tasks := []func() error{}
	for _, r := range manifest.Releases {
		release := r
		tasks = append(tasks, func() error {
			patchOps, err := m.createAndUploadRelease(release)
			if err != nil {
				return err
			}
			patchOpsChan <- patchOps
			return nil
		})
	}

	err := pool.ParallelDo(tasks...)
	if err != nil {
		return nil, err
	}
	close(patchOpsChan)

	var opss patch.Ops
	for result := range patchOpsChan {
		opss = append(opss, result)
	}

	return opss, nil
}

func (m ReleaseManager) createAndUploadRelease(rel boshdir.ManifestRelease) (patch.Ops, error) {
	var ops patch.Ops

	if len(rel.URL) == 0 {
		return nil, nil
	}

	ver, err := semver.NewVersionFromString(rel.Version)
	if err != nil {
		return nil, err
	}

	uploadOpts := cmdopts.UploadReleaseOpts{
		Name:    rel.Name,
		Version: cmdopts.VersionArg(ver),

		Args: cmdopts.UploadReleaseArgs{URL: cmdopts.URLArg(rel.URL)},
		SHA1: rel.SHA1,
		Fix:  m.uploadWithFix,
	}
	if len(rel.ExportedFrom) > 0 {
		// https://bosh.io/docs/locking-compiled-releases/#why-an-array It is an array but we only use the first item.
		uploadOpts.Stemcell = boshdir.NewOSVersionSlug(rel.ExportedFrom[0].OS, rel.ExportedFrom[0].Version)
	}
	if len(rel.Stemcell.OS) > 0 {
		uploadOpts.Stemcell = boshdir.NewOSVersionSlug(rel.Stemcell.OS, rel.Stemcell.Version)
	}

	if rel.Version == "create" {
		createOpts := cmdopts.CreateReleaseOpts{
			Name:             rel.Name,
			Directory:        cmdopts.DirOrCWDArg{Path: uploadOpts.Args.URL.FilePath()},
			TimestampVersion: true,
			Force:            true,
		}

		release, err := m.createReleaseCmd.Run(createOpts)
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Processing release '%s/%s'", rel.Name, rel.Version)
		}

		uploadOpts = cmdopts.UploadReleaseOpts{Release: release}

		replaceOp := patch.ReplaceOp{
			// equivalent to /releases/name=?/version
			Path: patch.NewPointer([]patch.Token{
				patch.RootToken{},
				patch.KeyToken{Key: "releases"},
				patch.MatchingIndexToken{Key: "name", Value: rel.Name},
				patch.KeyToken{Key: "version"},
			}),
			Value: release.Version(),
		}

		removeUrlOp := patch.RemoveOp{
			Path: patch.NewPointer([]patch.Token{
				patch.RootToken{},
				patch.KeyToken{Key: "releases"},
				patch.MatchingIndexToken{Key: "name", Value: rel.Name},
				patch.KeyToken{Key: "url"},
			}),
		}

		ops = append(ops, replaceOp, removeUrlOp)
	}

	err = m.uploadReleaseCmd.Run(uploadOpts)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Uploading release '%s/%s'", rel.Name, rel.Version)
	}

	return ops, nil
}
