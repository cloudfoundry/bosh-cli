package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshreldir "github.com/cloudfoundry/bosh-cli/v7/releasedir"
	fakereldir "github.com/cloudfoundry/bosh-cli/v7/releasedir/releasedirfakes"
)

var _ = Describe("InitReleaseCmd", func() {
	var (
		releaseDir *fakereldir.FakeReleaseDir
		command    cmd.InitReleaseCmd
	)

	BeforeEach(func() {
		releaseDir = &fakereldir.FakeReleaseDir{}
		command = cmd.NewInitReleaseCmd(releaseDir)
	})

	Describe("Run", func() {
		var (
			initReleaseOpts opts.InitReleaseOpts
		)

		BeforeEach(func() {
			initReleaseOpts = opts.InitReleaseOpts{}
		})

		act := func() error { return command.Run(initReleaseOpts) }

		It("inits release", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(releaseDir.InitCallCount()).To(Equal(1))
			Expect(releaseDir.InitArgsForCall(0)).To(Equal(boshreldir.InitOptions{}))
		})

		It("passes init options through", func() {
			initReleaseOpts.Args.Name = "my-release"
			initReleaseOpts.Git = true
			initReleaseOpts.Sample = true
			initReleaseOpts.Force = true
			initReleaseOpts.Blobstore = "s3"
			initReleaseOpts.BlobstoreOptions = []string{"bucket_name=my-bucket"}
			initReleaseOpts.BlobstorePrivateOptions = []string{"access_key_id=key"}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(releaseDir.InitArgsForCall(0)).To(Equal(boshreldir.InitOptions{
				Name:                    "my-release",
				Git:                     true,
				Sample:                  true,
				Force:                   true,
				BlobstoreProvider:       "s3",
				BlobstoreOptions:        map[string]interface{}{"bucket_name": "my-bucket"},
				BlobstorePrivateOptions: map[string]interface{}{"access_key_id": "key"},
			}))
		})

		It("returns error if blobstore options are given without a provider", func() {
			initReleaseOpts.BlobstoreOptions = []string{"bucket_name=my-bucket"}

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("without --blobstore"))
			Expect(releaseDir.InitCallCount()).To(Equal(0))
		})

		It("allows blobstore options when generating a sample", func() {
			initReleaseOpts.Sample = true
			initReleaseOpts.BlobstoreOptions = []string{"blobstore_path=/tmp/blobs"}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(releaseDir.InitArgsForCall(0).BlobstoreOptions).To(Equal(
				map[string]interface{}{"blobstore_path": "/tmp/blobs"}))
		})

		It("types boolean and numeric blobstore option values", func() {
			initReleaseOpts.Blobstore = "s3"
			initReleaseOpts.BlobstoreOptions = []string{
				"use_ssl=true",
				"port=443",
				"ratio=1.5",
				"bucket_name=123abc",
			}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(releaseDir.InitArgsForCall(0).BlobstoreOptions).To(Equal(map[string]interface{}{
				"use_ssl":     true,
				"port":        int64(443),
				"ratio":       1.5,
				"bucket_name": "123abc",
			}))
		})

		It("returns error if a blobstore option is malformed", func() {
			initReleaseOpts.Blobstore = "s3"
			initReleaseOpts.BlobstoreOptions = []string{"not-a-key-value"}

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected KEY=VALUE option"))
		})

		It("returns error if initing release fails", func() {
			releaseDir.InitReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
