package release_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gomegafmt "github.com/onsi/gomega/format"

	. "github.com/cloudfoundry/bosh-micro-cli/release"
)

var _ = Describe("NewDependencyAnalylis", func() {
	var (
		release Release
		da      DependencyAnalysis
	)

	gomegafmt.UseStringerRepresentation = true

	Context("DeterminePackageCompilationOrder", func() {
		var package1, package2 Package
		BeforeEach(func() {
			package1 = Package{
				Name: "fake-package-name-1",
			}
			package2 = Package{
				Name: "fake-package-name-2",
			}
			release = Release{
				Packages: []*Package{&package1, &package2},
			}
			da = NewDependencyAnalylis()
		})
		Context("disjoint packages have a valid compilation sequence", func() {
			It("returns an ordered set of package compilation", func() {
				compilationOrder, err := da.DeterminePackageCompilationOrder(release)
				Expect(err).NotTo(HaveOccurred())
				Expect(compilationOrder).To(Equal([]*Package{&package1, &package2}))
			})
		})

		Context("dependent packages", func() {
			BeforeEach(func() {
				package1.Dependencies = []*Package{&package2}
			})

			It("returns an ordered set of package compilation", func() {
				compilationOrder, err := da.DeterminePackageCompilationOrder(release)
				Expect(err).NotTo(HaveOccurred())
				Expect(compilationOrder).To(Equal([]*Package{&package2, &package1}))
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
				release.Packages = append(release.Packages, &package3, &package4)
			})

			It("returns an ordered set of package compilation", func() {
				compilationOrder, err := da.DeterminePackageCompilationOrder(release)
				Expect(err).NotTo(HaveOccurred())
				Expect(compilationOrder).To(Equal([]*Package{&package2, &package3, &package1, &package4}))
			})
		})

		Context("graph from a BOSH release", func() {
			It("compiles BOSH release packages (example)", func() {
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

				release.Packages = []*Package{
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

				compilationOrder, err := da.DeterminePackageCompilationOrder(release)
				Expect(err).NotTo(HaveOccurred())

				for _, pkg := range release.Packages {
					compileOrder := indexOf(compilationOrder, pkg)
					for _, dependencyPkg := range pkg.Dependencies {
						errorMessage := fmt.Sprintf("Package '%s' should be compiled earlier than package '%s'", pkg.Name, dependencyPkg.Name)
						Expect(compileOrder).To(BeNumerically(">", indexOf(compilationOrder, dependencyPkg)), errorMessage)
					}
				}

			})
		})

		// Context("when having a cyclic dependency", func() {
		// 	It("fails with error", func() {
		// 		package1.Dependencies = []*Package{&package2}
		// 		package2.Dependencies = []*Package{&package1}
		// 		_, err := da.DeterminePackageCompilationOrder(release)
		// 		Expect(err).To(HaveOccurred())
		// 	})

		// 	It("fails with more cyclic", func() {
		// 		package1.Dependencies = []*Package{&package2}
		// 		package3 := Package{
		// 			Name:         "fake-package-name-3",
		// 			Dependencies: []*Package{&package1},
		// 		}
		// 		package2.Dependencies = []*Package{&package3}

		// 		_, err := da.DeterminePackageCompilationOrder(release)
		// 		Expect(err).To(HaveOccurred())
		// 	})
		// })
	})
})

func indexOf(packages []*Package, pkg *Package) int {
	for index, currentPkg := range packages {
		if currentPkg == pkg {
			return index
		}
	}

	return -1
}
