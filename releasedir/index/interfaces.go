package index

// You only need **one** of these per package!
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Index

type Index interface {
	Find(name, version string) (string, string, error)
	Add(name, version, path, sha1 string) (string, string, error)
}

//counterfeiter:generate . IndexBlobs

type IndexBlobs interface {
	Get(name, blobID, sha1 string) (string, error)
	Add(name, path, sha1 string) (string, string, error)
}

//counterfeiter:generate . Reporter

type Reporter interface {
	IndexEntryStartedAdding(type_, desc string)
	IndexEntryFinishedAdding(type_, desc string, err error)

	IndexEntryDownloadStarted(type_, desc string)
	IndexEntryDownloadFinished(type_, desc string, err error)

	IndexEntryUploadStarted(type_, desc string)
	IndexEntryUploadFinished(type_, desc string, err error)
}
