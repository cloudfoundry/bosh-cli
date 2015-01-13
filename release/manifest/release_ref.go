package manifest

type ReleaseRef struct {
	Name    string
	Version string
}

func (r *ReleaseRef) IsLatest() bool {
	return r.Version == "latest"
}
