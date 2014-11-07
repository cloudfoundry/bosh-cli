package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type CID string

func (c CID) String() string {
	return string(c)
}

type Manager interface {
	Upload(tarballPath string) (Stemcell, CID, error)
}

type manager struct {
	fs          boshsys.FileSystem
	reader      Reader
	repo        Repo
	eventLogger bmeventlog.EventLogger
	cloud       bmcloud.Cloud
}

// Upload stemcell to an IAAS. It does the following steps:
// 1) extracts the tarball & parses its manifest,
// 2) uploads the stemcell to the cloud (if needed),
// 3) saves a record of the uploaded stemcell in the repo
func (m *manager) Upload(tarballPath string) (stemcell Stemcell, cid CID, err error) {
	tmpDir, err := m.fs.TempDir("stemcell-manager")
	if err != nil {
		return stemcell, cid, bosherr.WrapError(err, "creating temp dir for stemcell extraction")
	}
	defer m.fs.RemoveAll(tmpDir)

	eventLoggerStage := m.eventLogger.NewStage("uploading stemcell")
	eventLoggerStage.Start()
	defer eventLoggerStage.Finish()

	eventStep := eventLoggerStage.NewStep("Unpacking")
	eventStep.Start()

	stemcell, err = m.reader.Read(tarballPath, tmpDir)
	if err != nil {
		eventStep.Fail(err.Error())
		return Stemcell{}, "", bosherr.WrapError(err, "reading extracted stemcell manifest in `%s'", tmpDir)
	}

	eventStep.Finish()
	cid, found, err := m.repo.Find(stemcell.Manifest)
	if err != nil {
		return Stemcell{}, "", bosherr.WrapError(err, "finding existing stemcell record in repo")
	}
	eventStep = eventLoggerStage.NewStep("Uploading")
	if found {
		eventStep.Skip("Stemcell already uploaded")
		return stemcell, cid, nil
	}

	eventStep.Start()
	cloudProperties, err := stemcell.Manifest.CloudProperties()
	if err != nil {
		return Stemcell{}, "", bosherr.WrapError(err, "Getting cloud properties from stemcell manifest")
	}

	stemcellCid, err := m.cloud.CreateStemcell(cloudProperties, stemcell.Manifest.ImagePath)
	if err != nil {
		eventStep.Fail(err.Error())
		return Stemcell{}, "", bosherr.WrapError(
			err,
			"creating stemcell (cloud=%s, stemcell=%s)",
			m.cloud,
			stemcell.Manifest,
		)
	}

	cid = CID(stemcellCid)
	err = m.repo.Save(stemcell.Manifest, cid)
	if err != nil {
		//TODO: delete stemcell from cloud when saving fails
		eventStep.Fail(err.Error())
		return Stemcell{}, "", bosherr.WrapError(
			err,
			"saving stemcell record in repo (record=%s, stemcell=%s)",
			cid,
			stemcell.Manifest,
		)
	}

	eventStep.Finish()

	return stemcell, cid, nil
}
