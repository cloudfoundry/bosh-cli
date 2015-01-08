package fakes

import (
	"fmt"
)

type ExtractInput struct {
	BlobID    string
	BlobSHA1  string
	TargetDir string
}
type extractOutput struct {
	err error
}

type FakeExtractor struct {
	ExtractInputs   []ExtractInput
	extractBehavior map[ExtractInput]extractOutput
}

func NewFakeExtractor() *FakeExtractor {
	return &FakeExtractor{
		ExtractInputs:   []ExtractInput{},
		extractBehavior: map[ExtractInput]extractOutput{},
	}
}

func (f *FakeExtractor) Extract(blobID string, blobSHA1 string, targetDir string) error {
	input := ExtractInput{BlobID: blobID, BlobSHA1: blobSHA1, TargetDir: targetDir}
	f.ExtractInputs = append(f.ExtractInputs, input)
	output, found := f.extractBehavior[input]

	if found {
		return output.err
	}
	return fmt.Errorf("Unsupported Input: Extract('%s', '%s', '%s')", blobID, blobSHA1, targetDir)
}

func (f *FakeExtractor) SetExtractBehavior(blobID string, blobSHA1 string, targetDir string, err error) {
	f.extractBehavior[ExtractInput{BlobID: blobID, BlobSHA1: blobSHA1, TargetDir: targetDir}] = extractOutput{err: err}
}
