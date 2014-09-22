package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/config"
)

var _ = Describe("Config", func() {
	Describe("Paths", func() {
		var c DeploymentConfig
		BeforeEach(func() {
			c = DeploymentConfig{
				ContainingDir:  "/home/fake",
				DeploymentUUID: "madcow",
			}
		})

		It("returns the blobstore path", func() {
			Expect(c.BlobstorePath()).To(Equal("/home/fake/madcow/blobs"))
		})

		It("returns the compiled packages index path", func() {
			Expect(c.CompiledPackagedIndexPath()).To(Equal("/home/fake/madcow/compiled_packages.json"))
		})

		It("returns the templates index path", func() {
			Expect(c.TemplatesIndexPath()).To(Equal("/home/fake/madcow/templates.json"))
		})

		It("returns the packages path", func() {
			Expect(c.PackagesPath()).To(Equal("/home/fake/madcow/packages"))
		})
	})
})
