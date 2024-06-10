package cmd_test

import (
	"errors"

	semver "github.com/cppforlife/go-semi-semantic/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	fakecmd "github.com/cloudfoundry/bosh-cli/v7/cmd/cmdfakes"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
	fakerel "github.com/cloudfoundry/bosh-cli/v7/release/releasefakes"
)

var _ = Describe("ReleaseManager", func() {
	var (
		createReleaseCmd *fakecmd.FakeReleaseCreatingCmd
		uploadReleaseCmd *fakecmd.FakeReleaseUploadingCmd
		releaseManager   cmd.ReleaseManager
	)

	BeforeEach(func() {
		createReleaseCmd = &fakecmd.FakeReleaseCreatingCmd{
			RunStub: func(createReleaseOpts opts.CreateReleaseOpts) (boshrel.Release, error) {
				release := &fakerel.FakeRelease{
					NameStub:    func() string { return createReleaseOpts.Name },
					VersionStub: func() string { return createReleaseOpts.Name + "-created-ver" },
				}
				return release, nil
			},
		}

		uploadReleaseCmd = &fakecmd.FakeReleaseUploadingCmd{}

		threadCount := 5
		releaseManager = cmd.NewReleaseManager(createReleaseCmd, uploadReleaseCmd, threadCount)
	})

	Describe("UploadReleases", func() {
		DescribeTable("uploads remote releases skipping releases without url", func(fix bool, subject func(bytes []byte) ([]byte, error)) {
			bytes := []byte(`
releases:
- name: capi
  sha1: capi-sha1
  url: https://capi-url
  version: 1+capi
- name: rel-without-upload
  version: 1+rel
- name: consul
  sha1: consul-sha1
  url: https://consul-url
  version: 1+consul
- name: compiled-release
  url: https://compiled-release-url
  sha1: compiled-release-sha1
  version: 1+compiled-release
  stemcell:
    os: ubuntu-trusty
    version: 3421
- name: local
  url: file:///local-dir
  version: create
- name: compiled-release-exported
  url: file:///compiled-release-url.tgz
  sha1: compiled-release-sha1
  version: 1+compiled-release
  exported_from:
  - os: ubuntu-trusty
    version: 3421
`)

			_, err := subject(bytes)
			Expect(err).ToNot(HaveOccurred())

			Expect(uploadReleaseCmd.RunCallCount()).To(Equal(5))
			runArgs := []opts.UploadReleaseOpts{
				uploadReleaseCmd.RunArgsForCall(0),
				uploadReleaseCmd.RunArgsForCall(1),
				uploadReleaseCmd.RunArgsForCall(2),
				uploadReleaseCmd.RunArgsForCall(3),
				uploadReleaseCmd.RunArgsForCall(4),
			}

			var capiRelease opts.UploadReleaseOpts
			var consulRelease opts.UploadReleaseOpts
			var compiledRelease opts.UploadReleaseOpts
			var localRelease opts.UploadReleaseOpts
			var compiledReleaseExported opts.UploadReleaseOpts
			for _, uploadReleaseOpts := range runArgs {
				switch uploadReleaseOpts.Name {
				case "capi":
					capiRelease = uploadReleaseOpts
				case "consul":
					consulRelease = uploadReleaseOpts
				case "compiled-release":
					compiledRelease = uploadReleaseOpts
				case "local":
					localRelease = uploadReleaseOpts
				case "compiled-release-exported":
					compiledReleaseExported = uploadReleaseOpts
				}
			}

			Expect(capiRelease).To(Equal(opts.UploadReleaseOpts{
				Name:    "capi",
				Args:    opts.UploadReleaseArgs{URL: opts.URLArg("https://capi-url")},
				SHA1:    "capi-sha1",
				Version: opts.VersionArg(semver.MustNewVersionFromString("1+capi")),
				Fix:     fix,
			}))
			Expect(consulRelease).To(Equal(opts.UploadReleaseOpts{
				Name:    "consul",
				Args:    opts.UploadReleaseArgs{URL: opts.URLArg("https://consul-url")},
				SHA1:    "consul-sha1",
				Version: opts.VersionArg(semver.MustNewVersionFromString("1+consul")),
				Fix:     fix,
			}))
			Expect(compiledRelease).To(Equal(opts.UploadReleaseOpts{
				Name:    "compiled-release",
				Args:    opts.UploadReleaseArgs{URL: opts.URLArg("https://compiled-release-url")},
				SHA1:    "compiled-release-sha1",
				Version: opts.VersionArg(semver.MustNewVersionFromString("1+compiled-release")),
				Fix:     fix,

				Stemcell: boshdir.NewOSVersionSlug("ubuntu-trusty", "3421"),
			}))
			Expect(localRelease).To(Equal(opts.UploadReleaseOpts{
				Release: localRelease.Release, // only Release should be set
			}))
			Expect(compiledReleaseExported).To(Equal(opts.UploadReleaseOpts{
				Name:     "compiled-release-exported",
				Args:     opts.UploadReleaseArgs{URL: opts.URLArg("file:///compiled-release-url.tgz")},
				SHA1:     "compiled-release-sha1",
				Version:  opts.VersionArg(semver.MustNewVersionFromString("1+compiled-release")),
				Fix:      fix,
				Stemcell: boshdir.NewOSVersionSlug("ubuntu-trusty", "3421"),
			}))
		},
			Entry("when without fix option", false, func(bytes []byte) ([]byte, error) {
				return releaseManager.UploadReleases(bytes)
			}),
			Entry("when with fix option", true, func(bytes []byte) ([]byte, error) {
				return releaseManager.UploadReleasesWithFix(bytes)
			}),
		)

		DescribeTable("skips uploading releases if url is not provided, even if the version is invalid", func(subject func(bytes []byte) ([]byte, error)) {
			bytes := []byte(`
releases:
- name: capi
  version: ((/blah_interpolate_me_with_config_server))
`)

			_, err := subject(bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(uploadReleaseCmd.RunCallCount()).To(Equal(0))
		},
			Entry("when without fix option", func(bytes []byte) ([]byte, error) {
				return releaseManager.UploadReleases(bytes)
			}),
			Entry("when with fix option", func(bytes []byte) ([]byte, error) {
				return releaseManager.UploadReleasesWithFix(bytes)
			}),
		)

		DescribeTable("creates releases if version is 'create' skipping others", func(subject func(bytes []byte) ([]byte, error)) {
			bytes := []byte(`
releases:
- name: capi
  url: file:///capi-dir
  version: create
- name: rel-without-upload
  version: 1+rel
- name: consul
  url: /consul-dir # doesn't require file://
  version: create
`)

			bytes, err := subject(bytes)
			Expect(err).ToNot(HaveOccurred())

			Expect(createReleaseCmd.RunCallCount()).To(Equal(2))
			runArgs := []opts.CreateReleaseOpts{
				createReleaseCmd.RunArgsForCall(0),
				createReleaseCmd.RunArgsForCall(1),
			}

			var capiRelease opts.CreateReleaseOpts
			var consulRelease opts.CreateReleaseOpts
			for _, createReleaseOpts := range runArgs {
				switch createReleaseOpts.Name {
				case "capi":
					capiRelease = createReleaseOpts
				case "consul":
					consulRelease = createReleaseOpts
				}
			}

			Expect(capiRelease).To(Equal(opts.CreateReleaseOpts{
				Name:             "capi",
				Directory:        opts.DirOrCWDArg{Path: "/capi-dir"},
				TimestampVersion: true,
				Force:            true,
			}))

			Expect(consulRelease).To(Equal(opts.CreateReleaseOpts{
				Name:             "consul",
				Directory:        opts.DirOrCWDArg{Path: "/consul-dir"},
				TimestampVersion: true,
				Force:            true,
			}))

			Expect(bytes).To(Equal([]byte(`releases:
- name: capi
  version: capi-created-ver
- name: rel-without-upload
  version: 1+rel
- name: consul
  version: consul-created-ver
`)))
		},
			Entry("when without fix option", func(bytes []byte) ([]byte, error) {
				return releaseManager.UploadReleases(bytes)
			}),
			Entry("when with fix option", func(bytes []byte) ([]byte, error) {
				return releaseManager.UploadReleasesWithFix(bytes)
			}),
		)

		DescribeTable("returns error and does not upload if creating release fails", func(subject func(bytes []byte) ([]byte, error)) {
			bytes := []byte(`
releases:
- name: capi
  url: file:///capi-dir
  version: create
`)
			createReleaseCmd.RunReturns(nil, errors.New("fake-err"))

			_, err := subject(bytes)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(uploadReleaseCmd.RunCallCount()).To(Equal(0))
		},
			Entry("when without fix option", func(bytes []byte) ([]byte, error) {
				return releaseManager.UploadReleases(bytes)
			}),
			Entry("when with fix option", func(bytes []byte) ([]byte, error) {
				return releaseManager.UploadReleasesWithFix(bytes)
			}),
		)

		DescribeTable("returns error if uploading release fails", func(subject func(bytes []byte) ([]byte, error)) {
			bytes := []byte(`
releases:
- name: capi
  sha1: capi-sha1
  url: https://capi-url
  version: 1+capi
`)
			uploadReleaseCmd.RunReturns(errors.New("fake-err"))

			_, err := subject(bytes)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		},
			Entry("when without fix", func(bytes []byte) ([]byte, error) {
				return releaseManager.UploadReleases(bytes)
			}),
			Entry("when fix", func(bytes []byte) ([]byte, error) {
				return releaseManager.UploadReleasesWithFix(bytes)
			}))

		DescribeTable("returns an error and does not upload if release version cannot be parsed", func(subject func(bytes []byte) ([]byte, error)) {
			bytes := []byte(`
releases:
- name: capi
  sha1: capi-sha1
  url: https://capi-url
  version: 1+capi+capi
`)

			_, err := subject(bytes)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected version '1+capi+capi' to match version format"))

			Expect(uploadReleaseCmd.RunCallCount()).To(Equal(0))
		},
			Entry("when without fix option", func(bytes []byte) ([]byte, error) {
				return releaseManager.UploadReleases(bytes)
			}),
			Entry("when with fix option", func(bytes []byte) ([]byte, error) {
				return releaseManager.UploadReleasesWithFix(bytes)
			}),
		)

		DescribeTable("returns an error if bytes cannot be parsed to find releases", func(subject func(bytes []byte) ([]byte, error)) {
			bytes := []byte(`-`)

			_, err := subject(bytes)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Parsing manifest"))

			Expect(createReleaseCmd.RunCallCount()).To(Equal(0))
			Expect(uploadReleaseCmd.RunCallCount()).To(Equal(0))
		},
			Entry("when without fix option", func(bytes []byte) ([]byte, error) {
				return releaseManager.UploadReleases(bytes)
			}),
			Entry("when with fix option", func(bytes []byte) ([]byte, error) {
				return releaseManager.UploadReleasesWithFix(bytes)
			}),
		)
	})
})
