package instance_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
)

var _ = Describe("State", func() {
	Describe("ToApplySpec", func() {
		It("translates from instance model to apply spec model", func() {
			networkInterfaces := []NetworkRef{
				{
					Name: "fake-network-name",
					Interface: map[string]interface{}{
						"ip":   "fake-ip",
						"type": "dynamic",
					},
				},
			}
			renderedJobs := []JobRef{
				{
					Name:    "fake-job-name",
					Version: "fake-job-fingerprint",
				},
			}
			compiledPackages := []PackageRef{
				{
					Name:    "vcloud_cpi",
					Version: "fake-fingerprint-cpi",
					Archive: BlobRef{
						SHA1:        "fake-sha1-cpi",
						BlobstoreID: "fake-package-blob-id-cpi",
					},
				},
				{
					Name:    "ruby",
					Version: "fake-fingerprint-ruby",
					Archive: BlobRef{
						SHA1:        "fake-sha1-ruby",
						BlobstoreID: "fake-package-blob-id-ruby",
					},
				},
			}
			renderedJobListBlob := BlobRef{
				BlobstoreID: "fake-rendered-job-list-archive-blob-id",
				SHA1:        "fake-rendered-job-list-archive-blob-sha1",
			}
			state := NewState(
				"fake-deployment-name",
				"fake-job-name",
				0,
				networkInterfaces,
				renderedJobs,
				compiledPackages,
				renderedJobListBlob,
				"fake-state-hash",
			)

			applySpec := state.ToApplySpec()

			Expect(applySpec).To(Equal(bmas.ApplySpec{
				Deployment: "fake-deployment-name",
				Index:      0,
				Networks: map[string]interface{}{
					"fake-network-name": map[string]interface{}{
						"ip":   "fake-ip",
						"type": "dynamic",
					},
				},
				Job: bmas.Job{
					Name: "fake-job-name",
					Templates: []bmas.Blob{
						{
							Name:    "fake-job-name",
							Version: "fake-job-fingerprint",
						},
					},
				},
				Packages: map[string]bmas.Blob{
					"vcloud_cpi": bmas.Blob{
						Name:        "vcloud_cpi",
						Version:     "fake-fingerprint-cpi",
						SHA1:        "fake-sha1-cpi",
						BlobstoreID: "fake-package-blob-id-cpi",
					},
					"ruby": bmas.Blob{
						Name:        "ruby",
						Version:     "fake-fingerprint-ruby",
						SHA1:        "fake-sha1-ruby",
						BlobstoreID: "fake-package-blob-id-ruby",
					},
				},
				RenderedTemplatesArchive: bmas.RenderedTemplatesArchiveSpec{
					BlobstoreID: "fake-rendered-job-list-archive-blob-id",
					SHA1:        "fake-rendered-job-list-archive-blob-sha1",
				},
				ConfigurationHash: "fake-state-hash",
			}))
		})
	})

})
