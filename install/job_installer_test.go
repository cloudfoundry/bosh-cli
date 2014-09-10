package install_test

import (
	"errors"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtempcomp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebminstall "github.com/cloudfoundry/bosh-micro-cli/install/fakes"
	fakebmtemcomp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/install"
)

var _ = Describe("JobInstaller", func() {
	var (
		fs               boshsys.FileSystem
		jobInstaller     JobInstaller
		job              bmrel.Job
		path             string
		packageInstaller *fakebminstall.FakePackageInstaller
		blobExtractor    *fakebminstall.FakeBlobExtractor
		templateRepo     *fakebmtemcomp.FakeTemplatesRepo
		packagesDir      string
	)

	Context("Installing the job", func() {
		BeforeEach(func() {
			fs = fakesys.NewFakeFileSystem()
			packageInstaller = fakebminstall.NewFakePackageInstaller()
			blobExtractor = fakebminstall.NewFakeBlobExtractor()
			templateRepo = fakebmtemcomp.NewFakeTemplatesRepo()

			jobInstaller = NewJobInstaller(fs, packageInstaller, blobExtractor, templateRepo)
			path = "fake/path"
			job = bmrel.Job{
				Name: "cpi",
			}
			packagesDir = filepath.Join(path, "packages")
			templateRepo.SetFindBehavior(job, bmtempcomp.TemplateRecord{BlobID: "fake-blob-id", BlobSha1: "fake-sha1"}, true, nil)
			blobExtractor.SetExtractBehavior("fake-blob-id", "fake-sha1", "fake/path/jobs/cpi", nil)
		})

		It("creates basic job layout", func() {
			err := jobInstaller.Install(job, path)
			Expect(err).ToNot(HaveOccurred())
			installedJobDir := filepath.Join(path, "jobs", job.Name)
			Expect(fs.FileExists(installedJobDir)).To(BeTrue())
			Expect(fs.FileExists(filepath.Join(path, "packages"))).To(BeTrue())
		})

		It("finds the rendered templates for the job from the repo", func() {
			err := jobInstaller.Install(job, path)
			Expect(err).ToNot(HaveOccurred())
			Expect(templateRepo.FindInputs).To(ContainElement(fakebmtemcomp.FindInput{Job: job}))
		})

		It("tells the blobExtractor to extract the templates into the installed job dir", func() {
			err := jobInstaller.Install(job, path)
			Expect(err).ToNot(HaveOccurred())
			Expect(blobExtractor.ExtractInputs).To(ContainElement(fakebminstall.ExtractInput{
				BlobID:    "fake-blob-id",
				BlobSha1:  "fake-sha1",
				TargetDir: filepath.Join(path, "jobs", job.Name),
			}))
		})

		Context("when the job has packages", func() {
			var pkg1 bmrel.Package

			BeforeEach(func() {
				pkg1 = bmrel.Package{Name: "fake-pkg-name"}
				job.Packages = []*bmrel.Package{&pkg1}
				packageInstaller.SetInstallBehavior(&pkg1, packagesDir, nil)
				templateRepo.SetFindBehavior(job, bmtempcomp.TemplateRecord{BlobID: "fake-blob-id", BlobSha1: "fake-sha1"}, true, nil)
			})

			It("install packages correctly", func() {

				err := jobInstaller.Install(job, path)
				Expect(err).ToNot(HaveOccurred())
				Expect(packageInstaller.InstallInputs).To(ContainElement(
					fakebminstall.InstallInput{Package: &pkg1, Target: packagesDir},
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
	})
})
