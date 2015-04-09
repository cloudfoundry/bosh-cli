package templatescompiler_test

import (
	. "github.com/cloudfoundry/bosh-init/templatescompiler"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"

	"code.google.com/p/gomock/gomock"
	mock_template "github.com/cloudfoundry/bosh-init/templatescompiler/mocks"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakeblobs "github.com/cloudfoundry/bosh-agent/blobstore/fakes"
	fakecmd "github.com/cloudfoundry/bosh-agent/platform/commands/fakes"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	bmproperty "github.com/cloudfoundry/bosh-init/common/property"
	bmreljob "github.com/cloudfoundry/bosh-init/release/job"

	fakebmtemp "github.com/cloudfoundry/bosh-init/templatescompiler/fakes"
	fakebmui "github.com/cloudfoundry/bosh-init/ui/fakes"
)

var _ = Describe("TemplatesCompiler", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		mockJobListRenderer *mock_template.MockJobListRenderer

		templatesCompiler TemplatesCompiler
		compressor        *fakecmd.FakeCompressor
		blobstore         *fakeblobs.FakeBlobstore
		templatesRepo     *fakebmtemp.FakeTemplatesRepo
		fs                *fakesys.FakeFileSystem
		compileDir        string
		jobs              []bmreljob.Job
		logger            boshlog.Logger

		jobProperties    bmproperty.Map
		globalProperties bmproperty.Map
		deploymentName   string

		fakeStage *fakebmui.FakeStage

		expectJobRender *gomock.Call

		renderedPath         = "/fake-temp-dir"
		renderedTemplatePath = "/fake-temp-dir/bin/cpi"
	)

	BeforeEach(func() {
		mockJobListRenderer = mock_template.NewMockJobListRenderer(mockCtrl)

		compressor = fakecmd.NewFakeCompressor()
		compressor.CompressFilesInDirTarballPath = "fake-tarball-path"

		blobstore = fakeblobs.NewFakeBlobstore()
		fs = fakesys.NewFakeFileSystem()

		templatesRepo = fakebmtemp.NewFakeTemplatesRepo()

		jobProperties = bmproperty.Map{
			"fake-property-key": "fake-job-property-value",
		}

		// installation ignores global properties
		globalProperties = bmproperty.Map{}

		deploymentName = "fake-deployment-name"

		logger = boshlog.NewLogger(boshlog.LevelNone)

		templatesCompiler = NewTemplatesCompiler(
			mockJobListRenderer,
			compressor,
			blobstore,
			templatesRepo,
			fs,
			logger,
		)

		var err error
		compileDir, err = fs.TempDir("bosh-init-tests")
		Expect(err).ToNot(HaveOccurred())
		fs.TempDirDir = compileDir
	})

	JustBeforeEach(func() {
		job := bmreljob.Job{
			Name:          "fake-job-1",
			Fingerprint:   "",
			SHA1:          "",
			ExtractedPath: "fake-extracted-path",
			Templates: map[string]string{
				"cpi.erb": "/bin/cpi",
			},
			PackageNames: nil,
			Packages:     nil,
			Properties:   nil,
		}
		jobs := []bmreljob.Job{job}

		fakeStage = fakebmui.NewFakeStage()

		renderedJob := NewRenderedJob(job, renderedPath, fs, logger)
		renderedJobList := NewRenderedJobList()
		renderedJobList.Add(renderedJob)

		expectJobRender = mockJobListRenderer.EXPECT().Render(jobs, jobProperties, globalProperties, deploymentName).Do(func(_, _, _, _ interface{}) {
			err := fs.MkdirAll(renderedPath, os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(renderedTemplatePath, "fake-bin/cpi-content")
			Expect(err).ToNot(HaveOccurred())
		}).Return(renderedJobList, nil).AnyTimes()
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
				BlobSHA1: "fake-sha1",
			}
			templatesRepo.SetSaveBehavior(jobs[0], record, nil)
		})

		It("renders job templates", func() {
			expectJobRender.Times(1)

			fs.TempDirDir = "/fake-temp-dir"
			err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
			Expect(err).ToNot(HaveOccurred())
		})

		It("logs an event step", func() {
			err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeStage.PerformCalls).To(Equal([]fakebmui.PerformCall{
				{Name: "Rendering job templates"},
			}))
		})

		It("cleans the temp folder to hold the compile result for job", func() {
			err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
			Expect(err).ToNot(HaveOccurred())
			Expect(fs.FileExists(renderedPath)).To(BeFalse())
			Expect(fs.FileExists(renderedTemplatePath)).To(BeFalse())
		})

		It("generates templates archive", func() {
			err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
			Expect(err).ToNot(HaveOccurred())
			Expect(compressor.CompressFilesInDirDir).To(Equal(renderedPath))
			Expect(compressor.CleanUpTarballPath).To(Equal("fake-tarball-path"))
		})

		It("saves archive in blobstore", func() {
			err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
			Expect(err).ToNot(HaveOccurred())
			Expect(blobstore.CreateFileNames).To(Equal([]string{"fake-tarball-path"}))
		})

		It("stores the compiled package blobID and fingerprint into the compile package repo", func() {
			err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			record := TemplateRecord{
				BlobID:   "fake-blob-id",
				BlobSHA1: "fake-sha1",
			}

			Expect(templatesRepo.SaveInputs).To(ContainElement(
				fakebmtemp.SaveInput{Job: jobs[0], Record: record},
			))
		})

		Context("when rendering fails", func() {
			var (
				renderError = bosherr.Error("fake-render-error")
			)

			JustBeforeEach(func() {
				expectJobRender.Return(nil, renderError).Times(1)
			})

			It("returns an error", func() {
				err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-render-error"))
			})

			It("logs an event step", func() {
				err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.PerformCalls).To(Equal([]fakebmui.PerformCall{
					{
						Name:  "Rendering job templates",
						Error: renderError,
					},
				}))
			})
		})

		Context("when generating templates archive fails", func() {
			BeforeEach(func() {
				compressor.CompressFilesInDirErr = bosherr.Error("fake-compress-error")
			})

			It("returns an error", func() {
				err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-compress-error"))
			})

			It("logs an event step", func() {
				err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.PerformCalls[0].Name).To(Equal("Rendering job templates"))
				Expect(fakeStage.PerformCalls[0].Error).To(HaveOccurred())
				Expect(fakeStage.PerformCalls[0].Error.Error()).To(Equal("Compressing rendered job templates: fake-compress-error"))
			})
		})

		Context("when saving to blobstore fails", func() {
			BeforeEach(func() {
				blobstore.CreateErr = bosherr.Error("fake-blobstore-error")
			})

			It("returns an error", func() {
				err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-blobstore-error"))
			})

			It("logs an event step", func() {
				err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.PerformCalls[0].Name).To(Equal("Rendering job templates"))
				Expect(fakeStage.PerformCalls[0].Error).To(HaveOccurred())
				Expect(fakeStage.PerformCalls[0].Error.Error()).To(Equal("Creating blob: fake-blobstore-error"))
			})
		})

		Context("when saving to templates repo fails", func() {
			BeforeEach(func() {
				record := TemplateRecord{
					BlobID:   "fake-blob-id",
					BlobSHA1: "fake-sha1",
				}

				err := bosherr.Error("fake-template-error")
				templatesRepo.SetSaveBehavior(jobs[0], record, err)
			})

			It("returns an error", func() {
				err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-template-error"))
			})

			It("logs an event step", func() {
				err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.PerformCalls[0].Name).To(Equal("Rendering job templates"))
				Expect(fakeStage.PerformCalls[0].Error).To(HaveOccurred())
				Expect(fakeStage.PerformCalls[0].Error.Error()).To(Equal("Saving job to templates repo: fake-template-error"))
			})
		})

		Context("when one of the job fails to compile", func() {
			var (
				renderError = bosherr.Error("fake-render-2-error")
			)

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

				mockJobListRenderer.EXPECT().Render(jobs, jobProperties, globalProperties, deploymentName).Return(nil, renderError)

				record := TemplateRecord{
					BlobID:   "fake-blob-id",
					BlobSHA1: "fake-sha1",
				}
				templatesRepo.SetSaveBehavior(jobs[0], record, nil)
				templatesRepo.SetSaveBehavior(jobs[1], record, nil)
			})

			It("returns an error", func() {
				err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-render-2-error"))
			})

			It("logs an event step", func() {
				err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.PerformCalls).To(Equal([]fakebmui.PerformCall{
					{
						Name:  "Rendering job templates",
						Error: renderError,
					},
				}))
			})
		})
	})
})
