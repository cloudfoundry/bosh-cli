package stemcell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
)

var _ = Describe("CloudStemcell", func() {
	var (
		stemcellRepo      bmconfig.StemcellRepo
		fakeUUIDGenerator *fakeuuid.FakeGenerator
		fakeCloud         *fakebmcloud.FakeCloud
		cloudStemcell     CloudStemcell
	)

	BeforeEach(func() {
		stemcellRecord := bmconfig.StemcellRecord{
			CID:     "fake-stemcell-cid",
			Name:    "fake-stemcell-name",
			Version: "fake-stemcell-version",
		}
		fs := fakesys.NewFakeFileSystem()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fakeUUIDGenerator = &fakeuuid.FakeGenerator{}
		configService := bmconfig.NewFileSystemDeploymentConfigService("/fake/path", fs, fakeUUIDGenerator, logger)
		stemcellRepo = bmconfig.NewStemcellRepo(configService, fakeUUIDGenerator)
		fakeCloud = fakebmcloud.NewFakeCloud()
		cloudStemcell = NewCloudStemcell(stemcellRecord, stemcellRepo, fakeCloud)
	})

	Describe("PromoteAsCurrent", func() {
		Context("when stemcell is in the repo", func() {
			BeforeEach(func() {
				fakeUUIDGenerator.GeneratedUuid = "fake-stemcell-id"
				_, err := stemcellRepo.Save("fake-stemcell-name", "fake-stemcell-version", "fake-stemcell-cid")
				Expect(err).ToNot(HaveOccurred())
			})

			It("sets stemcell as current in the repo", func() {
				err := cloudStemcell.PromoteAsCurrent()
				Expect(err).ToNot(HaveOccurred())

				currentStemcell, found, err := stemcellRepo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(currentStemcell).To(Equal(bmconfig.StemcellRecord{
					ID:      "fake-stemcell-id",
					CID:     "fake-stemcell-cid",
					Name:    "fake-stemcell-name",
					Version: "fake-stemcell-version",
				}))
			})
		})

		Context("when stemcell is not in the repo", func() {
			It("returns an error", func() {
				err := cloudStemcell.PromoteAsCurrent()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Stemcell does not exist in repo"))
			})
		})
	})

	Describe("Delete", func() {
		It("deletes stemcell from cloud", func() {
			err := cloudStemcell.Delete()
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeCloud.DeleteStemcellInputs).To(Equal([]fakebmcloud.DeleteStemcellInput{
				{
					StemcellCID: "fake-stemcell-cid",
				},
			}))
		})

		It("deletes stemcell from repo", func() {
			_, err := stemcellRepo.Save("fake-stemcell-name", "fake-stemcell-version", "fake-stemcell-cid")
			Expect(err).ToNot(HaveOccurred())

			err = cloudStemcell.Delete()
			Expect(err).ToNot(HaveOccurred())
			stemcellRecords, err := stemcellRepo.All()
			Expect(stemcellRecords).To(BeEmpty())
		})

		Context("when deleted stemcell is the current stemcell", func() {
			BeforeEach(func() {
				stemcellRecord, err := stemcellRepo.Save("fake-stemcell-name", "fake-stemcell-version", "fake-stemcell-cid")
				Expect(err).ToNot(HaveOccurred())

				err = stemcellRepo.UpdateCurrent(stemcellRecord.ID)
				Expect(err).ToNot(HaveOccurred())
			})

			It("clears current stemcell in the repo", func() {
				err := cloudStemcell.Delete()
				Expect(err).ToNot(HaveOccurred())

				_, found, err := stemcellRepo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when deleting stemcell in the cloud fails", func() {
			BeforeEach(func() {
				fakeCloud.DeleteStemcellErr = errors.New("fake-delete-stemcell-error")
			})

			It("returns an error", func() {
				err := cloudStemcell.Delete()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-delete-stemcell-error"))
			})
		})

		Context("when deleting stemcell in the cloud fails with StemcellNotFoundError", func() {
			var deleteErr = bmcloud.NewCPIError("delete_stemcell", bmcloud.CmdError{
				Type:    bmcloud.StemcellNotFoundError,
				Message: "fake-stemcell-not-found-message",
			})

			BeforeEach(func() {
				stemcellRecord, err := stemcellRepo.Save("fake-stemcell-name", "fake-stemcell-version", "fake-stemcell-cid")
				Expect(err).ToNot(HaveOccurred())

				err = stemcellRepo.UpdateCurrent(stemcellRecord.ID)
				Expect(err).ToNot(HaveOccurred())

				fakeCloud.DeleteStemcellErr = deleteErr
			})

			It("deletes stemcell in the cloud", func() {
				err := cloudStemcell.Delete()
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(deleteErr))

				Expect(fakeCloud.DeleteStemcellInputs).To(Equal([]fakebmcloud.DeleteStemcellInput{
					{
						StemcellCID: "fake-stemcell-cid",
					},
				}))
			})

			It("deletes stemcell in the disk repo", func() {
				err := cloudStemcell.Delete()
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(deleteErr))

				stemcellRecords, err := stemcellRepo.All()
				Expect(stemcellRecords).To(BeEmpty())
			})

			It("clears current stemcell in the stemcell repo", func() {
				err := cloudStemcell.Delete()
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(deleteErr))

				_, found, err := stemcellRepo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})
})
