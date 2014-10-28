package fakes

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bminsup "github.com/cloudfoundry/bosh-micro-cli/microdeployer/instanceupdater"
	bmretrystrategy "github.com/cloudfoundry/bosh-micro-cli/retrystrategy"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type DeployInput struct {
	StemcellCID            bmstemcell.CID
	Cloud                  bmcloud.Cloud
	Deployment             bmdepl.Deployment
	Registry               bmdepl.Registry
	SSHTunnelConfig        bmdepl.SSHTunnel
	AgentPingRetryStrategy bmretrystrategy.RetryStrategy
	InstanceUpdater        bminsup.InstanceUpdater
}

type deployOutput struct {
	err error
}

type FakeMicroDeployer struct {
	DeployInput  DeployInput
	DeployOutput deployOutput
}

func NewFakeMicroDeployer() *FakeMicroDeployer {
	return &FakeMicroDeployer{
		DeployInput: DeployInput{},
	}
}

func (m *FakeMicroDeployer) Deploy(
	cloud bmcloud.Cloud,
	deployment bmdepl.Deployment,
	registry bmdepl.Registry,
	sshTunnelConfig bmdepl.SSHTunnel,
	agentPingRetryStrategy bmretrystrategy.RetryStrategy,
	stemcellCID bmstemcell.CID,
	instanceUpdater bminsup.InstanceUpdater,
) error {
	input := DeployInput{
		StemcellCID:            stemcellCID,
		Cloud:                  cloud,
		Deployment:             deployment,
		Registry:               registry,
		SSHTunnelConfig:        sshTunnelConfig,
		AgentPingRetryStrategy: agentPingRetryStrategy,
		InstanceUpdater:        instanceUpdater,
	}
	m.DeployInput = input

	if (m.DeployOutput != deployOutput{}) {
		return m.DeployOutput.err
	}

	return nil
}

func (m *FakeMicroDeployer) SetDeployBehavior(err error) {
	m.DeployOutput = deployOutput{err: err}
}
