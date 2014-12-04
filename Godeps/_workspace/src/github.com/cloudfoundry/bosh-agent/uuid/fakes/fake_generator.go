package fakes

import (
	"fmt"
)

type FakeGenerator struct {
	GeneratedUuid string
	NextUuid      int
	GenerateError error
}

func NewFakeGenerator() *FakeGenerator {
	return &FakeGenerator{}
}

func (gen *FakeGenerator) Generate() (uuid string, err error) {
	if gen.GeneratedUuid == "" && gen.GenerateError == nil {
		uuidString := fmt.Sprintf("fake-uuid-%d", gen.NextUuid)
		gen.NextUuid++
		return uuidString, nil
	}
	return gen.GeneratedUuid, gen.GenerateError
}
