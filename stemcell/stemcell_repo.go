package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type Repo interface {
	Save(stemcellPath string) (stemcell Stemcell, extractedPath string, err error)
}

type repo struct {
	fs     boshsys.FileSystem
	reader Reader
}

func NewRepo(fs boshsys.FileSystem, reader Reader) repo {
	return repo{
		fs:     fs,
		reader: reader,
	}
}

func (s repo) Save(stemcellPath string) (Stemcell, string, error) {
	stemcellExtractedPath, err := s.fs.TempDir("extracted-stemcell")
	if err != nil {
		return Stemcell{}, "", bosherr.WrapError(err, "Creating tempDir")
	}

	stemcell, err := s.reader.Read(stemcellPath, stemcellExtractedPath)
	if err != nil {
		s.fs.RemoveAll(stemcellExtractedPath)
		return Stemcell{}, "", bosherr.WrapError(err, "Reading stemcell")
	}

	//TODO: call the CPI to store info
	// save info locally

	return stemcell, stemcellExtractedPath, nil
}
