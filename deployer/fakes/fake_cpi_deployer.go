package fakes

import (
	"fmt"

	bmdeploy "github.com/cloudfoundry/bosh-micro-cli/deployer"
)

type DeployInput struct {
	DeploymentManifestPath string
	ReleaseTarballPath     string
}

type deployOutput struct {
	cloud bmdeploy.Cloud
	err   error
}

type FakeCpiDeployer struct {
	DeployInputs   []DeployInput
	deployBehavior map[DeployInput]deployOutput
}

func NewFakeCpiDeployer() *FakeCpiDeployer {
	return &FakeCpiDeployer{
		DeployInputs:   []DeployInput{},
		deployBehavior: map[DeployInput]deployOutput{},
	}
}

func (f *FakeCpiDeployer) ParseManifest() string {
	deploymentManifestPath := ""
	return deploymentManifestPath
}

func (f *FakeCpiDeployer) Deploy(deploymentManifestPath string, releaseTarballPath string) (bmdeploy.Cloud, error) {
	input := DeployInput{
		DeploymentManifestPath: deploymentManifestPath,
		ReleaseTarballPath:     releaseTarballPath,
	}
	f.DeployInputs = append(f.DeployInputs, input)
	output, found := f.deployBehavior[input]

	if found {
		return output.cloud, output.err
	}
	return bmdeploy.Cloud{}, fmt.Errorf("Unsupported Input: Deploy('%s', '%s')", deploymentManifestPath, releaseTarballPath)
}

func (f *FakeCpiDeployer) SetDeployBehavior(
	deploymentManifestPath string,
	releaseTarballPath string,
	cloud bmdeploy.Cloud,
	err error,
) {
	input := DeployInput{
		DeploymentManifestPath: deploymentManifestPath,
		ReleaseTarballPath:     releaseTarballPath,
	}
	f.deployBehavior[input] = deployOutput{cloud: cloud, err: err}
}
