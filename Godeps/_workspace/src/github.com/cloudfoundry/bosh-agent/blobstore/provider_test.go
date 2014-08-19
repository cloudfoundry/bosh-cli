package blobstore_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-agent/blobstore"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakeplatform "github.com/cloudfoundry/bosh-agent/platform/fakes"
	boshsettings "github.com/cloudfoundry/bosh-agent/settings"
	boshdir "github.com/cloudfoundry/bosh-agent/settings/directories"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"
)

var _ = Describe("Provider", func() {
	var (
		platform *fakeplatform.FakePlatform
		logger   boshlog.Logger
		provider Provider
	)

	BeforeEach(func() {
		platform = fakeplatform.NewFakePlatform()
		dirProvider := boshdir.NewProvider("/var/vcap")
		logger = boshlog.NewLogger(boshlog.LevelNone)
		provider = NewProvider(platform, dirProvider, logger)
	})

	Describe("Get", func() {
		It("get dummy", func() {
			blobstore, err := provider.Get(boshsettings.Blobstore{
				Type: boshsettings.BlobstoreTypeDummy,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(blobstore).ToNot(BeNil())
		})

		It("get external when external command in path", func() {
			options := map[string]interface{}{"key": "value"}

			platform.Runner.CommandExistsValue = true

			blobstore, err := provider.Get(boshsettings.Blobstore{
				Type:    "fake-external-type",
				Options: options,
			})
			Expect(err).ToNot(HaveOccurred())

			expectedBlobstore := NewExternalBlobstore(
				"fake-external-type",
				options,
				platform.GetFs(),
				platform.GetRunner(),
				boshuuid.NewGenerator(),
				"/var/vcap/bosh/etc/blobstore-fake-external-type.json",
			)
			expectedBlobstore = NewSHA1VerifiableBlobstore(expectedBlobstore)
			expectedBlobstore = NewRetryableBlobstore(expectedBlobstore, 3, logger)
			Expect(blobstore).To(Equal(expectedBlobstore))

			err = expectedBlobstore.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("get external errs when external command not in path", func() {
			options := map[string]interface{}{"key": "value"}

			platform.Runner.CommandExistsValue = false

			_, err := provider.Get(boshsettings.Blobstore{
				Type:    "fake-external-type",
				Options: options,
			})
			Expect(err).To(HaveOccurred())
		})
	})
})
