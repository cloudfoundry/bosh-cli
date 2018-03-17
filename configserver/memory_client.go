package configserver

import (
	"fmt"
)

type MemoryClient struct {
	store map[string]interface{}
}

var _ Client = &MemoryClient{}

func NewMemoryClient() *MemoryClient {
	return &MemoryClient{map[string]interface{}{}}
}

func (m *MemoryClient) Read(name string) (interface{}, error) {
	val, found := m.store[name]
	if !found {
		return nil, fmt.Errorf("Expected to find '%s'", name)
	}
	return val, nil
}

func (m *MemoryClient) Exists(name string) (bool, error) {
	_, found := m.store[name]
	return found, nil
}

func (m *MemoryClient) Write(name string, value interface{}) error {
	m.store[name] = value
	return nil
}

func (m *MemoryClient) Delete(name string) error {
	delete(m.store, name)
	return nil
}

func (m *MemoryClient) Generate(name, type_ string, params interface{}) (interface{}, error) {
	return nil, fmt.Errorf("Memory config server client does not support value generation")
}
