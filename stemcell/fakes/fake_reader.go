package fakes

import (
	"fmt"

	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type ReadInput struct {
	stemcellPath string
	destPath     string
}

type ReadOutput struct {
	stemcell bmstemcell.Stemcell
	err      error
}

type FakeStemcellReader struct {
	ReadBehavior map[ReadInput]ReadOutput
}

func NewFakeReader() *FakeStemcellReader {
	return &FakeStemcellReader{ReadBehavior: map[ReadInput]ReadOutput{}}
}

func (fr *FakeStemcellReader) Read(stemcellPath, destPath string) (bmstemcell.Stemcell, error) {
	input := ReadInput{
		stemcellPath: stemcellPath,
		destPath:     destPath,
	}
	output, found := fr.ReadBehavior[input]
	if !found {
		return bmstemcell.Stemcell{}, fmt.Errorf("Unsupported Input: Read('%#v', '%#v')", stemcellPath, destPath)
	}

	return output.stemcell, output.err
}

func (fr *FakeStemcellReader) SetReadBehavior(stemcellPath, destPath string, stemcell bmstemcell.Stemcell, err error) {
	input := ReadInput{
		stemcellPath: stemcellPath,
		destPath:     destPath,
	}
	fr.ReadBehavior[input] = ReadOutput{stemcell: stemcell, err: err}
}
