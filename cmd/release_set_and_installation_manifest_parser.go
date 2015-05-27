package cmd

import (
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	birelsetmanifest "github.com/cloudfoundry/bosh-init/release/set/manifest"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type releaseSetAndInstallationManifestParser struct {
	releaseSetParser   birelsetmanifest.Parser
	installationParser biinstallmanifest.Parser
}

func (y releaseSetAndInstallationManifestParser) ReleaseSetAndInstallationManifest(deploymentManifestPath string) (birelsetmanifest.Manifest, biinstallmanifest.Manifest, error) {
	releaseSetManifest, err := y.releaseSetParser.Parse(deploymentManifestPath)
	if err != nil {
		return birelsetmanifest.Manifest{}, biinstallmanifest.Manifest{}, bosherr.WrapErrorf(err, "Parsing release set manifest '%s'", deploymentManifestPath)
	}

	installationManifest, err := y.installationParser.Parse(deploymentManifestPath, releaseSetManifest)
	if err != nil {
		return birelsetmanifest.Manifest{}, biinstallmanifest.Manifest{}, bosherr.WrapErrorf(err, "Parsing installation manifest '%s'", deploymentManifestPath)
	}
	return releaseSetManifest, installationManifest, nil
}
