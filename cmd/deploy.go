package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	semver "github.com/cppforlife/go-semi-semantic/version"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type DeployCmd struct {
	ui               boshui.UI
	deployment       boshdir.Deployment
	uploadReleaseCmd ReleaseUploadingCmd
}

type ReleaseUploadingCmd interface {
	Run(UploadReleaseOpts) error
}

func NewDeployCmd(ui boshui.UI, deployment boshdir.Deployment, uploadReleaseCmd ReleaseUploadingCmd) DeployCmd {
	return DeployCmd{ui: ui, deployment: deployment, uploadReleaseCmd: uploadReleaseCmd}
}

func (c DeployCmd) Run(opts DeployOpts) error {
	tpl := boshtpl.NewTemplate(opts.Args.Manifest.Bytes)

	bytes, err := tpl.Evaluate(opts.VarFlags.AsVariables())
	if err != nil {
		return bosherr.WrapErrorf(err, "Evaluating manifest")
	}

	man, err := boshdir.NewManifestFromBytes(bytes)
	if err != nil {
		return bosherr.WrapErrorf(err, "Checking manifest")
	}

	if man.Name != c.deployment.Name() {
		errMsg := "Expected manifest to specify deployment name '%s' but was '%s'"
		return bosherr.Errorf(errMsg, c.deployment.Name(), man.Name)
	}

	err = c.printManifestDiff(bytes, opts)
	if err != nil {
		return bosherr.WrapError(err, "Diffing manifest")
	}

	err = c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	for _, rel := range man.Releases {
		err = c.uploadRelease(rel)
		if err != nil {
			return bosherr.WrapErrorf(err, "Uploading release '%s/%s'", rel.Name, rel.Version)
		}
	}

	updateOpts := boshdir.UpdateOpts{
		Recreate:  opts.Recreate,
		SkipDrain: opts.SkipDrain,
	}

	return c.deployment.Update(bytes, updateOpts)
}

func (c DeployCmd) uploadRelease(rel boshdir.ManifestRelease) error {
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

func (c DeployCmd) printManifestDiff(bytes []byte, opts DeployOpts) error {
	diff, err := c.deployment.Diff(bytes, opts.NoRedact)
	if err != nil {
		return err
	}

	for _, line := range diff {
		lineMod, _ := line[1].(string)

		if lineMod == "added" {
			c.ui.BeginLinef("+ %s\n", line[0])
		} else if lineMod == "removed" {
			c.ui.BeginLinef("- %s\n", line[0])
		} else {
			c.ui.BeginLinef("  %s\n", line[0])
		}
	}

	return nil
}
