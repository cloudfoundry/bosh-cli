package applyspec_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmas "github.com/cloudfoundry/bosh-micro-cli/microdeployer/applyspec"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"

	. "github.com/cloudfoundry/bosh-micro-cli/microdeployer/applyspec"
)

var _ = Describe("Factory", func() {
	var (
		originalApplySpec bmstemcell.ApplySpec
		networksSpec      map[string]interface{}
		applySpecFactory  Factory
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

		applySpecFactory = NewFactory()
	})

	Describe("Create", func() {
		It("creates an apply spec", func() {
			applySpec := applySpecFactory.Create(
				originalApplySpec,
				"fake-deployment-name",
				"fake-job-name",
				networksSpec,
				"fake-archived-templates-blob-id",
				"fake-archived-templates-sha1",
				"fake-templates-dir-sha1",
			)
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
	})
})
