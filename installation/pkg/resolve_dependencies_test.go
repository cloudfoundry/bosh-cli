package pkg_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	. "github.com/cloudfoundry/bosh-micro-cli/installation/pkg"
)

var _ = Describe("DependencyResolver", func() {
	It("supports a single dependency", func() {
		a := bmrel.Package{Name: "a"}
		b := bmrel.Package{Name: "b"}
		a.Dependencies = []*bmrel.Package{&b}

		deps := ResolveDependencies(&a)
		Expect(deps).To(Equal([]*bmrel.Package{&b}))
	})

	It("supports a transitive dependency", func() {
		a := bmrel.Package{Name: "a"}
		b := bmrel.Package{Name: "b"}
		a.Dependencies = []*bmrel.Package{&b}
		c := bmrel.Package{Name: "c"}
		b.Dependencies = []*bmrel.Package{&c}

		deps := ResolveDependencies(&a)
		Expect(deps).To(ContainElement(&b))
		Expect(deps).To(ContainElement(&c))
		Expect(len(deps)).To(Equal(2))
	})

	It("supports simple cycles", func() {
		a := bmrel.Package{Name: "a"}
		b := bmrel.Package{Name: "b"}
		a.Dependencies = []*bmrel.Package{&b}
		b.Dependencies = []*bmrel.Package{&a}

		deps := ResolveDependencies(&a)
		Expect(deps).ToNot(ContainElement(&a))
		Expect(deps).To(ContainElement(&b))
		Expect(len(deps)).To(Equal(1))
	})

	It("supports triangular cycles", func() {
		a := bmrel.Package{Name: "a"}
		b := bmrel.Package{Name: "b"}
		a.Dependencies = []*bmrel.Package{&b}
		c := bmrel.Package{Name: "c"}
		b.Dependencies = []*bmrel.Package{&c}
		c.Dependencies = []*bmrel.Package{&a}

		deps := ResolveDependencies(&a)
		Expect(deps).ToNot(ContainElement(&a))
		Expect(deps).To(ContainElement(&b))
		Expect(deps).To(ContainElement(&c))
		Expect(len(deps)).To(Equal(2))
	})
})
