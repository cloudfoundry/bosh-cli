package deployment

import (
	"strconv"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
)

type WatchTime struct {
	Start int
	End   int
}

func NewWatchTime(timeRange string) (WatchTime, error) {
	parts := strings.Split(timeRange, "-")
	if len(parts) != 2 {
		return WatchTime{}, bosherr.New("Invalid watch time range '%s'", timeRange)
	}

	start, err := strconv.Atoi(strings.Trim(parts[0], " "))
	if err != nil {
		return WatchTime{}, bosherr.WrapError(
			err, "Non-positive number as watch time minimum %s", parts[0])
	}

	end, err := strconv.Atoi(strings.Trim(parts[1], " "))
	if err != nil {
		return WatchTime{}, bosherr.WrapError(
			err, "Non-positive number as watch time maximum %s", parts[1])
	}

	if end < start {
		return WatchTime{}, bosherr.New(
			"Watch time must have maximum greater than or equal minimum %s", timeRange)
	}

	return WatchTime{
		Start: start,
		End:   end,
	}, nil
}
