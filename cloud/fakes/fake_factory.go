package fakes

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmcpiinstall "github.com/cloudfoundry/bosh-micro-cli/cpi/install"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type NewCloudInput struct {
	InstalledCPIJob bmcpiinstall.InstalledJob
	DirectorID      string
}

type newCloudOutput struct {
	cloud bmcloud.Cloud
	err   error
}

type FakeFactory struct {
	NewCloudInputs   []NewCloudInput
	newCloudBehavior map[string]newCloudOutput
}

func NewFakeFactory() *FakeFactory {
	return &FakeFactory{
		NewCloudInputs:   []NewCloudInput{},
		newCloudBehavior: map[string]newCloudOutput{},
	}
}

func (f *FakeFactory) NewCloud(installedCPIJob bmcpiinstall.InstalledJob, directorID string) (bmcloud.Cloud, error) {
	input := NewCloudInput{
		InstalledCPIJob: installedCPIJob,
		DirectorID:      directorID,
	}
	f.NewCloudInputs = append(f.NewCloudInputs, input)

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		panic(bosherr.WrapError(marshalErr, "Marshaling NewCloud input"))
	}

	output, found := f.newCloudBehavior[inputString]
	if !found {
		panic(fmt.Errorf("Unsupported NewCloud Input: %#v\nExpected Behavior: %#v", input, f.newCloudBehavior))
	}

	return output.cloud, output.err
}

func (f *FakeFactory) SetNewCloudBehavior(installedCPIJob bmcpiinstall.InstalledJob, directorID string, cloud bmcloud.Cloud, err error) {
	input := NewCloudInput{
		InstalledCPIJob: installedCPIJob,
		DirectorID:      directorID,
	}

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		panic(bosherr.WrapError(marshalErr, "Marshaling NewCloud input"))
	}

	f.newCloudBehavior[inputString] = newCloudOutput{cloud: cloud, err: err}
}
