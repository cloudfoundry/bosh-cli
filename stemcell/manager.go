package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogging"
)

type CID string

func (c CID) String() string {
	return string(c)
}

type Manager interface {
	Upload(tarballPath string) (Stemcell, CID, error)
}

type manager struct {
	fs             boshsys.FileSystem
	reader         Reader
	repo           Repo
	eventLogger    bmeventlog.EventLogger
	infrastructure Infrastructure
}

// Upload stemcell to an IAAS. It does the following steps:
// 1) extracts the tarball & parses its manifest,
// 2) uploads the stemcell to the infrastructure (if needed),
// 3) saves a record of the uploaded stemcell in the repo
func (m *manager) Upload(tarballPath string) (stemcell Stemcell, cid CID, err error) {
	tmpDir, err := m.fs.TempDir("stemcell-manager")
	if err != nil {
		return stemcell, cid, bosherr.WrapError(err, "creating temp dir for stemcell extraction")
	}
	defer m.fs.RemoveAll(tmpDir)

	event := bmeventlog.Event{
		Stage: "uploading stemcell",
		Total: 2,
		Task:  "Unpacking",
		Index: 1,
		State: bmeventlog.Started,
	}
	m.eventLogger.AddEvent(event)

	stemcell, err = m.reader.Read(tarballPath, tmpDir)
	if err != nil {
		event = bmeventlog.Event{
			Stage: "uploading stemcell",
			Total: 2,
			Task:  "Unpacking",
			Index: 1,
			State: bmeventlog.Failed,
		}
		m.eventLogger.AddEvent(event)

		return stemcell, cid, bosherr.WrapError(err, "reading extracted stemcell manifest in `%s'", tmpDir)
	}

	event = bmeventlog.Event{
		Stage: "uploading stemcell",
		Total: 2,
		Task:  "Unpacking",
		Index: 1,
		State: bmeventlog.Finished,
	}
	m.eventLogger.AddEvent(event)

	cid, found, err := m.repo.Find(stemcell)
	if err != nil {
		return stemcell, cid, bosherr.WrapError(err, "finding existing stemcell record in repo")
	}
	if found {
		event = bmeventlog.Event{
			Stage:   "uploading stemcell",
			Total:   2,
			Task:    "Uploading",
			Index:   2,
			State:   bmeventlog.Skipped,
			Message: "stemcell already uploaded",
		}
		m.eventLogger.AddEvent(event)

		return stemcell, cid, nil
	}

	event = bmeventlog.Event{
		Stage: "uploading stemcell",
		Total: 2,
		Task:  "Uploading",
		Index: 2,
		State: bmeventlog.Started,
	}
	m.eventLogger.AddEvent(event)

	cid, err = m.infrastructure.CreateStemcell(stemcell)
	if err != nil {
		event = bmeventlog.Event{
			Stage: "uploading stemcell",
			Total: 2,
			Task:  "Uploading",
			Index: 2,
			State: bmeventlog.Failed,
		}
		m.eventLogger.AddEvent(event)

		return Stemcell{}, "", bosherr.WrapError(
			err,
			"creating stemcell (infrastructure=%s, stemcell=%s)",
			m.infrastructure,
			stemcell,
		)
	}

	err = m.repo.Save(stemcell, cid)
	if err != nil {
		//TODO: delete stemcell from infrastructure when saving fails
		event = bmeventlog.Event{
			Stage: "uploading stemcell",
			Total: 2,
			Task:  "Uploading",
			Index: 2,
			State: bmeventlog.Failed,
		}
		m.eventLogger.AddEvent(event)

		return Stemcell{}, "", bosherr.WrapError(
			err,
			"saving stemcell record in repo (record=%s, stemcell=%s)",
			cid,
			stemcell,
		)
	}

	event = bmeventlog.Event{
		Stage: "uploading stemcell",
		Total: 2,
		Task:  "Uploading",
		Index: 2,
		State: bmeventlog.Finished,
	}
	m.eventLogger.AddEvent(event)

	return stemcell, cid, nil
}
