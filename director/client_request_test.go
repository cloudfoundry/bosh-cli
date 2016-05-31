package director_test

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	. "github.com/cloudfoundry/bosh-init/director"
	fakedir "github.com/cloudfoundry/bosh-init/director/fakes"
)

var _ = Describe("ClientRequest", func() {
	var (
		server *ghttp.Server
		resp   []string

		buildReq func(FileReporter) ClientRequest
		req      ClientRequest

		locationHeader http.Header
	)

	BeforeEach(func() {
		_, server = BuildServer()

		buildReq = func(fileReporter FileReporter) ClientRequest {
			httpTransport := &http.Transport{
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
				TLSHandshakeTimeout: 10 * time.Second,
			}

			rawClient := &http.Client{Transport: httpTransport}
			logger := boshlog.NewLogger(boshlog.LevelNone)
			httpClient := boshhttp.NewHTTPClient(rawClient, logger)
			return NewClientRequest(server.URL(), httpClient, fileReporter, logger)
		}

		resp = nil
		req = buildReq(NewNoopFileReporter())

		locationHeader = http.Header{}
		locationHeader.Add("Location", "/redirect")
	})

	AfterEach(func() {
		server.Close()
	})

	successCodes := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusPartialContent,
	}

	Describe("Get", func() {
		act := func() error { return req.Get("/path", &resp) }

		for _, code := range successCodes {
			code := code

			Describe(fmt.Sprintf("'%d' response", code), func() {
				It("makes request, succeeds and unmarshals response", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/path", ""),
							ghttp.RespondWith(code, `["val"]`),
						),
					)

					err := act()
					Expect(err).ToNot(HaveOccurred())
					Expect(resp).To(Equal([]string{"val"}))
				})

				It("returns error if cannot be unmarshalled", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/path"),
							ghttp.RespondWith(code, ""),
						),
					)

					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Unmarshaling Director response"))
				})
			})
		}

		Describe("'302' response", func() {
			It("makes request, follows redirect, succeeds and unmarshals response", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/path", ""),
						ghttp.RespondWith(http.StatusFound, "", locationHeader),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/redirect"),
						ghttp.RespondWith(http.StatusOK, `["val"]`),
					),
				)

				err := act()
				Expect(err).ToNot(HaveOccurred())
				Expect(resp).To(Equal([]string{"val"}))
			})

			It("returns error if redirect response cannot be unmarshalled", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/path"),
						ghttp.RespondWith(http.StatusFound, "", locationHeader),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/redirect"),
						ghttp.RespondWith(http.StatusOK, `-`),
					),
				)

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unmarshaling Director response"))
			})
		})

		It("returns error if response in non-successful response code", func() {
			AppendBadRequest(ghttp.VerifyRequest("GET", "/path"), server)

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Director responded with non-successful status code"))
		})
	})

	Describe("RawGet", func() {
		Context("when custom writer is not set", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/path"),
						ghttp.RespondWith(http.StatusOK, "body"),
					),
				)
			})

			It("returns full response body", func() {
				body, resp, err := req.RawGet("/path", nil, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(body).To(Equal([]byte("body")))
				Expect(resp).ToNot(BeNil())
			})

			It("does not track downloading", func() {
				fileReporter := &fakedir.FakeFileReporter{}
				req = buildReq(fileReporter)

				_, _, err := req.RawGet("/path", nil, nil)
				Expect(err).ToNot(HaveOccurred())

				Expect(fileReporter.TrackDownloadCallCount()).To(Equal(0))
			})
		})

		Context("when custom writer is set", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/path"),
						ghttp.RespondWith(http.StatusOK, "body"),
					),
				)
			})

			It("returns response body", func() {
				buf := bytes.NewBufferString("")

				body, resp, err := req.RawGet("/path", buf, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(body).To(BeEmpty())
				Expect(resp).ToNot(BeNil())

				Expect(buf.String()).To(Equal("body"))
			})

			It("tracks downloading based on content length", func() {
				buf := bytes.NewBufferString("")
				otherBuf := bytes.NewBufferString("")

				fileReporter := &fakedir.FakeFileReporter{
					TrackDownloadStub: func(size int64, out io.Writer) io.Writer {
						Expect(size).To(Equal(int64(4)))
						Expect(out).To(Equal(buf))
						return otherBuf
					},
				}

				req = buildReq(fileReporter)

				_, _, err := req.RawGet("/path", buf, nil)
				Expect(err).ToNot(HaveOccurred())

				Expect(otherBuf.String()).To(Equal("body"))
			})
		})
	})

	Describe("Post", func() {
		act := func() error { return req.Post("/path", []byte("req-body"), nil, &resp) }

		for _, code := range successCodes {
			code := code

			Describe(fmt.Sprintf("'%d' response", code), func() {
				It("makes request, succeeds and unmarshals response", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/path", ""),
							ghttp.VerifyBody([]byte("req-body")),
							ghttp.RespondWith(code, `["val"]`),
						),
					)

					err := act()
					Expect(err).ToNot(HaveOccurred())
					Expect(resp).To(Equal([]string{"val"}))
				})

				It("returns error if cannot be unmarshalled", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/path"),
							ghttp.VerifyBody([]byte("req-body")),
							ghttp.RespondWith(code, ""),
						),
					)

					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Unmarshaling Director response"))
				})
			})
		}

		Describe("'302' response", func() {
			It("makes request, follows redirect, succeeds and unmarshals response", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/path", ""),
						ghttp.VerifyBody([]byte("req-body")),
						ghttp.RespondWith(http.StatusFound, "", locationHeader),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/redirect"),
						ghttp.RespondWith(http.StatusOK, `["val"]`),
					),
				)

				err := act()
				Expect(err).ToNot(HaveOccurred())
				Expect(resp).To(Equal([]string{"val"}))
			})

			It("returns error if redirect response cannot be unmarshalled", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/path"),
						ghttp.RespondWith(http.StatusFound, "", locationHeader),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/redirect"),
						ghttp.RespondWith(http.StatusOK, `-`),
					),
				)

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unmarshaling Director response"))
			})
		})

		It("returns error if response in non-successful response code", func() {
			AppendBadRequest(ghttp.VerifyRequest("POST", "/path"), server)

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Director responded with non-successful status code"))
		})
	})

	Describe("RawPost", func() {
		Context("when request body is 'application/x-compressed'", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/path"),
						ghttp.VerifyBody([]byte("req-body")),
						ghttp.VerifyHeader(http.Header{"Content-Type": []string{"application/x-compressed"}}),
						ghttp.RespondWith(http.StatusOK, "body"),
					),
				)
			})

			setHeaders := func(req *http.Request) {
				req.Header.Add("Content-Type", "application/x-compressed")
				req.Body = ioutil.NopCloser(bytes.NewBufferString("req-body"))
				req.ContentLength = 8
			}

			It("uploads request body and returns response", func() {
				body, resp, err := req.RawPost("/path", nil, setHeaders)
				Expect(err).ToNot(HaveOccurred())
				Expect(body).To(Equal([]byte("body")))
				Expect(resp).ToNot(BeNil())
			})

			It("tracks uploading", func() {
				fileReporter := &fakedir.FakeFileReporter{
					TrackUploadStub: func(size int64, reader io.ReadCloser) io.ReadCloser {
						Expect(size).To(Equal(int64(8)))
						Expect(ioutil.ReadAll(reader)).To(Equal([]byte("req-body")))
						return ioutil.NopCloser(bytes.NewBufferString("req-body"))
					},
				}
				req = buildReq(fileReporter)

				_, _, err := req.RawPost("/path", nil, setHeaders)
				Expect(err).ToNot(HaveOccurred())

				Expect(fileReporter.TrackUploadCallCount()).To(Equal(1))
			})
		})

		Context("when request body is not 'application/x-compressed'", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/path"),
						ghttp.VerifyBody([]byte("req-body")),
						ghttp.VerifyHeader(http.Header{"Content-Type": []string{"application/json"}}),
						ghttp.RespondWith(http.StatusOK, "body"),
					),
				)
			})

			setHeaders := func(req *http.Request) {
				req.Header.Add("Content-Type", "application/json")
				req.Body = ioutil.NopCloser(bytes.NewBufferString("req-body"))
				req.ContentLength = 8
			}

			It("uploads request body and returns response", func() {
				body, resp, err := req.RawPost("/path", nil, setHeaders)
				Expect(err).ToNot(HaveOccurred())
				Expect(body).To(Equal([]byte("body")))
				Expect(resp).ToNot(BeNil())
			})

			It("does not track uploading", func() {
				fileReporter := &fakedir.FakeFileReporter{}
				req = buildReq(fileReporter)

				_, _, err := req.RawPost("/path", nil, setHeaders)
				Expect(err).ToNot(HaveOccurred())

				Expect(fileReporter.TrackUploadCallCount()).To(Equal(0))
			})
		})
	})

	Describe("Put", func() {
		act := func() error { return req.Put("/path", []byte("req-body"), nil, &resp) }

		for _, code := range successCodes {
			code := code

			Describe(fmt.Sprintf("'%d' response", code), func() {
				It("makes request, succeeds and unmarshals response", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/path", ""),
							ghttp.VerifyBody([]byte("req-body")),
							ghttp.RespondWith(code, `["val"]`),
						),
					)

					err := act()
					Expect(err).ToNot(HaveOccurred())
					Expect(resp).To(Equal([]string{"val"}))
				})

				It("returns error if cannot be unmarshalled", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PUT", "/path"),
							ghttp.VerifyBody([]byte("req-body")),
							ghttp.RespondWith(code, ""),
						),
					)

					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Unmarshaling Director response"))
				})
			})
		}

		Describe("'302' response", func() {
			It("makes request, follows redirect, succeeds and unmarshals response", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/path", ""),
						ghttp.VerifyBody([]byte("req-body")),
						ghttp.RespondWith(http.StatusFound, "", locationHeader),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/redirect"),
						ghttp.RespondWith(http.StatusOK, `["val"]`),
					),
				)

				err := act()
				Expect(err).ToNot(HaveOccurred())
				Expect(resp).To(Equal([]string{"val"}))
			})

			It("returns error if redirect response cannot be unmarshalled", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/path"),
						ghttp.RespondWith(http.StatusFound, "", locationHeader),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/redirect"),
						ghttp.RespondWith(http.StatusOK, `-`),
					),
				)

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unmarshaling Director response"))
			})
		})

		It("returns error if response in non-successful response code", func() {
			AppendBadRequest(ghttp.VerifyRequest("PUT", "/path"), server)

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Director responded with non-successful status code"))
		})
	})

	Describe("Delete", func() {
		act := func() error { return req.Delete("/path", &resp) }

		for _, code := range successCodes {
			code := code

			Describe(fmt.Sprintf("'%d' response", code), func() {
				It("makes request, succeeds and unmarshals response", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("DELETE", "/path", ""),
							ghttp.VerifyBody([]byte("")),
							ghttp.RespondWith(code, `["val"]`),
						),
					)

					err := act()
					Expect(err).ToNot(HaveOccurred())
					Expect(resp).To(Equal([]string{"val"}))
				})

				It("returns error if cannot be unmarshalled", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("DELETE", "/path"),
							ghttp.VerifyBody([]byte("")),
							ghttp.RespondWith(code, ""),
						),
					)

					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Unmarshaling Director response"))
				})
			})
		}

		Describe("'302' response", func() {
			It("makes request, follows redirect, succeeds and unmarshals response", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/path", ""),
						ghttp.VerifyBody([]byte("")),
						ghttp.RespondWith(http.StatusFound, "", locationHeader),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/redirect"),
						ghttp.RespondWith(http.StatusOK, `["val"]`),
					),
				)

				err := act()
				Expect(err).ToNot(HaveOccurred())
				Expect(resp).To(Equal([]string{"val"}))
			})

			It("returns error if redirect response cannot be unmarshalled", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/path"),
						ghttp.RespondWith(http.StatusFound, "", locationHeader),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/redirect"),
						ghttp.RespondWith(http.StatusOK, `-`),
					),
				)

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unmarshaling Director response"))
			})
		})

		It("returns error if response in non-successful response code", func() {
			AppendBadRequest(ghttp.VerifyRequest("DELETE", "/path"), server)

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Director responded with non-successful status code"))
		})
	})
})
