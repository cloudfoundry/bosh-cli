package manifest_test

import (
	. "github.com/cloudfoundry/bosh-init/deployment/manifest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	biproperty "github.com/cloudfoundry/bosh-init/common/property"
)

var _ = Describe("Manifest", func() {
	var (
		deploymentManifest Manifest
	)

	Describe("NetworksInterfaces", func() {
		Context("when the deployment has networks", func() {
			BeforeEach(func() {
				deploymentManifest = Manifest{
					Networks: []Network{
						{
							Name:            "fake-network-name",
							Type:            "dynamic",
							CloudProperties: biproperty.Map{},
						},
						{
							Name: "fake-manual-network-name",
							Type: "manual",
							Subnets: []Subnet{
								{
									Range:           "1.2.3.0/22",
									Gateway:         "1.1.1.1",
									CloudProperties: biproperty.Map{},
								},
							},
						},
						{
							Name:            "vip",
							Type:            "vip",
							CloudProperties: biproperty.Map{},
						},
						{
							Name:            "fake",
							Type:            "dynamic",
							CloudProperties: biproperty.Map{},
						},
					},
					Jobs: []Job{
						{
							Name: "fake-job-name",
							Networks: []JobNetwork{
								{
									Name:      "vip",
									StaticIPs: []string{"1.2.3.4"},
								},
								{
									Name:      "fake-network-name",
									StaticIPs: []string{"5.6.7.8"},
								},
								{
									Name:      "fake-manual-network-name",
									StaticIPs: []string{"5.6.7.9"},
								},
							},
						},
					},
				}
			})

			It("is a map of the network names to network interfaces", func() {
				Expect(deploymentManifest.NetworkInterfaces("fake-job-name")).To(Equal(map[string]biproperty.Map{
					"fake-network-name": biproperty.Map{
						"type":             "dynamic",
						"ip":               "5.6.7.8",
						"cloud_properties": biproperty.Map{},
					},
					"fake-manual-network-name": biproperty.Map{
						"type":             "manual",
						"ip":               "5.6.7.9",
						"netmask":          "255.255.252.0",
						"gateway":          "1.1.1.1",
						"cloud_properties": biproperty.Map{},
					},
					"vip": biproperty.Map{
						"type":             "vip",
						"ip":               "1.2.3.4",
						"cloud_properties": biproperty.Map{},
					},
				}))
			})

			Context("when the deployment does not have networks", func() {
				BeforeEach(func() {
					deploymentManifest = Manifest{
						Jobs: []Job{
							{
								Name: "fake-job-name",
							},
						},
						Networks: []Network{},
					}
				})

				It("is an empty map", func() {
					Expect(deploymentManifest.NetworkInterfaces("fake-job-name")).To(Equal(map[string]biproperty.Map{}))
				})
			})

			Context("when the deployment does not have a job with requested name", func() {
				BeforeEach(func() {
					deploymentManifest = Manifest{}
				})

				It("returns an error", func() {
					networkInterfaces, err := deploymentManifest.NetworkInterfaces("fake-job-name")
					Expect(networkInterfaces).To(Equal(map[string]biproperty.Map{}))
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Could not find job with name: fake-job-name"))
				})
			})
		})
	})

	Describe("ResourcePool", func() {
		BeforeEach(func() {
			deploymentManifest = Manifest{
				ResourcePools: []ResourcePool{
					{
						Name: "fake-resource-pool-name-1",
					},
					{
						Name: "fake-resource-pool-name-2",
					},
				},
				Jobs: []Job{
					{
						Name:         "fake-job-name",
						ResourcePool: "fake-resource-pool-name-2",
					},
					{
						Name:         "job-with-invalid-resource-pool",
						ResourcePool: "invalid-resource-pool",
					},
				},
			}
		})

		It("returns resource pool defined on a job", func() {
			resourcePool, err := deploymentManifest.ResourcePool("fake-job-name")
			Expect(err).ToNot(HaveOccurred())

			Expect(resourcePool).To(Equal(ResourcePool{
				Name: "fake-resource-pool-name-2",
			}))
		})

		Context("when resource pool specified on a job is not defined", func() {
			It("returns an error", func() {
				_, err := deploymentManifest.ResourcePool("job-with-invalid-resource-pool")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Could not find resource pool 'invalid-resource-pool' for job 'job-with-invalid-resource-pool'"))
			})
		})
	})

	Describe("DiskPool", func() {
		Context("when the deployment has disk_pools", func() {
			BeforeEach(func() {
				deploymentManifest = Manifest{
					DiskPools: []DiskPool{
						{
							Name:     "fake-disk-pool-name-1",
							DiskSize: 1024,
							CloudProperties: biproperty.Map{
								"fake-disk-prop-key-1": "fake-disk-prop-value-1",
							},
						},
						{
							Name:     "fake-disk-pool-name-2",
							DiskSize: 2048,
							CloudProperties: biproperty.Map{
								"fake-disk-prop-key-2": "fake-disk-prop-value-1",
							},
						},
					},
					Jobs: []Job{
						{
							Name:               "fake-job-name",
							PersistentDiskPool: "fake-disk-pool-name-2",
						},
					},
				}
			})

			It("is the disk pool", func() {
				diskPool, err := deploymentManifest.DiskPool("fake-job-name")
				Expect(err).ToNot(HaveOccurred())

				Expect(diskPool).To(Equal(DiskPool{
					Name:     "fake-disk-pool-name-2",
					DiskSize: 2048,
					CloudProperties: biproperty.Map{
						"fake-disk-prop-key-2": "fake-disk-prop-value-1",
					},
				}))
			})
		})

		Context("when job has persistent_disk and there are no disk_pools", func() {
			BeforeEach(func() {
				deploymentManifest = Manifest{
					Jobs: []Job{
						{
							Name:           "fake-job-name",
							PersistentDisk: 1024,
						},
					},
				}
			})

			It("is a new disk pool with the specified persistent disk size", func() {
				diskPool, err := deploymentManifest.DiskPool("fake-job-name")
				Expect(err).ToNot(HaveOccurred())

				Expect(diskPool).To(Equal(DiskPool{
					Name:            "",
					DiskSize:        1024,
					CloudProperties: biproperty.Map{},
				}))
			})
		})

		Context("when job has persistent_disk_pool and persistent_disk", func() {
			BeforeEach(func() {
				deploymentManifest = Manifest{
					DiskPools: []DiskPool{
						{
							Name:     "fake-disk-pool-name-1",
							DiskSize: 1024,
							CloudProperties: biproperty.Map{
								"fake-disk-prop-key-1": "fake-disk-prop-value-1",
							},
						},
					},
					Jobs: []Job{
						{
							Name:               "fake-job-name",
							PersistentDisk:     1024,
							PersistentDiskPool: "fake-disk-pool-name-1",
						},
					},
				}
			})

			It("returns the deployment disk pool", func() {
				diskPool, err := deploymentManifest.DiskPool("fake-job-name")
				Expect(err).ToNot(HaveOccurred())

				Expect(diskPool).To(Equal(DiskPool{
					Name:     "fake-disk-pool-name-1",
					DiskSize: 1024,
					CloudProperties: biproperty.Map{
						"fake-disk-prop-key-1": "fake-disk-prop-value-1",
					},
				}))
			})
		})

		Context("when job has persistent_disk_pool but no matching disk pool exists", func() {
			BeforeEach(func() {
				deploymentManifest = Manifest{
					Jobs: []Job{
						{
							Name:               "fake-job-name",
							PersistentDiskPool: "fake-disk-pool-name-1",
						},
					},
				}
			})

			It("returns an error", func() {
				_, err := deploymentManifest.DiskPool("fake-job-name")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Could not find persistent disk pool 'fake-disk-pool-name-1' for job 'fake-job-name'"))
			})
		})

		Context("when job does not have persistent_disk_pool or persistent_disk", func() {
			BeforeEach(func() {
				deploymentManifest = Manifest{
					Jobs: []Job{
						{
							Name: "fake-job-name",
						},
					},
				}
			})

			It("returns an empty disk pool", func() {
				diskPool, err := deploymentManifest.DiskPool("fake-job-name")
				Expect(err).ToNot(HaveOccurred())
				Expect(diskPool).To(Equal(DiskPool{}))
			})
		})
	})
})
