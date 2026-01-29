package opts

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type TimeArg struct {
	time.Time
}

func (a *TimeArg) UnmarshalFlag(data string) error {
	t, err := time.Parse(time.RFC3339, data)
	if err != nil {
		return bosherr.Errorf("Invalid RFC 3339 timestamp '%s': %s", data, err)
	}
	a.Time = t
	return nil
}

func (a TimeArg) IsSet() bool {
	return !a.IsZero()
}

func (a TimeArg) AsString() string {
	if a.IsSet() {
		return a.Format(time.RFC3339)
	}
	return ""
}
