package compile_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmindex "github.com/cloudfoundry/bosh-micro-cli/index"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelcomp "github.com/cloudfoundry/bosh-micro-cli/release/compile"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/release/compile"
)

var _ = Describe("CompiledPackageRepo", func() {
	var (
		index               bmindex.Index
		compiledPackageRepo CompiledPackageRepo
		fs                  *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		index = bmindex.NewFileIndex("/index_file", fs)
		compiledPackageRepo = NewCompiledPackageRepo(index)
	})

	Context("Save and Find", func() {
		It("saves the compiled package to the index", func() {
			record := bmrelcomp.CompiledPackageRecord{
				BlobID: "fake-blob-id",
				SHA1:   "fake-sha1",
			}

			pkg := bmrel.Package{
				Name:        "fake-package-name",
				Version:     "fake-version",
				Fingerprint: "fake-finger-print",
			}
			err := compiledPackageRepo.Save(pkg, record)
			Expect(err).ToNot(HaveOccurred())

			result, found, err := compiledPackageRepo.Find(pkg)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(result).To(Equal(record))
		})

		It("returns false when it finding before saving", func() {
			pkg := bmrel.Package{
				Name: "fake-package-name",
			}
			_, found, err := compiledPackageRepo.Find(pkg)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
		})

		Context("when storing the packages", func() {
			var (
				record bmrelcomp.CompiledPackageRecord
				pkg    bmrel.Package
			)

			BeforeEach(func() {
				record = bmrelcomp.CompiledPackageRecord{
					BlobID: "fake-blob-id",
					SHA1:   "fake-sha1",
				}
				pkg = bmrel.Package{
					Name:        "fake-package-name",
					Version:     "fake-version",
					Fingerprint: "fake-finger-print",
				}
				err := compiledPackageRepo.Save(pkg, record)
				Expect(err).ToNot(HaveOccurred())
			})

			It("considers package name in the key", func() {
				pkg.Name = "new-fake-name"
				_, found, err := compiledPackageRepo.Find(pkg)
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})

			It("considers package version in the key", func() {
				pkg.Version = "new-fake-version"
				_, found, err := compiledPackageRepo.Find(pkg)
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})

			It("considers package fingerprint in the key", func() {
				pkg.Fingerprint = "new-fake-fingerprint"
				_, found, err := compiledPackageRepo.Find(pkg)
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when saving to index fails", func() {
			It("returns error", func() {
				fs.WriteToFileError = errors.New("Could not save")
				record := bmrelcomp.CompiledPackageRecord{
					BlobID: "fake-blob-id",
					SHA1:   "fake-sha1",
				}

				pkg := bmrel.Package{
					Name: "fake-package-name",
				}

				err := compiledPackageRepo.Save(pkg, record)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Saving compiled package"))
			})
		})

		Context("when reading from index fails", func() {
			It("returns error", func() {
				record := bmrelcomp.CompiledPackageRecord{
					BlobID: "fake-blob-id",
					SHA1:   "fake-sha1",
				}
				pkg := bmrel.Package{
					Name: "fake-package-name",
				}
				err := compiledPackageRepo.Save(pkg, record)
				fs.ReadFileError = errors.New("fake-error")

				_, _, err = compiledPackageRepo.Find(pkg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Finding compiled package"))
			})
		})
	})
})
