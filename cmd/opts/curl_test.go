package opts_test

import (
	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CurlHeader", func() {
	Describe("UnmarshalFlag", func() {
		var (
			arg CurlHeader
		)

		BeforeEach(func() {
			arg = CurlHeader{}
		})

		It("sets name and value", func() {
			err := (&arg).UnmarshalFlag("name: val")
			Expect(err).ToNot(HaveOccurred())
			Expect(arg).To(Equal(CurlHeader{Name: "name", Value: "val"}))
		})

		It("sets name and value when value contains a `: `", func() {
			err := (&arg).UnmarshalFlag("name: val: ue")
			Expect(err).ToNot(HaveOccurred())
			Expect(arg).To(Equal(CurlHeader{Name: "name", Value: "val: ue"}))
		})

		It("returns error if string does not have 2 pieces", func() {
			err := (&arg).UnmarshalFlag("val")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected header 'val' to be in format 'name: value'"))
		})

		It("returns error if name is empty", func() {
			err := (&arg).UnmarshalFlag(": val")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected header ': val' to specify non-empty name"))
		})

		It("returns error if value is empty", func() {
			err := (&arg).UnmarshalFlag("name: ")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected header 'name: ' to specify non-empty value"))
		})
	})
})
