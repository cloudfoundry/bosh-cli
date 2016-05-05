package pkg_test

import (
	birelpkg "github.com/cloudfoundry/bosh-init/release/pkg"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/state/pkg"
)

var _ = Describe("DependencyResolver", func() {
	It("supports a single dependency", func() {
		a := birelpkg.Package{Name: "a"}
		b := birelpkg.Package{Name: "b"}
		a.Dependencies = []*birelpkg.Package{&b}

		deps := ResolveDependencies(&a)
		Expect(deps).To(Equal([]*birelpkg.Package{&b}))
	})

	It("supports a transitive dependency", func() {
		a := birelpkg.Package{Name: "a"}
		b := birelpkg.Package{Name: "b"}
		a.Dependencies = []*birelpkg.Package{&b}
		c := birelpkg.Package{Name: "c"}
		b.Dependencies = []*birelpkg.Package{&c}

		deps := ResolveDependencies(&a)
		Expect(deps).To(ContainElement(&b))
		Expect(deps).To(ContainElement(&c))
		Expect(len(deps)).To(Equal(2))
	})

	It("supports simple cycles", func() {
		a := birelpkg.Package{Name: "a"}
		b := birelpkg.Package{Name: "b"}
		a.Dependencies = []*birelpkg.Package{&b}
		b.Dependencies = []*birelpkg.Package{&a}

		deps := ResolveDependencies(&a)
		Expect(deps).ToNot(ContainElement(&a))
		Expect(deps).To(ContainElement(&b))
		Expect(len(deps)).To(Equal(1))
	})

	It("supports triangular cycles", func() {
		a := birelpkg.Package{Name: "a"}
		b := birelpkg.Package{Name: "b"}
		a.Dependencies = []*birelpkg.Package{&b}
		c := birelpkg.Package{Name: "c"}
		b.Dependencies = []*birelpkg.Package{&c}
		c.Dependencies = []*birelpkg.Package{&a}

		deps := ResolveDependencies(&a)
		Expect(deps).ToNot(ContainElement(&a))
		Expect(deps).To(ContainElement(&b))
		Expect(deps).To(ContainElement(&c))
		Expect(len(deps)).To(Equal(2))
	})
})
