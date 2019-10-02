package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/cmd"
	"github.com/cloudfoundry/bosh-cli/cmd/opts"
	"github.com/cloudfoundry/bosh-cli/release"

	"github.com/cloudfoundry/bosh-cli/release/job"
	"github.com/cloudfoundry/bosh-cli/release/pkg"
	fakerel "github.com/cloudfoundry/bosh-cli/release/releasefakes"
	"github.com/cloudfoundry/bosh-cli/release/resource/resourcefakes"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
)

var _ = Describe("MergeReleasesCmd", func() {
	var (
		command cmd.MergeReleasesCmd
		options opts.MergeReleasesOpts

		release1, release2                               *fakerel.FakeRelease
		job1, job2, duplicateJob                         *job.Job
		resource1, resource2                             *resourcefakes.FakeResource
		compiledPkg1, compiledPkg2, duplicateCompiledPkg *pkg.CompiledPackage

		reader *fakerel.FakeReader
		writer *fakerel.FakeWriter
		fs     *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		release1 = &fakerel.FakeRelease{}
		release2 = &fakerel.FakeRelease{}

		resource1 = &resourcefakes.FakeResource{}
		resource1.NameReturns("job-1")
		resource1.FingerprintReturns("job-1-fingerprint")

		resource2 = &resourcefakes.FakeResource{}
		resource2.NameReturns("job-2")
		resource2.FingerprintReturns("job-2-fingerprint")

		job1 = job.NewJob(resource1)
		job2 = job.NewJob(resource2)
		duplicateJob = job.NewJob(resource2)

		release1.JobsReturns([]*job.Job{job1, duplicateJob})
		release2.JobsReturns([]*job.Job{job2})

		compiledPkg1 = pkg.NewCompiledPackageWithoutArchive(
			"compiled-package-1",
			"compiled-package-1-fingerprint",
			"compiled-package-1-os-version-slug",
			"compiled-package-1-sha1",
			[]string{},
		)

		compiledPkg2 = pkg.NewCompiledPackageWithoutArchive(
			"compiled-package-2",
			"compiled-package-2-fingerprint",
			"compiled-package-2-os-version-slug",
			"compiled-package-2-sha1",
			[]string{},
		)

		duplicateCompiledPkg = pkg.NewCompiledPackageWithoutArchive(
			"compiled-package-2",
			"compiled-package-2-fingerprint",
			"compiled-package-2-os-version-slug",
			"compiled-package-2-sha1",
			[]string{},
		)

		release1.CompiledPackagesReturns([]*pkg.CompiledPackage{compiledPkg1, duplicateCompiledPkg})
		release2.CompiledPackagesReturns([]*pkg.CompiledPackage{compiledPkg2})

		reader = &fakerel.FakeReader{}
		reader.ReadStub = func(path string) (release.Release, error) {
			if path == "release-1" {
				return release1, nil
			} else if path == "release-2" {
				return release2, nil
			} else {
				return nil, errors.New("invalid-release")
			}
		}

		writer = &fakerel.FakeWriter{}
		fs = fakesys.NewFakeFileSystem()

		options = opts.MergeReleasesOpts{
			Args: opts.MergeReleasesArgs{
				ReleasePath1: "release-1",
				ReleasePath2: "release-2",
				TargetPath:   "target-path",
			},
		}

		writer.WriteReturns("temporary-path", nil)

		fs.WriteFile("temporary-path", []byte("merged-release-content"))

		command = cmd.NewMergeReleasesCmd(reader, writer, fs)
	})

	It("writes the merged release to the target path specified", func() {
		err := command.Run(options)
		Expect(err).NotTo(HaveOccurred())

		data, err := fs.ReadFile(options.Args.TargetPath)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(data)).To(Equal("merged-release-content"))
	})

	It("cleans up the releases after merging", func() {
		err := command.Run(options)
		Expect(err).NotTo(HaveOccurred())

		Expect(release1.CleanUpCallCount()).To(Equal(1))
		Expect(release2.CleanUpCallCount()).To(Equal(1))
	})

	It("merges and dedupes jobs and compiled packages from both releases", func() {
		err := command.Run(options)
		Expect(err).NotTo(HaveOccurred())

		Expect(writer.WriteCallCount()).To(Equal(1))
		mergedRelease, _ := writer.WriteArgsForCall(0)
		Expect(mergedRelease.Jobs()).To(ConsistOf(job1, job2))
		Expect(mergedRelease.CompiledPackages()).To(ConsistOf(compiledPkg1, compiledPkg2))
	})

	Context("when the releases do not have the same name", func() {
		BeforeEach(func() {
			release1.NameReturns("release-1")
			release2.NameReturns("release-2")
		})

		It("returns an error", func() {
			err := command.Run(options)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the releases do not have the same version", func() {
		BeforeEach(func() {
			release1.VersionReturns("version-1")
			release2.VersionReturns("version-2")
		})

		It("returns an error", func() {
			err := command.Run(options)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when there are conflicting jobs", func() {
		var (
			conflictingResource *resourcefakes.FakeResource
			conflictingJob      *job.Job
		)

		BeforeEach(func() {
			conflictingResource = &resourcefakes.FakeResource{}
			conflictingResource.NameReturns("job-2")
			conflictingResource.FingerprintReturns("different-fingerprint")
			conflictingJob = job.NewJob(conflictingResource)
			release1.JobsReturns([]*job.Job{job1, conflictingJob})
		})

		It("returns an error", func() {
			err := command.Run(options)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when there are conflicting compiled packages", func() {
		var (
			conflictingCompiledPkg *pkg.CompiledPackage
		)

		BeforeEach(func() {
			conflictingCompiledPkg = pkg.NewCompiledPackageWithoutArchive(
				"compiled-package-2",
				"different-fingerprint",
				"compiled-package-2-os-version-slug",
				"compiled-package-2-sha1",
				[]string{},
			)

			release1.CompiledPackagesReturns([]*pkg.CompiledPackage{compiledPkg1, conflictingCompiledPkg})
		})

		It("returns an error", func() {
			err := command.Run(options)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when one of the releases is not compiled", func() {
		var package1 *pkg.Package

		BeforeEach(func() {
			package1 = &pkg.Package{}
			release1.PackagesReturns([]*pkg.Package{package1})
		})

		It("returns an error", func() {
			err := command.Run(options)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when writing the merged release fails", func() {
		BeforeEach(func() {
			writer.WriteReturns("", errors.New("writing failed"))
		})

		It("returns an error", func() {
			err := command.Run(options)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when copying the merged release to the target path fails", func() {
		BeforeEach(func() {
			fs.RenameError = errors.New("copying file failed")
		})

		It("returns an error", func() {
			err := command.Run(options)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when reading the release fails", func() {
		BeforeEach(func() {
			reader.ReadReturns(nil, errors.New("bad release"))
		})

		It("returns an error", func() {
			err := command.Run(options)
			Expect(err).To(HaveOccurred())
		})
	})
})
