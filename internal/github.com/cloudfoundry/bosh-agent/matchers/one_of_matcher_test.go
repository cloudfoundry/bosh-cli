package matchers_test

import (
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/matchers"

	"fmt"

	"github.com/cloudfoundry/bosh-init/internal/github.com/onsi/gomega/internal/fakematcher"
)

var _ = Describe("matchers", func() {

	var _ = Describe("Match", func() {

		Context("when no sub-matchers match", func() {
			var fakematcher1 = &fakematcher.FakeMatcher{
				MatchesToReturn: false,
				ErrToReturn:     nil,
			}
			var fakematcher2 = &fakematcher.FakeMatcher{
				MatchesToReturn: false,
				ErrToReturn:     nil,
			}
			var oneOf = MatchOneOf(fakematcher1, fakematcher2)

			It("calls Match on each sub-matcher", func() {
				success, err := oneOf.Match("Fake Test Value")

				Expect(success).To(BeFalse())
				Expect(err).ToNot(HaveOccurred())

				Expect(fakematcher1.ReceivedActual).To(Equal("Fake Test Value"))
				Expect(fakematcher2.ReceivedActual).To(Equal("Fake Test Value"))
			})
		})

		Context("when at least one sub-matcher matches", func() {
			var fakematcher1 = &fakematcher.FakeMatcher{
				MatchesToReturn: false,
				ErrToReturn:     nil,
			}
			var fakematcher2 = &fakematcher.FakeMatcher{
				MatchesToReturn: true,
				ErrToReturn:     nil,
			}
			var fakematcher3 = &fakematcher.FakeMatcher{
				MatchesToReturn: false,
				ErrToReturn:     nil,
			}
			var oneOf = MatchOneOf(fakematcher1, fakematcher2, fakematcher3)

			It("calls Match on each sub-matcher until a match is found", func() {
				success, err := oneOf.Match("Fake Test Value")

				Expect(success).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())

				Expect(fakematcher1.ReceivedActual).To(Equal("Fake Test Value"))
				Expect(fakematcher2.ReceivedActual).To(Equal("Fake Test Value"))
				Expect(fakematcher3.ReceivedActual).To(BeNil())
			})
		})

		Context("when at least one sub-matcher errors", func() {
			var error = fmt.Errorf("Fake Error")
			var fakematcher1 = &fakematcher.FakeMatcher{
				MatchesToReturn: false,
				ErrToReturn:     nil,
			}
			var fakematcher2 = &fakematcher.FakeMatcher{
				MatchesToReturn: false,
				ErrToReturn:     error,
			}
			var fakematcher3 = &fakematcher.FakeMatcher{
				MatchesToReturn: true,
				ErrToReturn:     nil,
			}
			var oneOf = MatchOneOf(fakematcher1, fakematcher2, fakematcher3)

			It("calls Match on each sub-matcher until an error is returned", func() {
				success, err := oneOf.Match("Fake Test Value")

				Expect(success).To(BeFalse())
				Expect(err).To(Equal(error))

				Expect(fakematcher1.ReceivedActual).To(Equal("Fake Test Value"))
				Expect(fakematcher2.ReceivedActual).To(Equal("Fake Test Value"))
				Expect(fakematcher3.ReceivedActual).To(BeNil())
			})
		})

		Context("when an element is not a matcher", func() {
			var oneOf = MatchOneOf("abc", 123, []string{"x", "y", "z"}, Equal("foo"))

			It("uses an Equal matcher", func() {
				success, err := oneOf.Match("abc")
				Expect(err).ToNot(HaveOccurred())
				Expect(success).To(BeTrue())

				success, err = oneOf.Match(123)
				Expect(err).ToNot(HaveOccurred())
				Expect(success).To(BeTrue())

				success, err = oneOf.Match([]string{"x", "y", "z"})
				Expect(err).ToNot(HaveOccurred())
				Expect(success).To(BeTrue())
			})

			It("matchers still work", func() {
				success, err := oneOf.Match("foo")
				Expect(err).ToNot(HaveOccurred())
				Expect(success).To(BeTrue())
			})
		})
	})

	var _ = Describe("FailureMessage", func() {
		var oneOf = MatchOneOf(Equal("a"), BeNumerically(">", 1))

		It("concatonates the failure message of all matchers", func() {
			msg := oneOf.FailureMessage("Fake Test Value")

			expectedMessagePattern := `Expected
		<string>: Fake Test Value
to match one of
		<\*matchers.EqualMatcher | 0x[[:xdigit:]]+>: {Expected: "a"}
or
		<\*matchers.BeNumericallyMatcher | 0x[[:xdigit:]]+>: {Comparator: ">", CompareTo: \[1\]}`

			Expect(msg).To(MatchRegexp(expectedMessagePattern))
		})
	})

	var _ = Describe("NegatedFailureMessage", func() {
		var oneOf = MatchOneOf("a", BeNumerically(">", 1))

		It("concatonates the failure message of all matchers", func() {
			msg := oneOf.NegatedFailureMessage("Fake Test Value")

			expectedMessagePattern := `Expected
		<string>: Fake Test Value
not to match one of
		<string>: a
or
		<\*matchers\.BeNumericallyMatcher | 0x[[:xdigit:]]+>: {Comparator: ">", CompareTo: \[1\]}`

			Expect(msg).To(MatchRegexp(expectedMessagePattern))
		})
	})
})
