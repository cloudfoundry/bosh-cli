package fakes

import (
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
)

type FakeService struct {
	Saved bmconfig.Config
}

func (s *FakeService) Load() (bmconfig.Config, error) {
	return bmconfig.Config{}, nil
}

func (s *FakeService) Save(config bmconfig.Config) error {
	s.Saved = config
	return nil
}
