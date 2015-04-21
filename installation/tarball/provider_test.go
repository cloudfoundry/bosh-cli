package tarball_test

import (
	"errors"
	"io/ioutil"
	"os"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebicrypto "github.com/cloudfoundry/bosh-init/crypto/fakes"
	fakebihttpclient "github.com/cloudfoundry/bosh-init/deployment/httpclient/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/installation/tarball"
)

var _ = Describe("Provider", func() {
	var (
		provider       Provider
		cache          Cache
		fs             *fakesys.FakeFileSystem
		httpClient     *fakebihttpclient.FakeHTTPClient
		sha1Calculator *fakebicrypto.FakeSha1Calculator
		source         *fakeSource
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		cache = NewCache("/fake-base-path", fs, logger)
		sha1Calculator = fakebicrypto.NewFakeSha1Calculator()
		httpClient = fakebihttpclient.NewFakeHTTPClient()
		provider = NewProvider(cache, fs, httpClient, sha1Calculator, logger)
	})

	Describe("Get", func() {
		Context("when URL starts with file://", func() {
			BeforeEach(func() {
				source = newFakeSource("file://fake-file", "fake-sha1")
				fs.WriteFileString("expanded-file-path", "")
				fs.ExpandPathExpanded = "expanded-file-path"
			})

			It("returns expanded path to file", func() {
				path, err := provider.Get(source)
				Expect(err).ToNot(HaveOccurred())
				Expect(path).To(Equal("expanded-file-path"))
			})

			Context("when file does not exist", func() {
				BeforeEach(func() {
					fs.RemoveAll("expanded-file-path")
				})

				It("returns an error", func() {
					_, err := provider.Get(source)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("File path 'expanded-file-path' does not exist"))
				})
			})
		})

		Context("when URL starts with http(s)://", func() {
			BeforeEach(func() {
				source = newFakeSource("http://fake-url", "fake-sha1")
			})

			Context("when tarball is present in cache", func() {
				BeforeEach(func() {
					fs.WriteFileString("fake-source-path", "")
					cache.Save("fake-source-path", "fake-sha1")
				})

				It("returns cached tarball path", func() {
					path, err := provider.Get(source)
					Expect(err).ToNot(HaveOccurred())
					Expect(path).To(Equal("/fake-base-path/fake-sha1"))
				})
			})

			Context("when tarball is not present in cache", func() {
				var (
					tempDownloadFilePath string
				)

				BeforeEach(func() {
					tempDownloadFile, err := ioutil.TempFile("", "temp-download-file")
					Expect(err).ToNot(HaveOccurred())
					fs.ReturnTempFile = tempDownloadFile
					tempDownloadFilePath = tempDownloadFile.Name()
					sha1Calculator.SetCalculateBehavior(map[string]fakebicrypto.CalculateInput{
						tempDownloadFilePath: {Sha1: "fake-sha1"},
					})
				})

				AfterEach(func() {
					os.RemoveAll(tempDownloadFilePath)
				})

				Context("when downloading succeds", func() {
					BeforeEach(func() {
						httpClient.SetGetBehavior("fake-body", 200, nil)
					})

					It("downloads tarball from given URL and returns saved cache tarball path", func() {
						path, err := provider.Get(source)
						Expect(err).ToNot(HaveOccurred())
						Expect(path).To(Equal("/fake-base-path/fake-sha1"))

						Expect(httpClient.GetInputs).To(HaveLen(1))
						Expect(httpClient.GetInputs[0].Endpoint).To(Equal("http://fake-url"))
					})

					Context("when sha1 does not match", func() {
						BeforeEach(func() {
							sha1Calculator.SetCalculateBehavior(map[string]fakebicrypto.CalculateInput{
								tempDownloadFilePath: {Sha1: "fake-sha2"},
							})
						})

						It("returns an error", func() {
							_, err := provider.Get(source)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("'fake-sha2' does not match source SHA1 'fake-sha1'"))
						})

						It("removes the downloaded file", func() {
							_, err := provider.Get(source)
							Expect(err).To(HaveOccurred())
							Expect(fs.FileExists(tempDownloadFilePath)).To(BeFalse())
						})
					})

					Context("when saving to cache fails", func() {
						BeforeEach(func() {
							// Creating cache base directory fails
							fs.MkdirAllError = errors.New("fake-mkdir-error")
						})

						It("returns an error", func() {
							_, err := provider.Get(source)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("fake-mkdir-error"))
						})

						It("removes the downloaded file", func() {
							_, err := provider.Get(source)
							Expect(err).To(HaveOccurred())
							Expect(fs.FileExists(tempDownloadFilePath)).To(BeFalse())
						})
					})
				})

				Context("when downloading fails", func() {
					BeforeEach(func() {
						httpClient.SetGetBehavior("", 500, errors.New("fake-download-error"))
					})

					It("returns an error", func() {
						_, err := provider.Get(source)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-download-error"))
					})

					It("removes the downloaded file", func() {
						_, err := provider.Get(source)
						Expect(err).To(HaveOccurred())
						Expect(fs.FileExists(tempDownloadFilePath)).To(BeFalse())
					})
				})
			})
		})

		Context("when URL does not start with either file:// or http(s)://", func() {
			BeforeEach(func() {
				source = newFakeSource("invalid-url", "fake-sha1")
			})

			It("returns an error", func() {
				_, err := provider.Get(source)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Invalid source URL: 'invalid-url'"))
			})
		})
	})
})

type fakeSource struct {
	url  string
	sha1 string
}

func newFakeSource(url, sha1 string) *fakeSource {
	return &fakeSource{url, sha1}
}

func (s *fakeSource) GetURL() string  { return s.url }
func (s *fakeSource) GetSHA1() string { return s.sha1 }
