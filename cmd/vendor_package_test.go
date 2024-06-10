package cmd_test

import (
	"errors"

	semver "github.com/cppforlife/go-semi-semantic/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
	fakerel "github.com/cloudfoundry/bosh-cli/v7/release/releasefakes"
	. "github.com/cloudfoundry/bosh-cli/v7/release/resource"
	boshreldir "github.com/cloudfoundry/bosh-cli/v7/releasedir"
	fakereldir "github.com/cloudfoundry/bosh-cli/v7/releasedir/releasedirfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("VendorPackageCmd", func() {
	var (
		srcReleaseDir *fakereldir.FakeReleaseDir
		dstReleaseDir *fakereldir.FakeReleaseDir
		ui            *fakeui.FakeUI
		command       cmd.VendorPackageCmd
	)

	BeforeEach(func() {
		srcReleaseDir = &fakereldir.FakeReleaseDir{}
		dstReleaseDir = &fakereldir.FakeReleaseDir{}

		releaseDirFactory := func(dir opts.DirOrCWDArg) boshreldir.ReleaseDir {
			switch dir {
			case opts.DirOrCWDArg{Path: "/src-dir"}:
				return srcReleaseDir
			case opts.DirOrCWDArg{Path: "/dst-dir"}:
				return dstReleaseDir
			default:
				panic("Unexpected release dir")
			}
		}

		ui = &fakeui.FakeUI{}
		command = cmd.NewVendorPackageCmd(releaseDirFactory, ui)
	})

	Describe("Run", func() {
		var (
			vendorPackageOpts opts.VendorPackageOpts
		)

		BeforeEach(func() {
			vendorPackageOpts = opts.VendorPackageOpts{
				Args: opts.VendorPackageArgs{
					PackageName: "pkg1-name",
					URL:         opts.DirOrCWDArg{Path: "/src-dir"},
				},
				Directory: opts.DirOrCWDArg{Path: "/dst-dir"},
			}
		})

		act := func() error { return command.Run(vendorPackageOpts) }

		It("vendors package by name from source release", func() {
			pkg0 := boshpkg.NewPackage(NewResourceWithBuiltArchive(
				"pkg0-name", "pkg0-fp", "pkg0-path", "pkg0-sha1"), nil)
			pkg1 := boshpkg.NewPackage(NewResourceWithBuiltArchive(
				"pkg1-name", "pkg1-fp", "pkg1-path", "pkg1-sha1"), nil)

			srcRelease := &fakerel.FakeRelease{}
			srcRelease.PackagesReturns([]*boshpkg.Package{pkg0, pkg1})

			srcReleaseDir.FindReleaseReturns(srcRelease, nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			name, ver := srcReleaseDir.FindReleaseArgsForCall(0)
			Expect(name).To(Equal(""))
			Expect(ver).To(Equal(semver.Version{}))

			Expect(dstReleaseDir.VendorPackageCallCount()).To(Equal(1))
			Expect(dstReleaseDir.VendorPackageArgsForCall(0)).To(Equal(pkg1))
		})

		It("returns error if vendoring fails", func() {
			pkg1 := boshpkg.NewPackage(NewResourceWithBuiltArchive(
				"pkg1-name", "pkg1-fp", "pkg1-path", "pkg1-sha1"), nil)

			srcRelease := &fakerel.FakeRelease{}
			srcRelease.PackagesReturns([]*boshpkg.Package{pkg1})

			srcReleaseDir.FindReleaseReturns(srcRelease, nil)
			dstReleaseDir.VendorPackageReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("fake-err"))

			Expect(dstReleaseDir.VendorPackageCallCount()).To(Equal(1))
		})

		It("returns error if package does not exist within source release", func() {
			pkg1 := boshpkg.NewPackage(NewResourceWithBuiltArchive(
				"pkg1-other-name", "pkg1-fp", "pkg1-path", "pkg1-sha1"), nil)

			srcRelease := &fakerel.FakeRelease{}
			srcRelease.PackagesReturns([]*boshpkg.Package{pkg1})

			srcReleaseDir.FindReleaseReturns(srcRelease, nil)

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected to find package 'pkg1-name'"))
		})

		It("returns error if finding release fails", func() {
			srcReleaseDir.FindReleaseReturns(nil, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
