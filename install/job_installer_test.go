package install_test

import (
	"errors"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmlog "github.com/cloudfoundry/bosh-micro-cli/logging"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtempcomp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	faketime "github.com/cloudfoundry/bosh-agent/time/fakes"
	fakebminstall "github.com/cloudfoundry/bosh-micro-cli/install/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/logging/fakes"
	fakebmtemcomp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/install"
)

var _ = Describe("JobInstaller", func() {
	var (
		fs               *fakesys.FakeFileSystem
		jobInstaller     JobInstaller
		job              bmrel.Job
		packageInstaller *fakebminstall.FakePackageInstaller
		blobExtractor    *fakebminstall.FakeBlobExtractor
		templateRepo     *fakebmtemcomp.FakeTemplatesRepo
		jobsPath         string
		packagesPath     string
		eventLogger      *fakebmlog.FakeEventLogger
		timeService      *faketime.FakeService
	)

	Context("Installing the job", func() {
		BeforeEach(func() {
			fs = fakesys.NewFakeFileSystem()
			packageInstaller = fakebminstall.NewFakePackageInstaller()
			blobExtractor = fakebminstall.NewFakeBlobExtractor()
			templateRepo = fakebmtemcomp.NewFakeTemplatesRepo()

			jobsPath = "/fake/jobs"
			packagesPath = "/fake/packages"
			eventLogger = fakebmlog.NewFakeEventLogger()
			timeService = &faketime.FakeService{}

			jobInstaller = NewJobInstaller(fs, packageInstaller, blobExtractor, templateRepo, jobsPath, packagesPath, eventLogger, timeService)
			job = bmrel.Job{
				Name: "cpi",
			}

			templateRepo.SetFindBehavior(job, bmtempcomp.TemplateRecord{BlobID: "fake-blob-id", BlobSha1: "fake-sha1"}, true, nil)
			blobExtractor.SetExtractBehavior("fake-blob-id", "fake-sha1", "/fake/jobs/cpi", nil)
		})

		It("creates basic job layout", func() {
			err := jobInstaller.Install(job)
			Expect(err).ToNot(HaveOccurred())
			Expect(fs.FileExists(filepath.Join(jobsPath, job.Name))).To(BeTrue())
			Expect(fs.FileExists(packagesPath)).To(BeTrue())
		})

		It("finds the rendered templates for the job from the repo", func() {
			err := jobInstaller.Install(job)
			Expect(err).ToNot(HaveOccurred())
			Expect(templateRepo.FindInputs).To(ContainElement(fakebmtemcomp.FindInput{Job: job}))
		})

		It("tells the blobExtractor to extract the templates into the installed job dir", func() {
			err := jobInstaller.Install(job)
			Expect(err).ToNot(HaveOccurred())
			Expect(blobExtractor.ExtractInputs).To(ContainElement(fakebminstall.ExtractInput{
				BlobID:    "fake-blob-id",
				BlobSha1:  "fake-sha1",
				TargetDir: filepath.Join(jobsPath, job.Name),
			}))
		})

		It("logs events to the event logger", func() {
			installStart := time.Now()
			installFinish := installStart.Add(1 * time.Second)
			timeService.NowTimes = []time.Time{installStart, installFinish}

			err := jobInstaller.Install(job)
			Expect(err).ToNot(HaveOccurred())

			expectedStartEvent := bmlog.Event{
				Time:  installStart,
				Stage: "installing CPI jobs",
				Total: 1,
				Task:  "cpi",
				Index: 1,
				State: bmlog.Started,
			}

			expectedFinishEvent := bmlog.Event{
				Time:  installFinish,
				Stage: "installing CPI jobs",
				Total: 1,
				Task:  "cpi",
				Index: 1,
				State: bmlog.Finished,
			}

			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFinishEvent))
		})

		It("logs failure event", func() {
			fs.MkdirAllError = errors.New("fake-mkdir-error")

			installStart := time.Now()
			installFail := installStart.Add(1 * time.Second)
			timeService.NowTimes = []time.Time{installStart, installFail}

			err := jobInstaller.Install(job)
			Expect(err).To(HaveOccurred())

			expectedStartEvent := bmlog.Event{
				Time:  installStart,
				Stage: "installing CPI jobs",
				Total: 1,
				Task:  "cpi",
				Index: 1,
				State: bmlog.Started,
			}

			expectedFailEvent := bmlog.Event{
				Time:    installFail,
				Stage:   "installing CPI jobs",
				Total:   1,
				Task:    "cpi",
				Index:   1,
				State:   bmlog.Failed,
				Message: "Creating jobs directory `/fake/jobs/cpi': fake-mkdir-error",
			}

			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFailEvent))
		})

		Context("when the job has packages", func() {
			var pkg1 bmrel.Package

			BeforeEach(func() {
				pkg1 = bmrel.Package{Name: "fake-pkg-name"}
				job.Packages = []*bmrel.Package{&pkg1}
				packageInstaller.SetInstallBehavior(&pkg1, packagesPath, nil)
				templateRepo.SetFindBehavior(job, bmtempcomp.TemplateRecord{BlobID: "fake-blob-id", BlobSha1: "fake-sha1"}, true, nil)
			})

			It("install packages correctly", func() {
				err := jobInstaller.Install(job)
				Expect(err).ToNot(HaveOccurred())
				Expect(packageInstaller.InstallInputs).To(ContainElement(
					fakebminstall.InstallInput{Package: &pkg1, Target: packagesPath},
				))
			})

			It("return err when package installation fails", func() {
				packageInstaller.SetInstallBehavior(
					&pkg1,
					packagesPath,
					errors.New("Installation failed, yo"),
				)
				err := jobInstaller.Install(job)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Installation failed"))
			})
		})
	})
})
