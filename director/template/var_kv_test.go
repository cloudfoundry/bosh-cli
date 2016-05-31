package template_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/director/template"
)

var _ = Describe("VarKV", func() {
	Describe("UnmarshalFlag", func() {
		var (
			arg VarKV
		)

		BeforeEach(func() {
			arg = VarKV{}
		})

		It("sets name and value", func() {
			err := (&arg).UnmarshalFlag("name=val")
			Expect(err).ToNot(HaveOccurred())
			Expect(arg).To(Equal(VarKV{Name: "name", Value: "val"}))
		})

		It("returns error if string does not have 2 pieces", func() {
			err := (&arg).UnmarshalFlag("val")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected var 'val' to be in format 'name=value'"))
		})

		It("returns error if name is empty", func() {
			err := (&arg).UnmarshalFlag("=val")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected var '=val' to specify non-empty name"))
		})

		It("returns error if value is empty", func() {
			err := (&arg).UnmarshalFlag("name=")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected var 'name=' to specify non-empty value"))
		})
	})
})
