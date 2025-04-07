package cmd_test

import (
	"github.com/cloudfoundry/bosh-utils/errors"
	fakefu "github.com/cloudfoundry/bosh-utils/fileutil/fakes"
	fakes2 "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	fakecrypto "github.com/cloudfoundry/bosh-cli/v7/crypto/fakes"
	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
	boshjob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	"github.com/cloudfoundry/bosh-cli/v7/release/license"
	boshpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
	fakerel "github.com/cloudfoundry/bosh-cli/v7/release/releasefakes"
	. "github.com/cloudfoundry/bosh-cli/v7/release/resource"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

var _ = Describe("RedigestRelease", func() {
	var (
		releaseReader                *fakerel.FakeReader
		ui                           *fakeui.FakeUI
		fmv                          *fakefu.FakeMover
		releaseWriter                *fakerel.FakeWriter
		command                      cmd.RedigestReleaseCmd
		args                         opts.RedigestReleaseArgs
		fakeDigestCalculator         *fakecrypto.FakeDigestCalculator
		releaseWriterTempDestination string
		fakeSha128Release            *fakerel.FakeRelease
		fs                           *fakes2.FakeFileSystem
	)

	job1ResourcePath := "/job-resource-1-path"
	pkg1ResourcePath := "/pkg-resource-1-path"
	compiledPackage1ResourcePath := "/compiled-pkg-resource-path"
	licenseResourcePath := "/license-resource-path"
	fileContentSha1 := "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed"

	BeforeEach(func() {
		releaseReader = &fakerel.FakeReader{}
		releaseWriter = &fakerel.FakeWriter{}
		ui = &fakeui.FakeUI{}
		fmv = &fakefu.FakeMover{}
		fs = fakes2.NewFakeFileSystem()

		fakeDigestCalculator = fakecrypto.NewFakeDigestCalculator()
		command = cmd.NewRedigestReleaseCmd(releaseReader, releaseWriter, fakeDigestCalculator, fmv, fs, ui)
		args = opts.RedigestReleaseArgs{
			Path:        "/some/release_128.tgz",
			Destination: opts.FileArg{ExpandedPath: "/some/release_256.tgz"},
		}

		err := fs.WriteFileString(job1ResourcePath, "hello world")
		Expect(err).ToNot(HaveOccurred())
		err = fs.WriteFileString(pkg1ResourcePath, "hello world")
		Expect(err).ToNot(HaveOccurred())
		err = fs.WriteFileString(compiledPackage1ResourcePath, "hello world")
		Expect(err).ToNot(HaveOccurred())
		err = fs.WriteFileString(licenseResourcePath, "hello world")
		Expect(err).ToNot(HaveOccurred())

		fakeSha128Release = &fakerel.FakeRelease{}
		jobSha128 := boshjob.NewJob(NewResourceWithBuiltArchive("job-resource-1", "job-sha128-fp", job1ResourcePath, fileContentSha1))
		packageSha128 := boshpkg.NewPackage(NewResourceWithBuiltArchive("pkg-resource-1", "pkg-sha128-fp", pkg1ResourcePath, fileContentSha1), nil)
		compiledPackageSha128 := boshpkg.NewCompiledPackageWithArchive("compiledpkg-resource-1", "compiledpkg-sha128-fp", "1", compiledPackage1ResourcePath, fileContentSha1, nil)

		fakeSha128Release.JobsReturns([]*boshjob.Job{jobSha128})
		fakeSha128Release.PackagesReturns([]*boshpkg.Package{packageSha128})
		fakeSha128Release.LicenseReturns(license.NewLicense(NewResourceWithBuiltArchive("license-resource-path", "lic-sha128-fp", licenseResourcePath, fileContentSha1)))
		fakeSha128Release.CompiledPackagesReturns([]*boshpkg.CompiledPackage{compiledPackageSha128})

		fakeSha128Release.CopyWithStub = func(jobs []*boshjob.Job, pkgs []*boshpkg.Package, lic *license.License, compiledPackages []*boshpkg.CompiledPackage) boshrel.Release {
			fakeSha256Release := &fakerel.FakeRelease{}
			fakeSha256Release.NameReturns("custom-name")
			fakeSha256Release.VersionReturns("custom-ver")
			fakeSha256Release.CommitHashWithMarkReturns("commit")
			fakeSha256Release.JobsReturns(jobs)
			fakeSha256Release.PackagesReturns(pkgs)
			fakeSha256Release.LicenseReturns(lic)
			fakeSha256Release.CompiledPackagesReturns(compiledPackages)
			return fakeSha256Release
		}

		fakeDigestCalculator.SetCalculateBehavior(map[string]fakecrypto.CalculateInput{
			job1ResourcePath:             {DigestStr: "sha256:jobsha256"},
			pkg1ResourcePath:             {DigestStr: "sha256:pkgsha256"},
			licenseResourcePath:          {DigestStr: "sha256:licsha256"},
			compiledPackage1ResourcePath: {DigestStr: "sha256:compiledpkgsha256"},
		})

		releaseReader.ReadReturns(fakeSha128Release, nil)
		releaseWriterTempDestination = "/some/temp/release_256.tgz"
		releaseWriter.WriteReturns(releaseWriterTempDestination, nil)
	})

	Context("Given a valid sha128 release tar", func() {
		It("Should convert it to a sha256 release tar", func() {
			err := command.Run(args)
			Expect(err).ToNot(HaveOccurred())

			Expect(releaseReader.ReadCallCount()).ToNot(Equal(0))

			readPathArg := releaseReader.ReadArgsForCall(0)
			Expect(readPathArg).To(Equal("/some/release_128.tgz"))

			Expect(releaseWriter.WriteCallCount()).To(Equal(1))
			sha2ifyRelease, _ := releaseWriter.WriteArgsForCall(0)

			Expect(sha2ifyRelease).NotTo(BeNil())

			Expect(sha2ifyRelease.License()).ToNot(BeNil())
			Expect(sha2ifyRelease.License().ArchiveDigest()).To(Equal("sha256:licsha256"))

			Expect(sha2ifyRelease.Jobs()).To(HaveLen(1))
			Expect(sha2ifyRelease.Jobs()[0].ArchiveDigest()).To(Equal("sha256:jobsha256"))

			Expect(sha2ifyRelease.Packages()).To(HaveLen(1))
			Expect(sha2ifyRelease.Packages()[0].ArchiveDigest()).To(Equal("sha256:pkgsha256"))

			Expect(sha2ifyRelease.CompiledPackages()).To(HaveLen(1))
			Expect(sha2ifyRelease.CompiledPackages()[0].ArchiveDigest()).To(Equal("sha256:compiledpkgsha256"))

			Expect(fmv.MoveCallCount()).To(Equal(1))

			src, dst := fmv.MoveArgsForCall(0)
			Expect(src).To(Equal(releaseWriterTempDestination))
			Expect(dst).To(Equal(args.Destination.ExpandedPath))

			Expect(ui.Tables[0]).To(Equal(boshtbl.Table{
				Header: []boshtbl.Header{
					boshtbl.NewHeader("Name"),
					boshtbl.NewHeader("Version"),
					boshtbl.NewHeader("Commit Hash"),
					boshtbl.NewHeader("Archive"),
				},
				Rows: [][]boshtbl.Value{
					{
						boshtbl.NewValueString("custom-name"),
						boshtbl.NewValueString("custom-ver"),
						boshtbl.NewValueString("commit"),
						boshtbl.NewValueString("/some/release_256.tgz"),
					},
				},
				Transpose: true,
			}))
		})

		Context("when unable to write the sha256 tarball", func() {
			BeforeEach(func() {
				releaseWriter.WriteReturns("", errors.Error("disaster"))
			})

			It("should return an error", func() {
				err := command.Run(args)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("disaster"))
			})
		})

		Context("when rehashing a licence fails", func() {
			BeforeEach(func() {
				fakeDigestCalculator.SetCalculateBehavior(map[string]fakecrypto.CalculateInput{
					job1ResourcePath:             {DigestStr: "sha256:jobsha256"},
					pkg1ResourcePath:             {DigestStr: "sha256:pkgsha256"},
					compiledPackage1ResourcePath: {DigestStr: "sha256:compiledpkgsha256"},
					licenseResourcePath:          {Err: errors.Error("Unknown algorithm")},
				})
			})

			It("should return an error", func() {
				err := command.Run(args)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unknown algorithm"))
			})
		})

		Context("when rehashing compiled packages fails", func() {
			BeforeEach(func() {
				fakeDigestCalculator.SetCalculateBehavior(map[string]fakecrypto.CalculateInput{
					job1ResourcePath:             {DigestStr: "sha256:jobsha256"},
					pkg1ResourcePath:             {DigestStr: "sha256:pkgsha256"},
					compiledPackage1ResourcePath: {Err: errors.Error("Unknown algorithm")},
					licenseResourcePath:          {DigestStr: "sha256:licsha256"},
				})
			})

			It("should return an error", func() {
				err := command.Run(args)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unknown algorithm"))
			})
		})

		Context("when rehashing packages fails", func() {
			BeforeEach(func() {
				fakeDigestCalculator.SetCalculateBehavior(map[string]fakecrypto.CalculateInput{
					job1ResourcePath:             {DigestStr: "sha256:jobsha256"},
					pkg1ResourcePath:             {Err: errors.Error("Unknown algorithm")},
					compiledPackage1ResourcePath: {DigestStr: "sha256:compiledpkgsha256"},
					licenseResourcePath:          {DigestStr: "sha256:licsha256"},
				})
			})

			It("should return an error", func() {
				err := command.Run(args)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unknown algorithm"))
			})
		})

		Context("when rehashing jobs fails", func() {
			BeforeEach(func() {
				fakeDigestCalculator.SetCalculateBehavior(map[string]fakecrypto.CalculateInput{
					job1ResourcePath:             {Err: errors.Error("Unknown algorithm")},
					pkg1ResourcePath:             {DigestStr: "sha256:pkgsha256"},
					compiledPackage1ResourcePath: {DigestStr: "sha256:compiledpkgsha256"},
					licenseResourcePath:          {DigestStr: "sha256:licsha256"},
				})
			})

			It("should return an error", func() {
				err := command.Run(args)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unknown algorithm"))
			})
		})

		Context("when no licence is provided", func() {
			BeforeEach(func() {
				fakeSha128Release.LicenseReturns(nil)
				fakeDigestCalculator.SetCalculateBehavior(map[string]fakecrypto.CalculateInput{
					job1ResourcePath:             {DigestStr: "sha256:jobsha256"},
					pkg1ResourcePath:             {DigestStr: "sha256:pkgsha256"},
					compiledPackage1ResourcePath: {DigestStr: "sha256:compiledpkgsha256"},
				})
			})

			It("should not return an error", func() {
				err := command.Run(args)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("When unable to move sha2fyied release to destination", func() {
			BeforeEach(func() {
				fmv.MoveReturns(errors.Error("disaster"))
			})

			It("Should return an error", func() {
				err := command.Run(args)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("disaster"))
			})
		})
	})

	Context("Given an invalid sha128 release tar", func() {
		Context("Given a job that does not verify", func() {
			BeforeEach(func() {
				err := fs.WriteFileString(job1ResourcePath, "content that does not match expected sha1")
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return an error", func() {
				err := command.Run(args)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Expected stream to have digest"))

			})
		})

		Context("Given a package that does not verify", func() {
			BeforeEach(func() {
				err := fs.WriteFileString(pkg1ResourcePath, "content that does not match expected sha1")
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return an error", func() {
				err := command.Run(args)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Expected stream to have digest"))

			})
		})

		Context("Given a compiled package that does not verify", func() {
			BeforeEach(func() {
				err := fs.WriteFileString(compiledPackage1ResourcePath, "content that does not match expected sha1")
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return an error", func() {
				err := command.Run(args)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Expected stream to have digest"))

			})
		})

		Context("Given a license that does not verify", func() {
			BeforeEach(func() {
				err := fs.WriteFileString(licenseResourcePath, "content that does not match expected sha1")
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return an error", func() {
				err := command.Run(args)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Expected stream to have digest"))

			})
		})
	})

	Context("Given a bad file path", func() {
		BeforeEach(func() {
			args = opts.RedigestReleaseArgs{
				Path:        "/some/release_128.tgz",
				Destination: opts.FileArg{ExpandedPath: "/some/release_256.tgz"},
			}

			releaseReader.ReadReturns(nil, errors.Error("disaster"))
		})

		It("Should return an error", func() {
			err := command.Run(args)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("disaster"))
		})
	})
})
