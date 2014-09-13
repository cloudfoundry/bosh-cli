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
		manifestParser = NewCpiDeploymentParser(fakeFs)
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
			Context("when converting properties succeeds", func() {
				BeforeEach(func() {
					contents := `
---
name: fake-deployment-name
cloud_provider:
  properties:
    fake-property-name:
      nested-property: fake-property-value
`
					fakeFs.WriteFileString(deploymentPath, contents)
				})

				It("parses deployment manifest", func() {
					deployment, err := manifestParser.Parse(deploymentPath)
					Expect(err).ToNot(HaveOccurred())

					Expect(deployment.Name).To(Equal("fake-deployment-name"))
					Expect(deployment.Properties["fake-property-name"]).To(Equal(map[string]interface{}{
						"nested-property": "fake-property-value",
					}))
				})

				It("sets a CPI job into the deployment", func() {
					deployment, err := manifestParser.Parse(deploymentPath)
					Expect(err).ToNot(HaveOccurred())
					expectedJobs := []Job{
						Job{
							Name:      "cpi",
							Instances: 1,
							Templates: []ReleaseJobRef{
								ReleaseJobRef{
									Name:    "cpi",
									Release: "unknown-cpi-release-name",
								},
							},
						},
					}
					Expect(deployment.Jobs).To(Equal(expectedJobs))
				})
			})

			Context("when parsing properties fails", func() {
				BeforeEach(func() {
					contents := `
---
name: fake-deployment-name
cloud_provider:
  properties:
    123: fake-property-value
`
					fakeFs.WriteFileString(deploymentPath, contents)
				})

				It("returns an error", func() {
					_, err := manifestParser.Parse(deploymentPath)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Converting manifest cloud properties"))
				})
			})
		})
	})
})
