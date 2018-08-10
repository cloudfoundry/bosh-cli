package stemcellfakes

type FakeCloudStemcell struct {
	cid        string
	name       string
	version    string
	apiVersion int

	PromoteAsCurrentCalledTimes int
	PromoteAsCurrentErr         error

	DeleteCalledTimes int
	DeleteErr         error
}

func NewFakeCloudStemcell(cid, name, version string, apiVersion int) *FakeCloudStemcell {
	return &FakeCloudStemcell{
		cid:        cid,
		name:       name,
		version:    version,
		apiVersion: apiVersion,
	}
}

func (s *FakeCloudStemcell) CID() string {
	return s.cid
}

func (s *FakeCloudStemcell) Name() string {
	return s.name
}

func (s *FakeCloudStemcell) Version() string {
	return s.version
}

func (s *FakeCloudStemcell) ApiVersion() int {
	return s.apiVersion
}

func (s *FakeCloudStemcell) PromoteAsCurrent() error {
	s.PromoteAsCurrentCalledTimes++
	return s.PromoteAsCurrentErr
}

func (s *FakeCloudStemcell) Delete() error {
	s.DeleteCalledTimes++
	return s.DeleteErr
}
