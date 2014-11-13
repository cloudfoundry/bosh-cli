package deployer_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmconfig "github.com/cloudfoundry/bosh-micro-cli/config/fakes"
	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer"
)

var _ = Describe("DeploymentRecord", func() {
	var (
		fakeRelease      *fakebmrel.FakeRelease
		stemcell         bmstemcell.ExtractedStemcell
		stemcellRepo     *fakebmconfig.FakeStemcellRepo
		deploymentRecord DeploymentRecord
	)

	BeforeEach(func() {
		fakeRelease = fakebmrel.NewFakeRelease()
		fakeFS := fakesys.NewFakeFileSystem()
		stemcell = bmstemcell.NewExtractedStemcell(
			bmstemcell.Manifest{
				Name:    "fake-stemcell-name",
				Version: "fake-stemcell-version",
			},
			bmstemcell.ApplySpec{},
			"fake-extracted-path",
			fakeFS,
		)
		stemcellRepo = fakebmconfig.NewFakeStemcellRepo()
		deploymentRecord = NewDeploymentRecord(stemcellRepo)
	})

	Describe("IsDeployed", func() {
		Context("when the specified stemcell is currently deployed", func() {
			BeforeEach(func() {
				currentRecord := bmconfig.StemcellRecord{
					ID:      "fake-stemcell-id",
					Name:    "fake-stemcell-name",
					Version: "fake-stemcell-version",
					CID:     "fake-stemcell-cid",
				}
				stemcellRepo.SetFindCurrentBehavior(currentRecord, true, nil)
			})

			It("returns true", func() {
				isDeployed, err := deploymentRecord.IsDeployed("fake-manifest-path", fakeRelease, stemcell)
				Expect(err).ToNot(HaveOccurred())
				Expect(isDeployed).To(BeTrue())
			})
		})

		Context("when a different stemcell is currently deployed", func() {
			BeforeEach(func() {
				currentRecord := bmconfig.StemcellRecord{
					ID:      "fake-stemcell-id-2",
					Name:    "fake-stemcell-name-2",
					Version: "fake-stemcell-version-2",
					CID:     "fake-stemcell-cid-2",
				}
				stemcellRepo.SetFindCurrentBehavior(currentRecord, true, nil)
			})

			It("returns false", func() {
				isDeployed, err := deploymentRecord.IsDeployed("fake-manifest-path", fakeRelease, stemcell)
				Expect(err).ToNot(HaveOccurred())
				Expect(isDeployed).To(BeFalse())
			})
		})

		Context("when no stemcell is currently deployed", func() {
			BeforeEach(func() {
				stemcellRepo.SetFindCurrentBehavior(bmconfig.StemcellRecord{}, false, nil)
			})

			It("returns false", func() {
				isDeployed, err := deploymentRecord.IsDeployed("fake-manifest-path", fakeRelease, stemcell)
				Expect(err).ToNot(HaveOccurred())
				Expect(isDeployed).To(BeFalse())
			})
		})

		Context("when finding the currently deployed stemcell fails", func() {
			BeforeEach(func() {
				stemcellRepo.SetFindCurrentBehavior(bmconfig.StemcellRecord{}, false, errors.New("fake-find-error"))
			})

			It("returns an error", func() {
				_, err := deploymentRecord.IsDeployed("fake-manifest-path", fakeRelease, stemcell)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-find-error"))
			})
		})
	})

	Describe("UpdateDeploymentRecord", func() {
		Context("when the specified stemcell is not in the stemcell repo", func() {
			BeforeEach(func() {
				stemcellRepo.SetFindBehavior("fake-stemcell-name", "fake-stemcell-version", bmconfig.StemcellRecord{}, false, nil)
			})

			It("returns an error", func() {
				err := deploymentRecord.Update("fake-manifest-path", fakeRelease, stemcell)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Stemcell record not found"))
			})
		})

		Context("when specified stemcell is in the stemcell repo", func() {
			var deployedStemcell bmconfig.StemcellRecord

			BeforeEach(func() {
				deployedStemcell = bmconfig.StemcellRecord{
					ID:      "fake-stemcell-id",
					Name:    "fake-stemcell-name",
					Version: "fake-stemcell-version",
					CID:     "fake-stemcell-cid",
				}
				stemcellRepo.SetFindBehavior("fake-stemcell-name", "fake-stemcell-version", deployedStemcell, true, nil)
			})

			It("updates currently deployed stemcell", func() {
				err := deploymentRecord.Update("fake-manifest-path", fakeRelease, stemcell)
				Expect(err).ToNot(HaveOccurred())
				Expect(stemcellRepo.UpdateCurrentRecordID).To(Equal("fake-stemcell-id"))
			})

			Context("when finding the currently deployed stemcell fails", func() {
				BeforeEach(func() {
					stemcellRepo.SetFindBehavior("fake-stemcell-name", "fake-stemcell-version", deployedStemcell, false, errors.New("fake-find-error"))
				})

				It("returns an error", func() {
					err := deploymentRecord.Update("fake-manifest-path", fakeRelease, stemcell)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-find-error"))
				})
			})

			Context("when updating currently deployed stemcell fails", func() {
				BeforeEach(func() {
					stemcellRepo.UpdateCurrentErr = errors.New("fake-update-error")
				})

				It("returns an error", func() {
					err := deploymentRecord.Update("fake-manifest-path", fakeRelease, stemcell)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-update-error"))
				})
			})
		})
	})
})
