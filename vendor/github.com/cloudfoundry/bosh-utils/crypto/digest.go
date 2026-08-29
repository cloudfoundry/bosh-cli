package crypto

import (
	"fmt"
	"io"
	"strings"

	"os"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type digestImpl struct {
	algorithm Algorithm
	digest    string
}

func NewDigest(algorithm Algorithm, digest string) digestImpl {
	return digestImpl{
		algorithm: algorithm,
		digest:    strings.TrimPrefix(digest, algorithm.Name()+":"),
	}
}

func (d digestImpl) Algorithm() Algorithm { return d.algorithm }

func (d digestImpl) String() string {
	if d.algorithm.Name() == DigestAlgorithmSHA1.Name() {
		return d.digest
	}

	return fmt.Sprintf("%s:%s", d.algorithm.Name(), d.digest)
}

func (d digestImpl) Verify(reader io.Reader) error {
	computedDigest, err := d.Algorithm().CreateDigest(reader)
	if err != nil {
		return bosherr.WrapError(err, "Computing digest from stream")
	}

	if d.String() != computedDigest.String() {
		return bosherr.Errorf("Expected stream to have digest '%s' but was '%s'", d.String(), computedDigest.String())
	}

	return nil
}

func (d digestImpl) VerifyFilePath(filePath string, fs boshsys.FileSystem) error {
	file, err := fs.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		return bosherr.WrapErrorf(err, "calculating digest of '%s'", filePath)
	}
	defer func() {
		_ = file.Close()
	}()
	return d.Verify(file)
}
