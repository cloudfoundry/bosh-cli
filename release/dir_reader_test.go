package release_test

import (
	"errors"
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/release"
	boshjob "github.com/cloudfoundry/bosh-init/release/job"
	fakejob "github.com/cloudfoundry/bosh-init/release/job/fakes"
	boshlic "github.com/cloudfoundry/bosh-init/release/license"
	fakelic "github.com/cloudfoundry/bosh-init/release/license/fakes"
	boshpkg "github.com/cloudfoundry/bosh-init/release/pkg"
	fakepkg "github.com/cloudfoundry/bosh-init/release/pkg/fakes"
	. "github.com/cloudfoundry/bosh-init/release/resource"
)

var _ = Describe("DirReader", func() {
	var (
		jobReader *fakejob.FakeDirReader
		pkgReader *fakepkg.FakeDirReader
		licReader *fakelic.FakeDirReader
		fs        *fakesys.FakeFileSystem
		reader    DirReader
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		fs.TempDirDir = "/release"

		logger := boshlog.NewLogger(boshlog.LevelNone)

		jobReader = &fakejob.FakeDirReader{}
		pkgReader = &fakepkg.FakeDirReader{}
		licReader = &fakelic.FakeDirReader{}
		reader = NewDirReader(jobReader, pkgReader, licReader, fs, logger)
	})

	Describe("Read", func() {
		act := func() (Release, error) { return reader.Read("/release") }

		BeforeEach(func() {
			fs.SetGlob("/release/jobs/*", []string{
				"/release/jobs/job1",
				"/release/jobs/job2",
			})

			fs.SetGlob("/release/packages/*", []string{
				"/release/packages/pkg1",
				"/release/packages/pkg2",
			})
		})

		It("returns a release from the given directory", func() {
			job1 := boshjob.NewJob(NewResource("job1", "job1-fp", nil))
			job1.PackageNames = []string{"pkg1"}
			job2 := boshjob.NewJob(NewResource("job2", "job2-fp", nil))

			pkg1 := boshpkg.NewPackage(NewResource("pkg1", "pkg1-fp", nil), []string{"pkg2"})
			pkg2 := boshpkg.NewPackage(NewResource("pkg2", "pkg2-fp", nil), nil)

			lic := boshlic.NewLicense(NewResource("lic", "lic-fp", nil))

			jobReader.ReadStub = func(path string) (*boshjob.Job, error) {
				if path == "/release/jobs/job1" {
					return job1, nil
				}
				if path == "/release/jobs/job2" {
					return job2, nil
				}
				panic("Unexpected job")
			}

			pkgReader.ReadStub = func(path string) (*boshpkg.Package, error) {
				if path == "/release/packages/pkg1" {
					return pkg1, nil
				}
				if path == "/release/packages/pkg2" {
					return pkg2, nil
				}
				panic("Unexpected package")
			}

			licReader.ReadStub = func(path string) (*boshlic.License, error) {
				if path == "/release" {
					return lic, nil
				}
				panic("Unexpected license")
			}

			release, err := act()
			Expect(err).NotTo(HaveOccurred())

			Expect(release.Name()).To(BeEmpty())
			Expect(release.Version()).To(BeEmpty())
			Expect(release.CommitHashWithMark("*")).To(BeEmpty())
			Expect(release.Jobs()).To(Equal([]*boshjob.Job{job1, job2}))
			Expect(release.Packages()).To(Equal([]*boshpkg.Package{pkg1, pkg2}))
			Expect(release.CompiledPackages()).To(BeEmpty())
			Expect(release.IsCompiled()).To(BeFalse())
			Expect(release.License()).To(Equal(lic))

			// job and pkg dependencies are resolved
			Expect(job1.Packages).To(Equal([]boshpkg.Compilable{pkg1}))
			Expect(pkg1.Dependencies).To(Equal([]*boshpkg.Package{pkg2}))
		})

		It("returns empty release if there are no jobs or packages", func() {
			fs.SetGlob("/release/jobs/*", []string{})
			fs.SetGlob("/release/packages/*", []string{})

			release, err := act()
			Expect(err).NotTo(HaveOccurred())

			Expect(release.Name()).To(BeEmpty())
			Expect(release.Version()).To(BeEmpty())
			Expect(release.CommitHashWithMark("*")).To(BeEmpty())
			Expect(release.Jobs()).To(BeEmpty())
			Expect(release.Packages()).To(BeEmpty())
			Expect(release.CompiledPackages()).To(BeEmpty())
			Expect(release.IsCompiled()).To(BeFalse())
			Expect(release.License()).To(BeNil())
		})

		It("returns errors for each invalid job and package", func() {
			jobReader.ReadReturns(nil, errors.New("job-err"))
			pkgReader.ReadReturns(nil, errors.New("pkg-err"))

			_, err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Reading job from '/release/jobs/job1'"))
			Expect(err.Error()).To(ContainSubstring("Reading job from '/release/jobs/job2'"))
			Expect(err.Error()).To(ContainSubstring("Reading package from '/release/packages/pkg1'"))
			Expect(err.Error()).To(ContainSubstring("Reading package from '/release/packages/pkg2'"))
		})

		It("returns error if job's pkg dependencies cannot be satisfied", func() {
			job1 := boshjob.NewJob(NewResource("job1", "job1-fp", nil))
			job1.PackageNames = []string{"pkg-with-other-name"}
			jobReader.ReadReturns(job1, nil)

			pkg1 := boshpkg.NewPackage(NewResource("pkg1", "pkg1-fp", nil), nil)
			pkgReader.ReadReturns(pkg1, nil)

			_, err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Expected to find package 'pkg-with-other-name' since it's a dependency of job 'job1'"))
		})

		It("returns error if pkg's pkg dependencies cannot be satisfied", func() {
			job1 := boshjob.NewJob(NewResource("job1", "job1-fp", nil))
			jobReader.ReadReturns(job1, nil)

			pkg1 := boshpkg.NewPackage(NewResource("pkg1", "pkg1-fp", nil), []string{"pkg-with-other-name"})
			pkgReader.ReadReturns(pkg1, nil)

			_, err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Expected to find package 'pkg-with-other-name' since it's a dependency of package 'pkg1'"))
		})

		It("returns a release that does nothing for cleanup", func() {
			fs.SetGlob("/release/jobs/*", []string{})
			fs.SetGlob("/release/packages/*", []string{})

			fs.MkdirAll("/release", os.ModeDir)

			release, err := reader.Read("/release")
			Expect(err).NotTo(HaveOccurred())

			Expect(release.CleanUp()).ToNot(HaveOccurred())
			Expect(fs.FileExists("/release")).To(BeTrue())
		})
	})
})
