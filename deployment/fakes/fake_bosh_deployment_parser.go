package fakes

import (
	"fmt"

	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type parseInput struct {
	DeploymentPath string
}

type parseOutput struct {
	deployment bmdepl.Deployment
	err        error
}

type FakeBoshDeploymentParser struct {
	ParseInputs   []parseInput
	parseBehavior map[parseInput]parseOutput
}

func NewFakeBoshDeploymentParser() *FakeBoshDeploymentParser {
	return &FakeBoshDeploymentParser{
		ParseInputs:   []parseInput{},
		parseBehavior: map[parseInput]parseOutput{},
	}
}

func (f *FakeBoshDeploymentParser) Parse(deploymentPath string) (bmdepl.Deployment, error) {
	input := parseInput{DeploymentPath: deploymentPath}
	f.ParseInputs = append(f.ParseInputs, input)
	output, found := f.parseBehavior[input]

	if found {
		return output.deployment, output.err
	}

	return NewFakeDeployment(), fmt.Errorf("Unsupported Input: Parse('%s'), available behaviors '%#v'", deploymentPath, f.parseBehavior)
}

func (f *FakeBoshDeploymentParser) SetParseBehavior(deploymentPath string, deployment bmdepl.Deployment, err error) {
	f.parseBehavior[parseInput{DeploymentPath: deploymentPath}] = parseOutput{deployment: deployment, err: err}
}
