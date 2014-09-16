package stemcell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"

	fakebmconfig "github.com/cloudfoundry/bosh-micro-cli/config/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

var _ = Describe("Repo", func() {
	Describe("Save", func() {
		var (
			repo          Repo
			configService *fakebmconfig.FakeService
		)

		BeforeEach(func() {
			configService = fakebmconfig.NewFakeService()
			repo = NewRepo(configService)
		})

		It("saves the stemcell using config service", func() {
			stemcell := Stemcell{
				Name:    "fake-name",
				Version: "fake-version",
				Sha1:    "fake-sha1",
			}
			cid := CID("fake-cid")
			repo.Save(stemcell, cid)
			expectedStemcell := bmconfig.StemcellRecord{
				Name:    "fake-name",
				Version: "fake-version",
				SHA1:    "fake-sha1",
				CID:     cid.String(),
			}

			expectedConfig := bmconfig.Config{
				Stemcells: []bmconfig.StemcellRecord{expectedStemcell},
			}
			Expect(configService.SaveInputs).To(Equal(
				[]fakebmconfig.SaveInput{
					fakebmconfig.SaveInput{
						Config: expectedConfig,
					},
				},
			))
		})
	})
})
