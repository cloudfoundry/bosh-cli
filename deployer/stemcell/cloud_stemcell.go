package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
)

type CloudStemcell interface {
	CID() string
	Name() string
	Version() string
	PromoteAsCurrent() error
	Delete() error
}

type cloudStemcell struct {
	cid     string
	name    string
	version string
	repo    bmconfig.StemcellRepo
	cloud   bmcloud.Cloud
}

func NewCloudStemcell(
	stemcellRecord bmconfig.StemcellRecord,
	repo bmconfig.StemcellRepo,
	cloud bmcloud.Cloud,
) CloudStemcell {
	return &cloudStemcell{
		cid:     stemcellRecord.CID,
		name:    stemcellRecord.Name,
		version: stemcellRecord.Version,
		repo:    repo,
		cloud:   cloud,
	}
}

func (s *cloudStemcell) CID() string {
	return s.cid
}

func (s *cloudStemcell) Name() string {
	return s.name
}

func (s *cloudStemcell) Version() string {
	return s.version
}

func (s *cloudStemcell) PromoteAsCurrent() error {
	stemcellRecord, found, err := s.repo.Find(s.name, s.version)
	if err != nil {
		return bosherr.WrapError(err, "Finding current stemcell")
	}

	if !found {
		return bosherr.New("Stemcell does not exist in repo")
	}

	err = s.repo.UpdateCurrent(stemcellRecord.ID)
	if err != nil {
		return bosherr.WrapError(err, "Updating current stemcell")
	}

	return nil
}

func (s *cloudStemcell) Delete() error {
	err := s.cloud.DeleteStemcell(s.cid)
	if err != nil {
		return bosherr.WrapError(err, "Deleting stemcell from cloud")
	}

	stemcellRecord, found, err := s.repo.Find(s.name, s.version)
	if err != nil {
		return bosherr.WrapError(err, "Finding stemcell record (name=%s, version=%s)", s.name, s.version)
	}

	if !found {
		return nil
	}

	err = s.repo.Delete(stemcellRecord)
	if err != nil {
		return bosherr.WrapError(err, "Deleting stemcell record")
	}

	return nil
}
