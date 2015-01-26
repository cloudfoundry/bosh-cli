package manifest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
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

			It("is a map of the network names to network interfaces", func() {
				Expect(deploymentManifest.NetworkInterfaces("fake-job-name")).To(Equal(map[string]map[string]interface{}{
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
				Expect(deploymentManifest.NetworkInterfaces("fake-job-name")).To(Equal(map[string]map[string]interface{}{}))
			})
		})

		Context("when the deployment does not have a job with requested name", func() {
			BeforeEach(func() {
				deploymentManifest = Manifest{}
			})

			It("returns an error", func() {
				networkInterfaces, err := deploymentManifest.NetworkInterfaces("fake-job-name")
				Expect(networkInterfaces).To(Equal(map[string]map[string]interface{}{}))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Could not find job with name: fake-job-name"))
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
							RawCloudProperties: map[interface{}]interface{}{
								"fake-disk-prop-key-1": "fake-disk-prop-value-1",
							},
						},
						{
							Name:     "fake-disk-pool-name-2",
							DiskSize: 2048,
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
				diskPool, err := deploymentManifest.DiskPool("fake-job-name")
				Expect(err).ToNot(HaveOccurred())

				Expect(diskPool).To(Equal(DiskPool{
					Name:     "fake-disk-pool-name-2",
					DiskSize: 2048,
					RawCloudProperties: map[interface{}]interface{}{
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
					Name:               "",
					DiskSize:           1024,
					RawCloudProperties: map[interface{}]interface{}{},
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
				diskPool, err := deploymentManifest.DiskPool("fake-job-name")
				Expect(err).ToNot(HaveOccurred())

				Expect(diskPool).To(Equal(DiskPool{
					Name:     "fake-disk-pool-name-1",
					DiskSize: 1024,
					RawCloudProperties: map[interface{}]interface{}{
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
