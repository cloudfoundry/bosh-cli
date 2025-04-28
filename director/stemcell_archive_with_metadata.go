package director

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"

	"gopkg.in/yaml.v2"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type StemcellArchiveWithMetadata struct {
	path     string
	fileName string
	fs       boshsys.FileSystem
}

func NewFSStemcellArchive(path string, fs boshsys.FileSystem) (stemcellArchive StemcellArchive) {
	return StemcellArchiveWithMetadata{path, "stemcell.MF", fs}
}

func (a StemcellArchiveWithMetadata) Info() (StemcellMetadata, error) {
	bytes, err := a.readMFBytes()
	if err != nil {
		return StemcellMetadata{}, err
	}

	return a.extractNameAndVersion(bytes)
}

func (a StemcellArchiveWithMetadata) File() (UploadFile, error) {
	file, err := a.fs.OpenFile(a.path, os.O_RDONLY, 0)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Opening archive")
	}

	return file, nil
}

func (a StemcellArchiveWithMetadata) readMFBytes() ([]byte, error) {
	file, err := a.fs.OpenFile(a.path, os.O_RDONLY, 0)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Opening archive")
	}

	defer file.Close() //nolint:errcheck

	gr, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}

	defer gr.Close() //nolint:errcheck

	tr := tar.NewReader(gr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Reading next tar entry")
		}

		if hdr.Name == a.fileName || hdr.Name == "./"+a.fileName {
			bytes, err := io.ReadAll(tr)
			if err != nil {
				return nil, bosherr.WrapErrorf(err, "Reading '%s' entry", a.fileName)
			}

			return bytes, nil
		}
	}

	return nil, bosherr.Errorf("Missing '%s'", a.fileName)
}

func (a StemcellArchiveWithMetadata) extractNameAndVersion(bytes []byte) (metadata StemcellMetadata, err error) {
	err = yaml.Unmarshal(bytes, &metadata)
	if err != nil {
		return metadata, bosherr.WrapErrorf(err, "Unmarshalling '%s'", a.fileName)
	}

	return metadata, nil
}
