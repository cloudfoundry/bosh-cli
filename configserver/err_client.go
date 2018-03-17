package configserver

import (
	"fmt"
)

type ErrClient struct {
	err error
}

var _ Client = ErrClient{}

func NewErrClient() ErrClient {
	return ErrClient{fmt.Errorf("Expected to have configured config server")}
}

func (c ErrClient) Read(path string) (interface{}, error) { return nil, c.err }
func (c ErrClient) Generate(path, type_ string, params interface{}) (interface{}, error) {
	return nil, c.err
}
func (c ErrClient) Write(path string, value interface{}) error { return c.err }
func (c ErrClient) Delete(path string) error                   { return c.err }
func (c ErrClient) Exists(path string) (bool, error)           { return false, c.err }
