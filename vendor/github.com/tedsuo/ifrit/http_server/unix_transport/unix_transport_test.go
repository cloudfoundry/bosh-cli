package unix_transport

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Unix transport", func() {

	var (
		socket string
		client http.Client
	)

	Context("with server listening", func() {

		var (
			unixSocketListener net.Listener
			unixSocketServer   *ghttp.Server
			resp               *http.Response
			err                error
		)

		BeforeEach(func() {
			uuid, err := uuid.NewV4()
			Expect(err).NotTo(HaveOccurred())

			socket = fmt.Sprintf("/tmp/%s.sock", uuid)
			unixSocketListener, err = net.Listen("unix", socket)
			Expect(err).NotTo(HaveOccurred())

			unixSocketServer = ghttp.NewUnstartedServer()

			unixSocketServer.HTTPTestServer = &httptest.Server{
				Listener: unixSocketListener,
				Config:   &http.Server{Handler: unixSocketServer},
			}
			unixSocketServer.Start()

			client = http.Client{Transport: New(socket)}
		})

		Context("when a simple GET request is sent", func() {
			BeforeEach(func() {
				unixSocketServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/_ping"),
						ghttp.RespondWith(http.StatusOK, "true"),
					),
				)

				resp, err = client.Get("unix://" + socket + "/_ping")
			})

			It("responds with correct status", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

			})

			It("responds with correct body", func() {
				bytes, err := ioutil.ReadAll(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(bytes)).To(Equal("true"))
			})
		})

		Context("when a POST request is sent", func() {
			const (
				ReqBody  = `"id":"some-id"`
				RespBody = `{"Image" : "ubuntu"}`
			)

			assertBodyEquals := func(body io.ReadCloser, expectedContent string) {
				bytes, err := ioutil.ReadAll(body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(bytes)).To(Equal(expectedContent))

			}

			asserHeaderContains := func(header http.Header, key, value string) {
				Expect(header[key]).To(ConsistOf(value))
			}

			BeforeEach(func() {
				validateBody := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					assertBodyEquals(req.Body, ReqBody)
				})

				validateQueryParams := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					Expect(req.URL.RawQuery).To(Equal("fromImage=ubunut&tag=latest"))
				})

				handleRequest := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(RespBody))
				})

				unixSocketServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/containers/create"),
						ghttp.VerifyContentType("application/json"),
						validateBody,
						validateQueryParams,
						handleRequest,
					),
				)
				body := strings.NewReader(ReqBody)
				req, err := http.NewRequest("POST", "unix://"+socket+"/containers/create?fromImage=ubunut&tag=latest", body)
				req.Header.Add("Content-Type", "application/json")
				Expect(err).NotTo(HaveOccurred())

				resp, err = client.Do(req)
				Expect(err).NotTo(HaveOccurred())

			})

			It("responds with correct status", func() {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})

			It("responds with correct headers", func() {
				asserHeaderContains(resp.Header, "Content-Type", "application/json")
			})

			It("responds with correct body", func() {
				assertBodyEquals(resp.Body, RespBody)
			})

		})

		Context("when socket in reques URI is incorrect", func() {
			It("errors", func() {
				resp, err = client.Get("unix:///fake/socket.sock/_ping")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Wrong unix socket"))
			})
		})

		AfterEach(func() {
			unixSocketServer.Close()
		})
	})

	Context("with no server listening", func() {
		BeforeEach(func() {
			socket = "/not/existing.sock"
			client = http.Client{Transport: New(socket)}
		})

		It("errors", func() {
			_, err := client.Get("unix:///not/existing.sock/_ping")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Or(
				ContainSubstring(fmt.Sprintf("dial unix %s: connect: no such file or directory", socket)),
				ContainSubstring(fmt.Sprintf("dial unix %s: no such file or directory", socket)),
			))
		})
	})
})
