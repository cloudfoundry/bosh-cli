package installation_test

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	fakeboshsys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/installation"
	bireljob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	. "github.com/cloudfoundry/bosh-cli/v7/release/resource"
	bistatejob "github.com/cloudfoundry/bosh-cli/v7/state/job"
	"github.com/cloudfoundry/bosh-cli/v7/state/job/jobfakes"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("PackageCompiler", func() {
	var (
		mockDependencyCompiler *jobfakes.FakeDependencyCompiler

		fs       *fakeboshsys.FakeFileSystem
		compiler installation.PackageCompiler

		releaseJob  bireljob.Job
		releaseJobs []bireljob.Job
		stage       *testui.Stage
	)

	BeforeEach(func() {
		mockDependencyCompiler = &jobfakes.FakeDependencyCompiler{}
		fs = fakeboshsys.NewFakeFileSystem()
		stage = &testui.Stage{}

		job := bireljob.NewJob(NewResource("cpi", "fake-release-job-fingerprint", nil))
		releaseJob = *job
	})

	JustBeforeEach(func() {
		compiler = installation.NewPackageCompiler(mockDependencyCompiler, fs)

		releaseJobs = []bireljob.Job{releaseJob}
		compiledPackageRefs := []bistatejob.CompiledPackageRef{
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
		}
		mockDependencyCompiler.CompileReturns(compiledPackageRefs, nil)
	})

	Describe("From", func() {
		It("returns compiled packages and release jobs", func() {
			packages, err := compiler.For(releaseJobs, stage)
			Expect(err).ToNot(HaveOccurred())

			Expect(packages).To(ConsistOf([]installation.CompiledPackageRef{
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

			jobs, gotStage := mockDependencyCompiler.CompileArgsForCall(0)
			Expect(jobs).To(Equal(releaseJobs))
			Expect(gotStage).To(Equal(stage))
		})

		Context("when package compilation fails", func() {
			JustBeforeEach(func() {
				mockDependencyCompiler.CompileReturns([]bistatejob.CompiledPackageRef{}, bosherr.Error("fake-compile-package-2-error"))
			})

			It("returns an error", func() {
				_, err := compiler.For(releaseJobs, stage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-compile-package-2-error"))
			})
		})
	})
})
