package stemcell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"

	fakebmconfig "github.com/cloudfoundry/bosh-micro-cli/config/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

var _ = Describe("Repo", func() {
	var (
		repo          Repo
		configService *fakebmconfig.FakeService
	)

	BeforeEach(func() {
		configService = fakebmconfig.NewFakeService()
		repo = NewRepo(configService)
	})

	Describe("Save", func() {
		It("saves the stemcell record using the config service", func() {
			configService.SetLoadBehavior(bmconfig.Config{}, nil)
			cid := CID("fake-cid")
			expectedStemcell := bmconfig.StemcellRecord{
				Name:    "fake-name",
				Version: "fake-version",
				SHA1:    "fake-sha1",
				CID:     cid.String(),
			}
			expectedConfig := bmconfig.Config{
				Stemcells: []bmconfig.StemcellRecord{expectedStemcell},
			}
			err := configService.SetSaveBehavior(expectedConfig, nil)
			Expect(err).ToNot(HaveOccurred())

			stemcell := Stemcell{
				Name:    "fake-name",
				Version: "fake-version",
				SHA1:    "fake-sha1",
			}
			err = repo.Save(stemcell, cid)
			Expect(err).ToNot(HaveOccurred())

			Expect(configService.SaveInputs).To(Equal(
				[]fakebmconfig.SaveInput{
					fakebmconfig.SaveInput{
						Config: expectedConfig,
					},
				},
			))
		})
	})

	Describe("Find", func() {
		It("finds existing stemcell records", func() {
			expectedCid := CID("fake-cid")
			expectedStemcell := bmconfig.StemcellRecord{
				Name:    "fake-name",
				Version: "fake-version",
				SHA1:    "fake-sha1",
				CID:     expectedCid.String(),
			}
			expectedConfig := bmconfig.Config{
				Stemcells: []bmconfig.StemcellRecord{expectedStemcell},
			}
			configService.SetLoadBehavior(expectedConfig, nil)

			stemcell := Stemcell{
				Name:    "fake-name",
				Version: "fake-version",
				SHA1:    "fake-sha1",
			}
			cid, found, err := repo.Find(stemcell)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(cid).To(Equal(expectedCid))
		})
	})
})
