package fakes

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bminstall "github.com/cloudfoundry/bosh-micro-cli/cpideployer/install"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type NewCloudInput struct {
	InstalledJobs []bminstall.InstalledJob
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

func (f *FakeFactory) NewCloud(installedJobs []bminstall.InstalledJob) (bmcloud.Cloud, error) {
	input := NewCloudInput{
		InstalledJobs: installedJobs,
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

func (f *FakeFactory) SetNewCloudBehavior(installedJobs []bminstall.InstalledJob, cloud bmcloud.Cloud, err error) {
	input := NewCloudInput{
		InstalledJobs: installedJobs,
	}

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		panic(bosherr.WrapError(marshalErr, "Marshaling NewCloud input"))
	}

	f.newCloudBehavior[inputString] = newCloudOutput{cloud: cloud, err: err}
}
