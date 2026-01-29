package opts_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
)

var _ = Describe("TimeArg", func() {
	Describe("UnmarshalFlag", func() {
		It("parses valid RFC 3339 timestamps", func() {
			var arg TimeArg
			err := arg.UnmarshalFlag("2026-01-01T00:00:00Z")
			Expect(err).ToNot(HaveOccurred())
			Expect(arg.Time).To(Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
		})

		It("parses RFC 3339 timestamps with timezone offset", func() {
			var arg TimeArg
			err := arg.UnmarshalFlag("2026-06-15T14:30:00-07:00")
			Expect(err).ToNot(HaveOccurred())
			Expect(arg.Time.UTC()).To(Equal(time.Date(2026, 6, 15, 21, 30, 0, 0, time.UTC)))
		})

		It("returns error for invalid timestamps", func() {
			var arg TimeArg
			err := arg.UnmarshalFlag("not-a-timestamp")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Invalid RFC 3339 timestamp"))
		})

		It("returns error for non-RFC3339 date formats", func() {
			var arg TimeArg
			err := arg.UnmarshalFlag("2026-01-01")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Invalid RFC 3339 timestamp"))
		})
	})

	Describe("IsSet", func() {
		It("returns false for zero time", func() {
			var arg TimeArg
			Expect(arg.IsSet()).To(BeFalse())
		})

		It("returns true for non-zero time", func() {
			arg := TimeArg{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
			Expect(arg.IsSet()).To(BeTrue())
		})
	})

	Describe("AsString", func() {
		It("returns empty string for zero time", func() {
			var arg TimeArg
			Expect(arg.AsString()).To(Equal(""))
		})

		It("returns RFC 3339 formatted string for non-zero time", func() {
			arg := TimeArg{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
			Expect(arg.AsString()).To(Equal("2026-01-01T00:00:00Z"))
		})
	})
})
