package fakes

import (
	"fmt"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type InstallInput struct {
	Deployment bminstallmanifest.Manifest
	DirectorID string
}

type installOutput struct {
	cloud bmcloud.Cloud
	err   error
}

type ExtractInput struct {
	ReleaseTarballPath string
}

type extractCallback func(releaseTarballPath string) (bmrel.Release, error)

type FakeInstaller struct {
	InstallInputs   []InstallInput
	installBehavior map[string]installOutput
}

func NewFakeInstaller() *FakeInstaller {
	return &FakeInstaller{
		InstallInputs:   []InstallInput{},
		installBehavior: map[string]installOutput{},
	}
}

func (f *FakeInstaller) Install(deployment bminstallmanifest.Manifest, directorID string) (bmcloud.Cloud, error) {
	input := InstallInput{
		Deployment: deployment,
		DirectorID: directorID,
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
	return nil, fmt.Errorf("Unsupported Install Input: %s", value)
}

func (f *FakeInstaller) SetInstallBehavior(
	deployment bminstallmanifest.Manifest,
	directorID string,
	cloud bmcloud.Cloud,
	err error,
) error {
	input := InstallInput{
		Deployment: deployment,
		DirectorID: directorID,
	}

	value, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return fmt.Errorf("Could not serialize input %#v", input)
	}
	f.installBehavior[value] = installOutput{cloud: cloud, err: err}
	return nil
}
