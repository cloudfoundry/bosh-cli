package manifest_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmrelmanifest "github.com/cloudfoundry/bosh-micro-cli/release/manifest"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
)

var _ = Describe("Parser", func() {
	var (
		comboManifestPath string
		fakeFs            *fakesys.FakeFileSystem
		parser            Parser
	)

	BeforeEach(func() {
		comboManifestPath = "fake-deployment-path"
		fakeFs = fakesys.NewFakeFileSystem()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		parser = NewParser(fakeFs, logger)
	})

	Context("when combo manifest path does not exist", func() {
		BeforeEach(func() {
			err := fakeFs.RemoveAll(comboManifestPath)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error", func() {
			_, err := parser.Parse(comboManifestPath)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when parser fails to read the combo manifest file", func() {
		BeforeEach(func() {
			fakeFs.ReadFileError = errors.New("fake-read-file-error")
		})

		It("returns an error", func() {
			_, err := parser.Parse(comboManifestPath)
			Expect(err).To(HaveOccurred())
		})
	})

	BeforeEach(func() {
		contents := `
---
name: fake-deployment-name
releases:
- name: fake-release-name
  version: fake-release-version
update:
  update_watch_time: 2000-7000
resource_pools:
- name: fake-resource-pool-name
  env:
    bosh:
      password: secret
networks:
- name: fake-network-name
  type: dynamic
  cloud_properties:
    subnet: fake-subnet
    a:
      b: value
- name: vip
  type: vip
disk_pools:
- name: fake-disk-pool-name
  disk_size: 2048
  cloud_properties:
    fake-disk-pool-cloud-property-key: fake-disk-pool-cloud-property-value
jobs:
- name: bosh
  networks:
  - name: vip
    static_ips: [1.2.3.4]
  persistent_disk: 1024
  persistent_disk_pool: fake-disk-pool-name
  properties:
    fake-prop-key:
      nested-prop-key: fake-prop-value
`
		fakeFs.WriteFileString(comboManifestPath, contents)
	})

	It("parses deployment manifest from combo manifest file", func() {
		deploymentManifest, err := parser.Parse(comboManifestPath)
		Expect(err).ToNot(HaveOccurred())

		Expect(deploymentManifest).To(Equal(Manifest{
			Name: "fake-deployment-name",
			Releases: []bmrelmanifest.ReleaseRef{
				{
					Name:    "fake-release-name",
					Version: "fake-release-version",
				},
			},
			Update: Update{
				UpdateWatchTime: WatchTime{
					Start: 2000,
					End:   7000,
				},
			},
			Networks: []Network{
				{
					Name: "fake-network-name",
					Type: Dynamic,
					RawCloudProperties: map[interface{}]interface{}{
						"subnet": "fake-subnet",
						"a": map[interface{}]interface{}{
							"b": "value",
						},
					},
				},
				{
					Name: "vip",
					Type: VIP,
				},
			},
			ResourcePools: []ResourcePool{
				{
					Name: "fake-resource-pool-name",
					RawEnv: map[interface{}]interface{}{
						"bosh": map[interface{}]interface{}{
							"password": "secret",
						},
					},
				},
			},
			DiskPools: []DiskPool{
				{
					Name:     "fake-disk-pool-name",
					DiskSize: 2048,
					RawCloudProperties: map[interface{}]interface{}{
						"fake-disk-pool-cloud-property-key": "fake-disk-pool-cloud-property-value",
					},
				},
			},
			Jobs: []Job{
				{
					Name: "bosh",
					Networks: []JobNetwork{
						{
							Name:      "vip",
							StaticIPs: []string{"1.2.3.4"},
						},
					},
					PersistentDisk:     1024,
					PersistentDiskPool: "fake-disk-pool-name",
					RawProperties: map[interface{}]interface{}{
						"fake-prop-key": map[interface{}]interface{}{
							"nested-prop-key": "fake-prop-value",
						},
					},
				},
			},
		}))
	})

	Context("when update watch time is not set", func() {
		BeforeEach(func() {
			contents := `
---
name: fake-deployment-name
`
			fakeFs.WriteFileString(comboManifestPath, contents)
		})

		It("uses default values", func() {
			deploymentManifest, err := parser.Parse(comboManifestPath)
			Expect(err).ToNot(HaveOccurred())

			Expect(deploymentManifest.Name).To(Equal("fake-deployment-name"))
			Expect(deploymentManifest.Update.UpdateWatchTime.Start).To(Equal(0))
			Expect(deploymentManifest.Update.UpdateWatchTime.End).To(Equal(300000))
		})
	})
})
