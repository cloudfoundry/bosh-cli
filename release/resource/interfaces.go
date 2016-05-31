package resource

//go:generate counterfeiter . Archive

type Archive interface {
	Fingerprint() (string, error)
	Build(expectedFp string) (string, string, error)
}

type ArchiveFunc func([]File, []File, []string) Archive

//go:generate counterfeiter . ArchiveIndex

type ArchiveIndex interface {
	Find(name, fingerprint string) (string, string, error)
	Add(name, fingerprint, path, sha1 string) (string, string, error)
}

//go:generate counterfeiter . Resource

type Resource interface {
	Name() string
	Fingerprint() string

	ArchivePath() string
	ArchiveSHA1() string

	Build(dev, final ArchiveIndex) error
	Finalize(final ArchiveIndex) error
}

//go:generate counterfeiter . Fingerprinter

type Fingerprinter interface {
	Calculate([]File, []string) (string, error)
}
