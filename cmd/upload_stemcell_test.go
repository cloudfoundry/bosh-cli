package cmd_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	semver "github.com/cppforlife/go-semi-semantic/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
)

var _ = Describe("UploadStemcellCmd", func() {
	var (
		director         *fakedir.FakeDirector
		fs               *fakesys.FakeFileSystem
		archive          *fakedir.FakeStemcellArchive
		ui               *fakeui.FakeUI
		command          UploadStemcellCmd
		existingInfo     boshdir.StemcellInfo
		existingMetadata boshdir.StemcellMetadata
	)

	BeforeEach(func() {
		director = &fakedir.FakeDirector{}
		fs = fakesys.NewFakeFileSystem()
		archive = &fakedir.FakeStemcellArchive{}
		ui = &fakeui.FakeUI{}
		existingInfo = boshdir.StemcellInfo{Name: "existing-name", Version: "existing-ver"}
		existingMetadata = boshdir.StemcellMetadata{Name: "existing-name", Version: "existing-ver"}

		stemcellArchiveFactory := func(path string) boshdir.StemcellArchive {
			if archive.FileStub == nil {
				archive.FileStub = func() (boshdir.UploadFile, error) {
					return fakesys.NewFakeFile(path, fs), nil
				}
			}
			return archive
		}

		command = NewUploadStemcellCmd(director, stemcellArchiveFactory, ui)
	})

	Describe("Run", func() {
		var (
			opts UploadStemcellOpts
		)

		BeforeEach(func() {
			opts = UploadStemcellOpts{}
		})

		act := func() error { return command.Run(opts) }

		Context("when url is remote (http/https)", func() {
			BeforeEach(func() {
				opts.Args.URL = "https://some-file.tzg"
			})

			It("uploads given stemcell", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.UploadStemcellURLCallCount()).To(Equal(1))

				url, sha1, fix := director.UploadStemcellURLArgsForCall(0)
				Expect(url).To(Equal("https://some-file.tzg"))
				Expect(sha1).To(Equal(""))
				Expect(fix).To(BeFalse())
			})

			It("uploads given stemcell with a fix flag without checking if stemcell exists", func() {
				opts.Fix = true

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.HasStemcellCallCount()).To(Equal(0))
				Expect(director.StemcellNeedsUploadCallCount()).To(Equal(0))

				Expect(director.UploadStemcellURLCallCount()).To(Equal(1))

				url, sha1, fix := director.UploadStemcellURLArgsForCall(0)
				Expect(url).To(Equal("https://some-file.tzg"))
				Expect(sha1).To(Equal(""))
				Expect(fix).To(BeTrue())
			})

			It("uploads given stemcell with a specified sha1", func() {
				opts.SHA1 = "sha1"

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.UploadStemcellURLCallCount()).To(Equal(1))

				url, sha1, fix := director.UploadStemcellURLArgsForCall(0)
				Expect(url).To(Equal("https://some-file.tzg"))
				Expect(sha1).To(Equal("sha1"))
				Expect(fix).To(BeFalse())
			})

			Context("when a director supports StemcellMatches", func() {
				It("uploads a stemcell when any CPI is missing it", func() {
					opts.Name = "existing-name"
					opts.Version = VersionArg(semver.MustNewVersionFromString("existing-ver"))

					director.StemcellNeedsUploadReturns(
						true,
						true,
						nil,
					)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(director.UploadStemcellURLCallCount()).To(Equal(1))

					url, sha1, fix := director.UploadStemcellURLArgsForCall(0)
					Expect(url).To(Equal("https://some-file.tzg"))
					Expect(sha1).To(Equal(""))
					Expect(fix).To(BeFalse())

					Expect(director.StemcellNeedsUploadCallCount()).To(Equal(1))
					submission := director.StemcellNeedsUploadArgsForCall(0)
					Expect(submission).To(Equal(existingInfo))

					Expect(ui.Said).To(BeEmpty())
				})

				It("does not upload stemcell if no CPI needs that name and version", func() {
					opts.Name = "existing-name"
					opts.Version = VersionArg(semver.MustNewVersionFromString("existing-ver"))

					director.StemcellNeedsUploadReturns(false, true, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(director.UploadStemcellURLCallCount()).To(Equal(0))

					Expect(director.StemcellNeedsUploadCallCount()).To(Equal(1))
					submission := director.StemcellNeedsUploadArgsForCall(0)
					Expect(submission).To(Equal(existingInfo))

					Expect(ui.Said).To(Equal([]string{"Stemcell 'existing-name/existing-ver' already exists."}))
				})

				It("returns error if checking for stemcell existence fails", func() {
					director.StemcellNeedsUploadReturns(false, true, errors.New("fake-err"))

					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))

					Expect(director.UploadStemcellURLCallCount()).To(Equal(0))
				})
			})

			Context("when a legacy director does not support stemcell_uploads", func() {
				BeforeEach(func() {
					director.StemcellNeedsUploadReturns(true, false, nil)
				})

				It("does not upload stemcell if name and version match existing stemcell", func() {
					opts.Name = "existing-name"
					opts.Version = VersionArg(semver.MustNewVersionFromString("existing-ver"))

					director.HasStemcellReturns(true, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(director.UploadStemcellURLCallCount()).To(Equal(0))

					name, version := director.HasStemcellArgsForCall(0)
					Expect(name).To(Equal("existing-name"))
					Expect(version).To(Equal("existing-ver"))

					Expect(ui.Said).To(Equal([]string{"Stemcell 'existing-name/existing-ver' already exists."}))
				})

				It("uploads stemcell if name and version does not match existing stemcell", func() {
					opts.Name = "existing-name"
					opts.Version = VersionArg(semver.MustNewVersionFromString("existing-ver"))

					director.HasStemcellReturns(false, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(director.UploadStemcellURLCallCount()).To(Equal(1))

					url, sha1, fix := director.UploadStemcellURLArgsForCall(0)
					Expect(url).To(Equal("https://some-file.tzg"))
					Expect(sha1).To(Equal(""))
					Expect(fix).To(BeFalse())

					name, version := director.HasStemcellArgsForCall(0)
					Expect(name).To(Equal("existing-name"))
					Expect(version).To(Equal("existing-ver"))

					Expect(ui.Said).To(BeEmpty())
				})

				It("returns error if checking for stemcell existence fails", func() {
					director.HasStemcellReturns(false, errors.New("fake-err"))

					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))

					Expect(director.UploadStemcellURLCallCount()).To(Equal(0))
				})
			})

			It("returns error if uploading stemcell failed", func() {
				director.UploadStemcellURLReturns(errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})

		Context("when url is a local file (file or no prefix)", func() {
			BeforeEach(func() {
				opts.Args.URL = "./some-file.tgz"
			})

			It("uploads given stemcell", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.UploadStemcellFileCallCount()).To(Equal(1))

				file, fix := director.UploadStemcellFileArgsForCall(0)
				Expect(file.(*fakesys.FakeFile).Name()).To(Equal("./some-file.tgz"))
				Expect(fix).To(BeFalse())
			})

			It("uploads given stemcell with a fix flag without checking if stemcell exists", func() {
				opts.Fix = true

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.HasStemcellCallCount()).To(Equal(0))
				Expect(director.StemcellNeedsUploadCallCount()).To(Equal(0))

				Expect(director.UploadStemcellFileCallCount()).To(Equal(1))

				file, fix := director.UploadStemcellFileArgsForCall(0)
				Expect(file.(*fakesys.FakeFile).Name()).To(Equal("./some-file.tgz"))
				Expect(fix).To(BeTrue())
			})

			It("returns error if retrieving stemcell archive info fails", func() {
				archive.InfoReturns(boshdir.StemcellMetadata{}, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))

				Expect(director.UploadStemcellFileCallCount()).To(Equal(0))
			})

			It("returns error if opening file fails", func() {
				archive.FileStub = func() (boshdir.UploadFile, error) {
					return nil, errors.New("fake-err")
				}

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))

				Expect(director.UploadStemcellFileCallCount()).To(Equal(0))
			})

			It("returns error if uploading stemcell failed", func() {
				director.UploadStemcellFileReturns(errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})

			Context("when a director supports StemcellMatches", func() {
				It("uploads a stemcell when any CPI is missing it", func() {
					archive.InfoReturns(existingMetadata, nil)

					director.StemcellNeedsUploadReturns(
						true,
						true,
						nil,
					)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(director.UploadStemcellFileCallCount()).To(Equal(1))

					file, fix := director.UploadStemcellFileArgsForCall(0)
					Expect(file.(*fakesys.FakeFile).Name()).To(Equal("./some-file.tgz"))
					Expect(fix).To(BeFalse())

					Expect(director.StemcellNeedsUploadCallCount()).To(Equal(1))
					submission := director.StemcellNeedsUploadArgsForCall(0)
					Expect(submission).To(Equal(existingInfo))

					Expect(ui.Said).To(BeEmpty())
				})

				It("does not upload stemcell if no CPI needs that name and version", func() {
					archive.InfoReturns(existingMetadata, nil)

					director.StemcellNeedsUploadReturns(false, true, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(director.UploadStemcellFileCallCount()).To(Equal(0))

					Expect(director.StemcellNeedsUploadCallCount()).To(Equal(1))
					submission := director.StemcellNeedsUploadArgsForCall(0)
					Expect(submission).To(Equal(existingInfo))

					Expect(ui.Said).To(Equal([]string{"Stemcell 'existing-name/existing-ver' already exists."}))
				})

				It("returns error if checking for stemcell existence fails", func() {
					director.StemcellNeedsUploadReturns(false, true, errors.New("fake-err"))

					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))

					Expect(director.UploadStemcellFileCallCount()).To(Equal(0))
				})
			})

			Context("when a legacy director does not support stemcell_uploads", func() {
				BeforeEach(func() {
					director.StemcellNeedsUploadReturns(false, false, nil)
				})

				It("uploads stemcell if name and version does not match existing stemcell", func() {
					archive.InfoReturns(existingMetadata, nil)

					director.HasStemcellReturns(false, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(director.UploadStemcellFileCallCount()).To(Equal(1))

					file, fix := director.UploadStemcellFileArgsForCall(0)
					Expect(file.(*fakesys.FakeFile).Name()).To(Equal("./some-file.tgz"))
					Expect(fix).To(BeFalse())

					name, version := director.HasStemcellArgsForCall(0)
					Expect(name).To(Equal("existing-name"))
					Expect(version).To(Equal("existing-ver"))

					Expect(ui.Said).To(BeEmpty())
				})

				It("does not upload stemcell if name and version match existing stemcell", func() {
					archive.InfoReturns(existingMetadata, nil)

					director.HasStemcellReturns(true, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(director.UploadStemcellFileCallCount()).To(Equal(0))

					name, version := director.HasStemcellArgsForCall(0)
					Expect(name).To(Equal("existing-name"))
					Expect(version).To(Equal("existing-ver"))

					Expect(ui.Said).To(Equal(
						[]string{"Stemcell 'existing-name/existing-ver' already exists."}))
				})

				It("returns error if checking for stemcell existence fails", func() {
					director.HasStemcellReturns(false, errors.New("fake-err"))

					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))

					Expect(director.UploadStemcellFileCallCount()).To(Equal(0))
				})

			})
		})
	})
})
