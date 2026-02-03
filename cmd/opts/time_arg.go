package opts

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type TimeArg struct {
	time.Time
}

func (a *TimeArg) UnmarshalFlag(data string) error {
	// Try RFC3339 first (with timezone)
	t, err := time.Parse(time.RFC3339, data)
	if err != nil {
		// Try RFC3339 without timezone suffix, assume UTC
		// Format: "2006-01-02T15:04:05"
		t, err = time.Parse("2006-01-02T15:04:05", data)
		if err != nil {
			return bosherr.Errorf("Invalid timestamp '%s': expected RFC 3339 format (e.g., 2006-01-02T15:04:05Z or 2006-01-02T15:04:05)", data)
		}
		// Treat as UTC since no timezone was specified
		t = t.UTC()
	}
	// Always store as UTC internally
	a.Time = t.UTC()
	return nil
}

func (a TimeArg) IsSet() bool {
	return !a.IsZero()
}

func (a TimeArg) AsString() string {
	if a.IsSet() {
		// Always output in UTC with Z suffix for consistency
		return a.Format(time.RFC3339)
	}
	return ""
}
