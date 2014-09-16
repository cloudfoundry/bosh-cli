package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
)

type Repo interface {
	Save(stemcell Stemcell, cid CID) error
	Find(stemcell Stemcell) (CID, bool, error)
}

type repo struct {
	configService bmconfig.Service
}

func NewRepo(configService bmconfig.Service) repo {
	return repo{
		configService: configService,
	}
}

// Save extracts the stemcell archive, parses the stemcell manifest, and stores the stemcell archive in the repo.
// The repo stemcell record is indexed by name & sha1 (as specified by the manifest).
func (s repo) Save(stemcell Stemcell, cid CID) error {
	config, _ := s.configService.Load()

	records := config.Stemcells
	if records == nil {
		records = []bmconfig.StemcellRecord{}
	}

	newRecord := bmconfig.StemcellRecord{
		Name:    stemcell.Name,
		Version: stemcell.Version,
		SHA1:    stemcell.SHA1,
		CID:     cid.String(),
	}

	oldRecord, found := s.find(records, newRecord)
	if found {
		return bosherr.New("Failed to save stemcell record `%s', existing record found `%s'", newRecord, oldRecord)
	}

	records = append(records, newRecord)
	config.Stemcells = records

	_ = s.configService.Save(config)
	return nil
}

func (s repo) Find(stemcell Stemcell) (CID, bool, error) {
	return "", false, nil
}

func (s repo) find(records []bmconfig.StemcellRecord, record bmconfig.StemcellRecord) (bmconfig.StemcellRecord, bool) {
	for _, existingRecord := range records {
		// "key" excludes CID
		if record.Name == existingRecord.Name &&
			record.Version == existingRecord.Version &&
			record.SHA1 == existingRecord.SHA1 {
			return existingRecord, true
		}
	}
	return bmconfig.StemcellRecord{}, false
}
