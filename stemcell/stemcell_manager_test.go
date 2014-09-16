package stemcell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

var _ = Describe("Manager", func() {
	var (
		repo                *fakebmstemcell.FakeRepo
		manager             Manager
		fs                  *fakesys.FakeFileSystem
		reader              *fakebmstemcell.FakeStemcellReader
		infrastructure      *fakebmstemcell.FakeInfrastructure
		stemcellTarballPath string
		tempExtractionDir   string

		expectedStemcell Stemcell
		expectedCID      CID
	)

	BeforeEach(func() {
		reader = fakebmstemcell.NewFakeReader()
		fs = fakesys.NewFakeFileSystem()
		repo = fakebmstemcell.NewFakeRepo()
		infrastructure = fakebmstemcell.NewFakeInfrastructure()
		manager = NewManager(fs, reader, repo, infrastructure)

		stemcellTarballPath = "/stemcell/tarball/path"
		tempExtractionDir = "/path/to/dest"
		fs.TempDirDir = tempExtractionDir

		expectedStemcell = Stemcell{
			Name: "fake-stemcell-name",
		}
		reader.SetReadBehavior(stemcellTarballPath, tempExtractionDir, expectedStemcell, nil)

		//TODO: stemcell should only upload if not already in the repo
		//		repo.SetFindBehavior(expectedStemcell, StemcellRecord{}, false, nil)

		expectedCID = "fake-cid"
		infrastructure.SetCreateStemcellBehavior(expectedStemcell, expectedCID, nil)

		repo.SetSaveBehavior(expectedStemcell, expectedCID, nil)
	})

	It("cleans up the temp work dir", func() {
		Expect(fs.FileExists(tempExtractionDir)).To(Equal(false))

		_, _, err := manager.Upload(stemcellTarballPath)
		Expect(err).ToNot(HaveOccurred())

		Expect(fs.FileExists(tempExtractionDir)).To(Equal(false))
	})

	It("extracts and parses the stemcell manifest", func() {
		stemcell, _, err := manager.Upload(stemcellTarballPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(stemcell).To(Equal(expectedStemcell))

		Expect(reader.ReadInputs).To(Equal(
			[]fakebmstemcell.ReadInput{
				fakebmstemcell.ReadInput{
					StemcellTarballPath: stemcellTarballPath,
					DestPath:            tempExtractionDir,
				},
			},
		))
	})

	//	It("checks that the stemcell has not already been uploaded", func() {
	//		_, _, err := manager.Upload(stemcellTarballPath)
	//		Expect(err).ToNot(HaveOccurred())
	//
	//		Expect(repo.FindInputs).To(Equal(
	//			[]fakebmstemcell.FindInput{
	//				fakebmstemcell.FindInput{
	//					Stemcell: expectedStemcell,
	//				},
	//			},
	//		))
	//	})

	It("uploads the stemcell to the infrastructure and returns the cid", func() {
		_, cid, err := manager.Upload(stemcellTarballPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(cid).To(Equal(expectedCID))

		Expect(infrastructure.CreateInputs).To(Equal(
			[]fakebmstemcell.CreateInput{
				fakebmstemcell.CreateInput{
					Stemcell: expectedStemcell,
				},
			},
		))
	})

	It("saves the stemcell record in the repo", func() {
		_, cid, err := manager.Upload(stemcellTarballPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(cid).To(Equal(expectedCID))

		Expect(repo.SaveInputs).To(Equal(
			[]fakebmstemcell.SaveInput{
				fakebmstemcell.SaveInput{
					Stemcell: expectedStemcell,
					CID:      expectedCID,
				},
			},
		))
	})

	//	Context("when the stemcell record exists in the repo (having been previously uploaded)", func() {
	//
	//		var (
	//			existingRecord StemcellRecord
	//		)
	//
	//		BeforeEach(func() {
	//			expectedCID = "fake-cid"
	//			infrastructure.SetCreateStemcellBehavior(expectedStemcell, expectedCID, nil)
	//
	//			existingRecord = StemcellRecord{
	//
	//			}
	//			repo.SetFindBehavior(expectedStemcell, existingRecord, true, nil)
	//		})
	//
	//		It("does not re-upload the stemcell to the infrastructure", func() {
	//			_, cid, err := manager.Upload(stemcellTarballPath)
	//			Expect(err).ToNot(HaveOccurred())
	//
	//
	//		})
	//
	//		It("does not re-add the stemcell record to the infrastructure", func() {
	//
	//		})
	//
	//	})
})
