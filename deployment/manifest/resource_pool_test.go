package manifest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
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
			Expect(env).To(Equal(bmproperty.Map{
				"bosh": bmproperty.Map{
					"password": "secret",
				},
			}))
		})
	})

	Describe("CloudProperties", func() {
		It("returns a map of the resource pool's cloud properties", func() {
			cloudProperties, err := resourcePool.CloudProperties()
			Expect(err).ToNot(HaveOccurred())
			Expect(cloudProperties).To(Equal(bmproperty.Map{
				"fake-cloud-property-key": "fake-cloud-property-value",
			}))
		})
	})
})
