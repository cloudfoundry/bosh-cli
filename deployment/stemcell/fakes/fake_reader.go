package fakes

import (
	"fmt"

	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
)

type ReadInput struct {
	StemcellTarballPath string
	DestPath            string
}

type ReadOutput struct {
	stemcell bmstemcell.ExtractedStemcell
	err      error
}

type FakeStemcellReader struct {
	ReadBehavior map[ReadInput]ReadOutput
	ReadInputs   []ReadInput
}

func NewFakeReader() *FakeStemcellReader {
	return &FakeStemcellReader{
		ReadBehavior: map[ReadInput]ReadOutput{},
		ReadInputs:   []ReadInput{},
	}
}

func (fr *FakeStemcellReader) Read(stemcellTarballPath, destPath string) (bmstemcell.ExtractedStemcell, error) {
	input := ReadInput{
		StemcellTarballPath: stemcellTarballPath,
		DestPath:            destPath,
	}
	fr.ReadInputs = append(fr.ReadInputs, input)
	output, found := fr.ReadBehavior[input]
	if !found {
		return nil, fmt.Errorf("Unsupported Input: Read('%#v', '%#v')", stemcellTarballPath, destPath)
	}

	return output.stemcell, output.err
}

func (fr *FakeStemcellReader) SetReadBehavior(stemcellTarballPath, destPath string, stemcell bmstemcell.ExtractedStemcell, err error) {
	input := ReadInput{
		StemcellTarballPath: stemcellTarballPath,
		DestPath:            destPath,
	}
	fr.ReadBehavior[input] = ReadOutput{stemcell: stemcell, err: err}
}
