package deployer

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type Deployment interface {
	Deploy(
		cloud bmcloud.Cloud,
		extractedStemcell bmstemcell.ExtractedStemcell,
		registryConfig bmdepl.Registry,
		sshTunnelConfig bmdepl.SSHTunnel,
		mbusURL string,
	) error
}

type deployment struct {
	manifest bmdepl.Manifest
	deployer Deployer
}

func NewDeployment(manifest bmdepl.Manifest, deployer Deployer) Deployment {
	return &deployment{
		manifest: manifest,
		deployer: deployer,
	}
}

func (d *deployment) Manifest() bmdepl.Manifest {
	return d.manifest
}

func (d deployment) Deploy(
	cloud bmcloud.Cloud,
	extractedStemcell bmstemcell.ExtractedStemcell,
	registryConfig bmdepl.Registry,
	sshTunnelConfig bmdepl.SSHTunnel,
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
