package agentclient_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakebmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployer/agentclient/fakes"
	bmretrystrategy "github.com/cloudfoundry/bosh-micro-cli/retrystrategy"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer/agentclient"
)

var _ = Describe("GetStateRetryable", func() {
	Describe("Attempt", func() {
		var (
			fakeAgentClient   *fakebmagentclient.FakeAgentClient
			getStateRetryable bmretrystrategy.Retryable
		)

		BeforeEach(func() {
			fakeAgentClient = fakebmagentclient.NewFakeAgentClient()
			getStateRetryable = NewGetStateRetryable(fakeAgentClient)
		})

		Context("when get_state fails", func() {
			BeforeEach(func() {
				fakeAgentClient.SetGetStateBehavior(State{}, errors.New("fake-get-state-error"))
			})

			It("returns an error", func() {
				isRetryable, err := getStateRetryable.Attempt()
				Expect(err).To(HaveOccurred())
				Expect(isRetryable).To(BeFalse())
				Expect(err.Error()).To(ContainSubstring("fake-get-state-error"))
				Expect(fakeAgentClient.GetStateCalledTimes).To(Equal(1))
			})
		})

		Context("when get_state returns state as pending", func() {
			BeforeEach(func() {
				fakeAgentClient.SetGetStateBehavior(State{JobState: "pending"}, nil)
			})

			It("returns retryable error", func() {
				isRetryable, err := getStateRetryable.Attempt()
				Expect(err).To(HaveOccurred())
				Expect(isRetryable).To(BeTrue())
				Expect(fakeAgentClient.GetStateCalledTimes).To(Equal(1))
			})
		})

		Context("when get_state returns state as running", func() {
			BeforeEach(func() {
				fakeAgentClient.SetGetStateBehavior(State{JobState: "running"}, nil)
			})

			It("returns no error", func() {
				isRetryable, err := getStateRetryable.Attempt()
				Expect(err).ToNot(HaveOccurred())
				Expect(isRetryable).To(BeTrue())
				Expect(fakeAgentClient.GetStateCalledTimes).To(Equal(1))
			})
		})
	})
})
