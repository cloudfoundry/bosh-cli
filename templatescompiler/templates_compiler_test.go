package templatescompiler_test

import (
	"errors"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakeblobs "github.com/cloudfoundry/bosh-agent/blobstore/fakes"
	fakecmd "github.com/cloudfoundry/bosh-agent/platform/commands/fakes"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment/fakes"
	fakebmrender "github.com/cloudfoundry/bosh-micro-cli/erbrenderer/fakes"
	fakebmtemp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/fakes"

	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmrender "github.com/cloudfoundry/bosh-micro-cli/erbrenderer"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
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
		jobs              []bmrel.Job
		context           bmrender.TemplateEvaluationContext
		deployment        bmdepl.Deployment
		logger            boshlog.Logger
	)

	BeforeEach(func() {
		renderer = fakebmrender.NewFakeERBRender()
		compressor = fakecmd.NewFakeCompressor()
		compressor.CompressFilesInDirTarballPath = "fake-tarball-path"

		blobstore = fakeblobs.NewFakeBlobstore()
		fs = fakesys.NewFakeFileSystem()

		templatesRepo = fakebmtemp.NewFakeTemplatesRepo()

		deployment = fakebmdepl.NewFakeDeployment()
		deployment.Properties["fake-property-key"] = "fake-property-value"

		logger = boshlog.NewLogger(boshlog.LevelNone)

		templatesCompiler = NewTemplatesCompiler(
			renderer,
			compressor,
			blobstore,
			templatesRepo,
			fs,
			logger,
		)

		var err error
		compileDir, err = fs.TempDir("bosh-micro-cli-tests")
		Expect(err).ToNot(HaveOccurred())
		fs.TempDirDir = compileDir
	})

	Context("with a job", func() {
		BeforeEach(func() {
			jobs = []bmrel.Job{
				bmrel.Job{
					Name:          "fake-job-1",
					ExtractedPath: "fake-extracted-path",
					Templates: map[string]string{
						"cpi.erb": "/bin/cpi",
					},
				},
			}

			manifestProperties := map[string]interface{}{
				"fake-property-key": "fake-property-value",
			}

			context = NewJobEvaluationContext(jobs[0], manifestProperties, "fake-deployment-name", logger)
			renderer.SetRenderBehavior(
				"fake-extracted-path/templates/cpi.erb",
				filepath.Join(compileDir, "bin/cpi"),
				context,
				nil,
			)

			blobstore.CreateBlobID = "fake-blob-id"
			blobstore.CreateFingerprint = "fake-sha1"
			record := TemplateRecord{
				BlobID:   "fake-blob-id",
				BlobSha1: "fake-sha1",
			}
			templatesRepo.SetSaveBehavior(jobs[0], record, nil)
		})

		It("renders job templates", func() {
			err := templatesCompiler.Compile(jobs, deployment)
			Expect(err).ToNot(HaveOccurred())
			Expect(renderer.RenderInputs).To(ContainElement(
				fakebmrender.RenderInput{
					SrcPath: "fake-extracted-path/templates/cpi.erb",
					DstPath: filepath.Join(compileDir, "bin/cpi"),
					Context: context,
				}),
			)
		})

		It("cleans the temp folder to hold the compile result for job", func() {
			err := templatesCompiler.Compile(jobs, deployment)
			Expect(err).ToNot(HaveOccurred())
			Expect(fs.FileExists(compileDir)).To(BeFalse())
		})

		It("generates templates archive", func() {
			err := templatesCompiler.Compile(jobs, deployment)
			Expect(err).ToNot(HaveOccurred())
			Expect(compressor.CompressFilesInDirDir).To(Equal(compileDir))
			Expect(compressor.CleanUpTarballPath).To(Equal("fake-tarball-path"))
		})

		It("saves archive in blobstore", func() {
			err := templatesCompiler.Compile(jobs, deployment)
			Expect(err).ToNot(HaveOccurred())
			Expect(blobstore.CreateFileName).To(Equal("fake-tarball-path"))
		})

		It("stores the compiled package blobID and fingerprint into the compile package repo", func() {
			err := templatesCompiler.Compile(jobs, deployment)
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
				err := templatesCompiler.Compile(jobs, deployment)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-tempdir-error"))
			})
		})

		Context("when creating parent directory for templates fails", func() {
			BeforeEach(func() {
				fs.MkdirAllError = errors.New("fake-mkdirall-error")
			})

			It("returns an error", func() {
				err := templatesCompiler.Compile(jobs, deployment)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-mkdirall-error"))
			})
		})

		Context("when rendering fails", func() {
			BeforeEach(func() {
				renderer.SetRenderBehavior(
					"fake-extracted-path/templates/cpi.erb",
					filepath.Join(compileDir, "bin/cpi"),
					context,
					errors.New("fake-render-error"),
				)
			})

			It("returns an error", func() {
				err := templatesCompiler.Compile(jobs, deployment)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-render-error"))
			})
		})

		Context("when generating templates archive fails", func() {
			BeforeEach(func() {
				compressor.CompressFilesInDirErr = errors.New("fake-compress-error")
			})

			It("returns an error", func() {
				err := templatesCompiler.Compile(jobs, deployment)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-compress-error"))
			})
		})

		Context("when saving to blobstore fails", func() {
			BeforeEach(func() {
				blobstore.CreateErr = errors.New("fake-blobstore-error")
			})

			It("returns an error", func() {
				err := templatesCompiler.Compile(jobs, deployment)
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
				err := templatesCompiler.Compile(jobs, deployment)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-template-error"))
			})
		})

		Context("when one of the job fails to compile", func() {
			BeforeEach(func() {
				jobs = []bmrel.Job{
					bmrel.Job{
						Name:          "fake-job-1",
						ExtractedPath: "fake-extracted-path-1",
						Templates: map[string]string{
							"cpi.erb": "/bin/cpi",
						},
					},
					bmrel.Job{
						Name:          "fake-job-2",
						ExtractedPath: "fake-extracted-path-2",
						Templates: map[string]string{
							"cpi.erb": "/bin/cpi",
						},
					},
					bmrel.Job{
						Name:          "fake-job-3",
						ExtractedPath: "fake-extracted-path-3",
						Templates: map[string]string{
							"cpi.erb": "/bin/cpi",
						},
					},
				}

				renderer.SetRenderBehavior(
					"fake-extracted-path-1/templates/cpi.erb",
					filepath.Join(compileDir, "bin/cpi"),
					context,
					nil,
				)

				renderer.SetRenderBehavior(
					"fake-extracted-path-2/templates/cpi.erb",
					filepath.Join(compileDir, "bin/cpi"),
					context,
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
				err := templatesCompiler.Compile(jobs, deployment)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-render-2-error"))
			})
		})
	})
})
