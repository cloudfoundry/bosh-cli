package deployment_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

var _ = Describe("DeploymentRenderer", func() {
	var (
		deploymentPath string
		fakeFs         *fakesys.FakeFileSystem
		manifestParser ManifestParser
	)

	BeforeEach(func() {
		deploymentPath = "fake-deployment-path"
		fakeFs = fakesys.NewFakeFileSystem()
		manifestParser = NewBoshDeploymentParser(fakeFs)
	})

	Context("when deployment path does not exist", func() {
		It("returns an error", func() {
			_, err := manifestParser.Parse(deploymentPath)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when deployment path exists", func() {
		Context("when parser fails to read the deployment file", func() {
			BeforeEach(func() {
				fakeFs.ReadFileError = errors.New("fake-read-file-error")
			})

			It("returns an error", func() {
				_, err := manifestParser.Parse(deploymentPath)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when parser successfully reads the deployment file", func() {
			BeforeEach(func() {
				contents := `
---
name: fake-deployment-name
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
cloud_provider:
  properties:
    nested-property: fake-property-value
`
				fakeFs.WriteFileString(deploymentPath, contents)
			})

			It("parses deployment manifest", func() {
				deployment, err := manifestParser.Parse(deploymentPath)
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.Name).To(Equal("fake-deployment-name"))
				networks := deployment.Networks
				Expect(networks).To(Equal([]Network{
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
				}))
				resourcePools := deployment.ResourcePools
				Expect(resourcePools).To(Equal([]ResourcePool{
					{
						Name: "fake-resource-pool-name",
						RawEnv: map[interface{}]interface{}{
							"bosh": map[interface{}]interface{}{
								"password": "secret",
							},
						},
					},
				}))
			})
		})
	})
})
