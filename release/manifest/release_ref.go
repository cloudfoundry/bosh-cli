package manifest

import (
	"strings"
)

type ReleaseRef struct {
	Name string
	URL  string
}

func (r *ReleaseRef) Path() string {
	return strings.TrimPrefix(r.URL, "file://")
}
