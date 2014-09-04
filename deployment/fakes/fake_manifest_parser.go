package fakes

import (
	"fmt"

	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type parseInput struct {
	DeploymentPath string
}

type parseOutput struct {
	localDeployment bmdepl.LocalDeployment
	err             error
}

type FakeManifestParser struct {
	ParseInputs   []parseInput
	parseBehavior map[parseInput]parseOutput
}

func NewFakeManifestParser() *FakeManifestParser {
	return &FakeManifestParser{
		ParseInputs:   []parseInput{},
		parseBehavior: map[parseInput]parseOutput{},
	}
}

func (f *FakeManifestParser) Parse(deploymentPath string) (bmdepl.LocalDeployment, error) {
	input := parseInput{DeploymentPath: deploymentPath}
	f.ParseInputs = append(f.ParseInputs, input)
	output, found := f.parseBehavior[input]

	if found {
		return output.localDeployment, output.err
	}

	return bmdepl.LocalDeployment{}, fmt.Errorf("Unsupported Input: Parse('%s')", deploymentPath)
}

func (f *FakeManifestParser) SetParseBehavior(deploymentPath string, deployment bmdepl.LocalDeployment, err error) {
	f.parseBehavior[parseInput{DeploymentPath: deploymentPath}] = parseOutput{localDeployment: deployment, err: err}
}
