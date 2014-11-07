package deployment_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

var _ = Describe("Deployment", func() {
	var (
		deployment Deployment
	)

	Describe("NetworksSpec", func() {
		Context("when the deployment has networks", func() {
			BeforeEach(func() {
				deployment = Deployment{
					Networks: []Network{
						{
							Name: "fake-network-name",
							Type: "dynamic",
						},
						{
							Name: "fake-manual-network-name",
							Type: "manual",
						},
						{
							Name: "vip",
							Type: "vip",
						},
						{
							Name: "fake",
							Type: "dynamic",
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

			It("is a map of the networks in spec form", func() {
				Expect(deployment.NetworksSpec("fake-job-name")).To(Equal(map[string]interface{}{
					"fake-network-name": map[string]interface{}{
						"type":             "dynamic",
						"ip":               "5.6.7.8",
						"cloud_properties": map[string]interface{}{},
					},
					"fake-manual-network-name": map[string]interface{}{
						"type":             "manual",
						"ip":               "5.6.7.9",
						"cloud_properties": map[string]interface{}{},
					},
					"vip": map[string]interface{}{
						"type":             "vip",
						"ip":               "1.2.3.4",
						"cloud_properties": map[string]interface{}{},
					},
				}))
			})
		})

		Context("when the deployment does not have networks", func() {
			BeforeEach(func() {
				deployment = Deployment{
					Jobs: []Job{
						{
							Name: "fake-job-name",
						},
					},
					Networks: []Network{},
				}
			})

			It("is an empty map", func() {
				Expect(deployment.NetworksSpec("fake-job-name")).To(Equal(map[string]interface{}{}))
			})
		})

		Context("when the deployment does not have a job with requested name", func() {
			BeforeEach(func() {
				deployment = Deployment{}
			})

			It("returns an error", func() {
				networksSpec, err := deployment.NetworksSpec("fake-job-name")
				Expect(networksSpec).To(Equal(map[string]interface{}{}))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Could not find job with name: fake-job-name"))
			})
		})
	})

	Describe("DiskPool", func() {
		Context("when the deployment has disk_pools", func() {
			BeforeEach(func() {
				deployment = Deployment{
					DiskPools: []DiskPool{
						{
							Name: "fake-disk-pool-name-1",
							Size: 1024,
							RawCloudProperties: map[interface{}]interface{}{
								"fake-disk-prop-key-1": "fake-disk-prop-value-1",
							},
						},
						{
							Name: "fake-disk-pool-name-2",
							Size: 2048,
							RawCloudProperties: map[interface{}]interface{}{
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
				diskPool, err := deployment.DiskPool("fake-job-name")
				Expect(err).ToNot(HaveOccurred())

				Expect(diskPool).To(Equal(DiskPool{
					Name: "fake-disk-pool-name-2",
					Size: 2048,
					RawCloudProperties: map[interface{}]interface{}{
						"fake-disk-prop-key-2": "fake-disk-prop-value-1",
					},
				}))
			})
		})

		Context("when job has persistent_disk and there are no disk_pools", func() {
			BeforeEach(func() {
				deployment = Deployment{
					Jobs: []Job{
						{
							Name:           "fake-job-name",
							PersistentDisk: 1024,
						},
					},
				}
			})

			It("is a new disk pool with the specified persistent disk size", func() {
				diskPool, err := deployment.DiskPool("fake-job-name")
				Expect(err).ToNot(HaveOccurred())

				Expect(diskPool).To(Equal(DiskPool{
					Name:               "",
					Size:               1024,
					RawCloudProperties: map[interface{}]interface{}{},
				}))
			})
		})

		Context("when job has persistent_disk_pool and persistent_disk", func() {
			BeforeEach(func() {
				deployment = Deployment{
					DiskPools: []DiskPool{
						{
							Name: "fake-disk-pool-name-1",
							Size: 1024,
							RawCloudProperties: map[interface{}]interface{}{
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
				diskPool, err := deployment.DiskPool("fake-job-name")
				Expect(err).ToNot(HaveOccurred())

				Expect(diskPool).To(Equal(DiskPool{
					Name: "fake-disk-pool-name-1",
					Size: 1024,
					RawCloudProperties: map[interface{}]interface{}{
						"fake-disk-prop-key-1": "fake-disk-prop-value-1",
					},
				}))
			})
		})

		Context("when job has persistent_disk_pool but no matching disk pool exists", func() {
			BeforeEach(func() {
				deployment = Deployment{
					Jobs: []Job{
						{
							Name:               "fake-job-name",
							PersistentDiskPool: "fake-disk-pool-name-1",
						},
					},
				}
			})

			It("returns an error", func() {
				_, err := deployment.DiskPool("fake-job-name")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Could not find persistent disk pool 'fake-disk-pool-name-1' for job 'fake-job-name'"))
			})
		})

		Context("when job does not have persistent_disk_pool or persistent_disk", func() {
			BeforeEach(func() {
				deployment = Deployment{
					Jobs: []Job{
						{
							Name: "fake-job-name",
						},
					},
				}
			})

			It("returns an empty disk pool", func() {
				diskPool, err := deployment.DiskPool("fake-job-name")
				Expect(err).ToNot(HaveOccurred())
				Expect(diskPool).To(Equal(DiskPool{}))
			})
		})
	})
})
