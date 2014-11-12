package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type Manager interface {
	Upload(ExtractedStemcell) (CloudStemcell, error)
}

type manager struct {
	repo        Repo
	cloud       bmcloud.Cloud
	eventLogger bmeventlog.EventLogger
}

func NewManager(repo Repo, cloud bmcloud.Cloud, eventLogger bmeventlog.EventLogger) Manager {
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
	cloudStemcell, found, err := m.repo.Find(manifest)
	if err != nil {
		return CloudStemcell{}, bosherr.WrapError(err, "finding existing stemcell record in repo")
	}
	eventStep := eventLoggerStage.NewStep("Uploading")
	if found {
		eventStep.Skip("Stemcell already uploaded")
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

	cloudStemcell = CloudStemcell{CID: cid}
	err = m.repo.Save(manifest, cloudStemcell)
	if err != nil {
		//TODO: delete stemcell from cloud when saving fails
		eventStep.Fail(err.Error())
		return cloudStemcell, bosherr.WrapError(
			err,
			"saving stemcell record in repo (record=%s, stemcell=%s)",
			cid,
			manifest,
		)
	}

	eventStep.Finish()
	eventLoggerStage.Finish()

	return cloudStemcell, nil
}
