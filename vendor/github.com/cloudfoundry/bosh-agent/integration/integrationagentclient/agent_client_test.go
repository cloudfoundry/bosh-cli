package integrationagentclient_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-agent/agent/action"
	agentclienthttp "github.com/cloudfoundry/bosh-agent/agentclient/http"
	"github.com/cloudfoundry/bosh-agent/integration/integrationagentclient"
	fakehttpclient "github.com/cloudfoundry/bosh-utils/httpclient/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

var _ = Describe("AgentClient", func() {
	var (
		fakeHTTPClient *fakehttpclient.FakeHTTPClient
		agentClient    *integrationagentclient.IntegrationAgentClient

		agentAddress        string
		replyToAddress      string
		toleratedErrorCount int
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fakeHTTPClient = fakehttpclient.NewFakeHTTPClient()

		agentAddress = "http://localhost:6305"
		replyToAddress = "fake-reply-to-uuid"

		getTaskDelay := time.Duration(0)
		toleratedErrorCount = 2

		agentClient = integrationagentclient.NewIntegrationAgentClient(agentAddress, replyToAddress, getTaskDelay, toleratedErrorCount, fakeHTTPClient, logger)
	})

	Describe("SSH", func() {
		Context("when agent successfully executes ssh", func() {
			BeforeEach(func() {
				sshSuccess, err := json.Marshal(action.SSHResult{
					Command: "setup",
					Status:  "success",
				})
				Expect(err).ToNot(HaveOccurred())
				fakeHTTPClient.SetPostBehavior(string(sshSuccess), 200, nil)
			})

			It("makes a POST request to the endpoint", func() {
				params := action.SSHParams{
					User: "username",
				}

				err := agentClient.SSH("setup", params)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeHTTPClient.PostInputs).To(HaveLen(1))
				Expect(fakeHTTPClient.PostInputs[0].Endpoint).To(Equal("http://localhost:6305/agent"))

				var request agentclienthttp.AgentRequestMessage
				err = json.Unmarshal(fakeHTTPClient.PostInputs[0].Payload, &request)
				Expect(err).ToNot(HaveOccurred())

				Expect(request).To(Equal(agentclienthttp.AgentRequestMessage{
					Method:    "ssh",
					Arguments: []interface{}{"setup", map[string]interface{}{"user_regex": "", "User": "username", "public_key": ""}},
					ReplyTo:   "fake-reply-to-uuid",
				}))
			})
		})

		Context("when POST to agent returns error", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior("", http.StatusInternalServerError, errors.New("foo error"))
			})

			It("returns an error that wraps original error", func() {
				params := action.SSHParams{
					User: "username",
				}

				err := agentClient.SSH("setup", params)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Performing request to agent"))
				Expect(err.Error()).To(ContainSubstring("foo error"))
			})
		})
	})
})
