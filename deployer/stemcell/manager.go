package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type Manager interface {
	Upload(ExtractedStemcell) (CloudStemcell, error)
}

type manager struct {
	repo        bmconfig.StemcellRepo
	cloud       bmcloud.Cloud
	eventLogger bmeventlog.EventLogger
}

func NewManager(repo bmconfig.StemcellRepo, cloud bmcloud.Cloud, eventLogger bmeventlog.EventLogger) Manager {
	return &manager{
		repo:        repo,
		cloud:       cloud,
		eventLogger: eventLogger,
	}
}

// Upload stemcell to an IAAS. It does the following steps:
// 1) uploads the stemcell to the cloud (if needed),
// 2) saves a record of the uploaded stemcell in the repo
func (m *manager) Upload(extractedStemcell ExtractedStemcell) (CloudStemcell, error) {
	eventLoggerStage := m.eventLogger.NewStage("uploading stemcell")
	eventLoggerStage.Start()

	manifest := extractedStemcell.Manifest()
	foundStemcellRecord, found, err := m.repo.Find(manifest.Name, manifest.Version)
	if err != nil {
		return CloudStemcell{}, bosherr.WrapError(err, "finding existing stemcell record in repo")
	}
	eventStep := eventLoggerStage.NewStep("Uploading")
	if found {
		eventStep.Skip("Stemcell already uploaded")
		cloudStemcell := CloudStemcell{
			CID: foundStemcellRecord.CID,
		}
		return cloudStemcell, nil
	}

	eventStep.Start()
	cloudProperties, err := manifest.CloudProperties()
	if err != nil {
		return CloudStemcell{}, bosherr.WrapError(err, "Getting cloud properties from stemcell manifest")
	}

	cid, err := m.cloud.CreateStemcell(cloudProperties, manifest.ImagePath)
	if err != nil {
		eventStep.Fail(err.Error())
		return CloudStemcell{}, bosherr.WrapError(
			err,
			"creating stemcell (cloud=%s, stemcell=%s)",
			m.cloud,
			extractedStemcell.Manifest(),
		)
	}

	record := bmconfig.StemcellRecord{
		Name:    manifest.Name,
		Version: manifest.Version,
		CID:     cid,
	}

	err = m.repo.Save(record)
	if err != nil {
		//TODO: delete stemcell from cloud when saving fails
		eventStep.Fail(err.Error())
		return CloudStemcell{}, bosherr.WrapError(
			err,
			"saving stemcell record in repo (record=%s, stemcell=%s)",
			cid,
			manifest,
		)
	}

	cloudStemcell := CloudStemcell{CID: cid}

	eventStep.Finish()
	eventLoggerStage.Finish()

	return cloudStemcell, nil
}
