package pkg_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmrelpkg "github.com/cloudfoundry/bosh-init/release/pkg"

	. "github.com/cloudfoundry/bosh-init/state/pkg"
)

var _ = Describe("DependencyResolver", func() {
	It("supports a single dependency", func() {
		a := bmrelpkg.Package{Name: "a"}
		b := bmrelpkg.Package{Name: "b"}
		a.Dependencies = []*bmrelpkg.Package{&b}

		deps := ResolveDependencies(&a)
		Expect(deps).To(Equal([]*bmrelpkg.Package{&b}))
	})

	It("supports a transitive dependency", func() {
		a := bmrelpkg.Package{Name: "a"}
		b := bmrelpkg.Package{Name: "b"}
		a.Dependencies = []*bmrelpkg.Package{&b}
		c := bmrelpkg.Package{Name: "c"}
		b.Dependencies = []*bmrelpkg.Package{&c}

		deps := ResolveDependencies(&a)
		Expect(deps).To(ContainElement(&b))
		Expect(deps).To(ContainElement(&c))
		Expect(len(deps)).To(Equal(2))
	})

	It("supports simple cycles", func() {
		a := bmrelpkg.Package{Name: "a"}
		b := bmrelpkg.Package{Name: "b"}
		a.Dependencies = []*bmrelpkg.Package{&b}
		b.Dependencies = []*bmrelpkg.Package{&a}

		deps := ResolveDependencies(&a)
		Expect(deps).ToNot(ContainElement(&a))
		Expect(deps).To(ContainElement(&b))
		Expect(len(deps)).To(Equal(1))
	})

	It("supports triangular cycles", func() {
		a := bmrelpkg.Package{Name: "a"}
		b := bmrelpkg.Package{Name: "b"}
		a.Dependencies = []*bmrelpkg.Package{&b}
		c := bmrelpkg.Package{Name: "c"}
		b.Dependencies = []*bmrelpkg.Package{&c}
		c.Dependencies = []*bmrelpkg.Package{&a}

		deps := ResolveDependencies(&a)
		Expect(deps).ToNot(ContainElement(&a))
		Expect(deps).To(ContainElement(&b))
		Expect(deps).To(ContainElement(&c))
		Expect(len(deps)).To(Equal(2))
	})
})
