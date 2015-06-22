package settings_test

import (
	"encoding/json"

	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/matchers"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/settings"
)

func init() {
	Describe("Settings", func() {
		var settings Settings

		Describe("PersistentDiskSettings", func() {
			Context("when the disk settings are hash", func() {
				BeforeEach(func() {
					settings = Settings{
						Disks: Disks{
							Persistent: map[string]interface{}{
								"fake-disk-id": map[string]interface{}{
									"volume_id": "fake-disk-volume-id",
									"path":      "fake-disk-path",
								},
							},
						},
					}
				})

				It("returns disk settings", func() {
					diskSettings, found := settings.PersistentDiskSettings("fake-disk-id")
					Expect(found).To(BeTrue())
					Expect(diskSettings).To(Equal(DiskSettings{
						ID:       "fake-disk-id",
						VolumeID: "fake-disk-volume-id",
						Path:     "fake-disk-path",
					}))
				})
			})

			Context("when the disk settings is a string", func() {
				BeforeEach(func() {
					settings = Settings{
						Disks: Disks{
							Persistent: map[string]interface{}{
								"fake-disk-id": "fake-disk-value",
							},
						},
					}
				})

				It("converts it to disk settings", func() {
					diskSettings, found := settings.PersistentDiskSettings("fake-disk-id")
					Expect(found).To(BeTrue())
					Expect(diskSettings).To(Equal(DiskSettings{
						ID:       "fake-disk-id",
						VolumeID: "fake-disk-value",
						Path:     "fake-disk-value",
					}))
				})
			})

			Context("when disk with requested disk ID is not present", func() {
				BeforeEach(func() {
					settings = Settings{
						Disks: Disks{
							Persistent: map[string]interface{}{
								"fake-disk-id": "fake-disk-path",
							},
						},
					}
				})

				It("returns false", func() {
					diskSettings, found := settings.PersistentDiskSettings("fake-non-existent-disk-id")
					Expect(found).To(BeFalse())
					Expect(diskSettings).To(Equal(DiskSettings{}))
				})
			})
		})

		Describe("EphemeralDiskSettings", func() {
			BeforeEach(func() {
				settings = Settings{
					Disks: Disks{
						Ephemeral: "fake-disk-value",
					},
				}
			})

			It("converts disk settings", func() {
				Expect(settings.EphemeralDiskSettings()).To(Equal(DiskSettings{
					ID:       "",
					VolumeID: "fake-disk-value",
					Path:     "fake-disk-value",
				}))
			})
		})

		Describe("DefaultNetworkFor", func() {
			Context("when networks is empty", func() {
				It("returns found=false", func() {
					networks := Networks{}
					_, found := networks.DefaultNetworkFor("dns")
					Expect(found).To(BeFalse())
				})
			})

			Context("with a single network", func() {
				It("returns that network (found=true)", func() {
					networks := Networks{
						"first": Network{
							DNS: []string{"xx.xx.xx.xx"},
						},
					}

					network, found := networks.DefaultNetworkFor("dns")
					Expect(found).To(BeTrue())
					Expect(network).To(Equal(networks["first"]))
				})
			})

			Context("with multiple networks and default is found for dns", func() {
				It("returns the network marked default (found=true)", func() {
					networks := Networks{
						"first": Network{
							Default: []string{},
							DNS:     []string{"aa.aa.aa.aa"},
						},
						"second": Network{
							Default: []string{"something-else", "dns"},
							DNS:     []string{"xx.xx.xx.xx", "yy.yy.yy.yy", "zz.zz.zz.zz"},
						},
						"third": Network{
							Default: []string{},
							DNS:     []string{"aa.aa.aa.aa"},
						},
					}

					settings, found := networks.DefaultNetworkFor("dns")
					Expect(found).To(BeTrue())
					Expect(settings).To(Equal(networks["second"]))
				})
			})

			Context("with multiple networks and default is not found", func() {
				It("returns found=false", func() {
					networks := Networks{
						"first": Network{
							Default: []string{"foo"},
							DNS:     []string{"xx.xx.xx.xx", "yy.yy.yy.yy", "zz.zz.zz.zz"},
						},
						"second": Network{
							Default: []string{},
							DNS:     []string{"aa.aa.aa.aa"},
						},
					}

					_, found := networks.DefaultNetworkFor("dns")
					Expect(found).To(BeFalse())
				})
			})

			Context("with multiple networks marked as default", func() {
				It("returns one of them", func() {
					networks := Networks{
						"first": Network{
							Default: []string{"dns"},
							DNS:     []string{"xx.xx.xx.xx", "yy.yy.yy.yy", "zz.zz.zz.zz"},
						},
						"second": Network{
							Default: []string{"dns"},
							DNS:     []string{"aa.aa.aa.aa"},
						},
						"third": Network{
							DNS: []string{"bb.bb.bb.bb"},
						},
					}

					for i := 0; i < 100; i++ {
						settings, found := networks.DefaultNetworkFor("dns")
						Expect(found).To(BeTrue())
						Expect(settings).Should(MatchOneOf(networks["first"], networks["second"]))
					}
				})
			})
		})

		Describe("DefaultIP", func() {
			It("with two networks", func() {
				networks := Networks{
					"bosh": Network{
						IP: "xx.xx.xx.xx",
					},
					"vip": Network{
						IP: "aa.aa.aa.aa",
					},
				}

				ip, found := networks.DefaultIP()
				Expect(found).To(BeTrue())
				Expect(ip).To(MatchOneOf("xx.xx.xx.xx", "aa.aa.aa.aa"))
			})

			It("with two networks only with defaults", func() {
				networks := Networks{
					"bosh": Network{
						IP: "xx.xx.xx.xx",
					},
					"vip": Network{
						IP:      "aa.aa.aa.aa",
						Default: []string{"dns"},
					},
				}

				ip, found := networks.DefaultIP()
				Expect(found).To(BeTrue())
				Expect(ip).To(Equal("aa.aa.aa.aa"))
			})

			It("when none specified", func() {
				networks := Networks{
					"bosh": Network{},
					"vip": Network{
						Default: []string{"dns"},
					},
				}

				_, found := networks.DefaultIP()
				Expect(found).To(BeFalse())
			})
		})
	})

	Describe("Settings", func() {
		var expectSnakeCaseKeys func(map[string]interface{})

		expectSnakeCaseKeys = func(value map[string]interface{}) {
			for k, v := range value {
				Expect(k).To(MatchRegexp("\\A[a-z0-9_]+\\z"))

				tv, isMap := v.(map[string]interface{})
				if isMap {
					expectSnakeCaseKeys(tv)
				}
			}
		}

		It("marshals into JSON in snake case to stay consistent with CPI agent env formatting", func() {
			settings := Settings{}
			settingsJSON, err := json.Marshal(settings)
			Expect(err).NotTo(HaveOccurred())

			var settingsMap map[string]interface{}
			err = json.Unmarshal(settingsJSON, &settingsMap)
			Expect(err).NotTo(HaveOccurred())
			expectSnakeCaseKeys(settingsMap)
		})

		It("allows different types for blobstore option values", func() {
			var settings Settings
			settingsJSON := `{"blobstore":{"options":{"string":"value", "int":443, "bool":true, "map":{}}}}`

			err := json.Unmarshal([]byte(settingsJSON), &settings)
			Expect(err).NotTo(HaveOccurred())
			Expect(settings.Blobstore.Options).To(Equal(map[string]interface{}{
				"string": "value",
				"int":    443.0,
				"bool":   true,
				"map":    map[string]interface{}{},
			}))
		})
	})

	Describe("Network", func() {
		var network Network
		BeforeEach(func() {
			network = Network{}
		})

		Describe("IsDHCP", func() {
			Context("when network is VIP", func() {
				BeforeEach(func() {
					network.Type = NetworkTypeVIP
				})

				It("returns false", func() {
					Expect(network.IsDHCP()).To(BeFalse())
				})
			})

			Context("when network is Dynamic", func() {
				BeforeEach(func() {
					network.Type = NetworkTypeDynamic
				})

				It("returns true", func() {
					Expect(network.IsDHCP()).To(BeTrue())
				})
			})

			Context("when IP is not set", func() {
				BeforeEach(func() {
					network.Netmask = "255.255.255.0"
				})

				It("returns true", func() {
					Expect(network.IsDHCP()).To(BeTrue())
				})
			})

			Context("when Netmask is not set", func() {
				BeforeEach(func() {
					network.IP = "127.0.0.5"
				})

				It("returns true", func() {
					Expect(network.IsDHCP()).To(BeTrue())
				})
			})

			Context("when IP and Netmask are set", func() {
				BeforeEach(func() {
					network.IP = "127.0.0.5"
					network.Netmask = "255.255.255.0"
				})

				It("returns false", func() {
					Expect(network.IsDHCP()).To(BeFalse())
				})
			})

			Context("when network was previously resolved via DHCP", func() {
				BeforeEach(func() {
					network.Resolved = true
				})

				It("returns true", func() {
					Expect(network.IsDHCP()).To(BeTrue())
				})
			})
		})
	})
}
