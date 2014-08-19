package blobstore

import (
	"fmt"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshplatform "github.com/cloudfoundry/bosh-agent/platform"
	boshsettings "github.com/cloudfoundry/bosh-agent/settings"
	boshdir "github.com/cloudfoundry/bosh-agent/settings/directories"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"
)

type Provider struct {
	platform    boshplatform.Platform
	dirProvider boshdir.Provider
	uuidGen     boshuuid.Generator
	logger      boshlog.Logger
}

func NewProvider(
	platform boshplatform.Platform,
	dirProvider boshdir.Provider,
	logger boshlog.Logger,
) (p Provider) {
	p.uuidGen = boshuuid.NewGenerator()
	p.platform = platform
	p.dirProvider = dirProvider
	p.logger = logger
	return
}

func (p Provider) Get(settings boshsettings.Blobstore) (blobstore Blobstore, err error) {
	configName := fmt.Sprintf("blobstore-%s.json", settings.Type)
	externalConfigFile := filepath.Join(p.dirProvider.EtcDir(), configName)

	switch settings.Type {
	case boshsettings.BlobstoreTypeDummy:
		blobstore = newDummyBlobstore()

	case boshsettings.BlobstoreTypeLocal:
		blobstore = NewLocalBlobstore(
			p.platform.GetFs(),
			p.uuidGen,
			settings.Options,
		)

	default:
		blobstore = NewExternalBlobstore(
			settings.Type,
			settings.Options,
			p.platform.GetFs(),
			p.platform.GetRunner(),
			p.uuidGen,
			externalConfigFile,
		)
	}

	blobstore = NewSHA1VerifiableBlobstore(blobstore)

	blobstore = NewRetryableBlobstore(blobstore, 3, p.logger)

	err = blobstore.Validate()
	if err != nil {
		err = bosherr.WrapError(err, "Validating blobstore")
	}
	return
}
