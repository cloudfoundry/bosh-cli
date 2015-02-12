package time_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/ui/time"
)

var _ = Describe("DurationFmt", func() {
	Describe("Format", func() {
		var duration time.Duration

		Context("when given a duration less than one minute", func() {
			BeforeEach(func() {
				duration = 59 * time.Second
			})

			It("returns a string in 00:00:00 format", func() {
				Expect(Format(duration)).To(Equal("00:00:59"))
			})
		})

		Context("when given a duration greater than one minute", func() {
			BeforeEach(func() {
				duration = 69 * time.Second
			})

			It("returns a string in 00:00:00 format", func() {
				Expect(Format(duration)).To(Equal("00:01:09"))
			})
		})

		Context("when given a duration greater than one hour", func() {
			BeforeEach(func() {
				duration = 3669 * time.Second
			})

			It("returns a string in 00:00:00 format", func() {
				Expect(Format(duration)).To(Equal("01:01:09"))
			})
		})
	})
})
