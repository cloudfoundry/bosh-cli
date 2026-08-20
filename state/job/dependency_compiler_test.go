package job_test

import (
	"fmt"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	boshreljob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	boshrelpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
	. "github.com/cloudfoundry/bosh-cli/v7/release/resource"
	. "github.com/cloudfoundry/bosh-cli/v7/state/job"
	bistatepkg "github.com/cloudfoundry/bosh-cli/v7/state/pkg"
	"github.com/cloudfoundry/bosh-cli/v7/state/pkg/pkgfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("DependencyCompiler", func() {
	var (
		mockPackageCompiler *pkgfakes.FakeCompiler
		logger              boshlog.Logger

		dependencyCompiler DependencyCompiler
		stage              *fakeui.FakeStage

		pkg1 *boshrelpkg.Package
		pkg2 *boshrelpkg.Package

		job  *boshreljob.Job
		jobs []boshreljob.Job

		order []string

		pkg1AlreadyCompiled bool
		pkg2AlreadyCompiled bool
	)

	BeforeEach(func() {
		mockPackageCompiler = &pkgfakes.FakeCompiler{}
		order = nil

		logger = boshlog.NewLogger(boshlog.LevelNone)
		dependencyCompiler = NewDependencyCompiler(mockPackageCompiler, logger)

		stage = fakeui.NewFakeStage()

		pkg1 = newPkg("pkg1-name", "pkg1-fp", nil)
		pkg2 = newPkg("pkg2-name", "pkg2-fp", []string{"pkg1-name"})
		err := pkg2.AttachDependencies([]*boshrelpkg.Package{pkg1})
		Expect(err).ToNot(HaveOccurred())
		job = boshreljob.NewJob(NewResourceWithBuiltArchive("cpi", "job-fp", "path", "sha1"))
		job.PackageNames = []string{"pkg2-name"}
		err = job.AttachPackages([]*boshrelpkg.Package{pkg2})
		Expect(err).ToNot(HaveOccurred())
		jobs = []boshreljob.Job{*job}

		pkg1AlreadyCompiled = false
		pkg2AlreadyCompiled = false
	})

	JustBeforeEach(func() {
		compiledPackageRecord1 := bistatepkg.CompiledPackageRecord{
			BlobID:   "fake-compiled-package-blobstore-id-1",
			BlobSHA1: "fake-compiled-package-sha1-1",
		}
		compiledPackageRecord2 := bistatepkg.CompiledPackageRecord{
			BlobID:   "fake-compiled-package-blobstore-id-2",
			BlobSHA1: "fake-compiled-package-sha1-2",
		}

		mockPackageCompiler.CompileStub = func(pkg boshrelpkg.Compilable) (bistatepkg.CompiledPackageRecord, bool, error) {
			switch pkg {
			case pkg1:
				order = append(order, "pkg1-name")
				return compiledPackageRecord1, pkg1AlreadyCompiled, nil
			case pkg2:
				order = append(order, "pkg2-name")
				return compiledPackageRecord2, pkg2AlreadyCompiled, nil
			}
			return bistatepkg.CompiledPackageRecord{}, false, fmt.Errorf("unexpected package passed to Compile: %#v", pkg)
		}
	})

	It("compiles all the job dependencies (packages) such that no package is compiled before its dependencies", func() {
		_, err := dependencyCompiler.Compile(jobs, stage)
		Expect(err).ToNot(HaveOccurred())
		Expect(order).To(Equal([]string{"pkg1-name", "pkg2-name"}))
	})

	It("returns references to the compiled packages", func() {
		compiledPackageRefs, err := dependencyCompiler.Compile(jobs, stage)
		Expect(err).ToNot(HaveOccurred())

		Expect(compiledPackageRefs).To(Equal([]CompiledPackageRef{
			{
				Name:        "pkg1-name",
				Version:     "pkg1-fp",
				BlobstoreID: "fake-compiled-package-blobstore-id-1",
				SHA1:        "fake-compiled-package-sha1-1",
			},
			{
				Name:        "pkg2-name",
				Version:     "pkg2-fp",
				BlobstoreID: "fake-compiled-package-blobstore-id-2",
				SHA1:        "fake-compiled-package-sha1-2",
			},
		}))
	})

	It("logs compile stages", func() {
		_, err := dependencyCompiler.Compile(jobs, stage)
		Expect(err).ToNot(HaveOccurred())

		Expect(stage.PerformCalls).To(Equal([]*fakeui.PerformCall{
			{Name: "Compiling package 'pkg1-name/pkg1-fp'"},
			{Name: "Compiling package 'pkg2-name/pkg2-fp'"},
		}))
	})

	Context("when packages are in circular dependency", func() {
		var (
			pkg1, pkg2, pkg3 *boshrelpkg.Package
		)

		BeforeEach(func() {
			pkg1 = newPkg("pkg1-name", "pkg1-fp", []string{"pkg3-name"})
			pkg2 = newPkg("pkg2-name", "pkg2-fp", []string{"pkg1-name"})
			pkg3 = newPkg("pkg3-name", "pkg3-fp", []string{"pkg2-name"})
			err := pkg1.AttachDependencies([]*boshrelpkg.Package{pkg3})
			Expect(err).ToNot(HaveOccurred())
			err = pkg2.AttachDependencies([]*boshrelpkg.Package{pkg1})
			Expect(err).ToNot(HaveOccurred())
			err = pkg3.AttachDependencies([]*boshrelpkg.Package{pkg2})
			Expect(err).ToNot(HaveOccurred())

			job = boshreljob.NewJob(NewResourceWithBuiltArchive("cpi", "job-fp", "path", "sha1"))
			job.PackageNames = []string{"pkg2-name"}
			err = job.AttachPackages([]*boshrelpkg.Package{pkg1, pkg2, pkg3})
			Expect(err).ToNot(HaveOccurred())
			jobs = []boshreljob.Job{*job}
		})

		It("returns an error", func() {
			_, err := dependencyCompiler.Compile(jobs, stage)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when a compiled releases is provided", func() {
		BeforeEach(func() {
			pkg1AlreadyCompiled = true
			pkg2AlreadyCompiled = true
		})

		It("skips compiling the packages in the release", func() {
			_, err := dependencyCompiler.Compile(jobs, stage)
			Expect(err).ToNot(HaveOccurred())

			for _, call := range stage.PerformCalls {
				Expect(call.SkipError).To(HaveOccurred())
				Expect(call.SkipError.Error()).To(MatchRegexp("Package already compiled: Package 'pkg\\d-name' is already compiled. Skipped compilation"))
			}
		})
	})

	Context("when multiple jobs depend on the same package", func() {
		JustBeforeEach(func() {
			job2 := boshreljob.NewJob(NewResourceWithBuiltArchive("job2-name", "job2-fp", "", ""))
			job2.PackageNames = []string{"pkg2-name"}
			err := job2.AttachPackages([]*boshrelpkg.Package{pkg2})
			Expect(err).ToNot(HaveOccurred())
			jobs = append(jobs, *job2)
		})

		It("only compiles each package once", func() {
			_, err := dependencyCompiler.Compile(jobs, stage)
			Expect(err).ToNot(HaveOccurred())
			Expect(order).To(Equal([]string{"pkg1-name", "pkg2-name"}))
		})
	})

	Context("when multiple packages depend on the same package", func() {
		var (
			pkg3 *boshrelpkg.Package
		)

		BeforeEach(func() {
			pkg3 = newPkg("pkg3-name", "pkg3-fp", []string{"pkg1-name"})
			err := pkg3.AttachDependencies([]*boshrelpkg.Package{pkg1})
			Expect(err).ToNot(HaveOccurred())

			job.PackageNames = append(job.PackageNames, pkg3.Name())
			err = job.AttachPackages([]*boshrelpkg.Package{pkg1, pkg2, pkg3})
			Expect(err).ToNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			compiledPackageRecord3 := bistatepkg.CompiledPackageRecord{
				BlobID:   "fake-compiled-package-blobstore-id-3",
				BlobSHA1: "fake-compiled-package-sha1-3",
			}

			mockPackageCompiler.CompileStub = func(pkg boshrelpkg.Compilable) (bistatepkg.CompiledPackageRecord, bool, error) {
				switch pkg {
				case pkg1:
					order = append(order, "pkg1-name")
					return bistatepkg.CompiledPackageRecord{
						BlobID:   "fake-compiled-package-blobstore-id-1",
						BlobSHA1: "fake-compiled-package-sha1-1",
					}, false, nil
				case pkg2:
					order = append(order, "pkg2-name")
					return bistatepkg.CompiledPackageRecord{
						BlobID:   "fake-compiled-package-blobstore-id-2",
						BlobSHA1: "fake-compiled-package-sha1-2",
					}, false, nil
				case pkg3:
					order = append(order, "pkg3-name")
					return compiledPackageRecord3, false, nil
				}
				return bistatepkg.CompiledPackageRecord{}, false, fmt.Errorf("unexpected package passed to Compile: %#v", pkg)
			}
		})

		It("only compiles each package once", func() {
			_, err := dependencyCompiler.Compile(jobs, stage)
			Expect(err).ToNot(HaveOccurred())

			// pkg1 must be compiled exactly once, and before anything else (pkg2/pkg3 depend on it)
			Expect(order).ToNot(BeEmpty())
			Expect(order[0]).To(Equal("pkg1-name"))
			pkg1Count := 0
			for _, name := range order {
				if name == "pkg1-name" {
					pkg1Count++
				}
			}
			Expect(pkg1Count).To(Equal(1))
		})
	})
})
