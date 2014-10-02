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
networks:
- name: fake-network-name
  type: dynamic
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
				Expect(deployment.Networks).To(Equal([]Network{
					{
						Name: "fake-network-name",
						Type: Dynamic,
					},
				}))
			})
		})
	})
})
