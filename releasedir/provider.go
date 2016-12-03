package releasedir

import (
	gopath "path"

	boshblob "github.com/cloudfoundry/bosh-utils/blobstore"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"github.com/pivotal-golang/clock"

	bicrypto "github.com/cloudfoundry/bosh-cli/crypto"
	boshrel "github.com/cloudfoundry/bosh-cli/release"
	boshidx "github.com/cloudfoundry/bosh-cli/releasedir/index"
)

type Provider struct {
	indexReporter        boshidx.Reporter
	releaseIndexReporter ReleaseIndexReporter
	blobsReporter        BlobsDirReporter
	releaseProvider      boshrel.Provider
	sha1calc             bicrypto.SHA1Calculator

	cmdRunner   boshsys.CmdRunner
	uuidGen     boshuuid.Generator
	timeService clock.Clock
	fs          boshsys.FileSystem
	logger      boshlog.Logger
}

func NewProvider(
	indexReporter boshidx.Reporter,
	releaseIndexReporter ReleaseIndexReporter,
	blobsReporter BlobsDirReporter,
	releaseProvider boshrel.Provider,
	sha1calc bicrypto.SHA1Calculator,
	cmdRunner boshsys.CmdRunner,
	uuidGen boshuuid.Generator,
	timeService clock.Clock,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) Provider {
	return Provider{
		indexReporter:        indexReporter,
		releaseIndexReporter: releaseIndexReporter,
		blobsReporter:        blobsReporter,
		releaseProvider:      releaseProvider,
		sha1calc:             sha1calc,
		cmdRunner:            cmdRunner,
		uuidGen:              uuidGen,
		timeService:          timeService,
		fs:                   fs,
		logger:               logger,
	}
}

func (p Provider) NewFSReleaseDir(dirPath string) FSReleaseDir {
	gitRepo := NewFSGitRepo(dirPath, p.cmdRunner, p.fs)
	blobsDir := p.NewFSBlobsDir(dirPath)
	generator := NewFSGenerator(dirPath, p.fs)

	devRelsPath := gopath.Join(dirPath, "dev_releases")
	devReleases := NewFSReleaseIndex("dev", devRelsPath, p.releaseIndexReporter, p.uuidGen, p.fs)

	finalRelsPath := gopath.Join(dirPath, "releases")
	finalReleases := NewFSReleaseIndex("final", finalRelsPath, p.releaseIndexReporter, p.uuidGen, p.fs)

	indiciesProvider := boshidx.NewProvider(p.indexReporter, p.newBlobstore(dirPath), p.sha1calc, p.fs)
	_, finalIndex := indiciesProvider.DevAndFinalIndicies(dirPath)

	releaseReader := p.NewReleaseReader(dirPath)

	return NewFSReleaseDir(dirPath, p.newConfig(dirPath), gitRepo, blobsDir,
		generator, devReleases, finalReleases, finalIndex, releaseReader, p.timeService, p.fs)
}

func (p Provider) NewFSBlobsDir(dirPath string) FSBlobsDir {
	return NewFSBlobsDir(dirPath, p.blobsReporter, p.newBlobstore(dirPath), p.sha1calc, p.fs)
}

func (p Provider) NewReleaseReader(dirPath string) boshrel.BuiltReader {
	multiReader := p.releaseProvider.NewMultiReader(dirPath)
	indiciesProvider := boshidx.NewProvider(p.indexReporter, p.newBlobstore(dirPath), p.sha1calc, p.fs)
	devIndex, finalIndex := indiciesProvider.DevAndFinalIndicies(dirPath)
	return boshrel.NewBuiltReader(multiReader, devIndex, finalIndex)
}

func (p Provider) newBlobstore(dirPath string) boshblob.Blobstore {
	provider, options, err := p.newConfig(dirPath).Blobstore()
	if err != nil {
		return NewErrBlobstore(err)
	}

	var blobstore boshblob.Blobstore

	switch provider {
	case "local":
		blobstore = boshblob.NewLocalBlobstore(p.fs, p.uuidGen, options)
	case "s3":
		blobstore = NewS3Blobstore(p.fs, p.uuidGen, options)
	default:
		return NewErrBlobstore(bosherr.Error("Expected release blobstore to be configured"))
	}

	blobstore = boshblob.NewSHA1VerifiableBlobstore(blobstore)
	blobstore = boshblob.NewRetryableBlobstore(blobstore, 3, p.logger)

	err = blobstore.Validate()
	if err != nil {
		return NewErrBlobstore(err)
	}

	return blobstore
}

func (p Provider) newConfig(dirPath string) FSConfig {
	publicPath := gopath.Join(dirPath, "config", "final.yml")
	privatePath := gopath.Join(dirPath, "config", "private.yml")
	return NewFSConfig(publicPath, privatePath, p.fs)
}
