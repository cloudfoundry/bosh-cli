package templatescompiler_test

import (
	"errors"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakeblobs "github.com/cloudfoundry/bosh-agent/blobstore/fakes"
	fakecmd "github.com/cloudfoundry/bosh-agent/platform/commands/fakes"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmtemp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/fakes"

	fakebmrender "github.com/cloudfoundry/bosh-micro-cli/erbrenderer/fakes"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/jobs"
	. "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

var _ = Describe("TemplatesCompiler", func() {
	var (
		templatesCompiler TemplatesCompiler
		renderer          *fakebmrender.FakeERBRenderer
		compressor        *fakecmd.FakeCompressor
		blobstore         *fakeblobs.FakeBlobstore
		templatesRepo     *fakebmtemp.FakeTemplatesRepo
		fs                *fakesys.FakeFileSystem
		compileDir        string
		jobs              []bmreljob.Job
	)

	BeforeEach(func() {
		renderer = fakebmrender.NewFakeERBRender()
		compressor = fakecmd.NewFakeCompressor()
		compressor.CompressFilesInDirTarballPath = "fake-tarball-path"

		blobstore = fakeblobs.NewFakeBlobstore()
		fs = fakesys.NewFakeFileSystem()

		templatesRepo = fakebmtemp.NewFakeTemplatesRepo()

		templatesCompiler = NewTemplatesCompiler(renderer, compressor, blobstore, templatesRepo, fs)

		var err error
		compileDir, err = fs.TempDir("bosh-micro-cli-tests")
		Expect(err).ToNot(HaveOccurred())
		fs.TempDirDir = compileDir
		renderer.SetRenderBehavior(
			"fake-extracted-path/cpi.erb",
			filepath.Join(compileDir, "bin/cpi"),
			nil,
		)
	})

	Context("with a job", func() {
		BeforeEach(func() {
			jobs = []bmreljob.Job{
				bmreljob.Job{
					Name:          "fake-job-1",
					ExtractedPath: "fake-extracted-path",
					Templates: map[string]string{
						"cpi.erb": "/bin/cpi",
					},
				},
			}

			blobstore.CreateBlobID = "fake-blob-id"
			blobstore.CreateFingerprint = "fake-sha1"
			record := TemplateRecord{
				BlobID:   "fake-blob-id",
				BlobSha1: "fake-sha1",
			}
			templatesRepo.SetSaveBehavior(jobs[0], record, nil)
		})

		It("renders job templates", func() {
			err := templatesCompiler.Compile(jobs)
			Expect(err).ToNot(HaveOccurred())
			Expect(renderer.RenderInputs).To(ContainElement(
				fakebmrender.RenderInput{
					SrcPath: "fake-extracted-path/cpi.erb",
					DstPath: filepath.Join(compileDir, "bin/cpi"),
				}),
			)
		})

		It("cleans the temp folder to hold the compile result for job", func() {
			err := templatesCompiler.Compile(jobs)
			Expect(err).ToNot(HaveOccurred())
			Expect(fs.FileExists(compileDir)).To(BeFalse())
		})

		It("generates templates archive", func() {
			err := templatesCompiler.Compile(jobs)
			Expect(err).ToNot(HaveOccurred())
			Expect(compressor.CompressFilesInDirDir).To(Equal(compileDir))
			Expect(compressor.CleanUpTarballPath).To(Equal("fake-tarball-path"))
		})

		It("saves archive in blobstore", func() {
			err := templatesCompiler.Compile(jobs)
			Expect(err).ToNot(HaveOccurred())
			Expect(blobstore.CreateFileName).To(Equal("fake-tarball-path"))
		})

		It("stores the compiled package blobID and fingerprint into the compile package repo", func() {
			err := templatesCompiler.Compile(jobs)
			Expect(err).ToNot(HaveOccurred())

			record := TemplateRecord{
				BlobID:   "fake-blob-id",
				BlobSha1: "fake-sha1",
			}

			Expect(templatesRepo.SaveInputs).To(ContainElement(
				fakebmtemp.SaveInput{Job: jobs[0], Record: record},
			))
		})

		Context("when creating compilation directory fails", func() {
			BeforeEach(func() {
				fs.TempDirError = errors.New("fake-tempdir-error")
			})

			It("returns an error", func() {
				err := templatesCompiler.Compile(jobs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-tempdir-error"))
			})
		})

		Context("when rendering fails", func() {
			BeforeEach(func() {
				renderer.SetRenderBehavior(
					"fake-extracted-path/cpi.erb",
					filepath.Join(compileDir, "bin/cpi"),
					errors.New("fake-render-error"),
				)
			})

			It("returns an error", func() {
				err := templatesCompiler.Compile(jobs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-render-error"))
			})
		})

		Context("when generating templates archive fails", func() {
			BeforeEach(func() {
				compressor.CompressFilesInDirErr = errors.New("fake-compress-error")
			})

			It("returns an error", func() {
				err := templatesCompiler.Compile(jobs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-compress-error"))
			})
		})

		Context("when saving to blobstore fails", func() {
			BeforeEach(func() {
				blobstore.CreateErr = errors.New("fake-blobstore-error")
			})

			It("returns an error", func() {
				err := templatesCompiler.Compile(jobs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-blobstore-error"))
			})
		})

		Context("when saving to templates repo fails", func() {
			BeforeEach(func() {
				record := TemplateRecord{
					BlobID:   "fake-blob-id",
					BlobSha1: "fake-sha1",
				}

				err := errors.New("fake-template-error")
				templatesRepo.SetSaveBehavior(jobs[0], record, err)
			})

			It("returns an error", func() {
				err := templatesCompiler.Compile(jobs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-template-error"))
			})
		})

		Context("when one of the job fails to compile", func() {
			BeforeEach(func() {
				jobs = []bmreljob.Job{
					bmreljob.Job{
						Name:          "fake-job-1",
						ExtractedPath: "fake-extracted-path-1",
						Templates: map[string]string{
							"cpi.erb": "/bin/cpi",
						},
					},
					bmreljob.Job{
						Name:          "fake-job-2",
						ExtractedPath: "fake-extracted-path-2",
						Templates: map[string]string{
							"cpi.erb": "/bin/cpi",
						},
					},
					bmreljob.Job{
						Name:          "fake-job-3",
						ExtractedPath: "fake-extracted-path-3",
						Templates: map[string]string{
							"cpi.erb": "/bin/cpi",
						},
					},
				}

				renderer.SetRenderBehavior(
					"fake-extracted-path-1/cpi.erb",
					filepath.Join(compileDir, "bin/cpi"),
					nil,
				)

				renderer.SetRenderBehavior(
					"fake-extracted-path-2/cpi.erb",
					filepath.Join(compileDir, "bin/cpi"),
					errors.New("fake-render-2-error"),
				)

				record := TemplateRecord{
					BlobID:   "fake-blob-id",
					BlobSha1: "fake-sha1",
				}
				templatesRepo.SetSaveBehavior(jobs[0], record, nil)
				templatesRepo.SetSaveBehavior(jobs[1], record, nil)
			})

			It("returns an error", func() {
				err := templatesCompiler.Compile(jobs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-render-2-error"))
			})
		})
	})
})
