package opts_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
)

var _ = Describe("TimeArg", func() {
	Describe("UnmarshalFlag", func() {
		It("parses valid RFC 3339 timestamps with Z suffix", func() {
			var arg TimeArg
			err := arg.UnmarshalFlag("2026-01-01T00:00:00Z")
			Expect(err).ToNot(HaveOccurred())
			Expect(arg.Time).To(Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
		})

		It("parses RFC 3339 timestamps with timezone offset and converts to UTC", func() {
			var arg TimeArg
			err := arg.UnmarshalFlag("2026-06-15T14:30:00-07:00")
			Expect(err).ToNot(HaveOccurred())
			Expect(arg.Time).To(Equal(time.Date(2026, 6, 15, 21, 30, 0, 0, time.UTC)))
		})

		It("parses RFC 3339 timestamps with +00:00 offset and converts to UTC", func() {
			var arg TimeArg
			err := arg.UnmarshalFlag("2026-01-15T10:30:00+00:00")
			Expect(err).ToNot(HaveOccurred())
			Expect(arg.Time).To(Equal(time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)))
		})

		It("parses timestamps without timezone suffix and treats as UTC", func() {
			var arg TimeArg
			err := arg.UnmarshalFlag("2026-01-01T00:00:00")
			Expect(err).ToNot(HaveOccurred())
			Expect(arg.Time).To(Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
		})

		It("parses timestamps without timezone with specific time and treats as UTC", func() {
			var arg TimeArg
			err := arg.UnmarshalFlag("2026-06-15T14:30:45")
			Expect(err).ToNot(HaveOccurred())
			Expect(arg.Time).To(Equal(time.Date(2026, 6, 15, 14, 30, 45, 0, time.UTC)))
		})

		It("returns error for invalid timestamps", func() {
			var arg TimeArg
			err := arg.UnmarshalFlag("not-a-timestamp")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Invalid timestamp"))
		})

		It("returns error for date-only formats (no time component)", func() {
			var arg TimeArg
			err := arg.UnmarshalFlag("2026-01-01")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Invalid timestamp"))
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

		It("returns RFC 3339 formatted string in UTC for non-zero time", func() {
			arg := TimeArg{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
			Expect(arg.AsString()).To(Equal("2026-01-01T00:00:00Z"))
		})

		It("returns UTC formatted string even when time was parsed with offset", func() {
			var arg TimeArg
			// Parse with -07:00 offset (14:30 PST = 21:30 UTC)
			err := arg.UnmarshalFlag("2026-06-15T14:30:00-07:00")
			Expect(err).ToNot(HaveOccurred())
			Expect(arg.AsString()).To(Equal("2026-06-15T21:30:00Z"))
		})

		It("returns UTC formatted string for timestamp parsed without timezone", func() {
			var arg TimeArg
			err := arg.UnmarshalFlag("2026-06-15T14:30:00")
			Expect(err).ToNot(HaveOccurred())
			Expect(arg.AsString()).To(Equal("2026-06-15T14:30:00Z"))
		})
	})
})
