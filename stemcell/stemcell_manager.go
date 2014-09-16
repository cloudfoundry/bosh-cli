package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
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
	infrastructure Infrastructure
}

func NewManager(fs boshsys.FileSystem, reader Reader, repo Repo, infrastructure Infrastructure) Manager {
	return &manager{
		fs:             fs,
		reader:         reader,
		repo:           repo,
		infrastructure: infrastructure,
	}
}

// Upload
// 1) extracts the tarball & parses its manifest,
// 2) uploads the stemcell to the infrastructure (if needed),
// 3) saves a record of the uploaded stemcell in the repo
func (m *manager) Upload(tarballPath string) (Stemcell, CID, error) {
	// unpack stemcell tarball
	tmpDir, err := m.fs.TempDir("stemcell-manager")
	if err != nil {
		return Stemcell{}, "", bosherr.WrapError(err, "creating temp dir for stemcell extraction")
	}
	defer m.fs.RemoveAll(tmpDir)

	// parse/reads stemcell manifest into Stemcell object
	stemcell, err := m.reader.Read(tarballPath, tmpDir)
	if err != nil {
		return Stemcell{}, "", bosherr.WrapError(err, "reading extracted stemcell manifest in `%s'", tmpDir)
	}

	// TODO: check the stemcell repo to make sure the stemcell (with exact sha1) has not already been uploaded
	// m.repo.Find

	cid, err := m.infrastructure.CreateStemcell(stemcell)
	if err != nil {
		return Stemcell{}, "", bosherr.WrapError(
			err,
			"creating stemcell with infrastructure (infrastructure=%s, stemcell=%s)",
			m.infrastructure,
			stemcell,
		)
	}

	err = m.repo.Save(stemcell, cid)
	if err != nil {
		return Stemcell{}, "", bosherr.WrapError(
			err,
			"saving stemcell record in repo (record=%s, stemcell=%s)",
			cid,
			stemcell,
		)
	}

	return stemcell, cid, nil
}
