package license

import (
	. "github.com/cloudfoundry/bosh-cli/release/resource"
)

type License struct {
	resource Resource
}

func NewLicense(resource Resource) *License {
	return &License{resource: resource}
}

func (l License) Name() string        { return l.resource.Name() }
func (l License) Fingerprint() string { return l.resource.Fingerprint() }

func (l *License) ArchivePath() string { return l.resource.ArchivePath() }
func (l *License) ArchiveSHA1() string { return l.resource.ArchiveSHA1() }

func (l *License) Build(dev, final ArchiveIndex) error { return l.resource.Build(dev, final) }
func (l *License) Finalize(final ArchiveIndex) error   { return l.resource.Finalize(final) }
