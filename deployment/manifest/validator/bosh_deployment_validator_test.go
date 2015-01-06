package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest/validator"
)

var _ = Describe("BoshDeploymentValidator", func() {
	var (
		validator DeploymentValidator
	)

	BeforeEach(func() {
		validator = NewBoshDeploymentValidator()
	})

	Describe("Validate", func() {
		It("does not error if deployment is valid", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				Name: "fake-deployment-name",
				Networks: []bmdeplmanifest.Network{
					{
						Name: "fake-network-name",
						Type: "dynamic",
					},
				},
				ResourcePools: []bmdeplmanifest.ResourcePool{
					{
						Name:    "fake-resource-pool-name",
						Network: "fake-network-name",
						RawCloudProperties: map[interface{}]interface{}{
							"fake-prop-key": "fake-prop-value",
							"fake-prop-map-key": map[interface{}]interface{}{
								"fake-prop-key": "fake-prop-value",
							},
						},
					},
				},
				DiskPools: []bmdeplmanifest.DiskPool{
					{
						Name:     "fake-disk-pool-name",
						DiskSize: 1024,
						RawCloudProperties: map[interface{}]interface{}{
							"fake-prop-key": "fake-prop-value",
							"fake-prop-map-key": map[interface{}]interface{}{
								"fake-prop-key": "fake-prop-value",
							},
						},
					},
				},
				Jobs: []bmdeplmanifest.Job{
					{
						Name:           "fake-job-name",
						PersistentDisk: 1024,
						Networks: []bmdeplmanifest.JobNetwork{
							{
								Name:      "fake-network-name",
								StaticIPs: []string{"127.0.0.1"},
								Default:   []bmdeplmanifest.NetworkDefault{"dns", "gateway"},
							},
						},
						Lifecycle: "service",
						RawProperties: map[interface{}]interface{}{
							"fake-prop-key": "fake-prop-value",
							"fake-prop-map-key": map[interface{}]interface{}{
								"fake-prop-key": "fake-prop-value",
							},
						},
					},
				},
				RawProperties: map[interface{}]interface{}{
					"fake-prop-key": "fake-prop-value",
					"fake-prop-map-key": map[interface{}]interface{}{
						"fake-prop-key": "fake-prop-value",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).ToNot(HaveOccurred())
		})

		It("validates name is not empty", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				Name: "",
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("name must not be empty or blank"))
		})

		It("validates name is not blank", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				Name: "   \t",
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("name must not be empty or blank"))
		})

		It("validates that there is only one resource pool", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				ResourcePools: []bmdeplmanifest.ResourcePool{
					{},
					{},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools must be of size 1"))
		})

		It("validates resource pool name", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				ResourcePools: []bmdeplmanifest.ResourcePool{
					{
						Name: "",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools[0].name must not be empty or blank"))

			deploymentManifest = bmdeplmanifest.Manifest{
				ResourcePools: []bmdeplmanifest.ResourcePool{
					{
						Name: "not-blank",
					},
					{
						Name: "   \t",
					},
				},
			}

			err = validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools[1].name must not be empty or blank"))
		})

		It("validates resource pool network", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				ResourcePools: []bmdeplmanifest.ResourcePool{
					{
						Network: "",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools[0].network must not be empty or blank"))

			deploymentManifest = bmdeplmanifest.Manifest{
				Networks: []bmdeplmanifest.Network{
					{
						Name: "fake-network",
					},
				},
				ResourcePools: []bmdeplmanifest.ResourcePool{
					{
						Network: "other-network",
					},
				},
			}

			err = validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools[0].network must be the name of a network"))
		})

		It("validates resource pool cloud_properties", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				ResourcePools: []bmdeplmanifest.ResourcePool{
					{
						RawCloudProperties: map[interface{}]interface{}{
							123: "fake-property-value",
						},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools[0].cloud_properties must have only string keys"))
		})

		It("validates resource pool env", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				ResourcePools: []bmdeplmanifest.ResourcePool{
					{
						RawEnv: map[interface{}]interface{}{
							123: "fake-env-value",
						},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools[0].env must have only string keys"))
		})

		It("validates disk pool name", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				DiskPools: []bmdeplmanifest.DiskPool{
					{
						Name: "",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("disk_pools[0].name must not be empty or blank"))

			deploymentManifest = bmdeplmanifest.Manifest{
				DiskPools: []bmdeplmanifest.DiskPool{
					{
						Name: "not-blank",
					},
					{
						Name: "   \t",
					},
				},
			}

			err = validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("disk_pools[1].name must not be empty or blank"))
		})

		It("validates disk pool size", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				DiskPools: []bmdeplmanifest.DiskPool{
					{
						Name: "fake-disk",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("disk_pools[0].disk_size must be > 0"))
		})

		It("validates disk pool cloud_properties", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				DiskPools: []bmdeplmanifest.DiskPool{
					{
						RawCloudProperties: map[interface{}]interface{}{
							123: "fake-property-value",
						},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("disk_pools[0].cloud_properties must have only string keys"))
		})

		It("validates network name", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				Networks: []bmdeplmanifest.Network{
					{
						Name: "",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("networks[0].name must not be empty or blank"))

			deploymentManifest = bmdeplmanifest.Manifest{
				Networks: []bmdeplmanifest.Network{
					{
						Name: "not-blank",
					},
					{
						Name: "   \t",
					},
				},
			}

			err = validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("networks[1].name must not be empty or blank"))
		})

		It("validates network type", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				Networks: []bmdeplmanifest.Network{
					{
						Type: "unknown-type",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("networks[0].type must be 'manual', 'dynamic', or 'vip'"))
		})

		It("validates disk pool cloud_properties", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				Networks: []bmdeplmanifest.Network{
					{
						RawCloudProperties: map[interface{}]interface{}{
							123: "fake-property-value",
						},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("networks[0].cloud_properties must have only string keys"))
		})

		It("validates that there is only one job", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				Jobs: []bmdeplmanifest.Job{
					{},
					{},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs must be of size 1"))
		})

		It("validates job name", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				Jobs: []bmdeplmanifest.Job{
					{
						Name: "",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].name must not be empty or blank"))

			deploymentManifest = bmdeplmanifest.Manifest{
				Jobs: []bmdeplmanifest.Job{
					{
						Name: "not-blank",
					},
					{
						Name: "   \t",
					},
				},
			}

			err = validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[1].name must not be empty or blank"))
		})

		It("validates job persistent_disk", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				Jobs: []bmdeplmanifest.Job{
					{
						PersistentDisk: -1234,
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].persistent_disk must be >= 0"))
		})

		It("validates job persistent_disk_pool", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				Jobs: []bmdeplmanifest.Job{
					{
						PersistentDiskPool: "non-existent-disk-pool",
					},
				},
				DiskPools: []bmdeplmanifest.DiskPool{
					{
						Name: "fake-disk-pool",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].persistent_disk_pool must be the name of a disk pool"))
		})

		It("validates job instances", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				Jobs: []bmdeplmanifest.Job{
					{
						Instances: -1234,
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].instances must be >= 0"))
		})

		It("validates job networks", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				Jobs: []bmdeplmanifest.Job{
					{
						Networks: []bmdeplmanifest.JobNetwork{},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].networks must be a non-empty array"))
		})

		It("validates job network name", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				Jobs: []bmdeplmanifest.Job{
					{
						Networks: []bmdeplmanifest.JobNetwork{
							{
								Name: "",
							},
						},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].networks[0].name must not be empty or blank"))
		})

		It("validates job network static ips", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				Jobs: []bmdeplmanifest.Job{
					{
						Networks: []bmdeplmanifest.JobNetwork{
							{
								StaticIPs: []string{"non-ip"},
							},
						},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].networks[0].static_ips[0] must be a valid IP"))
		})

		It("validates job network default", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				Jobs: []bmdeplmanifest.Job{
					{
						Networks: []bmdeplmanifest.JobNetwork{
							{
								Default: []bmdeplmanifest.NetworkDefault{
									"non-dns-or-gateway",
								},
							},
						},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].networks[0].default[0] must be 'dns' or 'gateway'"))
		})

		It("validates job lifecycle", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				Jobs: []bmdeplmanifest.Job{
					{
						Lifecycle: "errand",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].lifecycle must be 'service' ('errand' not supported)"))
		})

		It("validates job properties", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				Jobs: []bmdeplmanifest.Job{
					{
						RawProperties: map[interface{}]interface{}{
							123: "fake-property-value",
						},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].properties must have only string keys"))
		})

		It("validates deployment properties", func() {
			deploymentManifest := bmdeplmanifest.Manifest{
				RawProperties: map[interface{}]interface{}{
					123: "fake-property-value",
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("properties must have only string keys"))
		})
	})
})
