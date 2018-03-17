package configserver_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/configserver"
)

var _ = Describe("MemoryClient", func() {
	var (
		client *MemoryClient
	)

	BeforeEach(func() {
		client = NewMemoryClient()
	})

	Describe("Read/Exists/Write", func() {
		It("saves and retrieves values", func() {
			_, err := client.Read("test-key")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected to find 'test-key'"))

			found, err := client.Exists("test-key")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())

			err = client.Write("test-key", "test-value")
			Expect(err).ToNot(HaveOccurred())

			val, err := client.Read("test-key")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal("test-value"))

			found, err = client.Exists("test-key")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())

			err = client.Write("test-key", "test-value2")
			Expect(err).ToNot(HaveOccurred())

			val, err = client.Read("test-key")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal("test-value2"))

			found, err = client.Exists("test-key")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
		})
	})

	Describe("Delete", func() {
		It("deletes values", func() {
			err := client.Write("test-key", "test-value")
			Expect(err).ToNot(HaveOccurred())

			found, err := client.Exists("test-key")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())

			err = client.Delete("test-key")
			Expect(err).ToNot(HaveOccurred())

			found, err = client.Exists("test-key")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
		})

		It("deletes values successfully that do not exist", func() {
			err := client.Delete("test-key-does-not-exist")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Generate", func() {
		It("returns error", func() {
			_, err := client.Generate("test-key", "test-type", nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Memory config server client does not support value generation"))
		})
	})
})
