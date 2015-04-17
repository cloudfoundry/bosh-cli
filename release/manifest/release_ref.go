package manifest

import (
	"strings"
)

type ReleaseRef struct {
	Name    string
	Version string
	URL     string
}

func (r *ReleaseRef) IsLatest() bool {
	return r.Version == "" || r.Version == "latest"
}

func (r *ReleaseRef) Path() string {
	return strings.TrimPrefix(r.URL, "file://")
}
