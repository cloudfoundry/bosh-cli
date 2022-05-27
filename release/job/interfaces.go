package job

import (
	boshman "github.com/cloudfoundry/bosh-cli/v7/release/manifest"
)

// You only need **one** of these per package!
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . ArchiveReader

type ArchiveReader interface {
	Read(boshman.JobRef, string) (*Job, error)
}

//counterfeiter:generate . DirReader

type DirReader interface {
	Read(string) (*Job, error)
}
