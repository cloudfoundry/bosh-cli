package manifest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/deployment/manifest"

	biproperty "github.com/cloudfoundry/bosh-init/common/property"
)

var _ = Describe("Network", func() {
	var (
		network Network
	)

	BeforeEach(func() {

	})

	Describe("Interface", func() {
		Context("when network type is manual", func() {
			BeforeEach(func() {
				network = Network{
					Name: "fake-manual-network-name",
					Type: "manual",
					Subnets: []Subnet{
						{
							Range:   "1.2.3.0/22",
							Gateway: "1.1.1.1",
							DNS:     []string{"1.2.3.4"},
							CloudProperties: biproperty.Map{
								"cp_key": "cp_value",
							},
						},
					},
				}
			})

			It("includes gateway, dns, ip from the job and netmask calculated from range", func() {
				Expect(network.Interface([]string{"5.6.7.9"})).To(Equal(biproperty.Map{
					"type":    "manual",
					"ip":      "5.6.7.9",
					"gateway": "1.1.1.1",
					"netmask": "255.255.252.0",
					"dns":     []string{"1.2.3.4"},
					"cloud_properties": biproperty.Map{
						"cp_key": "cp_value",
					},
				}))
			})
		})
	})

	Context("when network type is dynamic", func() {
		BeforeEach(func() {
			network = Network{
				Name: "fake-dynamic-network-name",
				Type: "dynamic",
				CloudProperties: biproperty.Map{
					"cp_key": "cp_value",
				},
				DNS: []string{"2.2.2.2"},
			}
		})

		It("includes dns and cloud_properties", func() {
			Expect(network.Interface([]string{})).To(Equal(biproperty.Map{
				"type": "dynamic",
				"dns":  []string{"2.2.2.2"},
				"cloud_properties": biproperty.Map{
					"cp_key": "cp_value",
				},
			}))
		})
	})
})
