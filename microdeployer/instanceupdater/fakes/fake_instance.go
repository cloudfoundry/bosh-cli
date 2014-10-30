package fakes

import (
	"time"

	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type FakeInstance struct {
	ApplyInputs []ApplyInput
	ApplyErr    error

	StartCalled bool
	StartErr    error

	WaitInputs       []WaitInput
	WaitToBeReadyErr error
}

type ApplyInput struct {
	StemcellApplySpec bmstemcell.ApplySpec
	Deployment        bmdepl.Deployment
}

type WaitInput struct {
	MaxAttempts int
	Delay       time.Duration
}

func NewFakeInstance() *FakeInstance {
	return &FakeInstance{
		ApplyInputs: []ApplyInput{},
		WaitInputs:  []WaitInput{},
	}
}

func (i *FakeInstance) WaitToBeReady(maxAttempts int, delay time.Duration) error {
	i.WaitInputs = append(i.WaitInputs, WaitInput{
		MaxAttempts: maxAttempts,
		Delay:       delay,
	})
	return i.WaitToBeReadyErr
}

func (i *FakeInstance) Apply(stemcellApplySpec bmstemcell.ApplySpec, deployment bmdepl.Deployment) error {
	i.ApplyInputs = append(i.ApplyInputs, ApplyInput{
		StemcellApplySpec: stemcellApplySpec,
		Deployment:        deployment,
	})

	return i.ApplyErr
}

func (i *FakeInstance) Start() error {
	i.StartCalled = true

	return i.StartErr
}
