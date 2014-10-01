package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmlog "github.com/cloudfoundry/bosh-micro-cli/logging"
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
	eventLogger    bmlog.EventLogger
	infrastructure Infrastructure
}

// Upload
// 1) extracts the tarball & parses its manifest,
// 2) uploads the stemcell to the infrastructure (if needed),
// 3) saves a record of the uploaded stemcell in the repo
func (m *manager) Upload(tarballPath string) (stemcell Stemcell, cid CID, err error) {
	// unpack stemcell tarball
	tmpDir, err := m.fs.TempDir("stemcell-manager")
	if err != nil {
		return stemcell, cid, bosherr.WrapError(err, "creating temp dir for stemcell extraction")
	}
	defer m.fs.RemoveAll(tmpDir)

	event := bmlog.Event{
		Stage: "uploading stemcell",
		Total: 2,
		Task:  "Unpacking",
		Index: 1,
		State: bmlog.Started,
	}
	m.eventLogger.AddEvent(event)

	// parse/reads stemcell manifest into Stemcell object
	stemcell, err = m.reader.Read(tarballPath, tmpDir)
	if err != nil {
		event = bmlog.Event{
			Stage: "uploading stemcell",
			Total: 2,
			Task:  "Unpacking",
			Index: 1,
			State: bmlog.Failed,
		}
		m.eventLogger.AddEvent(event)

		return stemcell, cid, bosherr.WrapError(err, "reading extracted stemcell manifest in `%s'", tmpDir)
	}

	event = bmlog.Event{
		Stage: "uploading stemcell",
		Total: 2,
		Task:  "Unpacking",
		Index: 1,
		State: bmlog.Finished,
	}
	m.eventLogger.AddEvent(event)

	cid, found, err := m.repo.Find(stemcell)
	if err != nil {
		return stemcell, cid, bosherr.WrapError(err, "finding existing stemcell record in repo")
	}
	if found {
		event = bmlog.Event{
			Stage:   "uploading stemcell",
			Total:   2,
			Task:    "Uploading",
			Index:   2,
			State:   bmlog.Skipped,
			Message: "stemcell already uploaded",
		}
		m.eventLogger.AddEvent(event)

		return stemcell, cid, nil
	}

	event = bmlog.Event{
		Stage: "uploading stemcell",
		Total: 2,
		Task:  "Uploading",
		Index: 2,
		State: bmlog.Started,
	}
	m.eventLogger.AddEvent(event)

	cid, err = m.infrastructure.CreateStemcell(stemcell)
	if err != nil {
		event = bmlog.Event{
			Stage: "uploading stemcell",
			Total: 2,
			Task:  "Uploading",
			Index: 2,
			State: bmlog.Failed,
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
		//TODO: delete stemcell from infrastructure

		event = bmlog.Event{
			Stage: "uploading stemcell",
			Total: 2,
			Task:  "Uploading",
			Index: 2,
			State: bmlog.Failed,
		}
		m.eventLogger.AddEvent(event)

		return Stemcell{}, "", bosherr.WrapError(
			err,
			"saving stemcell record in repo (record=%s, stemcell=%s)",
			cid,
			stemcell,
		)
	}

	event = bmlog.Event{
		Stage: "uploading stemcell",
		Total: 2,
		Task:  "Uploading",
		Index: 2,
		State: bmlog.Finished,
	}
	m.eventLogger.AddEvent(event)

	return stemcell, cid, nil
}
