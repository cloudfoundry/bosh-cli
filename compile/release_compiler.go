package compile

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmdep "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtemcomp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

type ReleaseCompiler interface {
	Compile(release bmrel.Release, manifestPath string) error
}

type releaseCompiler struct {
	packagesCompiler  ReleasePackagesCompiler
	manifestParser    bmdep.ManifestParser
	templatesCompiler bmtemcomp.TemplatesCompiler
}

func NewReleaseCompiler(
	packagesCompiler ReleasePackagesCompiler,
	manifestParser bmdep.ManifestParser,
	templatesCompiler bmtemcomp.TemplatesCompiler,
) ReleaseCompiler {
	return releaseCompiler{
		packagesCompiler:  packagesCompiler,
		manifestParser:    manifestParser,
		templatesCompiler: templatesCompiler,
	}
}

func (c releaseCompiler) Compile(release bmrel.Release, manifestPath string) error {
	err := c.packagesCompiler.Compile(release)
	if err != nil {
		return bosherr.WrapError(err, "Compiling release packages")
	}

	deployment, err := c.manifestParser.Parse(manifestPath)
	if err != nil {
		return bosherr.WrapError(err, "Parsing the deployment manifest")
	}

	err = c.templatesCompiler.Compile(release.Jobs, deployment)
	if err != nil {
		return bosherr.WrapError(err, "Compiling job templates")
	}
	return nil
}
