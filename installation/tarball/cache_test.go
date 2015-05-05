package tarball_test

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/installation/tarball"
)

type tarballSource struct {
	url         string
	sha1        string
	description string
}

func (ts tarballSource) GetURL() string {
	return ts.url
}

func (ts tarballSource) GetSHA1() string {
	return ts.sha1
}

func (ts tarballSource) Description() string {
	return ts.description
}

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
				path, found := cache.Get(tarballSource{sha1: "fake-sha1"})
				Expect(path).To(Equal("/fake-base-path/fake-sha1"))
				Expect(found).To(BeTrue())
			})
		})

		Context("when cached tarball does not exist", func() {
			It("returns not found", func() {
				path, found := cache.Get(tarballSource{sha1: "non-existent-fake-sha1"})
				Expect(path).To(Equal(""))
				Expect(found).To(BeFalse())
			})
		})
	})

	Describe("Save", func() {
		BeforeEach(func() {
			fs.WriteFileString("source-path", "")
		})

		It("saves the tarball", func() {
			err := cache.Save("source-path", tarballSource{sha1: "fake-sha1"})
			Expect(err).ToNot(HaveOccurred())
			Expect(fs.FileExists("/fake-base-path/fake-sha1")).To(BeTrue())
		})

		Context("when saving tarball fails", func() {
			It("returns error", func() {
				err := cache.Save("nonexistent-source-path", tarballSource{sha1: "fake-sha1"})
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
