package manifest

type ReleaseRef struct {
	Name string
	URL  string
	SHA1 string
}

func (r ReleaseRef) GetURL() string {
	return r.URL
}

func (r ReleaseRef) GetSHA1() string {
	return r.SHA1
}
