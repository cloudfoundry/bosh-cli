package errors_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/errors"
)

var _ = Describe("explainableError", func() {
	Describe("Error", func() {
		Context("when reasons are given", func() {
			It("returns each reason as bullet points", func() {
				err := NewExplainableError([]error{errors.New("reason 1"), errors.New("reason 2")})
				Expect(err.Error()).To(Equal("reason 1\nreason 2"))
			})
		})

		Context("when no reasons are given", func() {
			It("returns empty string", func() {
				err := NewExplainableError([]error{})
				Expect(err.Error()).To(Equal(""))
			})
		})
	})
})
