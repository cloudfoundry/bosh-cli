package agentclient_test

import (
	"errors"

	. "github.com/cloudfoundry/bosh-agent/agentclient"
	fakeagentclient "github.com/cloudfoundry/bosh-agent/agentclient/fakes"
	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"crypto/x509"
)

var _ = Describe("PingRetryable", func() {
	Describe("Attempt", func() {
		var (
			fakeAgentClient *fakeagentclient.FakeAgentClient
			pingRetryable   boshretry.Retryable
		)

		BeforeEach(func() {
			fakeAgentClient = &fakeagentclient.FakeAgentClient{}
			pingRetryable = NewPingRetryable(fakeAgentClient)
		})

		It("tells the agent client to ping", func() {
			isRetryable, err := pingRetryable.Attempt()
			Expect(err).ToNot(HaveOccurred())
			Expect(isRetryable).To(BeTrue())
			Expect(fakeAgentClient.PingCallCount()).To(Equal(1))
		})

		Context("when pinging fails", func() {
			BeforeEach(func() {
				fakeAgentClient.PingReturns("", errors.New("fake-agent-client-ping-error"))
			})

			It("returns an error", func() {
				isRetryable, err := pingRetryable.Attempt()
				Expect(err).To(HaveOccurred())
				Expect(isRetryable).To(BeTrue())
				Expect(err.Error()).To(ContainSubstring("fake-agent-client-ping-error"))
			})
		})

		Context("when failing with a certificate error", func() {
			BeforeEach(func() {
				fakeAgentClient.PingReturns("", x509.CertificateInvalidError{})
			})

			It("returns stops retrying and returns the error", func() {
				isRetryable, err := pingRetryable.Attempt()
				Expect(err).To(HaveOccurred())
				Expect(isRetryable).To(BeFalse())
				Expect(err.Error()).To(ContainSubstring("x509: certificate is not authorized to sign other certificates"))
			})

			Context("when the certificate error is wrapped", func() {
				BeforeEach(func() {
					certError := x509.CertificateInvalidError{}
					wrappedError := bosherr.WrapError(certError, "nope")
					doubleWrappedError := bosherr.WrapError(wrappedError, "nope nope")
					fakeAgentClient.PingReturns("", doubleWrappedError)
				})

				It("stops retrying and returns the error", func() {
					isRetryable, err := pingRetryable.Attempt()
					Expect(err).To(HaveOccurred())
					Expect(isRetryable).To(BeFalse())
					Expect(err.Error()).To(ContainSubstring("x509: certificate is not authorized to sign other certificates"))
				})
			})
		})
	})
})
