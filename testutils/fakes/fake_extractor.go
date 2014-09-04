package fakes

import "fmt"

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

type FakeMultiResponseExtractor struct {
	decompressBehavior map[decompressInput]decompressOutput
	compressBehavior   map[compressInput]compressOutput
	cleanUpBehavior    map[cleanUpInput]cleanUpOutput
}

func NewFakeMultiResponseExtractor() *FakeMultiResponseExtractor {
	return &FakeMultiResponseExtractor{decompressBehavior: map[decompressInput]decompressOutput{}}
}

func (e *FakeMultiResponseExtractor) DecompressFileToDir(srcFile, destDir string) error {
	input := decompressInput{srcFile: srcFile, destDir: destDir}
	output, found := e.decompressBehavior[input]

	if found {
		return output.err
	}
	return fmt.Errorf("Unsupported Input: DecompressFileToDir('%s', '%s')", srcFile, destDir)
}

func (e *FakeMultiResponseExtractor) SetDecompressBehavior(srcFile, destDir string, err error) {
	e.decompressBehavior[decompressInput{srcFile: srcFile, destDir: destDir}] = decompressOutput{err: err}
}

func (e *FakeMultiResponseExtractor) Behaviors() map[decompressInput]decompressOutput {
	return e.decompressBehavior
}

func (e *FakeMultiResponseExtractor) CompressFilesInDir(srcDir string) (string, error) {
	input := compressInput{srcDir: srcDir}
	output, found := e.compressBehavior[input]

	if found {
		return output.destFile, output.err
	}
	return "", fmt.Errorf("Unsupported Input: CompressFilesInDir('%s')", srcDir)
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
