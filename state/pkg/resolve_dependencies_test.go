package pkg_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshrelpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
	. "github.com/cloudfoundry/bosh-cli/v7/state/pkg"
)

var _ = Describe("DependencyResolver", func() {
	It("supports a single dependency", func() {
		a := newPkg("a", "", []string{"b"})
		b := newPkg("b", "", nil)
		err := a.AttachDependencies([]*boshrelpkg.Package{b})
		Expect(err).ToNot(HaveOccurred())

		deps := ResolveDependencies(a)
		Expect(deps).To(Equal([]boshrelpkg.Compilable{b}))
	})

	It("supports a transitive dependency", func() {
		a := newPkg("a", "", []string{"b"})
		b := newPkg("b", "", []string{"c"})
		c := newPkg("c", "", nil)
		err := a.AttachDependencies([]*boshrelpkg.Package{b})
		Expect(err).ToNot(HaveOccurred())
		err = b.AttachDependencies([]*boshrelpkg.Package{c})
		Expect(err).ToNot(HaveOccurred())

		deps := ResolveDependencies(a)
		Expect(deps).To(Equal([]boshrelpkg.Compilable{c, b}))
	})

	It("supports simple cycles", func() {
		a := newPkg("a", "", []string{"b"})
		b := newPkg("b", "", []string{"a"})
		err := a.AttachDependencies([]*boshrelpkg.Package{b})
		Expect(err).ToNot(HaveOccurred())
		err = b.AttachDependencies([]*boshrelpkg.Package{a})
		Expect(err).ToNot(HaveOccurred())

		deps := ResolveDependencies(a)
		Expect(deps).ToNot(ContainElement(a))
		Expect(deps).To(Equal([]boshrelpkg.Compilable{b}))
	})

	It("supports triangular cycles", func() {
		a := newPkg("a", "", []string{"b"})
		b := newPkg("b", "", []string{"c"})
		c := newPkg("c", "", []string{"a"})
		err := a.AttachDependencies([]*boshrelpkg.Package{b})
		Expect(err).ToNot(HaveOccurred())
		err = b.AttachDependencies([]*boshrelpkg.Package{c})
		Expect(err).ToNot(HaveOccurred())
		err = c.AttachDependencies([]*boshrelpkg.Package{a})
		Expect(err).ToNot(HaveOccurred())

		deps := ResolveDependencies(a)
		Expect(deps).ToNot(ContainElement(a))
		Expect(deps).To(Equal([]boshrelpkg.Compilable{c, b}))
	})

	It("supports no cycles", func() {
		a := newPkg("a", "", []string{"b", "c"})
		b := newPkg("b", "", nil)
		c := newPkg("c", "", []string{"b"})
		err := a.AttachDependencies([]*boshrelpkg.Package{b, c})
		Expect(err).ToNot(HaveOccurred())
		err = c.AttachDependencies([]*boshrelpkg.Package{b})
		Expect(err).ToNot(HaveOccurred())

		deps := ResolveDependencies(a)
		Expect(deps).ToNot(ContainElement(a))
		Expect(deps).To(Equal([]boshrelpkg.Compilable{c, b}))
	})

	It("supports diamond cycles", func() {
		a := newPkg("a", "", []string{"c"})
		b := newPkg("b", "", []string{"a"})
		c := newPkg("c", "", []string{"d"})
		d := newPkg("d", "", []string{"b"})
		err := a.AttachDependencies([]*boshrelpkg.Package{c})
		Expect(err).ToNot(HaveOccurred())
		err = b.AttachDependencies([]*boshrelpkg.Package{a})
		Expect(err).ToNot(HaveOccurred())
		err = c.AttachDependencies([]*boshrelpkg.Package{d})
		Expect(err).ToNot(HaveOccurred())
		err = d.AttachDependencies([]*boshrelpkg.Package{b})
		Expect(err).ToNot(HaveOccurred())

		deps := ResolveDependencies(a)
		Expect(deps).ToNot(ContainElement(a))
		Expect(deps).To(Equal([]boshrelpkg.Compilable{b, d, c}))
	})

	It("supports sibling dependencies", func() {
		a := newPkg("a", "", []string{"b", "c"})
		b := newPkg("b", "", []string{"c", "d"})
		c := newPkg("c", "", []string{"d"})
		d := newPkg("d", "", nil)
		err := a.AttachDependencies([]*boshrelpkg.Package{b, c})
		Expect(err).ToNot(HaveOccurred())
		err = b.AttachDependencies([]*boshrelpkg.Package{c, d})
		Expect(err).ToNot(HaveOccurred())
		err = c.AttachDependencies([]*boshrelpkg.Package{d})
		Expect(err).ToNot(HaveOccurred())

		deps := ResolveDependencies(a)
		Expect(deps).ToNot(ContainElement(a))
		Expect(deps).To(Equal([]boshrelpkg.Compilable{d, c, b}))
	})
})
