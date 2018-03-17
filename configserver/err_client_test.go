package configserver_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/configserver"
)

var _ = Describe("ErrClient", func() {
	var (
		client ErrClient
	)

	BeforeEach(func() {
		client = NewErrClient()
	})

	Describe("Read", func() {
		It("returns error", func() {
			_, err := client.Read("test-key")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected to have configured config server"))
		})
	})

	Describe("Exists", func() {
		It("returns error", func() {
			_, err := client.Exists("test-key")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected to have configured config server"))
		})
	})

	Describe("Write", func() {
		It("returns error", func() {
			err := client.Write("test-key", nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected to have configured config server"))
		})
	})

	Describe("Delete", func() {
		It("returns error", func() {
			err := client.Delete("test-key")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected to have configured config server"))
		})
	})

	Describe("Generate", func() {
		It("returns error", func() {
			_, err := client.Generate("test-key", "test-type", nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected to have configured config server"))
		})
	})
})
