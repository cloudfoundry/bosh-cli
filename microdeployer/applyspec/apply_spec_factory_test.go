package applyspec_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmas "github.com/cloudfoundry/bosh-micro-cli/microdeployer/applyspec"
	fakebmas "github.com/cloudfoundry/bosh-micro-cli/microdeployer/applyspec/fakes"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"

	. "github.com/cloudfoundry/bosh-micro-cli/microdeployer/applyspec"
)

var _ = Describe("Factory", func() {
	var (
		originalApplySpec  bmstemcell.ApplySpec
		networksSpec       map[string]interface{}
		applySpecFactory   Factory
		fakeSha1Calculator *fakebmas.FakeSha1Calculator
	)

	BeforeEach(func() {
		originalApplySpec = bmstemcell.ApplySpec{
			Packages: map[string]bmstemcell.Blob{
				"fake-first-package-name": bmstemcell.Blob{
					Name: "fake-first-package-name",
				},
			},
			Job: bmstemcell.Job{
				Templates: []bmstemcell.Blob{
					{
						Name: "fake-template-name",
					},
				},
			},
		}

		networksSpec = map[string]interface{}{
			"fake-network-name": "fake-network-value",
		}

		fakeSha1Calculator = fakebmas.NewFakeSha1Calculator()
		fakeSha1Calculator.SetCalculateBehavior(map[string]fakebmas.CalculateInput{
			"/fake-archived-templates-path": fakebmas.CalculateInput{
				Sha1: "fake-archived-templates-sha1",
				Err:  nil,
			},
			"/fake-templates-dir": fakebmas.CalculateInput{
				Sha1: "fake-templates-dir-sha1",
				Err:  nil,
			},
		})
		applySpecFactory = NewFactory(fakeSha1Calculator)
	})

	Describe("Create", func() {
		It("creates an apply spec", func() {
			applySpec, err := applySpecFactory.Create(
				originalApplySpec,
				"fake-deployment-name",
				"fake-job-name",
				networksSpec,
				"fake-archived-templates-blob-id",
				"/fake-archived-templates-path",
				"/fake-templates-dir",
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(applySpec).To(Equal(bmas.ApplySpec{
				Deployment: "fake-deployment-name",
				Index:      0,
				Packages: map[string]bmas.Blob{
					"fake-first-package-name": bmas.Blob{
						Name: "fake-first-package-name",
					},
				},
				ConfigurationHash: "fake-templates-dir-sha1",
				Networks: map[string]interface{}{
					"fake-network-name": "fake-network-value",
				},
				Job: bmas.Job{
					Name: "fake-job-name",
					Templates: []bmas.Blob{
						{
							Name: "fake-template-name",
						},
					},
				},
				RenderedTemplatesArchive: bmas.RenderedTemplatesArchiveSpec{
					BlobstoreID: "fake-archived-templates-blob-id",
					SHA1:        "fake-archived-templates-sha1",
				},
			}))
		})

		Context("when creating the apply spec fails", func() {
			BeforeEach(func() {
				calculateErr := errors.New("fake-calculate-error")
				fakeSha1Calculator.SetCalculateBehavior(map[string]fakebmas.CalculateInput{
					"/fake-archived-templates-path": fakebmas.CalculateInput{
						Sha1: "fake-archived-templates-sha1",
						Err:  calculateErr,
					},
					"/fake-templates-dir": fakebmas.CalculateInput{
						Sha1: "fake-templates-dir-sha1",
						Err:  nil,
					},
				})
			})

			It("returns an error", func() {
				_, err := applySpecFactory.Create(
					originalApplySpec,
					"fake-deployment-name",
					"fake-job-name",
					networksSpec,
					"fake-archived-templates-blob-id",
					"/fake-archived-templates-path",
					"/fake-templates-dir",
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-calculate-error"))
			})
		})
	})
})
