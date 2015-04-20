package tarball_test

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/installation/tarball"
)

var _ = Describe("Cache", func() {
	var (
		cache Cache
		fs    *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		cache = NewCache(
			"/fake-base-path",
			fs,
			logger,
		)
	})

	Describe("Get", func() {
		Context("when cached tarball exists", func() {
			BeforeEach(func() {
				fs.WriteFileString("/fake-base-path/fake-sha1", "")
			})

			It("returns path to tarball", func() {
				path, found := cache.Get("fake-sha1")
				Expect(path).To(Equal("/fake-base-path/fake-sha1"))
				Expect(found).To(BeTrue())
			})
		})

		Context("when cached tarball does not exist", func() {
			It("returns not found", func() {
				path, found := cache.Get("non-existent-fake-sha1")
				Expect(path).To(Equal(""))
				Expect(found).To(BeFalse())
			})
		})
	})

	Describe("Save", func() {
		Context("when saving tarball succeeds", func() {
			BeforeEach(func() {
				fs.WriteFileString("source-path", "")
			})

			It("returns path to tarball", func() {
				path, err := cache.Save("source-path", "fake-sha1")
				Expect(err).ToNot(HaveOccurred())
				Expect(path).To(Equal("/fake-base-path/fake-sha1"))
			})
		})

		Context("when saving tarball fails", func() {
			It("returns error", func() {
				_, err := cache.Save("source-path", "fake-sha1")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
