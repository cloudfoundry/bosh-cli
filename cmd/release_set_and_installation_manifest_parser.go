package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	boshtpl "github.com/cloudfoundry/bosh-init/director/template"
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	birelsetmanifest "github.com/cloudfoundry/bosh-init/release/set/manifest"
)

type ReleaseSetAndInstallationManifestParser struct {
	ReleaseSetParser   birelsetmanifest.Parser
	InstallationParser biinstallmanifest.Parser
}

func (y ReleaseSetAndInstallationManifestParser) ReleaseSetAndInstallationManifest(deploymentManifestPath string, vars boshtpl.Variables) (birelsetmanifest.Manifest, biinstallmanifest.Manifest, error) {
	releaseSetManifest, err := y.ReleaseSetParser.Parse(deploymentManifestPath, vars)
	if err != nil {
		return birelsetmanifest.Manifest{}, biinstallmanifest.Manifest{}, bosherr.WrapErrorf(err, "Parsing release set manifest '%s'", deploymentManifestPath)
	}

	installationManifest, err := y.InstallationParser.Parse(deploymentManifestPath, vars, releaseSetManifest)
	if err != nil {
		return birelsetmanifest.Manifest{}, biinstallmanifest.Manifest{}, bosherr.WrapErrorf(err, "Parsing installation manifest '%s'", deploymentManifestPath)
	}

	return releaseSetManifest, installationManifest, nil
}
