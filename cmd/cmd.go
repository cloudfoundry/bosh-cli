package cmd

import (
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type Cmd interface {
	Run([]string) error
	FileSystem() boshsys.FileSystem
}
