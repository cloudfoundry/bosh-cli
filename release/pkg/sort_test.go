package pkg_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gomegafmt "github.com/onsi/gomega/format"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	. "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
)

var _ = Describe("Sort", func() {
	var (
		packages []*bmrel.Package
	)

	gomegafmt.UseStringerRepresentation = true

	var indexOf = func(packages []*bmrel.Package, pkg *bmrel.Package) int {
		for index, currentPkg := range packages {
			if currentPkg == pkg {
				return index
			}
		}
		return -1
	}

	var expectSorted = func(sortedPackages []*bmrel.Package) {
		for _, pkg := range packages {
			sortedIndex := indexOf(sortedPackages, pkg)
			for _, dependencyPkg := range pkg.Dependencies {
				errorMessage := fmt.Sprintf("Package '%s' should be compiled after package '%s'", pkg.Name, dependencyPkg.Name)
				Expect(sortedIndex).To(BeNumerically(">", indexOf(sortedPackages, dependencyPkg)), errorMessage)
			}
		}
	}

	var package1, package2 bmrel.Package

	BeforeEach(func() {
		package1 = bmrel.Package{
			Name: "fake-package-name-1",
		}
		package2 = bmrel.Package{
			Name: "fake-package-name-2",
		}
		packages = []*bmrel.Package{&package1, &package2}
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
			package1.Dependencies = []*bmrel.Package{&package2}
		})

		It("returns an ordered set of package compilation", func() {
			sortedPackages := Sort(packages)

			Expect(sortedPackages).To(Equal([]*bmrel.Package{&package2, &package1}))
		})
	})

	Context("complex graph of dependent packages", func() {
		var package3, package4 bmrel.Package

		BeforeEach(func() {
			package1.Dependencies = []*bmrel.Package{&package2, &package3}
			package3 = bmrel.Package{
				Name: "fake-package-name-3",
			}
			package4 = bmrel.Package{
				Name:         "fake-package-name-4",
				Dependencies: []*bmrel.Package{&package3, &package2},
			}
			packages = []*bmrel.Package{&package1, &package2, &package3, &package4}
		})

		It("returns an ordered set of package compilation", func() {
			sortedPackages := Sort(packages)

			expectSorted(sortedPackages)
		})
	})

	Context("graph with transitively dependent packages", func() {
		var package3, package4, package5 bmrel.Package

		BeforeEach(func() {
			package3 = bmrel.Package{
				Name: "fake-package-name-3",
			}
			package4 = bmrel.Package{
				Name: "fake-package-name-4",
			}
			package5 = bmrel.Package{
				Name: "fake-package-name-5",
			}

			package3.Dependencies = []*bmrel.Package{&package2}
			package2.Dependencies = []*bmrel.Package{&package1}

			package5.Dependencies = []*bmrel.Package{&package2}

			packages = []*bmrel.Package{&package1, &package2, &package3, &package4, &package5}
		})

		It("returns an ordered set of package compilation", func() {
			sortedPackages := Sort(packages)

			expectSorted(sortedPackages)
		})
	})

	Context("graph from a BOSH release", func() {
		BeforeEach(func() {
			nginx := bmrel.Package{Name: "nginx"}
			genisoimage := bmrel.Package{Name: "genisoimage"}
			powerdns := bmrel.Package{Name: "powerdns"}
			ruby := bmrel.Package{Name: "ruby"}

			blobstore := bmrel.Package{
				Name:         "blobstore",
				Dependencies: []*bmrel.Package{&ruby},
			}

			mysql := bmrel.Package{Name: "mysql"}

			nats := bmrel.Package{
				Name:         "nats",
				Dependencies: []*bmrel.Package{&ruby},
			}

			common := bmrel.Package{Name: "common"}
			redis := bmrel.Package{Name: "redis"}
			libpq := bmrel.Package{Name: "libpq"}
			postgres := bmrel.Package{Name: "postgres"}

			registry := bmrel.Package{
				Name:         "registry",
				Dependencies: []*bmrel.Package{&libpq, &mysql, &ruby},
			}

			director := bmrel.Package{
				Name:         "director",
				Dependencies: []*bmrel.Package{&libpq, &mysql, &ruby},
			}

			healthMonitor := bmrel.Package{
				Name:         "health_monitor",
				Dependencies: []*bmrel.Package{&ruby},
			}

			packages = []*bmrel.Package{
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

	//	Context("when having a cyclic dependency", func() {
	//		It("fails with error", func() {
	//			package1.Dependencies = []*bmrel.Package{&package2}
	//			package2.Dependencies = []*bmrel.Package{&package1}
	//
	//			sort.Sort(packages)
	//		})
	//
	//		It("fails with more cyclic", func() {
	//			package1.Dependencies = []*bmrel.Package{&package2}
	//			package3 := bmrel.Package{
	//				Name:         "fake-package-name-3",
	//				Dependencies: []*bmrel.Package{&package1},
	//			}
	//			package2.Dependencies = []*bmrel.Package{&package3}
	//
	//			sort.Sort(packages)
	//		})
	//	})
})
