package job

import (
	boshman "github.com/cloudfoundry/bosh-init/release/manifest"
)

//go:generate counterfeiter . ArchiveReader

type ArchiveReader interface {
	Read(boshman.JobRef, string) (*Job, error)
}

//go:generate counterfeiter . DirReader

type DirReader interface {
	Read(string) (*Job, error)
}
