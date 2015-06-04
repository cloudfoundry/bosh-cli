package fakes

import (
	"fmt"

	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
)

type decompressInput struct {
	srcFile string
	destDir string
}
type decompressOutput struct {
	err error
}
type compressInput struct {
	srcDir string
}
type compressOutput struct {
	destFile string
	err      error
}
type cleanUpInput struct {
	tarballPath string
}
type cleanUpOutput struct {
	err error
}

type decompressError error

type FakeMultiResponseExtractor struct {
	decompressBehavior map[decompressInput]decompressError
	compressBehavior   map[compressInput]compressOutput
	cleanUpBehavior    map[cleanUpInput]cleanUpOutput
	decompressedInputs []decompressInput
}

func NewFakeMultiResponseExtractor() *FakeMultiResponseExtractor {
	return &FakeMultiResponseExtractor{decompressBehavior: map[decompressInput]decompressError{}}
}

func (e *FakeMultiResponseExtractor) DecompressFileToDir(srcFile, destDir string, options boshcmd.CompressorOptions) error {
	decompressInput := decompressInput{srcFile: srcFile, destDir: destDir}
	decompressError := e.decompressBehavior[decompressInput]

	if decompressError != nil {
		return decompressError
	}

	e.decompressedInputs = append(e.decompressedInputs, decompressInput)

	return nil
}

func (e *FakeMultiResponseExtractor) SetDecompressBehavior(srcFile, destDir string, err decompressError) {
	e.decompressBehavior[decompressInput{srcFile: srcFile, destDir: destDir}] = err
}

func (e *FakeMultiResponseExtractor) DecompressedFiles() []string {
	files := make([]string, 0, len(e.decompressedInputs))
	for _, decompressedInput := range e.decompressedInputs {
		file := fmt.Sprintf("%s/%s", decompressedInput.destDir, decompressedInput.srcFile)
		files = append(files, file)
	}
	return files
}

func (e *FakeMultiResponseExtractor) CompressFilesInDir(srcDir string) (string, error) {
	input := compressInput{srcDir: srcDir}
	output, found := e.compressBehavior[input]

	if found {
		return output.destFile, output.err
	}
	return "", fmt.Errorf("Unsupported Input: CompressFilesInDir('%s')\nAvailable inputs: %#v", srcDir, e.compressBehavior)
}

func (e *FakeMultiResponseExtractor) SetCompressBehavior(srcDir, destFile string, err error) {
	e.compressBehavior[compressInput{srcDir: srcDir}] = compressOutput{destFile: destFile, err: err}
}

func (e *FakeMultiResponseExtractor) CleanUp(tarballPath string) error {
	input := cleanUpInput{tarballPath: tarballPath}
	output, found := e.cleanUpBehavior[input]

	if found {
		return output.err
	}
	return fmt.Errorf("Unsupported Input: CleanUp('%s')", tarballPath)
}

func (e *FakeMultiResponseExtractor) SetCleanUpBehavior(tarballPath string, err error) {
	e.cleanUpBehavior[cleanUpInput{tarballPath: tarballPath}] = cleanUpOutput{err: err}
}
