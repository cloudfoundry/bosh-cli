package deployer_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmconfig "github.com/cloudfoundry/bosh-micro-cli/config/fakes"
	fakebmcrypto "github.com/cloudfoundry/bosh-micro-cli/crypto/fakes"
	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

var _ = Describe("DeploymentRecord", func() {
	var (
		fakeRelease        *fakebmrel.FakeRelease
		stemcell           bmstemcell.ExtractedStemcell
		deploymentRepo     *fakebmconfig.FakeDeploymentRepo
		releaseRepo        *fakebmconfig.FakeReleaseRepo
		stemcellRepo       *fakebmconfig.FakeStemcellRepo
		fakeSHA1Calculator *fakebmcrypto.FakeSha1Calculator
		deploymentRecord   DeploymentRecord
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
		deploymentRepo = fakebmconfig.NewFakeDeploymentRepo()
		releaseRepo = fakebmconfig.NewFakeReleaseRepo()
		stemcellRepo = fakebmconfig.NewFakeStemcellRepo()
		fakeSHA1Calculator = fakebmcrypto.NewFakeSha1Calculator()
		deploymentRecord = NewDeploymentRecord(deploymentRepo, releaseRepo, stemcellRepo, fakeSHA1Calculator)
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

			deploymentRepo.SetFindCurrentBehavior("fake-manifest-sha1", true, nil)
			fakeSHA1Calculator.SetCalculateBehavior(map[string]fakebmcrypto.CalculateInput{
				"fake-manifest-path": fakebmcrypto.CalculateInput{
					Sha1: "fake-manifest-sha1",
					Err:  nil,
				},
			})
		})

		It("returns true", func() {
			isDeployed, err := deploymentRecord.IsDeployed("fake-manifest-path", fakeRelease, stemcell)
			Expect(err).ToNot(HaveOccurred())
			Expect(isDeployed).To(BeTrue())
		})

		Context("when getting current deployment manifest sha1 fails", func() {
			BeforeEach(func() {
				deploymentRepo.SetFindCurrentBehavior("fake-manifest-path", true, errors.New("fake-find-error"))
			})

			It("returns an error", func() {
				_, err := deploymentRecord.IsDeployed("fake-manifest-path", fakeRelease, stemcell)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-find-error"))
			})
		})

		Context("when no deployment is set", func() {
			BeforeEach(func() {
				deploymentRepo.SetFindCurrentBehavior("", false, nil)
			})

			It("returns false", func() {
				isDeployed, err := deploymentRecord.IsDeployed("fake-manifest-path", fakeRelease, stemcell)
				Expect(err).ToNot(HaveOccurred())
				Expect(isDeployed).To(BeFalse())
			})
		})

		Context("when calculating the deployment manifest sha1 fails", func() {
			BeforeEach(func() {
				fakeSHA1Calculator.SetCalculateBehavior(map[string]fakebmcrypto.CalculateInput{
					"fake-manifest-path": fakebmcrypto.CalculateInput{
						Sha1: "",
						Err:  errors.New("fake-calculate-error"),
					},
				})
			})

			It("returns an error", func() {
				_, err := deploymentRecord.IsDeployed("fake-manifest-path", fakeRelease, stemcell)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-calculate-error"))
			})
		})

		Context("when a different deployment manifest is currently deployed", func() {
			BeforeEach(func() {
				deploymentRepo.SetFindCurrentBehavior("fake-manifest-sha1-2", true, nil)
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
	})

	Describe("Update", func() {
		var (
			deployedRelease bmconfig.ReleaseRecord
		)

		BeforeEach(func() {
			fakeSHA1Calculator.SetCalculateBehavior(map[string]fakebmcrypto.CalculateInput{
				"fake-manifest-path": fakebmcrypto.CalculateInput{
					Sha1: "fake-manifest-sha1",
					Err:  nil,
				},
			})

			deployedRelease = bmconfig.ReleaseRecord{
				ID:      "fake-release-id",
				Name:    "fake-release-name",
				Version: "fake-release-version",
			}
			releaseRepo.SetFindBehavior("fake-release-name", "fake-release-version", deployedRelease, true, nil)
		})

		It("calculates and updates sha1 of currently deployed manifest", func() {
			err := deploymentRecord.Update("fake-manifest-path", fakeRelease)
			Expect(err).ToNot(HaveOccurred())
			Expect(deploymentRepo.UpdateCurrentManifestSHA1).To(Equal("fake-manifest-sha1"))
		})

		It("updates currently deployed release", func() {
			err := deploymentRecord.Update("fake-manifest-path", fakeRelease)
			Expect(err).ToNot(HaveOccurred())
			Expect(releaseRepo.UpdateCurrentRecordID).To(Equal("fake-release-id"))
		})

		Context("when release is not in release repo", func() {
			BeforeEach(func() {
				releaseRepo.SetFindBehavior("fake-release-name", "fake-release-version", bmconfig.ReleaseRecord{}, false, nil)
				savedRelease := bmconfig.ReleaseRecord{
					ID:      "fake-saved-release-id",
					Name:    "fake-release-name",
					Version: "fake-release-version",
				}

				releaseRepo.SetSaveBehavior("fake-release-name", "fake-release-version", savedRelease, nil)
			})

			It("saves release to release repo", func() {
				err := deploymentRecord.Update("fake-manifest-path", fakeRelease)
				Expect(err).ToNot(HaveOccurred())
				Expect(releaseRepo.UpdateCurrentRecordID).To(Equal("fake-saved-release-id"))

				Expect(releaseRepo.SaveInputs).To(Equal([]fakebmconfig.ReleaseRepoSaveInput{
					{
						Name:    "fake-release-name",
						Version: "fake-release-version",
					},
				}))
			})

			Context("when saving release record fails", func() {
				BeforeEach(func() {
					releaseRepo.SetSaveBehavior("fake-release-name", "fake-release-version", bmconfig.ReleaseRecord{}, errors.New("fake-save-error"))
				})

				It("returns an error", func() {
					err := deploymentRecord.Update("fake-manifest-path", fakeRelease)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-save-error"))
				})
			})
		})

		Context("when calculating the deployment manifest sha1 fails", func() {
			BeforeEach(func() {
				fakeSHA1Calculator.SetCalculateBehavior(map[string]fakebmcrypto.CalculateInput{
					"fake-manifest-path": fakebmcrypto.CalculateInput{
						Sha1: "",
						Err:  errors.New("fake-calculate-error"),
					},
				})
			})

			It("returns an error", func() {
				err := deploymentRecord.Update("fake-manifest-path", fakeRelease)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-calculate-error"))
			})
		})

		Context("when updating currently deployed manifest sha1 fails", func() {
			BeforeEach(func() {
				deploymentRepo.UpdateCurrentErr = errors.New("fake-update-error")
			})

			It("returns an error", func() {
				err := deploymentRecord.Update("fake-manifest-path", fakeRelease)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-update-error"))
			})
		})

		Context("when updating current release record fails", func() {
			BeforeEach(func() {
				releaseRepo.UpdateCurrentErr = errors.New("fake-update-error")
			})

			It("returns an error", func() {
				err := deploymentRecord.Update("fake-manifest-path", fakeRelease)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-update-error"))
			})
		})
	})
})
