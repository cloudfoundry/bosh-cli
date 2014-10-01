package fakes

import (
	"fmt"

	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/vm"
)

type CreateVMInput struct {
	StemcellCID bmstemcell.CID
}

type createVMOutput struct {
	err error
}

type FakeManager struct {
	CreateVMInputs   []CreateVMInput
	createVMBehavior map[CreateVMInput]createVMOutput
}

func NewFakeManager() *FakeManager {
	return &FakeManager{
		CreateVMInputs:   []CreateVMInput{},
		createVMBehavior: map[CreateVMInput]createVMOutput{},
	}
}

func (m *FakeManager) CreateVM(stemcellCID bmstemcell.CID) (bmvm.CID, error) {
	input := CreateVMInput{
		StemcellCID: stemcellCID,
	}
	m.CreateVMInputs = append(m.CreateVMInputs, input)
	output, found := m.createVMBehavior[input]
	if !found {
		return "", fmt.Errorf("Unsupported CreateVM Input: %s", stemcellCID)
	}

	return "", output.err
}

func (m *FakeManager) SetCreateVMBehavior(cid bmstemcell.CID, err error) {
	input := CreateVMInput{
		StemcellCID: cid,
	}
	m.createVMBehavior[input] = createVMOutput{err: err}
}
