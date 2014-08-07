package fakes

import "errors"

type FakeExtractor struct {
	ExtractExpectedArchives map[string]string
}

func NewFakeExtractor() *FakeExtractor {
	return &FakeExtractor{
		ExtractExpectedArchives: make(map[string]string),
	}
}

func (e *FakeExtractor) Extract(source string, destination string) error {
	_, ok := e.ExtractExpectedArchives[source]
	if !ok {
		return errors.New("Missing archive")
	}

	return nil
}

func (e *FakeExtractor) AddExpectedArchive(tarFilePath string) {
	e.ExtractExpectedArchives[tarFilePath] = ""
}
