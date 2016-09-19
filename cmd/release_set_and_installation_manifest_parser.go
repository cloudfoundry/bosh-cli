package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/cppforlife/go-patch/patch"

	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/installation/manifest"
	birelsetmanifest "github.com/cloudfoundry/bosh-cli/release/set/manifest"
)

type ReleaseSetAndInstallationManifestParser struct {
	ReleaseSetParser   birelsetmanifest.Parser
	InstallationParser biinstallmanifest.Parser
}

func (y ReleaseSetAndInstallationManifestParser) ReleaseSetAndInstallationManifest(deploymentManifestPath string, vars boshtpl.Variables, ops patch.Ops) (birelsetmanifest.Manifest, biinstallmanifest.Manifest, error) {
	releaseSetManifest, err := y.ReleaseSetParser.Parse(deploymentManifestPath, vars, ops)
	if err != nil {
		return birelsetmanifest.Manifest{}, biinstallmanifest.Manifest{}, bosherr.WrapErrorf(err, "Parsing release set manifest '%s'", deploymentManifestPath)
	}

	installationManifest, err := y.InstallationParser.Parse(deploymentManifestPath, vars, ops, releaseSetManifest)
	if err != nil {
		return birelsetmanifest.Manifest{}, biinstallmanifest.Manifest{}, bosherr.WrapErrorf(err, "Parsing installation manifest '%s'", deploymentManifestPath)
	}

	return releaseSetManifest, installationManifest, nil
}
