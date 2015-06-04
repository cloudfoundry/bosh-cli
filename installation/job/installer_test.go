package job_test

import (
	"errors"
	"os"
	"path/filepath"

	fakebiinstallblob "github.com/cloudfoundry/bosh-init/installation/blob/fakes"
	. "github.com/cloudfoundry/bosh-init/installation/job"
	fakesys "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/system/fakes"
	fakebitemplate "github.com/cloudfoundry/bosh-init/templatescompiler/fakes"
	fakebiui "github.com/cloudfoundry/bosh-init/ui/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Installer", func() {
	var (
		fs             *fakesys.FakeFileSystem
		jobInstaller   Installer
		renderedJobRef RenderedJobRef
		blobExtractor  *fakebiinstallblob.FakeExtractor
		templateRepo   *fakebitemplate.FakeTemplatesRepo
		jobsPath       string
		fakeStage      *fakebiui.FakeStage
	)

	Context("Installing the job", func() {
		BeforeEach(func() {
			fs = fakesys.NewFakeFileSystem()
			blobExtractor = fakebiinstallblob.NewFakeExtractor()
			templateRepo = fakebitemplate.NewFakeTemplatesRepo()

			jobsPath = "/fake/jobs"
			fakeStage = fakebiui.NewFakeStage()

			jobInstaller = NewInstaller(fs, blobExtractor, templateRepo, jobsPath)

			renderedJobRef = RenderedJobRef{
				Name:        "cpi",
				Version:     "fake-job-version-cpi",
				BlobstoreID: "fake-job-blobstore-id-cpi",
				SHA1:        "fake-job-sha1-cpi",
			}
		})

		JustBeforeEach(func() {
			blobExtractor.SetExtractBehavior("fake-job-blobstore-id-cpi", "fake-job-sha1-cpi", "/fake/jobs/cpi", nil)
		})

		It("makes the files in the job's bin directory executable", func() {
			cpiExecutablePath := "/fake/jobs/cpi/bin/cpi"
			fs.SetGlob("/fake/jobs/cpi/bin/*", []string{cpiExecutablePath})
			fs.WriteFileString(cpiExecutablePath, "contents")

			_, err := jobInstaller.Install(renderedJobRef, fakeStage)
			Expect(err).ToNot(HaveOccurred())
			Expect(fs.GetFileTestStat(cpiExecutablePath).FileMode).To(Equal(os.FileMode(0755)))
		})

		It("returns a record of the installed job", func() {
			installedJob, err := jobInstaller.Install(renderedJobRef, fakeStage)
			Expect(err).ToNot(HaveOccurred())
			Expect(installedJob).To(Equal(
				InstalledJob{
					Name: "cpi",
					Path: "/fake/jobs/cpi",
				},
			))
		})

		It("creates basic job layout", func() {
			_, err := jobInstaller.Install(renderedJobRef, fakeStage)
			Expect(err).ToNot(HaveOccurred())
			Expect(fs.FileExists(filepath.Join(jobsPath, renderedJobRef.Name))).To(BeTrue())
		})

		It("tells the blobExtractor to extract the templates into the installed job dir", func() {
			_, err := jobInstaller.Install(renderedJobRef, fakeStage)
			Expect(err).ToNot(HaveOccurred())
			Expect(blobExtractor.ExtractInputs).To(ContainElement(fakebiinstallblob.ExtractInput{
				BlobID:    "fake-job-blobstore-id-cpi",
				BlobSHA1:  "fake-job-sha1-cpi",
				TargetDir: filepath.Join(jobsPath, renderedJobRef.Name),
			}))
		})

		It("logs events to the event logger", func() {
			_, err := jobInstaller.Install(renderedJobRef, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeStage.PerformCalls).To(Equal([]*fakebiui.PerformCall{
				{Name: "Installing job 'cpi'"},
			}))
		})

		It("logs failure event", func() {
			fs.MkdirAllError = errors.New("fake-mkdir-error")

			_, err := jobInstaller.Install(renderedJobRef, fakeStage)
			Expect(err).To(HaveOccurred())

			Expect(fakeStage.PerformCalls[0].Name).To(Equal("Installing job 'cpi'"))
			Expect(fakeStage.PerformCalls[0].Error).To(HaveOccurred())
			Expect(fakeStage.PerformCalls[0].Error.Error()).To(Equal("Creating job directory '/fake/jobs/cpi': fake-mkdir-error"))
		})
	})

	Context("Cleanup", func() {
		It("cleans up files left under the jobPath when done", func() {
			fs.MkdirAll("/some/job/dir", os.ModePerm)
			fs.WriteFileString("/some/job/dir/file", "contents")

			job := InstalledJob{Name: "job-name", Path: "/some/job/dir"}

			err := jobInstaller.Cleanup(job)
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.FileExists("/some/job/dir")).To(BeFalse())
		})

		It("returns the error if deleting the job dir fails", func() {
			job := InstalledJob{Name: "job-name", Path: "/some/job/dir"}

			fs.RemoveAllError = errors.New("couldn't delete that")
			err := jobInstaller.Cleanup(job)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("couldn't delete that"))
		})
	})
})
