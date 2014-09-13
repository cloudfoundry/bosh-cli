package fakes

import (
	"fmt"

	bmdeploy "github.com/cloudfoundry/bosh-micro-cli/deployer"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type DeployInput struct {
	Deployment         bmdepl.Deployment
	ReleaseTarballPath string
}

type deployOutput struct {
	cloud bmdeploy.Cloud
	err   error
}

type FakeCpiDeployer struct {
	DeployInputs   []DeployInput
	deployBehavior map[string]deployOutput
}

func NewFakeCpiDeployer() *FakeCpiDeployer {
	return &FakeCpiDeployer{
		DeployInputs:   []DeployInput{},
		deployBehavior: map[string]deployOutput{},
	}
}

func (f *FakeCpiDeployer) Deploy(deployment bmdepl.Deployment, releaseTarballPath string) (bmdeploy.Cloud, error) {
	input := DeployInput{
		Deployment:         deployment,
		ReleaseTarballPath: releaseTarballPath,
	}
	f.DeployInputs = append(f.DeployInputs, input)

	value, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return bmdeploy.Cloud{}, fmt.Errorf("Could not serialize input %#v", input)
	}

	output, found := f.deployBehavior[value]
	if found {
		return output.cloud, output.err
	}
	return bmdeploy.Cloud{}, fmt.Errorf("Unsupported Input: %s", value)
}

func (f *FakeCpiDeployer) SetDeployBehavior(
	deployment bmdepl.Deployment,
	releaseTarballPath string,
	cloud bmdeploy.Cloud,
	err error,
) error {
	input := DeployInput{
		Deployment:         deployment,
		ReleaseTarballPath: releaseTarballPath,
	}

	value, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return fmt.Errorf("Could not serialize input %#v", input)
	}
	f.deployBehavior[value] = deployOutput{cloud: cloud, err: err}
	return nil
}
