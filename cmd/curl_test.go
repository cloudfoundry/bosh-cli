package cmd_test

import (
	"net/http"
	"strings"

	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
)

var _ = Describe("CurlCmd", func() {
	var (
		ui      *fakeui.FakeUI
		server  *ghttp.Server
		command CurlCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		server = ghttp.NewServer()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		command = NewCurlCmd(ui, boshdir.NewClientRequest(
			server.URL(),
			boshhttp.NewHTTPClient(boshhttp.CreateDefaultClient(nil), logger),
			boshdir.NewNoopFileReporter(),
			logger,
		))
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Run", func() {
		var (
			opts CurlOpts
		)

		BeforeEach(func() {
			opts = CurlOpts{}
		})

		act := func() error { return command.Run(opts) }

		Describe("GET requests", func() {
			BeforeEach(func() {
				opts.Method = "GET"
			})

			It("does not return error and prints response body", func() {
				opts.Args.Path = "/path?query=query-val"
				opts.Headers = []CurlHeader{
					{Name: "Header1", Value: "header1-val"},
					{Name: "Header2", Value: "header2-val1"},
					{Name: "Header2", Value: "header2-val2"},
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/path", "query=query-val"),
						ghttp.VerifyBody([]byte{}),
						ghttp.VerifyHeader(http.Header{"Header1": []string{"header1-val"}}),
						ghttp.VerifyHeader(http.Header{"Header2": []string{"header2-val1", "header2-val2"}}),
						ghttp.RespondWith(http.StatusOK, "resp-body"),
					),
				)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Blocks).To(Equal([]string{"resp-body"}))
			})

			It("returns error if client request considers response as failure", func() {
				opts.Args.Path = "/path"

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/path"),
						ghttp.RespondWith(http.StatusInternalServerError, `{"code":12345,"description":"Some Error"}`),
					),
				)

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(
					`Executing HTTP request: Director responded with non-successful status code '500' response '{"code":12345,"description":"Some Error"}'`))

				Expect(ui.Blocks).To(Equal([]string{`{"code":12345,"description":"Some Error"}`}))
			})

			It("shows response headers if requested", func() {
				opts.Args.Path = "/path"
				opts.ShowHeaders = true

				respHeaders := http.Header{}
				respHeaders.Add("Date", "date") // dont want date to change
				respHeaders.Add("Header1", "header1-val")
				respHeaders.Add("Header2", "header2-val1")
				respHeaders.Add("Header2", "header2-val2")

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/path"),
						ghttp.RespondWith(http.StatusOK, "resp-body", respHeaders),
					),
				)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Blocks).To(Equal([]string{
					strings.Join([]string{
						"HTTP/1.1 200 OK\r\n",
						"Content-Length: 9\r\n",
						"Content-Type: text/plain; charset=utf-8\r\n",
						"Date: date\r\n",
						"Header1: header1-val\r\n",
						"Header2: header2-val1\r\n",
						"Header2: header2-val2\r\n",
						"\r\n",
					}, ""),
					"resp-body",
				}))
			})
		})

		Describe("POST requests", func() {
			BeforeEach(func() {
				opts.Method = "POST"
			})

			It("does not return error and prints response body", func() {
				opts.Args.Path = "/path?query=query-val"
				opts.Headers = []CurlHeader{
					{Name: "Header1", Value: "header1-val"},
					{Name: "Header2", Value: "header2-val1"},
					{Name: "Header2", Value: "header2-val2"},
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/path", "query=query-val"),
						ghttp.VerifyBody([]byte{}),
						ghttp.VerifyHeader(http.Header{"Header1": []string{"header1-val"}}),
						ghttp.VerifyHeader(http.Header{"Header2": []string{"header2-val1", "header2-val2"}}),
						ghttp.RespondWith(http.StatusOK, "resp-body"),
					),
				)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Blocks).To(Equal([]string{"resp-body"}))
			})

			It("accepts request body", func() {
				opts.Args.Path = "/path?query=query-val"
				opts.Body = FileBytesArg{
					Bytes: []byte("req-body"),
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/path"),
						ghttp.VerifyBody([]byte("req-body")),
						ghttp.RespondWith(http.StatusOK, "resp-body"),
					),
				)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Blocks).To(Equal([]string{"resp-body"}))
			})

			It("returns error if client request considers response as failure", func() {
				opts.Args.Path = "/path"

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/path"),
						ghttp.RespondWith(http.StatusInternalServerError, `{"code":12345,"description":"Some Error"}`),
					),
				)

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(
					`Executing HTTP request: Director responded with non-successful status code '500' response '{"code":12345,"description":"Some Error"}'`))

				Expect(ui.Blocks).To(Equal([]string{`{"code":12345,"description":"Some Error"}`}))
			})

			It("shows response headers if requested", func() {
				opts.Args.Path = "/path"
				opts.ShowHeaders = true

				respHeaders := http.Header{}
				respHeaders.Add("Date", "date") // dont want date to change
				respHeaders.Add("Header1", "header1-val")
				respHeaders.Add("Header2", "header2-val1")
				respHeaders.Add("Header2", "header2-val2")

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/path"),
						ghttp.RespondWith(http.StatusOK, "resp-body", respHeaders),
					),
				)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Blocks).To(Equal([]string{
					strings.Join([]string{
						"HTTP/1.1 200 OK\r\n",
						"Content-Length: 9\r\n",
						"Content-Type: text/plain; charset=utf-8\r\n",
						"Date: date\r\n",
						"Header1: header1-val\r\n",
						"Header2: header2-val1\r\n",
						"Header2: header2-val2\r\n",
						"\r\n",
					}, ""),
					"resp-body",
				}))
			})
		})

		Describe("PUT requests", func() {
			BeforeEach(func() {
				opts.Method = "PUT"
			})

			It("does not return error and prints response body", func() {
				opts.Args.Path = "/path?query=query-val"
				opts.Headers = []CurlHeader{
					{Name: "Header1", Value: "header1-val"},
					{Name: "Header2", Value: "header2-val1"},
					{Name: "Header2", Value: "header2-val2"},
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/path", "query=query-val"),
						ghttp.VerifyBody([]byte{}),
						ghttp.VerifyHeader(http.Header{"Header1": []string{"header1-val"}}),
						ghttp.VerifyHeader(http.Header{"Header2": []string{"header2-val1", "header2-val2"}}),
						ghttp.RespondWith(http.StatusOK, "resp-body"),
					),
				)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Blocks).To(Equal([]string{"resp-body"}))
			})

			It("accepts request body", func() {
				opts.Args.Path = "/path?query=query-val"
				opts.Body = FileBytesArg{
					Bytes: []byte("req-body"),
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/path"),
						ghttp.VerifyBody([]byte("req-body")),
						ghttp.RespondWith(http.StatusOK, "resp-body"),
					),
				)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Blocks).To(Equal([]string{"resp-body"}))
			})

			It("returns error if client request considers response as failure", func() {
				opts.Args.Path = "/path"

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/path"),
						ghttp.RespondWith(http.StatusInternalServerError, `{"code":12345,"description":"Some Error"}`),
					),
				)

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(
					`Executing HTTP request: Director responded with non-successful status code '500' response '{"code":12345,"description":"Some Error"}'`))

				Expect(ui.Blocks).To(Equal([]string{`{"code":12345,"description":"Some Error"}`}))
			})

			It("shows response headers if requested", func() {
				opts.Args.Path = "/path"
				opts.ShowHeaders = true

				respHeaders := http.Header{}
				respHeaders.Add("Date", "date") // dont want date to change
				respHeaders.Add("Header1", "header1-val")
				respHeaders.Add("Header2", "header2-val1")
				respHeaders.Add("Header2", "header2-val2")

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/path"),
						ghttp.RespondWith(http.StatusOK, "resp-body", respHeaders),
					),
				)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Blocks).To(Equal([]string{
					strings.Join([]string{
						"HTTP/1.1 200 OK\r\n",
						"Content-Length: 9\r\n",
						"Content-Type: text/plain; charset=utf-8\r\n",
						"Date: date\r\n",
						"Header1: header1-val\r\n",
						"Header2: header2-val1\r\n",
						"Header2: header2-val2\r\n",
						"\r\n",
					}, ""),
					"resp-body",
				}))
			})
		})

		Describe("DELETE requests", func() {
			BeforeEach(func() {
				opts.Method = "DELETE"
			})

			It("does not return error and prints response body", func() {
				opts.Args.Path = "/path?query=query-val"

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/path", "query=query-val"),
						ghttp.VerifyBody([]byte{}),
						ghttp.RespondWith(http.StatusOK, "resp-body"),
					),
				)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Blocks).To(Equal([]string{"resp-body"}))
			})

			It("returns error if any headers are provided (currently no supported)", func() {
				opts.Args.Path = "/path"
				opts.Headers = []CurlHeader{
					{Name: "Header1", Value: "header1-val"},
				}

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Expected no headers"))

				Expect(server.ReceivedRequests()).To(BeEmpty())

				Expect(ui.Blocks).To(BeEmpty())
			})

			It("returns error if client request considers response as failure", func() {
				opts.Args.Path = "/path"

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/path"),
						ghttp.RespondWith(http.StatusInternalServerError, `{"code":12345,"description":"Some Error"}`),
					),
				)

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(
					`Executing HTTP request: Director responded with non-successful status code '500' response '{"code":12345,"description":"Some Error"}'`))

				Expect(ui.Blocks).To(Equal([]string{`{"code":12345,"description":"Some Error"}`}))
			})

			It("shows response headers if requested", func() {
				opts.Args.Path = "/path"
				opts.ShowHeaders = true

				respHeaders := http.Header{}
				respHeaders.Add("Date", "date") // dont want date to change
				respHeaders.Add("Header1", "header1-val")
				respHeaders.Add("Header2", "header2-val1")
				respHeaders.Add("Header2", "header2-val2")

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/path"),
						ghttp.RespondWith(http.StatusOK, "resp-body", respHeaders),
					),
				)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Blocks).To(Equal([]string{
					strings.Join([]string{
						"HTTP/1.1 200 OK\r\n",
						"Content-Length: 9\r\n",
						"Content-Type: text/plain; charset=utf-8\r\n",
						"Date: date\r\n",
						"Header1: header1-val\r\n",
						"Header2: header2-val1\r\n",
						"Header2: header2-val2\r\n",
						"\r\n",
					}, ""),
					"resp-body",
				}))
			})
		})

		Describe("unknown method requests", func() {
			BeforeEach(func() {
				opts.Method = "UNKNOWN"
			})

			It("returns error", func() {
				opts.Args.Path = "/path"

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Unknown method 'UNKNOWN'"))

				Expect(server.ReceivedRequests()).To(BeEmpty())

				Expect(ui.Blocks).To(BeEmpty())
			})
		})
	})
})
