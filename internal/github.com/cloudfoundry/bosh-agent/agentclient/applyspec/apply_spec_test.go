package applyspec_test

import (
	"encoding/json"

	. "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/agentclient/applyspec"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/gomega"

	biproperty "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/property"
)

var _ = Describe("ApplySpec", func() {
	var (
		applySpec ApplySpec
	)

	BeforeEach(func() {
		applySpec = ApplySpec{
			Deployment: "fake-deployment-name",
			Index:      0,
			Packages: map[string]Blob{
				"first-package-name": Blob{
					Name:        "first-package-name",
					Version:     "first-package-version",
					SHA1:        "first-package-sha1",
					BlobstoreID: "first-package-blobstore-id",
				},
			},
			Networks: map[string]biproperty.Map{
				"fake-network-name": biproperty.Map{
					"fake-network-key": "fake-network-value",
				},
			},
			Job: Job{
				Name: "fake-job-name",
				Templates: []Blob{
					{
						Name:        "first-template-name",
						Version:     "first-template-version",
						SHA1:        "first-template-sha1",
						BlobstoreID: "first-template-blobstore-id",
					},
				},
			},
			RenderedTemplatesArchive: RenderedTemplatesArchiveSpec{
				BlobstoreID: "fake-rendered-template-blob-id",
				SHA1:        "fake-rendered-template-sha1",
			},
			ConfigurationHash: "fake-configuration-hash",
		}
	})

	Describe("Marshal", func() {
		It("returns correct json", func() {
			applySpecJSON, err := json.Marshal(applySpec)
			Expect(err).ToNot(HaveOccurred())

			var applySpecMap map[string]interface{}
			err = json.Unmarshal(applySpecJSON, &applySpecMap)
			Expect(err).ToNot(HaveOccurred())

			Expect(applySpecMap).To(Equal(map[string]interface{}{
				"deployment": "fake-deployment-name",
				"index":      0.0, //json.Unmarshal ultimately converts all ints to floats. type must be float for comparisons to work
				"packages": map[string]interface{}{
					"first-package-name": map[string]interface{}{
						"name":         "first-package-name",
						"version":      "first-package-version",
						"sha1":         "first-package-sha1",
						"blobstore_id": "first-package-blobstore-id",
					},
				},
				"networks": map[string]interface{}{
					"fake-network-name": map[string]interface{}{
						"fake-network-key": "fake-network-value",
					},
				},
				"job": map[string]interface{}{
					"name": "fake-job-name",
					"templates": []interface{}{
						map[string]interface{}{
							"blobstore_id": "first-template-blobstore-id",
							"name":         "first-template-name",
							"version":      "first-template-version",
							"sha1":         "first-template-sha1",
						},
					},
				},
				"rendered_templates_archive": map[string]interface{}{
					"blobstore_id": "fake-rendered-template-blob-id",
					"sha1":         "fake-rendered-template-sha1",
				},
				"configuration_hash": "fake-configuration-hash",
			}))
		})
	})
})
