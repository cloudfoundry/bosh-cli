package deployer

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
)

type Deployment interface {
	Deploy(
		cloud bmcloud.Cloud,
		extractedStemcell bmstemcell.ExtractedStemcell,
		registryConfig bmmanifest.Registry,
		sshTunnelConfig bmmanifest.SSHTunnel,
		mbusURL string,
	) error
}

type deployment struct {
	manifest bmmanifest.Manifest
	deployer Deployer
}

func NewDeployment(manifest bmmanifest.Manifest, deployer Deployer) Deployment {
	return &deployment{
		manifest: manifest,
		deployer: deployer,
	}
}

func (d *deployment) Manifest() bmmanifest.Manifest {
	return d.manifest
}

func (d deployment) Deploy(
	cloud bmcloud.Cloud,
	extractedStemcell bmstemcell.ExtractedStemcell,
	registryConfig bmmanifest.Registry,
	sshTunnelConfig bmmanifest.SSHTunnel,
	mbusURL string,
) error {
	return d.deployer.Deploy(
		cloud,
		d.manifest,
		extractedStemcell,
		registryConfig,
		sshTunnelConfig,
		mbusURL,
	)
}
