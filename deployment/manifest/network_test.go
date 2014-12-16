package manifest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
)

var _ = Describe("Network", func() {
	var (
		network Network
	)

	BeforeEach(func() {
		network = Network{
			Name: "fake-name",
			Type: Dynamic,
			RawCloudProperties: map[interface{}]interface{}{
				"subnet": map[interface{}]interface{}{
					"name": "sg-1234",
				},
			},
			IP: "1.2.3.4",
		}
	})

	Describe("Spec", func() {
		It("returns a map of the network's spec", func() {
			spec, err := network.Spec()
			Expect(err).ToNot(HaveOccurred())
			Expect(spec).To(Equal(map[string]interface{}{
				"type": "dynamic",
				"ip":   "1.2.3.4",
				"cloud_properties": map[string]interface{}{
					"subnet": map[string]interface{}{
						"name": "sg-1234",
					},
				},
			}))
		})
	})
})
