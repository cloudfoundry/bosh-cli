package manifest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
)

var _ = Describe("ResourcePool", func() {
	var (
		resourcePool ResourcePool
	)

	BeforeEach(func() {
		resourcePool = ResourcePool{
			Name: "fake-resource-pool-name",
			RawCloudProperties: map[interface{}]interface{}{
				"fake-cloud-property-key": "fake-cloud-property-value",
			},
			RawEnv: map[interface{}]interface{}{
				"bosh": map[interface{}]interface{}{
					"password": "secret",
				},
			},
		}
	})

	Describe("Env", func() {
		It("returns a map of the resource pool's env", func() {
			env, err := resourcePool.Env()
			Expect(err).ToNot(HaveOccurred())
			Expect(env).To(Equal(map[string]interface{}{
				"bosh": map[string]interface{}{
					"password": "secret",
				},
			}))
		})
	})

	Describe("CloudProperties", func() {
		It("returns a map of the resource pool's cloud properties", func() {
			cloudProperties, err := resourcePool.CloudProperties()
			Expect(err).ToNot(HaveOccurred())
			Expect(cloudProperties).To(Equal(map[string]interface{}{
				"fake-cloud-property-key": "fake-cloud-property-value",
			}))
		})
	})
})
