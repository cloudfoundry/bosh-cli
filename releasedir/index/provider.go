package index

import (
	gopath "path"

	boshblob "github.com/cloudfoundry/bosh-utils/blobstore"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	bicrypto "github.com/cloudfoundry/bosh-cli/crypto"
	boshrel "github.com/cloudfoundry/bosh-cli/release"
)

type Provider struct {
	reporter  Reporter
	blobstore boshblob.Blobstore
	sha1calc  bicrypto.SHA1Calculator
	fs        boshsys.FileSystem
}

func NewProvider(
	reporter Reporter,
	blobstore boshblob.Blobstore,
	sha1calc bicrypto.SHA1Calculator,
	fs boshsys.FileSystem,
) Provider {
	return Provider{
		reporter:  reporter,
		blobstore: blobstore,
		sha1calc:  sha1calc,
		fs:        fs,
	}
}

func (p Provider) DevAndFinalIndicies(dirPath string) (boshrel.ArchiveIndicies, boshrel.ArchiveIndicies) {
	cachePath := gopath.Join("~", ".bosh", "cache")

	devBlobsCache := NewFSIndexBlobs(cachePath, p.reporter, nil, p.sha1calc, p.fs)
	finalBlobsCache := NewFSIndexBlobs(cachePath, p.reporter, p.blobstore, p.sha1calc, p.fs)

	devJobsPath := gopath.Join(dirPath, ".dev_builds", "jobs")
	devPkgsPath := gopath.Join(dirPath, ".dev_builds", "packages")
	devLicPath := gopath.Join(dirPath, ".dev_builds", "license")

	finalJobsPath := gopath.Join(dirPath, ".final_builds", "jobs")
	finalPkgsPath := gopath.Join(dirPath, ".final_builds", "packages")
	finalLicPath := gopath.Join(dirPath, ".final_builds", "license")

	devIndicies := boshrel.ArchiveIndicies{
		Jobs:     NewFSIndex("job", devJobsPath, true, false, p.reporter, devBlobsCache, p.fs),
		Packages: NewFSIndex("package", devPkgsPath, true, false, p.reporter, devBlobsCache, p.fs),
		Licenses: NewFSIndex("license", devLicPath, false, false, p.reporter, devBlobsCache, p.fs),
	}

	finalIndicies := boshrel.ArchiveIndicies{
		Jobs:     NewFSIndex("job", finalJobsPath, true, true, p.reporter, finalBlobsCache, p.fs),
		Packages: NewFSIndex("package", finalPkgsPath, true, true, p.reporter, finalBlobsCache, p.fs),
		Licenses: NewFSIndex("license", finalLicPath, false, true, p.reporter, finalBlobsCache, p.fs),
	}

	return devIndicies, finalIndicies
}
