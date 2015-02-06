package templatescompiler_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"os"

	"code.google.com/p/gomock/gomock"
	mock_template "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/mocks"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakeblobs "github.com/cloudfoundry/bosh-agent/blobstore/fakes"
	fakecmd "github.com/cloudfoundry/bosh-agent/platform/commands/fakes"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"

	fakebmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"
	fakebmtemp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/fakes"
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
		mockJobRenderer *mock_template.MockJobRenderer

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

		fakeStage *fakebmeventlog.FakeStage

		expectJobRender *gomock.Call

		renderedPath         = "/fake-temp-dir"
		renderedTemplatePath = "/fake-temp-dir/bin/cpi"
	)

	BeforeEach(func() {
		mockJobRenderer = mock_template.NewMockJobRenderer(mockCtrl)

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
			mockJobRenderer,
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

		fakeStage = fakebmeventlog.NewFakeStage()

		renderedJob := NewRenderedJob(job, renderedPath, fs, logger)

		expectJobRender = mockJobRenderer.EXPECT().Render(job, jobProperties, globalProperties, deploymentName).Do(func(_, _, _, _ interface{}) {
			err := fs.MkdirAll(renderedPath, os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(renderedTemplatePath, "fake-bin/cpi-content")
			Expect(err).ToNot(HaveOccurred())
		}).Return(renderedJob, nil).AnyTimes()
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

			Expect(fakeStage.Steps).To(ContainElement(&fakebmeventlog.FakeStep{
				Name: "Rendering job templates",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
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
			JustBeforeEach(func() {
				expectJobRender.Return(nil, errors.New("fake-render-error")).Times(1)
			})

			It("returns an error", func() {
				err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-render-error"))
			})

			It("logs an event step", func() {
				err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmeventlog.FakeStep{
					Name: "Rendering job templates",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Rendering templates for job 'fake-job-1': fake-render-error",
				}))
			})
		})

		Context("when generating templates archive fails", func() {
			BeforeEach(func() {
				compressor.CompressFilesInDirErr = errors.New("fake-compress-error")
			})

			It("returns an error", func() {
				err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-compress-error"))
			})

			It("logs an event step", func() {
				err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmeventlog.FakeStep{
					Name: "Rendering job templates",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Compressing rendered job templates: fake-compress-error",
				}))
			})
		})

		Context("when saving to blobstore fails", func() {
			BeforeEach(func() {
				blobstore.CreateErr = errors.New("fake-blobstore-error")
			})

			It("returns an error", func() {
				err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-blobstore-error"))
			})

			It("logs an event step", func() {
				err := templatesCompiler.Compile(jobs, "fake-deployment-name", jobProperties, fakeStage)
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmeventlog.FakeStep{
					Name: "Rendering job templates",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Creating blob: fake-blobstore-error",
				}))
			})
		})

		Context("when saving to templates repo fails", func() {
			BeforeEach(func() {
				record := TemplateRecord{
					BlobID:   "fake-blob-id",
					BlobSHA1: "fake-sha1",
				}

				err := errors.New("fake-template-error")
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

				Expect(fakeStage.Steps).To(ContainElement(&fakebmeventlog.FakeStep{
					Name: "Rendering job templates",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Saving job to templates repo: fake-template-error",
				}))
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

				renderedJob := NewRenderedJob(jobs[0], renderedPath, fs, logger)
				mockJobRenderer.EXPECT().Render(jobs[0], jobProperties, globalProperties, deploymentName).Return(renderedJob, nil)

				mockJobRenderer.EXPECT().Render(jobs[1], jobProperties, globalProperties, deploymentName).Return(nil, errors.New("fake-render-2-error"))

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

				Expect(fakeStage.Steps).To(ContainElement(&fakebmeventlog.FakeStep{
					Name: "Rendering job templates",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Rendering templates for job 'fake-job-2': fake-render-2-error",
				}))
			})
		})
	})
})
