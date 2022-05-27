package pkg

import (
	boshman "github.com/cloudfoundry/bosh-cli/v7/release/manifest"
)

// You only need **one** of these per package!
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Compilable

type Compilable interface {
	Name() string
	Fingerprint() string

	ArchivePath() string
	ArchiveDigest() string

	IsCompiled() bool

	Deps() []Compilable
}

//counterfeiter:generate . ArchiveReader

type ArchiveReader interface {
	Read(boshman.PackageRef, string) (*Package, error)
}

//counterfeiter:generate . DirReader

type DirReader interface {
	Read(string) (*Package, error)
}
