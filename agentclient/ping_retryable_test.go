package agentclient_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakebmagentclient "github.com/cloudfoundry/bosh-micro-cli/agentclient/fakes"
	bmretrystrategy "github.com/cloudfoundry/bosh-micro-cli/retrystrategy"

	. "github.com/cloudfoundry/bosh-micro-cli/agentclient"
)

var _ = Describe("PingRetryable", func() {
	Describe("Attempt", func() {
		var (
			fakeAgentClient *fakebmagentclient.FakeAgentClient
			pingRetryable   bmretrystrategy.Retryable
		)

		BeforeEach(func() {
			fakeAgentClient = fakebmagentclient.NewFakeAgentClient()
			pingRetryable = NewPingRetryable(fakeAgentClient)
		})

		It("tells the agent client to ping", func() {
			err := pingRetryable.Attempt()
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeAgentClient.PingCalledCount).To(Equal(1))
		})

		Context("when pinging fails", func() {
			BeforeEach(func() {
				fakeAgentClient.SetPingBehavior("", errors.New("fake-agent-client-ping-error"))
			})

			It("returns an error", func() {
				err := pingRetryable.Attempt()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-agent-client-ping-error"))
			})
		})
	})
})
