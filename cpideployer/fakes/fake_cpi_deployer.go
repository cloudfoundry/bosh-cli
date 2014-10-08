package fakes

import (
	"fmt"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type DeployInput struct {
	Deployment         bmdepl.Deployment
	ReleaseTarballPath string
}

type deployOutput struct {
	cloud bmcloud.Cloud
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

func (f *FakeCpiDeployer) Deploy(deployment bmdepl.Deployment, releaseTarballPath string) (bmcloud.Cloud, error) {
	input := DeployInput{
		Deployment:         deployment,
		ReleaseTarballPath: releaseTarballPath,
	}
	f.DeployInputs = append(f.DeployInputs, input)

	value, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return nil, fmt.Errorf("Could not serialize input %#v", input)
	}

	output, found := f.deployBehavior[value]
	if found {
		return output.cloud, output.err
	}
	return nil, fmt.Errorf("Unsupported Input: %s", value)
}

func (f *FakeCpiDeployer) SetDeployBehavior(
	deployment bmdepl.Deployment,
	releaseTarballPath string,
	cloud bmcloud.Cloud,
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
