package deployer

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
)

type Deployment interface {
	Deploy(
		cloud bmcloud.Cloud,
		extractedStemcell bmstemcell.ExtractedStemcell,
		registryConfig bminstallmanifest.Registry,
		sshTunnelConfig bminstallmanifest.SSHTunnel,
		vmManager bmvm.Manager,
	) error
}

type deployment struct {
	manifest bmdeplmanifest.Manifest
	deployer Deployer
}

func NewDeployment(manifest bmdeplmanifest.Manifest, deployer Deployer) Deployment {
	return &deployment{
		manifest: manifest,
		deployer: deployer,
	}
}

func (d *deployment) Manifest() bmdeplmanifest.Manifest {
	return d.manifest
}

func (d deployment) Deploy(
	cloud bmcloud.Cloud,
	extractedStemcell bmstemcell.ExtractedStemcell,
	registryConfig bminstallmanifest.Registry,
	sshTunnelConfig bminstallmanifest.SSHTunnel,
	vmManager bmvm.Manager,
) error {
	return d.deployer.Deploy(
		cloud,
		d.manifest,
		extractedStemcell,
		registryConfig,
		sshTunnelConfig,
		vmManager,
	)
}
