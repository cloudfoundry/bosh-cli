package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/config"
)

var _ = Describe("Config", func() {
	It("returns correct deploymentFile when deployment is set", func() {
		c := Config{Deployment: "/fake-path/manifest.yml"}
		Expect(c.DeploymentFile()).To(Equal("/fake-path/deployment.json"))
	})

	Describe("Paths", func() {
		var c Config
		BeforeEach(func() {
			c = Config{
				ContainingDir:  "/home/fake",
				DeploymentUUID: "madcow",
			}
		})

		It("returns the blobstore path", func() {
			Expect(c.BlobstorePath()).To(Equal("/home/fake/.bosh_micro/madcow/blobs"))
		})

		It("returns the compiled packages index path", func() {
			Expect(c.CompiledPackagedIndexPath()).To(Equal("/home/fake/.bosh_micro/madcow/compiled_packages.json"))
		})

		It("returns the templates index path", func() {
			Expect(c.TemplatesIndexPath()).To(Equal("/home/fake/.bosh_micro/madcow/templates.json"))
		})

		It("returns the packages path", func() {
			Expect(c.PackagesPath()).To(Equal("/home/fake/.bosh_micro/madcow/packages"))
		})
	})
})
