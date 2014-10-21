package agentclient_test

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"

	. "github.com/cloudfoundry/bosh-micro-cli/agentclient"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

var _ = Describe("AgentClient", func() {
	var (
		agentClient AgentClient
		agentServer *agentServer
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		agentClient = NewAgentClient("http://localhost:6305", "fake-uuid", logger)
		agentServer = NewAgentServer("localhost:6305")

		readyCh := make(chan error)
		go agentServer.Start(readyCh)
		err := <-readyCh
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		agentServer.Stop()
	})

	Describe("Ping", func() {
		Context("when agent responds with a value", func() {
			BeforeEach(func() {
				agentServer.SetResponseBody(`{"value":"pong"}`)
			})

			It("makes a POST request to the endpoint", func() {
				_, err := agentClient.Ping()
				Expect(err).ToNot(HaveOccurred())

				Expect(len(agentServer.ReceivedRequests)).To(Equal(1))
				receivedRequest := agentServer.ReceivedRequests[0]

				Expect(receivedRequest.Method).To(Equal("POST"))

				var request receivedRequestBody
				err = json.Unmarshal(receivedRequest.Body, &request)
				Expect(err).ToNot(HaveOccurred())

				Expect(request).To(Equal(receivedRequestBody{
					Method:    "ping",
					Arguments: nil,
					ReplyTo:   "fake-uuid",
				}))
			})

			It("returns the value", func() {
				responseValue, err := agentClient.Ping()
				Expect(err).ToNot(HaveOccurred())
				Expect(responseValue).To(Equal("pong"))
			})
		})

		Context("when agent does not respond with 200", func() {
			BeforeEach(func() {
				agentServer.SetResponseStatus(http.StatusInternalServerError)
			})

			It("returns an error", func() {
				_, err := agentClient.Ping()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when agent does not respond with response value", func() {
			BeforeEach(func() {
				agentServer.SetResponseBody(`{"exception":{"message":"bad request"}}`)
			})

			It("returns an error", func() {
				_, err := agentClient.Ping()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("bad request"))
			})
		})
	})
})

type receivedRequestBody struct {
	Method    string
	Arguments []string
	ReplyTo   string `json:"reply_to"`
}

type receivedRequest struct {
	Body   []byte
	Method string
}

type agentServer struct {
	listener         net.Listener
	endpoint         string
	ReceivedRequests []receivedRequest
	responseBody     string
	responseStatus   int
}

func NewAgentServer(endpoint string) *agentServer {
	return &agentServer{
		endpoint:         endpoint,
		responseStatus:   http.StatusOK,
		ReceivedRequests: []receivedRequest{},
	}
}

func (s *agentServer) Start(readyErrCh chan error) {
	var err error
	s.listener, err = net.Listen("tcp", s.endpoint)
	if err != nil {
		readyErrCh <- err
		return
	}

	readyErrCh <- nil

	fakeServer := http.Server{}
	fakeServer.SetKeepAlivesEnabled(false)
	mux := http.NewServeMux()
	fakeServer.Handler = mux

	mux.HandleFunc("/agent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(s.responseStatus)
		w.Write([]byte(s.responseBody))
		requestBody, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		receivedRequest := receivedRequest{
			Body:   requestBody,
			Method: r.Method,
		}
		s.ReceivedRequests = append(s.ReceivedRequests, receivedRequest)
	})

	fakeServer.Serve(s.listener)
}

func (s *agentServer) Stop() {
	s.listener.Close()
}

func (s *agentServer) SetResponseStatus(code int) {
	s.responseStatus = code
}

func (s *agentServer) SetResponseBody(body string) {
	s.responseBody = body
}
