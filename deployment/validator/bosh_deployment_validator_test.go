package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/validator"
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
			deployment := bmdepl.Deployment{
				Name: "fake-deployment-name",
				Networks: []bmdepl.Network{
					{
						Name: "fake-network-name",
						Type: "dynamic",
					},
				},
				ResourcePools: []bmdepl.ResourcePool{
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
				DiskPools: []bmdepl.DiskPool{
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
				Jobs: []bmdepl.Job{
					{
						Name:           "fake-job-name",
						PersistentDisk: 1024,
						Networks: []bmdepl.JobNetwork{
							{
								Name:      "fake-network-name",
								StaticIPs: []string{"127.0.0.1"},
								Default:   []bmdepl.NetworkDefault{"dns", "gateway"},
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

			err := validator.Validate(deployment)
			Expect(err).ToNot(HaveOccurred())
		})

		It("validates name is not empty", func() {
			deployment := bmdepl.Deployment{
				Name: "",
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("name must not be empty or blank"))
		})

		It("validates name is not blank", func() {
			deployment := bmdepl.Deployment{
				Name: "   \t",
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("name must not be empty or blank"))
		})

		It("validates that there is only one resource pool", func() {
			deployment := bmdepl.Deployment{
				ResourcePools: []bmdepl.ResourcePool{
					{},
					{},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools must be of size 1"))
		})

		It("validates resource pool name", func() {
			deployment := bmdepl.Deployment{
				ResourcePools: []bmdepl.ResourcePool{
					{
						Name: "",
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools[0].name must not be empty or blank"))

			deployment = bmdepl.Deployment{
				ResourcePools: []bmdepl.ResourcePool{
					{
						Name: "not-blank",
					},
					{
						Name: "   \t",
					},
				},
			}

			err = validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools[1].name must not be empty or blank"))
		})

		It("validates resource pool network", func() {
			deployment := bmdepl.Deployment{
				ResourcePools: []bmdepl.ResourcePool{
					{
						Network: "",
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools[0].network must not be empty or blank"))

			deployment = bmdepl.Deployment{
				Networks: []bmdepl.Network{
					{
						Name: "fake-network",
					},
				},
				ResourcePools: []bmdepl.ResourcePool{
					{
						Network: "other-network",
					},
				},
			}

			err = validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools[0].network must be the name of a network"))
		})

		It("validates resource pool cloud_properties", func() {
			deployment := bmdepl.Deployment{
				ResourcePools: []bmdepl.ResourcePool{
					{
						RawCloudProperties: map[interface{}]interface{}{
							123: "fake-property-value",
						},
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools[0].cloud_properties must have only string keys"))
		})

		It("validates resource pool env", func() {
			deployment := bmdepl.Deployment{
				ResourcePools: []bmdepl.ResourcePool{
					{
						RawEnv: map[interface{}]interface{}{
							123: "fake-env-value",
						},
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools[0].env must have only string keys"))
		})

		It("validates disk pool name", func() {
			deployment := bmdepl.Deployment{
				DiskPools: []bmdepl.DiskPool{
					{
						Name: "",
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("disk_pools[0].name must not be empty or blank"))

			deployment = bmdepl.Deployment{
				DiskPools: []bmdepl.DiskPool{
					{
						Name: "not-blank",
					},
					{
						Name: "   \t",
					},
				},
			}

			err = validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("disk_pools[1].name must not be empty or blank"))
		})

		It("validates disk pool size", func() {
			deployment := bmdepl.Deployment{
				DiskPools: []bmdepl.DiskPool{
					{
						Name: "fake-disk",
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("disk_pools[0].disk_size must be > 0"))
		})

		It("validates disk pool cloud_properties", func() {
			deployment := bmdepl.Deployment{
				DiskPools: []bmdepl.DiskPool{
					{
						RawCloudProperties: map[interface{}]interface{}{
							123: "fake-property-value",
						},
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("disk_pools[0].cloud_properties must have only string keys"))
		})

		It("validates network name", func() {
			deployment := bmdepl.Deployment{
				Networks: []bmdepl.Network{
					{
						Name: "",
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("networks[0].name must not be empty or blank"))

			deployment = bmdepl.Deployment{
				Networks: []bmdepl.Network{
					{
						Name: "not-blank",
					},
					{
						Name: "   \t",
					},
				},
			}

			err = validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("networks[1].name must not be empty or blank"))
		})

		It("validates network type", func() {
			deployment := bmdepl.Deployment{
				Networks: []bmdepl.Network{
					{
						Type: "unknown-type",
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("networks[0].type must be 'manual', 'dynamic', or 'vip'"))
		})

		It("validates disk pool cloud_properties", func() {
			deployment := bmdepl.Deployment{
				Networks: []bmdepl.Network{
					{
						RawCloudProperties: map[interface{}]interface{}{
							123: "fake-property-value",
						},
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("networks[0].cloud_properties must have only string keys"))
		})

		It("validates that there is only one job", func() {
			deployment := bmdepl.Deployment{
				Jobs: []bmdepl.Job{
					{},
					{},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs must be of size 1"))
		})

		It("validates job name", func() {
			deployment := bmdepl.Deployment{
				Jobs: []bmdepl.Job{
					{
						Name: "",
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].name must not be empty or blank"))

			deployment = bmdepl.Deployment{
				Jobs: []bmdepl.Job{
					{
						Name: "not-blank",
					},
					{
						Name: "   \t",
					},
				},
			}

			err = validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[1].name must not be empty or blank"))
		})

		It("validates job persistent_disk", func() {
			deployment := bmdepl.Deployment{
				Jobs: []bmdepl.Job{
					{
						PersistentDisk: -1234,
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].persistent_disk must be >= 0"))
		})

		It("validates job persistent_disk_pool", func() {
			deployment := bmdepl.Deployment{
				Jobs: []bmdepl.Job{
					{
						PersistentDiskPool: "non-existent-disk-pool",
					},
				},
				DiskPools: []bmdepl.DiskPool{
					{
						Name: "fake-disk-pool",
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].persistent_disk_pool must be the name of a disk pool"))
		})

		It("validates job instances", func() {
			deployment := bmdepl.Deployment{
				Jobs: []bmdepl.Job{
					{
						Instances: -1234,
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].instances must be >= 0"))
		})

		It("validates job networks", func() {
			deployment := bmdepl.Deployment{
				Jobs: []bmdepl.Job{
					{
						Networks: []bmdepl.JobNetwork{},
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].networks must be a non-empty array"))
		})

		It("validates job network name", func() {
			deployment := bmdepl.Deployment{
				Jobs: []bmdepl.Job{
					{
						Networks: []bmdepl.JobNetwork{
							{
								Name: "",
							},
						},
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].networks[0].name must not be empty or blank"))
		})

		It("validates job network static ips", func() {
			deployment := bmdepl.Deployment{
				Jobs: []bmdepl.Job{
					{
						Networks: []bmdepl.JobNetwork{
							{
								StaticIPs: []string{"non-ip"},
							},
						},
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].networks[0].static_ips[0] must be a valid IP"))
		})

		It("validates job network default", func() {
			deployment := bmdepl.Deployment{
				Jobs: []bmdepl.Job{
					{
						Networks: []bmdepl.JobNetwork{
							{
								Default: []bmdepl.NetworkDefault{
									"non-dns-or-gateway",
								},
							},
						},
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].networks[0].default[0] must be 'dns' or 'gateway'"))
		})

		It("validates job lifecycle", func() {
			deployment := bmdepl.Deployment{
				Jobs: []bmdepl.Job{
					{
						Lifecycle: "errand",
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].lifecycle must be 'service' ('errand' not supported)"))
		})

		It("validates job properties", func() {
			deployment := bmdepl.Deployment{
				Jobs: []bmdepl.Job{
					{
						RawProperties: map[interface{}]interface{}{
							123: "fake-property-value",
						},
					},
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].properties must have only string keys"))
		})

		It("validates deployment properties", func() {
			deployment := bmdepl.Deployment{
				RawProperties: map[interface{}]interface{}{
					123: "fake-property-value",
				},
			}

			err := validator.Validate(deployment)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("properties must have only string keys"))
		})
	})
})
