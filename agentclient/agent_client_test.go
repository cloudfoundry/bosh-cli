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
		agentClient = NewAgentClient("http://localhost:6305", "fake-uuid", 0, logger)
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
					Arguments: []interface{}{},
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

		Context("when agent responds with exception", func() {
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

	Describe("Stop", func() {
		Context("when agent responds with a value", func() {
			BeforeEach(func() {
				agentServer.SetResponseBody(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`)
				agentServer.SetTaskStates([]string{"running", "running", "stoppped"})
			})

			It("makes a POST request to the endpoint", func() {
				err := agentClient.Stop()
				Expect(err).ToNot(HaveOccurred())

				receivedStopRequest := agentServer.ReceivedRequests[0]
				Expect(receivedStopRequest.Method).To(Equal("POST"))

				var request receivedRequestBody
				err = json.Unmarshal(receivedStopRequest.Body, &request)
				Expect(err).ToNot(HaveOccurred())

				Expect(request).To(Equal(receivedRequestBody{
					Method:    "stop",
					Arguments: []interface{}{},
					ReplyTo:   "fake-uuid",
				}))
			})

			It("waits for the task to be finished", func() {
				err := agentClient.Stop()
				Expect(err).ToNot(HaveOccurred())

				Expect(len(agentServer.ReceivedRequests)).To(Equal(4))
				for _, receivedGetTaskRequest := range agentServer.ReceivedRequests[1:] {
					Expect(receivedGetTaskRequest.Method).To(Equal("POST"))

					var request receivedRequestBody
					err = json.Unmarshal(receivedGetTaskRequest.Body, &request)
					Expect(err).ToNot(HaveOccurred())

					Expect(request).To(Equal(receivedRequestBody{
						Method:    "get_task",
						Arguments: []interface{}{"fake-agent-task-id"},
						ReplyTo:   "fake-uuid",
					}))
				}
			})
		})

		Context("when agent does not respond with 200", func() {
			BeforeEach(func() {
				agentServer.SetResponseStatus(http.StatusInternalServerError)
			})

			It("returns an error", func() {
				err := agentClient.Stop()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when agent responds with exception", func() {
			BeforeEach(func() {
				agentServer.SetResponseBody(`{"exception":{"message":"bad request"}}`)
			})

			It("returns an error", func() {
				err := agentClient.Stop()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("bad request"))
			})
		})
	})

	Describe("Apply", func() {
		var (
			specJSON []byte
			spec     ApplySpec
		)

		BeforeEach(func() {
			spec = ApplySpec{
				Deployment: "fake-deployment-name",
			}
			var err error
			specJSON, err = json.Marshal(spec)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when agent responds with a value", func() {
			BeforeEach(func() {
				agentServer.SetResponseBody(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`)
				agentServer.SetTaskStates([]string{"running", "running", "stopped"})
			})

			It("makes a POST request to the endpoint", func() {
				err := agentClient.Apply(spec)
				Expect(err).ToNot(HaveOccurred())

				receivedApplyRequest := agentServer.ReceivedRequests[0]
				Expect(receivedApplyRequest.Method).To(Equal("POST"))

				var request receivedRequestBody
				err = json.Unmarshal(receivedApplyRequest.Body, &request)
				Expect(err).ToNot(HaveOccurred())

				var specArgument interface{}
				err = json.Unmarshal(specJSON, &specArgument)
				Expect(err).ToNot(HaveOccurred())

				Expect(request).To(Equal(receivedRequestBody{
					Method:    "apply",
					Arguments: []interface{}{specArgument},
					ReplyTo:   "fake-uuid",
				}))
			})

			It("waits for the task to be finished", func() {
				err := agentClient.Apply(spec)
				Expect(err).ToNot(HaveOccurred())

				Expect(len(agentServer.ReceivedRequests)).To(Equal(4))
				for _, receivedGetTaskRequest := range agentServer.ReceivedRequests[1:] {
					Expect(receivedGetTaskRequest.Method).To(Equal("POST"))

					var request receivedRequestBody
					err = json.Unmarshal(receivedGetTaskRequest.Body, &request)
					Expect(err).ToNot(HaveOccurred())

					Expect(request).To(Equal(receivedRequestBody{
						Method:    "get_task",
						Arguments: []interface{}{"fake-agent-task-id"},
						ReplyTo:   "fake-uuid",
					}))
				}
			})
		})

		Context("when agent does not respond with 200", func() {
			BeforeEach(func() {
				agentServer.SetResponseStatus(http.StatusInternalServerError)
			})

			It("returns an error", func() {
				err := agentClient.Apply(spec)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when agent responds with exception to get_task", func() {
			BeforeEach(func() {
				agentServer.SetResponseBody(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`)
				agentServer.SetTaskResponseBody(`{"exception":{"message":"bad request"}}`)
			})

			It("stops polling for task state", func() {
				err := agentClient.Apply(spec)
				Expect(err).To(HaveOccurred())

				Expect(agentServer.GetTaskRequests).To(Equal(1))
			})
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

type agentServer struct {
	listener         net.Listener
	endpoint         string
	ReceivedRequests []receivedRequest
	responseBody     string
	responseStatus   int
	taskResponseBody string
	taskStates       []string
	GetTaskRequests  int
}

func NewAgentServer(endpoint string) *agentServer {
	return &agentServer{
		endpoint:         endpoint,
		responseStatus:   http.StatusOK,
		ReceivedRequests: []receivedRequest{},
	}
}

type taskStateResponse struct {
	Value string
}

type runningTaskStateResponse struct {
	Value agentTaskState
}

type agentTaskState struct {
	AgentTaskID string `json:"agent_task_id"`
	State       string `json:"state"`
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

		requestBody, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		receivedRequest := receivedRequest{
			Body:   requestBody,
			Method: r.Method,
		}

		s.ReceivedRequests = append(s.ReceivedRequests, receivedRequest)

		var agentRequest AgentRequest
		json.Unmarshal(requestBody, &agentRequest)
		if agentRequest.Method == "get_task" {
			s.GetTaskRequests++
			if s.taskResponseBody != "" {
				w.Write([]byte(s.taskResponseBody))
				return
			}

			if len(s.taskStates) > 0 {
				state := s.taskStates[0]
				s.taskStates = s.taskStates[1:]
				var responseBody []byte
				if state == "running" {
					responseBody, _ = json.Marshal(runningTaskStateResponse{
						Value: agentTaskState{
							State: "running",
						},
					})
				} else {
					responseBody, _ = json.Marshal(taskStateResponse{
						Value: state,
					})
				}

				w.Write([]byte(responseBody))
			}
		} else {
			w.Write([]byte(s.responseBody))
		}
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

func (s *agentServer) SetTaskStates(taskStates []string) {
	s.taskStates = taskStates
}

func (s *agentServer) SetTaskResponseBody(taskResponseBody string) {
	s.taskResponseBody = taskResponseBody
}
