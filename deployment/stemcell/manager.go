package stemcell

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type Manager interface {
	FindCurrent() ([]CloudStemcell, error)
	Upload(ExtractedStemcell, bmeventlog.Stage) (CloudStemcell, error)
	FindUnused() ([]CloudStemcell, error)
	DeleteUnused(bmeventlog.Stage) error
}

type manager struct {
	repo  bmconfig.StemcellRepo
	cloud bmcloud.Cloud
}

func NewManager(repo bmconfig.StemcellRepo, cloud bmcloud.Cloud) Manager {
	return &manager{
		repo:  repo,
		cloud: cloud,
	}
}

func (m *manager) FindCurrent() ([]CloudStemcell, error) {
	stemcells := []CloudStemcell{}

	stemcellRecord, found, err := m.repo.FindCurrent()
	if err != nil {
		return stemcells, bosherr.WrapError(err, "Reading stemcell record")
	}

	if found {
		stemcell := NewCloudStemcell(stemcellRecord, m.repo, m.cloud)
		stemcells = append(stemcells, stemcell)
	}

	return stemcells, nil
}

// Upload stemcell to an IAAS. It does the following steps:
// 1) uploads the stemcell to the cloud (if needed),
// 2) saves a record of the uploaded stemcell in the repo
func (m *manager) Upload(extractedStemcell ExtractedStemcell, uploadStage bmeventlog.Stage) (cloudStemcell CloudStemcell, err error) {
	err = uploadStage.PerformStep("Uploading", func() error {
		manifest := extractedStemcell.Manifest()
		foundStemcellRecord, found, err := m.repo.Find(manifest.Name, manifest.Version)
		if err != nil {
			return bosherr.WrapError(err, "Finding existing stemcell record in repo")
		}

		if found {
			cloudStemcell = NewCloudStemcell(foundStemcellRecord, m.repo, m.cloud)
			return bmeventlog.NewSkippedStepError("Stemcell already uploaded")
		}

		cloudProperties, err := manifest.CloudProperties()
		if err != nil {
			return bosherr.WrapError(err, "Getting cloud properties from stemcell manifest")
		}

		cid, err := m.cloud.CreateStemcell(manifest.ImagePath, cloudProperties)
		if err != nil {
			return bosherr.WrapErrorf(err, "creating stemcell (%s %s)", manifest.Name, manifest.Version)
		}

		stemcellRecord, err := m.repo.Save(manifest.Name, manifest.Version, cid)
		if err != nil {
			//TODO: delete stemcell from cloud when saving fails
			return bosherr.WrapErrorf(err, "saving stemcell record in repo (cid=%s, stemcell=%s)", cid, extractedStemcell)
		}

		cloudStemcell = NewCloudStemcell(stemcellRecord, m.repo, m.cloud)
		return nil
	})
	if err != nil {
		return cloudStemcell, err
	}

	return cloudStemcell, nil
}

func (m *manager) FindUnused() ([]CloudStemcell, error) {
	unusedStemcells := []CloudStemcell{}

	stemcellRecords, err := m.repo.All()
	if err != nil {
		return unusedStemcells, bosherr.WrapError(err, "Getting all stemcell records")
	}

	currentStemcellRecord, found, err := m.repo.FindCurrent()
	if err != nil {
		return unusedStemcells, bosherr.WrapError(err, "Finding current disk record")
	}

	for _, stemcellRecord := range stemcellRecords {
		if !found || stemcellRecord.ID != currentStemcellRecord.ID {
			stemcell := NewCloudStemcell(stemcellRecord, m.repo, m.cloud)
			unusedStemcells = append(unusedStemcells, stemcell)
		}
	}

	return unusedStemcells, nil
}

func (m *manager) DeleteUnused(deleteStage bmeventlog.Stage) error {
	stemcells, err := m.FindUnused()
	if err != nil {
		return bosherr.WrapError(err, "Finding unused stemcells")
	}

	for _, stemcell := range stemcells {
		stepName := fmt.Sprintf("Deleting unused stemcell '%s'", stemcell.CID())
		err = deleteStage.PerformStep(stepName, func() error {
			err := stemcell.Delete()
			cloudErr, ok := err.(bmcloud.Error)
			if ok && cloudErr.Type() == bmcloud.StemcellNotFoundError {
				return bmeventlog.NewSkippedStepError(cloudErr.Error())
			}
			return err
		})
		if err != nil {
			return err
		}
	}

	return nil
}
