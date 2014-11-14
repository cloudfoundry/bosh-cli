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
		releaseRepo      *fakebmconfig.FakeReleaseRepo
		deploymentRecord DeploymentRecord
	)

	BeforeEach(func() {
		fakeRelease = &fakebmrel.FakeRelease{
			ReleaseName:    "fake-release-name",
			ReleaseVersion: "fake-release-version",
		}
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
		releaseRepo = fakebmconfig.NewFakeReleaseRepo()
		deploymentRecord = NewDeploymentRecord(releaseRepo, stemcellRepo)
	})

	Describe("IsDeployed", func() {
		BeforeEach(func() {
			stemcellRecord := bmconfig.StemcellRecord{
				ID:      "fake-stemcell-id",
				Name:    "fake-stemcell-name",
				Version: "fake-stemcell-version",
				CID:     "fake-stemcell-cid",
			}
			stemcellRepo.SetFindCurrentBehavior(stemcellRecord, true, nil)

			releaseRecord := bmconfig.ReleaseRecord{
				ID:      "fake-release-id",
				Name:    "fake-release-name",
				Version: "fake-release-version",
			}
			releaseRepo.SetFindCurrentBehavior(releaseRecord, true, nil)
		})

		Context("when the specified stemcell and release are currently deployed", func() {
			It("returns true", func() {
				isDeployed, err := deploymentRecord.IsDeployed("fake-manifest-path", fakeRelease, stemcell)
				Expect(err).ToNot(HaveOccurred())
				Expect(isDeployed).To(BeTrue())
			})
		})

		Context("when a different stemcell is currently deployed", func() {
			BeforeEach(func() {
				stemcellRecord := bmconfig.StemcellRecord{
					ID:      "fake-stemcell-id-2",
					Name:    "fake-stemcell-name-2",
					Version: "fake-stemcell-version-2",
					CID:     "fake-stemcell-cid-2",
				}
				stemcellRepo.SetFindCurrentBehavior(stemcellRecord, true, nil)
			})

			It("returns false", func() {
				isDeployed, err := deploymentRecord.IsDeployed("fake-manifest-path", fakeRelease, stemcell)
				Expect(err).ToNot(HaveOccurred())
				Expect(isDeployed).To(BeFalse())
			})
		})

		Context("when a different release is currently deployed", func() {
			BeforeEach(func() {
				releaseRecord := bmconfig.ReleaseRecord{
					ID:      "fake-release-id-2",
					Name:    "fake-release-name-2",
					Version: "fake-release-version-2",
				}
				releaseRepo.SetFindCurrentBehavior(releaseRecord, true, nil)
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

		Context("when no release is currently deployed", func() {
			BeforeEach(func() {
				releaseRepo.SetFindCurrentBehavior(bmconfig.ReleaseRecord{}, false, nil)
			})

			It("returns false", func() {
				isDeployed, err := deploymentRecord.IsDeployed("fake-manifest-path", fakeRelease, stemcell)
				Expect(err).ToNot(HaveOccurred())
				Expect(isDeployed).To(BeFalse())
			})
		})

		Context("when finding the currently deployed release fails", func() {
			BeforeEach(func() {
				releaseRepo.SetFindCurrentBehavior(bmconfig.ReleaseRecord{}, false, errors.New("fake-find-error"))
			})

			It("returns an error", func() {
				_, err := deploymentRecord.IsDeployed("fake-manifest-path", fakeRelease, stemcell)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-find-error"))
			})
		})
	})

	Describe("Update", func() {
		var (
			deployedStemcell bmconfig.StemcellRecord
			deployedRelease  bmconfig.ReleaseRecord
		)

		BeforeEach(func() {
			deployedStemcell = bmconfig.StemcellRecord{
				ID:      "fake-stemcell-id",
				Name:    "fake-stemcell-name",
				Version: "fake-stemcell-version",
				CID:     "fake-stemcell-cid",
			}
			stemcellRepo.SetFindBehavior("fake-stemcell-name", "fake-stemcell-version", deployedStemcell, true, nil)

			deployedRelease = bmconfig.ReleaseRecord{
				ID:      "fake-release-id",
				Name:    "fake-release-name",
				Version: "fake-release-version",
			}
			releaseRepo.SetSaveBehavior("fake-release-name", "fake-release-version", deployedRelease, nil)
		})

		It("updates currently deployed stemcell", func() {
			err := deploymentRecord.Update("fake-manifest-path", fakeRelease, stemcell)
			Expect(err).ToNot(HaveOccurred())
			Expect(stemcellRepo.UpdateCurrentRecordID).To(Equal("fake-stemcell-id"))
		})

		It("updates currently deployed release", func() {
			err := deploymentRecord.Update("fake-manifest-path", fakeRelease, stemcell)
			Expect(err).ToNot(HaveOccurred())
			Expect(releaseRepo.UpdateCurrentRecordID).To(Equal("fake-release-id"))
		})

		Context("when saving release record fails", func() {
			BeforeEach(func() {
				releaseRepo.SetSaveBehavior("fake-release-name", "fake-release-version", bmconfig.ReleaseRecord{}, errors.New("fake-save-error"))
			})

			It("returns an error", func() {
				err := deploymentRecord.Update("fake-manifest-path", fakeRelease, stemcell)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-save-error"))
			})
		})

		Context("when updating current release record fails", func() {
			BeforeEach(func() {
				releaseRepo.UpdateCurrentErr = errors.New("fake-update-error")
			})

			It("returns an error", func() {
				err := deploymentRecord.Update("fake-manifest-path", fakeRelease, stemcell)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-update-error"))
			})
		})

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
