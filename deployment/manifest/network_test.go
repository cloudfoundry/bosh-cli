package manifest_test

import (
	. "github.com/cloudfoundry/bosh-init/deployment/manifest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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
					Defaults: []string{"dns", "gateway"},
				}
			})

			It("includes gateway, dns, ip from the job and netmask calculated from range", func() {
				iface, err := network.Interface([]string{"5.6.7.9"})
				Expect(err).ToNot(HaveOccurred())
				Expect(iface).To(Equal(biproperty.Map{
					"type":    "manual",
					"ip":      "5.6.7.9",
					"gateway": "1.1.1.1",
					"netmask": "255.255.252.0",
					"dns":     []string{"1.2.3.4"},
					"cloud_properties": biproperty.Map{
						"cp_key": "cp_value",
					},
					"defaults": []string{"dns", "gateway"},
				}))
			})

			It("doesn't include defaults if they're empty", func() {
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

				iface, err := network.Interface([]string{"5.6.7.9"})
				Expect(err).ToNot(HaveOccurred())
				Expect(iface).To(Equal(biproperty.Map{
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

			Context("when range is invalid", func() {
				BeforeEach(func() {
					network.Subnets[0].Range = "invalid-range"
				})

				It("returns an error", func() {
					_, err := network.Interface([]string{"5.6.7.9"})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Failed to parse subnet range"))
				})
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
				DNS:      []string{"2.2.2.2"},
				Defaults: []string{"dns", "gateway"},
			}
		})

		It("includes dns and cloud_properties", func() {
			iface, err := network.Interface([]string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(iface).To(Equal(biproperty.Map{
				"type": "dynamic",
				"dns":  []string{"2.2.2.2"},
				"cloud_properties": biproperty.Map{
					"cp_key": "cp_value",
				},
				"defaults": []string{"dns", "gateway"},
			}))
		})
	})
})
