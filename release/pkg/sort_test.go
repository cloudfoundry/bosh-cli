package pkg_test

import (
	"fmt"

	. "github.com/cloudfoundry/bosh-init/release/pkg"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gomegafmt "github.com/onsi/gomega/format"
)

var _ = Describe("Sort", func() {
	var (
		packages []*Package
	)

	gomegafmt.UseStringerRepresentation = true

	var indexOf = func(packages []*Package, pkg *Package) int {
		for index, currentPkg := range packages {
			if currentPkg == pkg {
				return index
			}
		}
		return -1
	}

	var expectSorted = func(sortedPackages []*Package) {
		for _, pkg := range packages {
			sortedIndex := indexOf(sortedPackages, pkg)
			for _, dependencyPkg := range pkg.Dependencies {
				errorMessage := fmt.Sprintf("Package '%s' should be compiled after package '%s'", pkg.Name, dependencyPkg.Name)
				Expect(sortedIndex).To(BeNumerically(">", indexOf(sortedPackages, dependencyPkg)), errorMessage)
			}
		}
	}

	var package1, package2 Package

	BeforeEach(func() {
		package1 = Package{
			Name: "fake-package-name-1",
		}
		package2 = Package{
			Name: "fake-package-name-2",
		}
		packages = []*Package{&package1, &package2}
	})

	Context("disjoint packages have a valid compilation sequence", func() {
		It("returns an ordered set of package compilation", func() {
			sortedPackages := Sort(packages)

			Expect(sortedPackages).To(ContainElement(&package1))
			Expect(sortedPackages).To(ContainElement(&package2))
		})
	})

	Context("dependent packages", func() {
		BeforeEach(func() {
			package1.Dependencies = []*Package{&package2}
		})

		It("returns an ordered set of package compilation", func() {
			sortedPackages := Sort(packages)

			Expect(sortedPackages).To(Equal([]*Package{&package2, &package1}))
		})
	})

	Context("complex graph of dependent packages", func() {
		var package3, package4 Package

		BeforeEach(func() {
			package1.Dependencies = []*Package{&package2, &package3}
			package3 = Package{
				Name: "fake-package-name-3",
			}
			package4 = Package{
				Name:         "fake-package-name-4",
				Dependencies: []*Package{&package3, &package2},
			}
			packages = []*Package{&package1, &package2, &package3, &package4}
		})

		It("returns an ordered set of package compilation", func() {
			sortedPackages := Sort(packages)

			expectSorted(sortedPackages)
		})
	})

	Context("graph with transitively dependent packages", func() {
		var package3, package4, package5 Package

		BeforeEach(func() {
			package3 = Package{
				Name: "fake-package-name-3",
			}
			package4 = Package{
				Name: "fake-package-name-4",
			}
			package5 = Package{
				Name: "fake-package-name-5",
			}

			package3.Dependencies = []*Package{&package2}
			package2.Dependencies = []*Package{&package1}

			package5.Dependencies = []*Package{&package2}

			packages = []*Package{&package1, &package2, &package3, &package4, &package5}
		})

		It("returns an ordered set of package compilation", func() {
			sortedPackages := Sort(packages)

			expectSorted(sortedPackages)
		})
	})

	Context("graph from a BOSH release", func() {
		BeforeEach(func() {
			nginx := Package{Name: "nginx"}
			genisoimage := Package{Name: "genisoimage"}
			powerdns := Package{Name: "powerdns"}
			ruby := Package{Name: "ruby"}

			blobstore := Package{
				Name:         "blobstore",
				Dependencies: []*Package{&ruby},
			}

			mysql := Package{Name: "mysql"}

			nats := Package{
				Name:         "nats",
				Dependencies: []*Package{&ruby},
			}

			common := Package{Name: "common"}
			redis := Package{Name: "redis"}
			libpq := Package{Name: "libpq"}
			postgres := Package{Name: "postgres"}

			registry := Package{
				Name:         "registry",
				Dependencies: []*Package{&libpq, &mysql, &ruby},
			}

			director := Package{
				Name:         "director",
				Dependencies: []*Package{&libpq, &mysql, &ruby},
			}

			healthMonitor := Package{
				Name:         "health_monitor",
				Dependencies: []*Package{&ruby},
			}

			packages = []*Package{
				&nginx,
				&genisoimage,
				&powerdns,
				&blobstore, // before ruby
				&ruby,
				&mysql,
				&nats,
				&common,
				&director, // before libpq, postgres; after ruby
				&redis,
				&registry, // before libpq, postgres; after ruby
				&libpq,
				&postgres,
				&healthMonitor, // after ruby, libpq, postgres
			}
		})

		It("orders BOSH release packages for compilation (example)", func() {
			sortedPackages := Sort(packages)

			expectSorted(sortedPackages)
		})
	})
})
