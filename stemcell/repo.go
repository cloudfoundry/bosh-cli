package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
)

// Repo persists stemcells metadata
type Repo interface {
	Save(stemcellManifest Manifest, cid CID) error
	Find(stemcellManifest Manifest) (CID, bool, error)
}

type repo struct {
	configService bmconfig.DeploymentConfigService
}

func NewRepo(configService bmconfig.DeploymentConfigService) repo {
	return repo{
		configService: configService,
	}
}

func (r repo) Save(stemcellManifest Manifest, cid CID) error {
	config, err := r.configService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	records := config.Stemcells
	if records == nil {
		records = []bmconfig.StemcellRecord{}
	}

	newRecord := bmconfig.StemcellRecord{
		Name:    stemcellManifest.Name,
		Version: stemcellManifest.Version,
		SHA1:    stemcellManifest.SHA1,
	}

	oldRecord, found := r.find(records, newRecord)
	if found {
		return bosherr.New("Failed to save stemcell record `%s', existing record found `%s'", newRecord, oldRecord)
	}

	newRecord.CID = cid.String()
	records = append(records, newRecord)
	config.Stemcells = records

	err = r.configService.Save(config)
	if err != nil {
		//		r.logger.Error("Failed saving updated config: %s", config)
		return bosherr.WrapError(err, "Saving new config")
	}
	return nil
}

func (r repo) Find(stemcellManifest Manifest) (cid CID, found bool, err error) {
	config, err := r.configService.Load()
	if err != nil {
		return cid, false, bosherr.WrapError(err, "Loading existing config")
	}

	records := config.Stemcells
	if records == nil {
		records = []bmconfig.StemcellRecord{}
	}

	newRecord := bmconfig.StemcellRecord{
		Name:    stemcellManifest.Name,
		Version: stemcellManifest.Version,
		SHA1:    stemcellManifest.SHA1,
	}

	oldRecord, found := r.find(records, newRecord)
	return CID(oldRecord.CID), found, nil
}

func (r repo) find(records []bmconfig.StemcellRecord, record bmconfig.StemcellRecord) (bmconfig.StemcellRecord, bool) {
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
