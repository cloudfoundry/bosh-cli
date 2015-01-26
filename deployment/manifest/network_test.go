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
			IP:      "1.2.3.4",
			Netmask: "255.255.255.0",
			Gateway: "1.2.3.1",
			DNS:     []string{"1.1.1.1"},
		}
	})

	Describe("Interface", func() {
		It("returns an interface that can be used to connect to the network", func() {
			iface, err := network.Interface()
			Expect(err).ToNot(HaveOccurred())
			Expect(iface).To(Equal(NetworkInterface{
				"type":    "dynamic",
				"ip":      "1.2.3.4",
				"netmask": "255.255.255.0",
				"gateway": "1.2.3.1",
				"dns":     []string{"1.1.1.1"},
				"cloud_properties": map[string]interface{}{
					"subnet": map[string]interface{}{
						"name": "sg-1234",
					},
				},
			}))
		})

		It("leaves out missing optional fields", func() {
			network.Netmask = ""
			network.Gateway = ""
			network.DNS = []string{}

			iface, err := network.Interface()
			Expect(err).ToNot(HaveOccurred())
			Expect(iface).To(Equal(NetworkInterface{
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
