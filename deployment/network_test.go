package deployment_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment"
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
		}
	})

	Describe("Spec", func() {
		It("returns a map of the network's spec", func() {
			spec, err := network.Spec()
			Expect(err).ToNot(HaveOccurred())
			Expect(spec).To(Equal(map[string]interface{}{
				"fake-name": map[string]interface{}{
					"type": "dynamic",
					"cloud_properties": map[string]interface{}{
						"subnet": map[string]interface{}{
							"name": "sg-1234",
						},
					},
				},
			}))
		})
	})
})
