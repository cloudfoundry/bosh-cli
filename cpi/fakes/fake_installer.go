package fakes

import (
	"fmt"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type InstallInput struct {
	Deployment bmdepl.CPIDeployment
	Release    bmrel.Release
}

type installOutput struct {
	cloud bmcloud.Cloud
	err   error
}

type ExtractInput struct {
	ReleaseTarballPath string
}

type extractOutput struct {
	release bmrel.Release
	err     error
}

type FakeInstaller struct {
	InstallInputs   []InstallInput
	installBehavior map[string]installOutput

	ExtractInputs   []ExtractInput
	extractBehavior map[string]extractOutput
}

func NewFakeInstaller() *FakeInstaller {
	return &FakeInstaller{
		InstallInputs:   []InstallInput{},
		installBehavior: map[string]installOutput{},

		ExtractInputs:   []ExtractInput{},
		extractBehavior: map[string]extractOutput{},
	}
}

func (f *FakeInstaller) Extract(releaseTarballPath string) (bmrel.Release, error) {
	input := ExtractInput{
		ReleaseTarballPath: releaseTarballPath,
	}
	f.ExtractInputs = append(f.ExtractInputs, input)

	value, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return nil, fmt.Errorf("Could not serialize input %#v", input)
	}

	output, found := f.extractBehavior[value]
	if found {
		return output.release, output.err
	}
	return nil, fmt.Errorf("Unsupported Input: %s", value)
}

func (f *FakeInstaller) Install(deployment bmdepl.CPIDeployment, release bmrel.Release) (bmcloud.Cloud, error) {
	input := InstallInput{
		Deployment: deployment,
		Release:    release,
	}
	f.InstallInputs = append(f.InstallInputs, input)

	value, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return nil, fmt.Errorf("Could not serialize input %#v", input)
	}

	output, found := f.installBehavior[value]
	if found {
		return output.cloud, output.err
	}
	return nil, fmt.Errorf("Unsupported Input: %s", value)
}

func (f *FakeInstaller) SetInstallBehavior(
	deployment bmdepl.CPIDeployment,
	release bmrel.Release,
	cloud bmcloud.Cloud,
	err error,
) error {
	input := InstallInput{
		Deployment: deployment,
		Release:    release,
	}

	value, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return fmt.Errorf("Could not serialize input %#v", input)
	}
	f.installBehavior[value] = installOutput{cloud: cloud, err: err}
	return nil
}

func (f *FakeInstaller) SetExtractBehavior(
	releaseTarballPath string,
	release bmrel.Release,
	err error,
) error {
	input := ExtractInput{
		ReleaseTarballPath: releaseTarballPath,
	}

	value, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return fmt.Errorf("Could not serialize input %#v", input)
	}
	f.extractBehavior[value] = extractOutput{release: release, err: err}
	return nil
}
