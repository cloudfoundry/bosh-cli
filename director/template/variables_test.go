package template_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/director/template"
)

var _ = Describe("Variables", func() {
	Describe("Merge", func() {
		It("merges the two sets into one", func() {
			a := Variables{"a": "foo"}
			b := Variables{"b": "bar"}

			Expect(a.Merge(b)).To(Equal(Variables{
				"a": "foo",
				"b": "bar",
			}))
		})

		It("does not affect the original sets", func() {
			a := Variables{"a": "foo"}
			b := Variables{"b": "bar"}

			a.Merge(b)

			Expect(a).To(Equal(Variables{
				"a": "foo",
			}))
		})

		It("overwrites the LHS with the RHS", func() {
			a := Variables{"a": "foo", "b": "old"}
			b := Variables{"b": "new"}

			Expect(a.Merge(b)).To(Equal(Variables{
				"a": "foo",
				"b": "new",
			}))
		})
	})
})
