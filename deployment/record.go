package deployment

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	biconfig "github.com/cloudfoundry/bosh-init/config"
	bicrypto "github.com/cloudfoundry/bosh-init/crypto"
	birel "github.com/cloudfoundry/bosh-init/release"
	bistemcell "github.com/cloudfoundry/bosh-init/stemcell"
)

type Record interface {
	IsDeployed(manifestPath string, releases []birel.Release, stemcell bistemcell.ExtractedStemcell) (bool, error)
	Update(manifestPath string, releases []birel.Release) error
}

type deploymentRecord struct {
	deploymentRepo biconfig.DeploymentRepo
	releaseRepo    biconfig.ReleaseRepo
	stemcellRepo   biconfig.StemcellRepo
	sha1Calculator bicrypto.SHA1Calculator
}

func NewRecord(
	deploymentRepo biconfig.DeploymentRepo,
	releaseRepo biconfig.ReleaseRepo,
	stemcellRepo biconfig.StemcellRepo,
	sha1Calculator bicrypto.SHA1Calculator,
) Record {
	return &deploymentRecord{
		deploymentRepo: deploymentRepo,
		releaseRepo:    releaseRepo,
		stemcellRepo:   stemcellRepo,
		sha1Calculator: sha1Calculator,
	}
}

func (v *deploymentRecord) IsDeployed(manifestPath string, releases []birel.Release, stemcell bistemcell.ExtractedStemcell) (bool, error) {
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

	currentReleaseRecords, err := v.releaseRepo.List()
	if err != nil {
		return false, bosherr.WrapError(err, "Finding currently deployed release")
	}

	if len(currentReleaseRecords) == 0 {
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

func (v *deploymentRecord) Update(manifestPath string, releases []birel.Release) error {
	manifestSHA1, err := v.sha1Calculator.Calculate(manifestPath)
	if err != nil {
		return bosherr.WrapError(err, "Calculating sha1 of current deployment manifest")
	}

	err = v.deploymentRepo.UpdateCurrent(manifestSHA1)
	if err != nil {
		return bosherr.WrapError(err, "Saving sha1 of deployed manifest")
	}

	err = v.releaseRepo.Update(releases)
	if err != nil {
		return bosherr.WrapError(err, "Updating releases")
	}

	return nil
}
