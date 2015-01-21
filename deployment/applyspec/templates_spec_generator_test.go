package applyspec_test

import (
	"encoding/json"
	"errors"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fakecmd "github.com/cloudfoundry/bosh-agent/platform/commands/fakes"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"
	fakebmblobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore/fakes"
	fakebmcrypto "github.com/cloudfoundry/bosh-micro-cli/crypto/fakes"
	fakebmtemp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
)

var _ = Describe("TemplatesSpecGenerator", func() {
	var (
		templatesSpecGenerator TemplatesSpecGenerator
		fakeJobRenderer        *fakebmtemp.FakeJobRenderer
		fakeCompressor         *fakecmd.FakeCompressor
		fakeBlobstore          *fakebmblobstore.FakeBlobstore
		fakeBlobstoreFactory   *fakebmblobstore.FakeBlobstoreFactory
		fakeUUIDGenerator      *fakeuuid.FakeGenerator
		fakeSha1Calculator     *fakebmcrypto.FakeSha1Calculator
		deploymentJob          bmdeplmanifest.Job
		jobBlobs               []bmstemcell.Blob
		extractedJob           bmrel.Job
		jobProperties          map[string]interface{}
		fs                     *fakesys.FakeFileSystem
		logger                 boshlog.Logger
		tempFile               boshsys.File
		compileDir             string
		extractDir             string
	)

	BeforeEach(func() {
		fakeJobRenderer = fakebmtemp.NewFakeJobRenderer()
		fakeCompressor = fakecmd.NewFakeCompressor()
		fakeCompressor.CompressFilesInDirTarballPath = "fake-tarball-path"
		jobBlobs = []bmstemcell.Blob{
			{
				Name:        "first-job-name",
				Version:     "first-job-version",
				SHA1:        "first-job-sha1",
				BlobstoreID: "first-job-blobstore-id",
			},
			{
				Name:        "second-job-name",
				Version:     "second-job-version",
				SHA1:        "second-job-sha1",
				BlobstoreID: "second-job-blobstore-id",
			},
			{
				Name:        "third-job-name",
				Version:     "third-job-version",
				SHA1:        "third-job-sha1",
				BlobstoreID: "third-job-blobstore-id",
			},
		}

		fakeBlobstore = fakebmblobstore.NewFakeBlobstore()
		fakeBlobstoreFactory = fakebmblobstore.NewFakeBlobstoreFactory()
		fakeBlobstoreFactory.CreateBlobstore = fakeBlobstore

		fakeUUIDGenerator = &fakeuuid.FakeGenerator{
			GeneratedUuid: "fake-blob-id",
		}

		jobProperties = map[string]interface{}{
			"fake-property-key": "fake-property-value",
		}

		fakeSha1Calculator = fakebmcrypto.NewFakeSha1Calculator()
		fs = fakesys.NewFakeFileSystem()
		logger = boshlog.NewLogger(boshlog.LevelNone)

		var err error
		tempFile, err = fs.TempFile("fake-blob-temp-file")
		Expect(err).ToNot(HaveOccurred())
		fs.ReturnTempFile = tempFile
		fs.TempDirDir = "/fake-tmp-dir"
		// fake file system only supports one temp dir
		compileDir = "/fake-tmp-dir"
		extractDir = "/fake-tmp-dir"
		deploymentJob = bmdeplmanifest.Job{
			Templates: []bmdeplmanifest.ReleaseJobRef{
				{
					Name: "first-job-name",
				},
				{
					Name: "third-job-name",
				},
			},
			RawProperties: map[interface{}]interface{}{
				"fake-property-key": "fake-property-value",
			},
		}

		templatesSpecGenerator = NewTemplatesSpecGenerator(
			fakeBlobstoreFactory,
			fakeCompressor,
			fakeJobRenderer,
			fakeUUIDGenerator,
			fakeSha1Calculator,
			fs,
			logger,
		)

		extractedJob = bmrel.Job{
			Templates: map[string]string{
				"director.yml.erb": "config/director.yml",
			},
			ExtractedPath: extractDir,
		}

		blobJobJSON, err := json.Marshal(extractedJob)
		Expect(err).ToNot(HaveOccurred())

		fakeCompressor.DecompressFileToDirCallBack = func() {
			fs.WriteFile("/fake-tmp-dir/job.MF", blobJobJSON)
			fs.WriteFile("/fake-tmp-dir/monit", []byte("fake-monit-contents"))
		}

		fakeCompressor.CompressFilesInDirTarballPath = "fake-tarball-path"

		fakeSha1Calculator.SetCalculateBehavior(map[string]fakebmcrypto.CalculateInput{
			compileDir: fakebmcrypto.CalculateInput{
				Sha1: "fake-configuration-hash",
			},
			"fake-tarball-path": fakebmcrypto.CalculateInput{
				Sha1: "fake-archive-sha1",
			},
		})
	})

	Describe("Create", func() {
		It("downloads only job template blobs from the blobstore that are specified in the manifest", func() {
			templatesSpec, err := templatesSpecGenerator.Create(deploymentJob, jobBlobs, "fake-deployment-name", jobProperties, "fake-blobstore-url")
			Expect(err).ToNot(HaveOccurred())
			Expect(templatesSpec).To(Equal(TemplatesSpec{
				BlobID:            "fake-blob-id",
				ArchiveSha1:       "fake-archive-sha1",
				ConfigurationHash: "fake-configuration-hash",
			}))
			Expect(fakeBlobstore.GetInputs).To(Equal([]fakebmblobstore.GetInput{
				{
					BlobID:          "first-job-blobstore-id",
					DestinationPath: tempFile.Name(),
				},
				{
					BlobID:          "third-job-blobstore-id",
					DestinationPath: tempFile.Name(),
				},
			}))
		})

		It("removes the tempfile for downloaded blobs", func() {
			tempFile, err := fs.TempFile("fake-blob-temp-file")
			Expect(err).ToNot(HaveOccurred())

			fs.ReturnTempFile = tempFile
			templatesSpec, err := templatesSpecGenerator.Create(deploymentJob, jobBlobs, "fake-deployment-name", jobProperties, "fake-blobstore-url")
			Expect(err).ToNot(HaveOccurred())
			Expect(templatesSpec).To(Equal(TemplatesSpec{
				BlobID:            "fake-blob-id",
				ArchiveSha1:       "fake-archive-sha1",
				ConfigurationHash: "fake-configuration-hash",
			}))

			Expect(fs.FileExists(tempFile.Name())).To(BeFalse())
		})

		It("decompressed job templates", func() {
			templatesSpec, err := templatesSpecGenerator.Create(deploymentJob, jobBlobs, "fake-deployment-name", jobProperties, "fake-blobstore-url")
			Expect(err).ToNot(HaveOccurred())
			Expect(templatesSpec).To(Equal(TemplatesSpec{
				BlobID:            "fake-blob-id",
				ArchiveSha1:       "fake-archive-sha1",
				ConfigurationHash: "fake-configuration-hash",
			}))

			Expect(fakeCompressor.DecompressFileToDirTarballPaths[0]).To(Equal(tempFile.Name()))
			Expect(fakeCompressor.DecompressFileToDirDirs[0]).To(Equal(extractDir))
		})

		It("renders job templates", func() {
			templatesSpec, err := templatesSpecGenerator.Create(deploymentJob, jobBlobs, "fake-deployment-name", jobProperties, "fake-blobstore-url")
			Expect(err).ToNot(HaveOccurred())
			Expect(templatesSpec).To(Equal(TemplatesSpec{
				BlobID:            "fake-blob-id",
				ArchiveSha1:       "fake-archive-sha1",
				ConfigurationHash: "fake-configuration-hash",
			}))

			Expect(fakeJobRenderer.RenderInputs).To(Equal([]fakebmtemp.RenderInput{
				{
					SourcePath:      extractDir,
					DestinationPath: filepath.Join(compileDir, "first-job-name"),
					Job:             extractedJob,
					Properties: map[string]interface{}{
						"fake-property-key": "fake-property-value",
					},
					DeploymentName: "fake-deployment-name",
				},
				{
					SourcePath:      extractDir,
					DestinationPath: filepath.Join(compileDir, "third-job-name"),
					Job:             extractedJob,
					Properties: map[string]interface{}{
						"fake-property-key": "fake-property-value",
					},
					DeploymentName: "fake-deployment-name",
				},
			}))
		})

		It("compresses rendered templates", func() {
			templatesSpec, err := templatesSpecGenerator.Create(deploymentJob, jobBlobs, "fake-deployment-name", jobProperties, "fake-blobstore-url")
			Expect(err).ToNot(HaveOccurred())
			Expect(templatesSpec).To(Equal(TemplatesSpec{
				BlobID:            "fake-blob-id",
				ArchiveSha1:       "fake-archive-sha1",
				ConfigurationHash: "fake-configuration-hash",
			}))

			Expect(fakeCompressor.CompressFilesInDirDir).To(Equal(compileDir))
		})

		It("cleans up rendered tarball", func() {
			templatesSpec, err := templatesSpecGenerator.Create(deploymentJob, jobBlobs, "fake-deployment-name", jobProperties, "fake-blobstore-url")
			Expect(err).ToNot(HaveOccurred())
			Expect(templatesSpec).To(Equal(TemplatesSpec{
				BlobID:            "fake-blob-id",
				ArchiveSha1:       "fake-archive-sha1",
				ConfigurationHash: "fake-configuration-hash",
			}))

			Expect(fs.FileExists("fake-tarball-path")).To(BeFalse())
		})

		It("uploads rendered jobs to the blobstore", func() {
			templatesSpec, err := templatesSpecGenerator.Create(deploymentJob, jobBlobs, "fake-deployment-name", jobProperties, "fake-blobstore-url")
			Expect(err).ToNot(HaveOccurred())
			Expect(templatesSpec).To(Equal(TemplatesSpec{
				BlobID:            "fake-blob-id",
				ArchiveSha1:       "fake-archive-sha1",
				ConfigurationHash: "fake-configuration-hash",
			}))

			Expect(fakeBlobstoreFactory.CreateBlobstoreURL).To(Equal("fake-blobstore-url"))

			Expect(fakeBlobstore.SaveInputs).To(Equal([]fakebmblobstore.SaveInput{
				{
					BlobID:     "fake-blob-id",
					SourcePath: "fake-tarball-path",
				},
			}))
		})

		Context("when creating temp directory fails", func() {
			It("returns an error", func() {
				fs.TempDirError = errors.New("fake-temp-dir-error")
				templatesSpec, err := templatesSpecGenerator.Create(deploymentJob, jobBlobs, "fake-deployment-name", jobProperties, "fake-blobstore-url")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-temp-dir-error"))
				Expect(templatesSpec).To(Equal(TemplatesSpec{}))
			})
		})

		Context("when creating blobstore fails", func() {
			It("returns an error", func() {
				fakeBlobstoreFactory.CreateErr = errors.New("fake-blobstore-factory-create-error")
				templatesSpec, err := templatesSpecGenerator.Create(deploymentJob, jobBlobs, "fake-deployment-name", jobProperties, "fake-blobstore-url")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-blobstore-factory-create-error"))
				Expect(templatesSpec).To(Equal(TemplatesSpec{}))
			})
		})

		Context("when getting blob from blobstore fails", func() {
			It("returns an error", func() {
				fakeBlobstore.GetErr = errors.New("fake-blobstore-get-error")
				templatesSpec, err := templatesSpecGenerator.Create(deploymentJob, jobBlobs, "fake-deployment-name", jobProperties, "fake-blobstore-url")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-blobstore-get-error"))
				Expect(templatesSpec).To(Equal(TemplatesSpec{}))
			})
		})

		Context("when rendering job fails", func() {
			It("returns an error", func() {
				fakeJobRenderer.SetRenderBehavior("/fake-tmp-dir", errors.New("fake-render-error"))
				templatesSpec, err := templatesSpecGenerator.Create(deploymentJob, jobBlobs, "fake-deployment-name", jobProperties, "fake-blobstore-url")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-render-error"))
				Expect(templatesSpec).To(Equal(TemplatesSpec{}))
			})
		})

		Context("when compressing rendered templates fails", func() {
			It("returns an error", func() {
				fakeJobRenderer.SetRenderBehavior("/fake-tmp-dir", errors.New("fake-render-error"))
				templatesSpec, err := templatesSpecGenerator.Create(deploymentJob, jobBlobs, "fake-deployment-name", jobProperties, "fake-blobstore-url")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-render-error"))
				Expect(templatesSpec).To(Equal(TemplatesSpec{}))
			})
		})

		Context("when calculating sha1 fails", func() {
			It("returns an error", func() {
				fakeSha1Calculator.SetCalculateBehavior(map[string]fakebmcrypto.CalculateInput{
					"/fake-tmp-dir": fakebmcrypto.CalculateInput{
						Sha1: "",
						Err:  errors.New("fake-sha1-error"),
					},
				})
				templatesSpec, err := templatesSpecGenerator.Create(deploymentJob, jobBlobs, "fake-deployment-name", jobProperties, "fake-blobstore-url")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-sha1-error"))
				Expect(templatesSpec).To(Equal(TemplatesSpec{}))
			})
		})
	})
})
