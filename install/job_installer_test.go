package install_test

import (
	"errors"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebminstall "github.com/cloudfoundry/bosh-micro-cli/install/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/install"
)

var _ = Describe("JobInstaller", func() {
	var (
		fs               boshsys.FileSystem
		jobInstaller     JobInstaller
		job              bmrel.Job
		path             string
		packageInstaller *fakebminstall.FakePackageInstaller
	)

	Context("Installing the job", func() {
		BeforeEach(func() {
			fs = fakesys.NewFakeFileSystem()
			packageInstaller = fakebminstall.NewFakePackageInstaller()
			jobInstaller = NewJobInstaller(fs, packageInstaller)
			path = "fake/path"
			job = bmrel.Job{
				Name: "cpi",
			}
		})

		It("creates basic job layout", func() {
			err := jobInstaller.Install(job, path)
			Expect(err).ToNot(HaveOccurred())
			installedJobDir := filepath.Join(path, "jobs", job.Name)
			Expect(fs.FileExists(installedJobDir)).To(BeTrue())
			Expect(fs.FileExists(filepath.Join(path, "packages"))).To(BeTrue())
		})

		Context("when the job has packages", func() {
			var pkg1 bmrel.Package

			BeforeEach(func() {
				pkg1 = bmrel.Package{Name: "fake-pkg-name"}
				job.Packages = []*bmrel.Package{&pkg1}
			})

			It("install packages correctly", func() {
				packageInstaller.SetInstallBehavior(
					&pkg1,
					filepath.Join(path, "packages"),
					nil,
				)
				err := jobInstaller.Install(job, path)
				Expect(err).ToNot(HaveOccurred())
				Expect(packageInstaller.InstallInputs).To(ContainElement(
					fakebminstall.InstallInput{Package: &pkg1, Target: filepath.Join(path, "packages")},
				))
			})

			It("return err when package installation fails", func() {
				packageInstaller.SetInstallBehavior(
					&pkg1,
					filepath.Join(path, "packages"),
					errors.New("Installation failed, yo"),
				)
				err := jobInstaller.Install(job, path)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Installation failed"))
			})
		})

		XIt("It cleans thing before running")

		XIt("overwrite the jobs and packages in the given path")
	})
})
