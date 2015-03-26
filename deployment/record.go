package deployment

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmcrypto "github.com/cloudfoundry/bosh-micro-cli/crypto"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type Record interface {
	IsDeployed(manifestPath string, releases []bmrel.Release, stemcell bmstemcell.ExtractedStemcell) (bool, error)
	Update(manifestPath string, releases []bmrel.Release) error
}

type deploymentRecord struct {
	deploymentRepo bmconfig.DeploymentRepo
	releaseRepo    bmconfig.ReleaseRepo
	stemcellRepo   bmconfig.StemcellRepo
	sha1Calculator bmcrypto.SHA1Calculator
}

func NewRecord(
	deploymentRepo bmconfig.DeploymentRepo,
	releaseRepo bmconfig.ReleaseRepo,
	stemcellRepo bmconfig.StemcellRepo,
	sha1Calculator bmcrypto.SHA1Calculator,
) Record {
	return &deploymentRecord{
		deploymentRepo: deploymentRepo,
		releaseRepo:    releaseRepo,
		stemcellRepo:   stemcellRepo,
		sha1Calculator: sha1Calculator,
	}
}

func (v *deploymentRecord) IsDeployed(manifestPath string, releases []bmrel.Release, stemcell bmstemcell.ExtractedStemcell) (bool, error) {
	manifestSHA1, found, err := v.deploymentRepo.FindCurrent()
	if err != nil {
		return false, bosherr.WrapError(err, "Finding sha1 of currently deployed manifest")
	}

	if !found {
		return false, nil
	}

	newSHA1, err := v.sha1Calculator.Calculate(manifestPath)
	if err != nil {
		return false, bosherr.WrapError(err, "Calculating sha1 of current deployment manifest")
	}

	if manifestSHA1 != newSHA1 {
		return false, nil
	}

	currentStemcell, found, err := v.stemcellRepo.FindCurrent()
	if err != nil {
		return false, bosherr.WrapError(err, "Finding currently deployed stemcell")
	}

	if !found {
		return false, nil
	}

	if currentStemcell.Name != stemcell.Manifest().Name || currentStemcell.Version != stemcell.Manifest().Version {
		return false, nil
	}

	currentReleaseRecords, found, err := v.releaseRepo.FindCurrent()
	if err != nil {
		return false, bosherr.WrapError(err, "Finding currently deployed release")
	}

	if !found {
		return false, nil
	}

	if len(releases) != len(currentReleaseRecords) {
		return false, nil
	}

	for _, release := range releases {
		found := false
		for _, releaseRecord := range currentReleaseRecords {
			if releaseRecord.Name == release.Name() && releaseRecord.Version == release.Version() {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	return true, nil
}

func (v *deploymentRecord) Update(manifestPath string, releases []bmrel.Release) error {
	manifestSHA1, err := v.sha1Calculator.Calculate(manifestPath)
	if err != nil {
		return bosherr.WrapError(err, "Calculating sha1 of current deployment manifest")
	}

	err = v.deploymentRepo.UpdateCurrent(manifestSHA1)
	if err != nil {
		return bosherr.WrapError(err, "Saving sha1 of deployed manifest")
	}

	var currentReleaseRecordIDs []string
	for _, release := range releases {
		releaseRecord, found, err := v.releaseRepo.Find(release.Name(), release.Version())
		if !found {
			releaseRecord, err = v.releaseRepo.Save(release.Name(), release.Version())
			if err != nil {
				return bosherr.WrapErrorf(err, "Saving release record with name: '%s', version: '%s'", release.Name(), release.Version())
			}
		}
		currentReleaseRecordIDs = append(currentReleaseRecordIDs, releaseRecord.ID)
	}

	err = v.releaseRepo.UpdateCurrent(currentReleaseRecordIDs)
	if err != nil {
		return bosherr.WrapError(err, "Updating current release record")
	}

	return nil
}
