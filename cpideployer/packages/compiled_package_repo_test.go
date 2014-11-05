package packages_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmindex "github.com/cloudfoundry/bosh-micro-cli/index"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/cpideployer/packages"
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
		var (
			record     CompiledPackageRecord
			dependency bmrel.Package
			pkg        bmrel.Package
		)

		BeforeEach(func() {
			record = CompiledPackageRecord{}
			dependency = bmrel.Package{
				Name:        "fake-dependency-package",
				Fingerprint: "fake-dependency-fingerprint",
			}
			pkg = bmrel.Package{
				Name:         "fake-package-name",
				Fingerprint:  "fake-package-fingerprint",
				Dependencies: []*bmrel.Package{&dependency},
			}
		})

		It("saves the compiled package to the index", func() {
			err := compiledPackageRepo.Save(pkg, record)
			Expect(err).ToNot(HaveOccurred())

			result, found, err := compiledPackageRepo.Find(pkg)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(result).To(Equal(record))
		})

		It("returns false when finding before saving", func() {
			pkg := bmrel.Package{
				Name: "fake-package-name",
			}
			_, found, err := compiledPackageRepo.Find(pkg)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
		})

		It("returns false if package dependencies have changed after saving", func() {
			err := compiledPackageRepo.Save(pkg, record)
			Expect(err).ToNot(HaveOccurred())

			_, found, err := compiledPackageRepo.Find(pkg)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())

			dependency.Fingerprint = "new-fake-dependency-fingerprint"

			_, found, err = compiledPackageRepo.Find(pkg)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
		})

		It("returns true if dependency order changed", func() {
			dependency1 := bmrel.Package{
				Name:        "fake-package-1",
				Fingerprint: "fake-dependency-fingerprint-1",
			}
			dependency2 := bmrel.Package{
				Name:        "fake-package-2",
				Fingerprint: "fake-dependency-fingerprint-2",
			}

			pkg.Dependencies = []*bmrel.Package{&dependency1, &dependency2}

			err := compiledPackageRepo.Save(pkg, record)
			Expect(err).ToNot(HaveOccurred())

			pkg.Dependencies = []*bmrel.Package{&dependency2, &dependency1}

			result, found, err := compiledPackageRepo.Find(pkg)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(result).To(Equal(record))
		})

		It("returns false if a transitive dependency has changed after saving", func() {
			transitive := bmrel.Package{
				Name:        "fake-transitive-package",
				Fingerprint: "fake-transitive-fingerprint",
			}
			dependency.Dependencies = []*bmrel.Package{&transitive}

			err := compiledPackageRepo.Save(pkg, record)
			Expect(err).ToNot(HaveOccurred())

			_, found, err := compiledPackageRepo.Find(pkg)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())

			transitive.Fingerprint = "new-fake-dependency-fingerprint"

			_, found, err = compiledPackageRepo.Find(pkg)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
		})

		Context("when saving to index fails", func() {
			It("returns error", func() {
				fs.WriteToFileError = errors.New("Could not save")
				record := CompiledPackageRecord{
					BlobID:   "fake-blob-id",
					BlobSHA1: "fake-sha1",
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
				err := compiledPackageRepo.Save(pkg, record)
				fs.ReadFileError = errors.New("fake-error")

				_, _, err = compiledPackageRepo.Find(pkg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Finding compiled package"))
			})
		})
	})
})
