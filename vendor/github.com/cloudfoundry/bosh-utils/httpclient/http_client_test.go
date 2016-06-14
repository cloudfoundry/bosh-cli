package httpclient_test

import (
	"io/ioutil"
	"net"
	"net/http"

	. "github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HttpClient", func() {
	var (
		httpClient HTTPClient
		fakeServer *fakeServer
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		httpClient = NewHTTPClient(DefaultClient, logger)
		fakeServer = newFakeServer("localhost:0")

		readyCh := make(chan error)
		go fakeServer.Start(readyCh)
		err := <-readyCh
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		fakeServer.Stop()
	})

	Describe("DefaultClient", func() {
		It("is a singleton http client", func() {
			var client http.Client
			client = DefaultClient

			Expect(client).To(Equal(DefaultClient))
		})

		It("disables keep alive", func() {
			var client http.Client
			client = DefaultClient

			Expect(client.Transport.(*http.Transport).DisableKeepAlives).To(Equal(true))
		})
	})

	Describe("CreateDefaultClient", func() {
		It("creates a new http client", func() {
			var (
				first  http.Client
				second http.Client
			)

			first = CreateDefaultClient()
			second = CreateDefaultClient()

			Expect(first).ToNot(Equal(second))
		})
	})

	Describe("Post", func() {
		It("makes a post request with given payload", func() {
			fakeServer.SetResponseBody("fake-post-response")
			fakeServer.SetResponseStatus(200)

			response, err := httpClient.Post("http://"+fakeServer.Listener.Addr().String()+"/fake-path", []byte("fake-post-request"))
			Expect(err).ToNot(HaveOccurred())

			defer response.Body.Close()
			responseBody, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())

			Expect(responseBody).To(Equal([]byte("fake-post-response")))
			Expect(response.StatusCode).To(Equal(200))

			Expect(fakeServer.ReceivedRequests).To(HaveLen(1))
			Expect(fakeServer.ReceivedRequests).To(ContainElement(
				receivedRequest{
					Body:   []byte("fake-post-request"),
					Method: "POST",
				},
			))
		})
	})

	Describe("Get", func() {
		It("makes a get request with given payload", func() {
			fakeServer.SetResponseBody("fake-get-response")
			fakeServer.SetResponseStatus(200)

			response, err := httpClient.Get("http://" + fakeServer.Listener.Addr().String() + "/fake-path")
			Expect(err).ToNot(HaveOccurred())

			defer response.Body.Close()
			responseBody, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())

			Expect(responseBody).To(Equal([]byte("fake-get-response")))
			Expect(response.StatusCode).To(Equal(200))

			Expect(fakeServer.ReceivedRequests).To(HaveLen(1))
			Expect(fakeServer.ReceivedRequests).To(ContainElement(
				receivedRequest{
					Body:   []byte(""),
					Method: "GET",
				},
			))
		})
	})
})

type receivedRequestBody struct {
	Method    string
	Arguments []interface{}
	ReplyTo   string `json:"reply_to"`
}

type receivedRequest struct {
	Body   []byte
	Method string
}

type fakeServer struct {
	Listener         net.Listener
	endpoint         string
	ReceivedRequests []receivedRequest
	responseBody     string
	responseStatus   int
}

func newFakeServer(endpoint string) *fakeServer {
	return &fakeServer{
		endpoint:         endpoint,
		responseStatus:   http.StatusOK,
		ReceivedRequests: []receivedRequest{},
	}
}

func (s *fakeServer) Start(readyErrCh chan error) {
	var err error
	s.Listener, err = net.Listen("tcp", s.endpoint)
	if err != nil {
		readyErrCh <- err
		return
	}

	readyErrCh <- nil

	httpServer := http.Server{}
	httpServer.SetKeepAlivesEnabled(false)
	mux := http.NewServeMux()
	httpServer.Handler = mux

	mux.HandleFunc("/fake-path", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(s.responseStatus)

		requestBody, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		receivedRequest := receivedRequest{
			Body:   requestBody,
			Method: r.Method,
		}

		s.ReceivedRequests = append(s.ReceivedRequests, receivedRequest)
		w.Write([]byte(s.responseBody))
	})

	httpServer.Serve(s.Listener)
}

func (s *fakeServer) Stop() {
	s.Listener.Close()
}

func (s *fakeServer) SetResponseStatus(code int) {
	s.responseStatus = code
}

func (s *fakeServer) SetResponseBody(body string) {
	s.responseBody = body
}
