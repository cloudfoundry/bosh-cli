package deployment_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

var _ = Describe("Parser", func() {
	var (
		deploymentPath string
		fakeFs         *fakesys.FakeFileSystem
		parser         Parser
	)

	BeforeEach(func() {
		deploymentPath = "fake-deployment-path"
		fakeFs = fakesys.NewFakeFileSystem()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		parser = NewParser(fakeFs, logger)
	})

	Context("when deployment path does not exist", func() {
		It("returns an error", func() {
			_, _, err := parser.Parse(deploymentPath)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when deployment path exists", func() {
		Context("when parser fails to read the deployment file", func() {
			BeforeEach(func() {
				fakeFs.ReadFileError = errors.New("fake-read-file-error")
			})

			It("returns an error", func() {
				_, _, err := parser.Parse(deploymentPath)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when parser successfully reads the deployment file", func() {
			BeforeEach(func() {
				contents := `
---
name: fake-deployment-name
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
cloud_provider:
  ssh_tunnel:
    host: 54.34.56.8
    port: 22
    user: fake-ssh-user
    private_key: /tmp/fake-ssh-key.pem
  agent_env_service: registry
  mbus: http://fake-mbus-user:fake-mbus-password@0.0.0.0:6868
  registry:
    username: fake-registry-username
    password: fake-registry-password
    host: fake-registry-host
    port: 123
  properties:
    fake-property-name:
      nested-property: fake-property-value
`
				fakeFs.WriteFileString(deploymentPath, contents)
			})

			It("parses deployment from manifest", func() {
				deployment, _, err := parser.Parse(deploymentPath)
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment).To(Equal(Deployment{
					Name: "fake-deployment-name",
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

			It("parses cpi deployment from manifest", func() {
				_, cpiDeploymentManifest, err := parser.Parse(deploymentPath)
				Expect(err).ToNot(HaveOccurred())

				Expect(cpiDeploymentManifest).To(Equal(CPIDeploymentManifest{
					Name: "fake-deployment-name",
					Registry: Registry{
						Username: "fake-registry-username",
						Password: "fake-registry-password",
						Host:     "fake-registry-host",
						Port:     123,
					},
					AgentEnvService: "registry",
					RawProperties: map[interface{}]interface{}{
						"fake-property-name": map[interface{}]interface{}{
							"nested-property": "fake-property-value",
						},
					},
					SSHTunnel: SSHTunnel{
						Host:       "54.34.56.8",
						Port:       22,
						User:       "fake-ssh-user",
						PrivateKey: "/tmp/fake-ssh-key.pem",
					},
					Mbus: "http://fake-mbus-user:fake-mbus-password@0.0.0.0:6868",
				}))
			})

			Context("when update watch time is not set", func() {
				BeforeEach(func() {
					contents := `
---
name: fake-deployment-name
`
					fakeFs.WriteFileString(deploymentPath, contents)
				})

				It("uses default values", func() {
					deployment, _, err := parser.Parse(deploymentPath)
					Expect(err).ToNot(HaveOccurred())

					Expect(deployment.Name).To(Equal("fake-deployment-name"))
					Expect(deployment.Update.UpdateWatchTime.Start).To(Equal(0))
					Expect(deployment.Update.UpdateWatchTime.End).To(Equal(300000))
				})
			})
		})
	})
})
