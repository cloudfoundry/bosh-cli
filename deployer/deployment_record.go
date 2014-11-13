package deployer

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type DeploymentRecord interface {
	IsDeployed(manifestPath string, release bmrel.Release, stemcell bmstemcell.ExtractedStemcell) (bool, error)
	Update(manifestPath string, release bmrel.Release, stemcell bmstemcell.ExtractedStemcell) error
}

type deploymentRecord struct {
	stemcellRepo bmconfig.StemcellRepo
}

func NewDeploymentRecord(stemcellRepo bmconfig.StemcellRepo) DeploymentRecord {
	return &deploymentRecord{
		stemcellRepo: stemcellRepo,
	}
}

func (v *deploymentRecord) IsDeployed(manifestPath string, release bmrel.Release, stemcell bmstemcell.ExtractedStemcell) (bool, error) {
	currentStemcell, found, err := v.stemcellRepo.FindCurrent()
	if err != nil {
		return false, err
	}

	if !found {
		return false, nil
	}

	if currentStemcell.Name != stemcell.Manifest().Name || currentStemcell.Version != stemcell.Manifest().Version {
		return false, nil
	}

	return true, nil
}

func (v *deploymentRecord) Update(manifestPath string, release bmrel.Release, stemcell bmstemcell.ExtractedStemcell) error {
	stemcellManifest := stemcell.Manifest()
	stemcellRecord, found, err := v.stemcellRepo.Find(stemcellManifest.Name, stemcellManifest.Version)
	if err != nil {
		return bosherr.WrapError(err, "Finding stemcell record with name: '%s', version: '%s'", stemcellManifest.Name, stemcellManifest.Version)
	}

	if !found {
		return bosherr.New("Stemcell record not found with name '%s', version: '%s'", stemcellManifest.Name, stemcellManifest.Version)
	}

	err = v.stemcellRepo.UpdateCurrent(stemcellRecord.ID)
	if err != nil {
		return bosherr.WrapError(err, "Updating current stemcell record")
	}

	return nil
}
