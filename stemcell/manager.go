package stemcell

import (
	"errors"
	"fmt"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bicloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Manager

type Manager interface {
	FindCurrent() ([]CloudStemcell, error)
	Upload(ExtractedStemcell, biui.Stage, bool) (CloudStemcell, error)
	FindUnused() ([]CloudStemcell, error)
	DeleteUnused(biui.Stage) error
}

type manager struct {
	repo  biconfig.StemcellRepo
	cloud bicloud.Cloud
}

func NewManager(repo biconfig.StemcellRepo, cloud bicloud.Cloud) Manager {
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
//
// The repo records stemcells by name and version only -- it has no notion of
// which IaaS, or which vCenter, the image was actually materialized in. When
// the deployment is repointed at different infrastructure the recorded CID
// names an image that does not exist there, but the name and version still
// match, so the upload is skipped and the stale CID is handed to create_vm.
// Passing fix forces a fresh create_stemcell against whatever the CPI is now
// talking to.
func (m *manager) Upload(extractedStemcell ExtractedStemcell, uploadStage biui.Stage, fix bool) (cloudStemcell CloudStemcell, err error) {
	manifest := extractedStemcell.Manifest()
	stageName := fmt.Sprintf("Uploading stemcell '%s/%s'", manifest.Name, manifest.Version)
	err = uploadStage.Perform(stageName, func() error {
		foundStemcellRecord, found, err := m.repo.Find(manifest.Name, manifest.Version)
		if err != nil {
			return bosherr.WrapError(err, "Finding existing stemcell record in repo")
		}

		if found && !fix {
			cloudStemcell = NewCloudStemcell(foundStemcellRecord, m.repo, m.cloud)
			return biui.NewSkipStageError(bosherr.Errorf("Found stemcell: %#v", foundStemcellRecord), "Stemcell already uploaded")
		}

		if found {
			// Drop the stale record before uploading: Save rejects a duplicate
			// name/version pair, so it would fail after the new image had
			// already been created.
			//
			// Only the record is removed, not the image. CloudStemcell.Delete
			// would ask the CPI to delete the old CID, which at best is a
			// no-op against infrastructure that never had it and at worst
			// destroys the image the deployment can still be rolled back onto.
			err = m.repo.Delete(foundStemcellRecord)
			if err != nil {
				return bosherr.WrapErrorf(err, "Deleting stale stemcell record (name=%s, version=%s, cid=%s)", foundStemcellRecord.Name, foundStemcellRecord.Version, foundStemcellRecord.CID)
			}
		}

		cid, err := m.cloud.CreateStemcell(filepath.Join(extractedStemcell.GetExtractedPath(), "image"), manifest.CloudProperties)
		if err != nil {
			return bosherr.WrapErrorf(err, "creating stemcell (%s %s)", manifest.Name, manifest.Version)
		}

		stemcellRecord, err := m.repo.Save(manifest.Name, manifest.Version, cid, manifest.ApiVersion)
		if err != nil {
			// The image now exists in the IaaS with nothing recording it, so
			// neither delete-env nor unused-stemcell cleanup can ever find it,
			// and a retry would create another one. Remove it, but report the
			// original save failure rather than the cleanup result.
			if deleteErr := m.cloud.DeleteStemcell(cid); deleteErr != nil {
				return bosherr.WrapErrorf(err, "saving stemcell record in repo (cid=%s, stemcell=%s); the orphaned stemcell could not be deleted either: %s", cid, extractedStemcell, deleteErr.Error())
			}
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

func (m *manager) DeleteUnused(deleteStage biui.Stage) error {
	stemcells, err := m.FindUnused()
	if err != nil {
		return bosherr.WrapError(err, "Finding unused stemcells")
	}

	for _, stemcell := range stemcells {
		stepName := fmt.Sprintf("Deleting unused stemcell '%s'", stemcell.CID())
		err = deleteStage.Perform(stepName, func() error {
			err := stemcell.Delete()
			var cloudErr bicloud.Error
			ok := errors.As(err, &cloudErr)
			if ok && cloudErr.Type() == bicloud.StemcellNotFoundError {
				return biui.NewSkipStageError(cloudErr, "Stemcell not found")
			}
			return err
		})
		if err != nil {
			return err
		}
	}

	return nil
}
