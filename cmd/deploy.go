package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"gopkg.in/yaml.v3"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type DeployCmd struct {
	ui              boshui.UI
	deployment      boshdir.Deployment
	releaseUploader ReleaseUploader
	director        boshdir.Director
}

type ReleaseUploader interface {
	UploadReleases([]byte) ([]byte, error)
	UploadReleasesWithFix([]byte) ([]byte, error)
}

type Conf struct {
	Flags              []string `yaml:"flags"`
	IncludeDeployments []string `yaml:"include"`
	ExcludeDeployments []string `yaml:"exclude"`
}

func NewDeployCmd(
	ui boshui.UI,
	deployment boshdir.Deployment,
	releaseUploader ReleaseUploader,
	director boshdir.Director,
) DeployCmd {
	return DeployCmd{ui, deployment, releaseUploader, director}
}

func (c DeployCmd) Run(opts DeployOpts) error {
	tpl := boshtpl.NewTemplate(opts.Args.Manifest.Bytes)

	configs, _ := c.director.ListConfigs(1, boshdir.ConfigsFilter{Type: "deploy"})

	for _, config := range configs {
		var conf Conf

		err := yaml.Unmarshal([]byte(config.Content), &conf)
		if err != nil {
			return err
		}

		deploymentIncluded := applies(conf.IncludeDeployments, c.deployment.Name())
		deploymentExcluded := applies(conf.ExcludeDeployments, c.deployment.Name())

		if conf.ExcludeDeployments != nil &&
			conf.IncludeDeployments != nil {
			c.ui.PrintLinef("Ignoring deployment flags from config of type '%s' (name: '%s'). Please use only 'include'- OR 'exclude'-property in the config.", config.Type, config.Name)
		} else {
			if (conf.IncludeDeployments == nil && conf.ExcludeDeployments == nil) ||
				deploymentIncluded ||
				(!deploymentExcluded && conf.ExcludeDeployments != nil) {
				c.ui.ErrorLinef("Using deployment flags from config of type '%s' (name: '%s')", config.Type, config.Name)

				opts = setFlags(conf.Flags, opts)
			}
		}
	}

	bytes, err := tpl.Evaluate(opts.VarFlags.AsVariables(), opts.OpsFlags.AsOp(), boshtpl.EvaluateOpts{})
	if err != nil {
		return bosherr.WrapErrorf(err, "Evaluating manifest")
	}

	err = c.checkDeploymentName(bytes)
	if err != nil {
		return err
	}

	if opts.FixReleases {
		bytes, err = c.releaseUploader.UploadReleasesWithFix(bytes)
	} else {
		bytes, err = c.releaseUploader.UploadReleases(bytes)
	}
	if err != nil {
		return err
	}

	deploymentDiff, err := c.deployment.Diff(bytes, opts.NoRedact)
	if err != nil {
		return err
	}

	diff := NewDiff(deploymentDiff.Diff)
	diff.Print(c.ui)

	err = c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	updateOpts := boshdir.UpdateOpts{
		RecreatePersistentDisks: opts.RecreatePersistentDisks,
		Recreate:                opts.Recreate,
		Fix:                     opts.Fix,
		SkipDrain:               opts.SkipDrain,
		DryRun:                  opts.DryRun,
		Canaries:                opts.Canaries,
		MaxInFlight:             opts.MaxInFlight,
		Diff:                    deploymentDiff,
		ForceLatestVariables:    opts.ForceLatestVariables,
	}

	return c.deployment.Update(bytes, updateOpts)
}

func setFlags(flags []string, opts DeployOpts) DeployOpts {
	for j := range flags {
		switch flags[j] {
		case "fix-releases":
			opts.FixReleases = true
		case "fix":
			opts.Fix = true
		case "recreate":
			opts.Recreate = true
		case "recreate-persistent-disks":
			opts.RecreatePersistentDisks = true
		}
	}

	return opts
}

func applies(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

func (c DeployCmd) checkDeploymentName(bytes []byte) error {
	manifest, err := boshdir.NewManifestFromBytes(bytes)
	if err != nil {
		return bosherr.WrapErrorf(err, "Parsing manifest")
	}

	if manifest.Name != c.deployment.Name() {
		errMsg := "Expected manifest to specify deployment name '%s' but was '%s'"
		return bosherr.Errorf(errMsg, c.deployment.Name(), manifest.Name)
	}

	return nil
}
