package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type Manager interface {
	Extract(tarballPath string) (ExtractedStemcell, error)
	Upload(ExtractedStemcell) (CloudStemcell, error)
}

type manager struct {
	fs          boshsys.FileSystem
	reader      Reader
	repo        Repo
	eventLogger bmeventlog.EventLogger
	cloud       bmcloud.Cloud
}

// Extract decompresses a stemcell tarball into a temp directory (stemcell.extractedPath)
// and parses and validates the stemcell manifest.
// Use stemcell.Delete() to clean up the temp directory.
func (m *manager) Extract(tarballPath string) (ExtractedStemcell, error) {
	tmpDir, err := m.fs.TempDir("stemcell-manager")
	if err != nil {
		return nil, bosherr.WrapError(err, "creating temp dir for stemcell extraction")
	}

	stemcell, err := m.reader.Read(tarballPath, tmpDir)
	if err != nil {
		return nil, bosherr.WrapError(err, "reading extracted stemcell manifest in `%s'", tmpDir)
	}

	return stemcell, nil
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
