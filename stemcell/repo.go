package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
)

// Repo persists stemcells metadata
type Repo interface {
	Save(stemcell Stemcell, cid CID) error
	Find(stemcell Stemcell) (CID, bool, error)
}

type repo struct {
	configService bmconfig.DeploymentConfigService
}

func NewRepo(configService bmconfig.DeploymentConfigService) repo {
	return repo{
		configService: configService,
	}
}

// Save extracts the stemcell archive,
// parses the stemcell manifest,
// and stores the stemcell archive in the repo.
// The repo stemcell record is indexed by name & sha1 (as specified by the manifest).
func (s repo) Save(stemcell Stemcell, cid CID) error {
	config, err := s.configService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	records := config.Stemcells
	if records == nil {
		records = []bmconfig.StemcellRecord{}
	}

	newRecord := bmconfig.StemcellRecord{
		Name:    stemcell.Name,
		Version: stemcell.Version,
		SHA1:    stemcell.SHA1,
	}

	oldRecord, found := s.find(records, newRecord)
	if found {
		return bosherr.New("Failed to save stemcell record `%s', existing record found `%s'", newRecord, oldRecord)
	}

	newRecord.CID = cid.String()
	records = append(records, newRecord)
	config.Stemcells = records

	err = s.configService.Save(config)
	if err != nil {
		//		s.logger.Error("Failed saving updated config: %s", config)
		return bosherr.WrapError(err, "Saving new config")
	}
	return nil
}

func (s repo) Find(stemcell Stemcell) (cid CID, found bool, err error) {
	config, err := s.configService.Load()
	if err != nil {
		return cid, false, bosherr.WrapError(err, "Loading existing config")
	}

	records := config.Stemcells
	if records == nil {
		records = []bmconfig.StemcellRecord{}
	}

	newRecord := bmconfig.StemcellRecord{
		Name:    stemcell.Name,
		Version: stemcell.Version,
		SHA1:    stemcell.SHA1,
	}

	oldRecord, found := s.find(records, newRecord)
	return CID(oldRecord.CID), found, nil
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
