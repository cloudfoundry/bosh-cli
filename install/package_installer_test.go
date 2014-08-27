package install_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmpkgs "github.com/cloudfoundry/bosh-micro-cli/packages"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fakeblobstore "github.com/cloudfoundry/bosh-agent/blobstore/fakes"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	fakebmpkgs "github.com/cloudfoundry/bosh-micro-cli/packages/fakes"
	fakebmtar "github.com/cloudfoundry/bosh-micro-cli/tar/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/install"
)

var _ = Describe("Install", func() {
	var (
		installer PackageInstaller
		repo      *fakebmpkgs.FakeCompiledPackageRepo
		blobstore *fakeblobstore.FakeBlobstore
		targetDir string
		extractor *fakebmtar.FakeExtractor
		pkg       *bmrel.Package
		logger    boshlog.Logger
		fs        *fakesys.FakeFileSystem
	)
	BeforeEach(func() {
		repo = fakebmpkgs.NewFakeCompiledPackageRepo()
		blobstore = fakeblobstore.NewFakeBlobstore()
		targetDir = "fake-target-dir"
		extractor = fakebmtar.NewFakeExtractor()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()

		installer = NewPackageInstaller(repo, blobstore, extractor, fs, logger)

		pkg = &bmrel.Package{
			Name:         "fake-package-name",
			Version:      "fake-package-version",
			Fingerprint:  "fake-package-fingerprint",
			Sha1:         "fake-package-sha1",
			Dependencies: []*bmrel.Package{},
		}
	})

	Context("when the package exists in the repo", func() {
		BeforeEach(func() {
			repo.FindCompiledPackageRecord = bmpkgs.CompiledPackageRecord{
				BlobID:      "fake-blob-id",
				Fingerprint: "fake-package-fingerprint",
			}
			blobstore.GetFileName = "/tmp/fake-blob-file"
			extractor.AddExpectedArchive("/tmp/fake-blob-file")
		})

		It("gets the package record from the repo", func() {
			err := installer.Install(pkg, targetDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(blobstore.GetBlobIDs).To(Equal([]string{"fake-blob-id"}))
			Expect(blobstore.GetFingerprints).To(Equal([]string{"fake-package-fingerprint"}))
		})

		It("creates the target dir if it does not exist", func() {
			Expect(fs.FileExists(targetDir)).To(BeFalse())
			err := installer.Install(pkg, targetDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(fs.FileExists(targetDir)).To(BeTrue())
		})

		It("extracts the blob into the target dir", func() {
			err := installer.Install(pkg, targetDir)
			Expect(err).ToNot(HaveOccurred())
			//TODO: expect that file is extracted to the targetDir
		})

		It("cleans up the blob file", func() {
			err := installer.Install(pkg, targetDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(blobstore.CleanUpFileName).To(Equal("/tmp/fake-blob-file"))
		})

		Context("when finding the package in the repo errors", func() {
			BeforeEach(func() {
				repo.FindCompiledPackageError = errors.New("fake-error")
			})

			It("returns an error", func() {
				err := installer.Install(pkg, targetDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Finding compiled package record"))
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})

		Context("when getting the blob from the blobstore errors", func() {
			BeforeEach(func() {
				blobstore.GetError = errors.New("fake-error")
			})

			It("returns an error", func() {
				err := installer.Install(pkg, targetDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Getting compiled package from blobstore"))
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})

		Context("when creating the target dir fails", func() {
			It("return an error", func() {
				fs.MkdirAllError = errors.New("fake-error")
				err := installer.Install(pkg, targetDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Creating target dir"))
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})

		//TODO: test extraction failure
		XContext("when extracting the blob fails", func() {
			BeforeEach(func() {
				//				extractor.ExtractError = errors.New("fake-error")
			})

			It("returns an error", func() {
				err := installer.Install(pkg, targetDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Extracting compiled package"))
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})

			It("cleans up the target dir if it was created by this installer", func() {})
		})

		Context("when cleaning up the downloaded blob errors", func() {
			BeforeEach(func() {
				blobstore.CleanUpErr = errors.New("fake-error")
			})

			It("does not return the error", func() {
				err := installer.Install(pkg, targetDir)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Context("when the package does not exist in the repo", func() {
		It("returns an error", func() {
			err := installer.Install(pkg, targetDir)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Compiled package record not found"))
		})
	})
})
