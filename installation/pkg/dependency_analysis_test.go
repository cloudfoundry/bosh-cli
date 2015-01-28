package pkg_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gomegafmt "github.com/onsi/gomega/format"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	. "github.com/cloudfoundry/bosh-micro-cli/installation/pkg"
)

var _ = Describe("NewDependencyAnalylis", func() {
	var (
		packages []*bmrel.Package
		da       DependencyAnalysis
	)

	gomegafmt.UseStringerRepresentation = true

	Context("DeterminePackageCompilationOrder", func() {
		var package1, package2 bmrel.Package
		BeforeEach(func() {
			package1 = bmrel.Package{
				Name: "fake-package-name-1",
			}
			package2 = bmrel.Package{
				Name: "fake-package-name-2",
			}
			packages = []*bmrel.Package{&package1, &package2}

			da = NewDependencyAnalysis()
		})
		Context("disjoint packages have a valid compilation sequence", func() {
			It("returns an ordered set of package compilation", func() {
				compilationOrder, err := da.DeterminePackageCompilationOrder(packages)
				Expect(err).NotTo(HaveOccurred())
				Expect(compilationOrder).To(ContainElement(&package1))
				Expect(compilationOrder).To(ContainElement(&package2))
			})
		})

		Context("dependent packages", func() {
			BeforeEach(func() {
				package1.Dependencies = []*bmrel.Package{&package2}
			})

			It("returns an ordered set of package compilation", func() {
				compilationOrder, err := da.DeterminePackageCompilationOrder(packages)
				Expect(err).NotTo(HaveOccurred())
				Expect(compilationOrder).To(ContainElement(&package1))
				Expect(compilationOrder).To(ContainElement(&package2))
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
				compilationOrder, err := da.DeterminePackageCompilationOrder(packages)
				Expect(err).NotTo(HaveOccurred())
				for _, pkg := range packages {
					compileOrder := indexOf(compilationOrder, pkg)
					for _, dependencyPkg := range pkg.Dependencies {
						errorMessage := fmt.Sprintf("Package '%s' should be compiled later than package '%s'", pkg.Name, dependencyPkg.Name)
						Expect(compileOrder).To(BeNumerically(">", indexOf(compilationOrder, dependencyPkg)), errorMessage)
					}
				}
			})
		})

		Context("graph from a BOSH release", func() {
			It("compiles BOSH release packages (example)", func() {
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

				compilationOrder, err := da.DeterminePackageCompilationOrder(packages)
				Expect(err).NotTo(HaveOccurred())

				for _, pkg := range packages {
					compileOrder := indexOf(compilationOrder, pkg)
					for _, dependencyPkg := range pkg.Dependencies {
						errorMessage := fmt.Sprintf("Package '%s' should be compiled later than package '%s'", pkg.Name, dependencyPkg.Name)
						Expect(compileOrder).To(BeNumerically(">", indexOf(compilationOrder, dependencyPkg)), errorMessage)
					}
				}
			})
		})

		// Context("when having a cyclic dependency", func() {
		// 	It("fails with error", func() {
		// 		package1.Dependencies = []*Package{&package2}
		// 		package2.Dependencies = []*Package{&package1}
		// 		_, err := da.DeterminePackageCompilationOrder(packages)
		// 		Expect(err).To(HaveOccurred())
		// 	})

		// 	It("fails with more cyclic", func() {
		// 		package1.Dependencies = []*Package{&package2}
		// 		package3 := Package{
		// 			Name:         "fake-package-name-3",
		// 			Dependencies: []*Package{&package1},
		// 		}
		// 		package2.Dependencies = []*Package{&package3}

		// 		_, err := da.DeterminePackageCompilationOrder(packages)
		// 		Expect(err).To(HaveOccurred())
		// 	})
		// })
	})
})

func indexOf(packages []*bmrel.Package, pkg *bmrel.Package) int {
	for index, currentPkg := range packages {
		if currentPkg == pkg {
			return index
		}
	}

	return -1
}
