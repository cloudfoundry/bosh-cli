package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	semver "github.com/cppforlife/go-semi-semantic/version"

	boshdir "github.com/cloudfoundry/bosh-init/director"
	boshtpl "github.com/cloudfoundry/bosh-init/director/template"
	boshui "github.com/cloudfoundry/bosh-init/ui"
)

type UpdateRuntimeConfigCmd struct {
	ui               boshui.UI
	director         boshdir.Director
	uploadReleaseCmd ReleaseUploadingCmd
}

func NewUpdateRuntimeConfigCmd(ui boshui.UI, director boshdir.Director, uploadReleaseCmd ReleaseUploadingCmd) UpdateRuntimeConfigCmd {
	return UpdateRuntimeConfigCmd{ui: ui, director: director, uploadReleaseCmd: uploadReleaseCmd}
}

func (c UpdateRuntimeConfigCmd) Run(opts UpdateRuntimeConfigOpts) error {
	tpl := boshtpl.NewTemplate(opts.Args.RuntimeConfig.Bytes)

	interpolatedRuntimeConfig, err := tpl.Evaluate(opts.VarFlags.AsVariables())
	if err != nil {
		return bosherr.WrapErrorf(err, "Evaluating runtime config")
	}

	runtimeConfigBytes := interpolatedRuntimeConfig.Content()
	rc, err := boshdir.NewRuntimeConfigManifestFromBytes(runtimeConfigBytes)
	if err != nil {
		return bosherr.WrapErrorf(err, "Checking runtime config")
	}

	err = c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	for _, rel := range rc.Releases {
		err = c.uploadRelease(rel)
		if err != nil {
			return bosherr.WrapErrorf(err, "Uploading release '%s/%s'", rel.Name, rel.Version)
		}
	}

	return c.director.UpdateRuntimeConfig(runtimeConfigBytes)
}

func (c UpdateRuntimeConfigCmd) uploadRelease(rel boshdir.RuntimeConfigManifestRelease) error {
	ver, err := semver.NewVersionFromString(rel.Version)
	if err != nil {
		return err
	}

	opts := UploadReleaseOpts{
		Name:    rel.Name,
		Version: VersionArg(ver),

		Args: UploadReleaseArgs{URL: URLArg(rel.URL)},
		SHA1: rel.SHA1,
	}

	if opts.Args.URL.IsEmpty() {
		return nil
	}

	return c.uploadReleaseCmd.Run(opts)
}
