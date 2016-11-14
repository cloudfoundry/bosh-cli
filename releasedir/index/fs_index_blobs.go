package index

import (
	"fmt"
	"os"
	gopath "path"

	boshblob "github.com/cloudfoundry/bosh-utils/blobstore"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshfu "github.com/cloudfoundry/bosh-utils/fileutil"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	bicrypto "github.com/cloudfoundry/bosh-cli/crypto"
)

type FSIndexBlobs struct {
	dirPath  string
	reporter Reporter

	blobstore boshblob.Blobstore
	sha1calc  bicrypto.SHA1Calculator
	fs        boshsys.FileSystem
}

func NewFSIndexBlobs(
	dirPath string,
	reporter Reporter,
	blobstore boshblob.Blobstore,
	sha1calc bicrypto.SHA1Calculator,
	fs boshsys.FileSystem,
) FSIndexBlobs {
	return FSIndexBlobs{
		dirPath:  dirPath,
		reporter: reporter,

		blobstore: blobstore,
		sha1calc:  sha1calc,
		fs:        fs,
	}
}

// Get gurantees that returned file matches requested SHA1.
func (c FSIndexBlobs) Get(name string, blobID string, sha1 string) (string, error) {
	dstPath, err := c.blobPath(sha1)
	if err != nil {
		return "", err
	}

	if c.fs.FileExists(dstPath) {
		actualSHA1, err := c.sha1calc.Calculate(dstPath)
		if err != nil {
			return "", bosherr.WrapErrorf(err, "Calculating SHA1 of local copy '%s'", dstPath)
		}

		if sha1 != actualSHA1 {
			errMsg := "Expected local copy ('%s') of blob '%s' to have SHA1 '%s' but was '%s'"
			return "", bosherr.Errorf(errMsg, dstPath, blobID, sha1, actualSHA1)
		}

		return dstPath, nil
	}

	if c.blobstore != nil && len(blobID) > 0 {
		desc := fmt.Sprintf("sha1=%s", sha1)

		c.reporter.IndexEntryDownloadStarted(name, desc)

		// SHA1 expected to be checked via blobstore
		path, err := c.blobstore.Get(blobID, sha1)
		if err != nil {
			c.reporter.IndexEntryDownloadFinished(name, desc, err)
			return "", bosherr.WrapErrorf(err, "Downloading blob '%s' with SHA1 '%s'", blobID, sha1)
		}

		err = boshfu.NewFileMover(c.fs).Move(path, dstPath)
		if err != nil {
			c.reporter.IndexEntryDownloadFinished(name, desc, err)
			return "", bosherr.WrapErrorf(err, "Moving blob '%s' into cache", blobID)
		}

		c.reporter.IndexEntryDownloadFinished(name, desc, nil)

		return dstPath, nil
	}

	if len(blobID) == 0 {
		return "", bosherr.Errorf("Cannot find blob named '%s' with SHA1 '%s'", name, sha1)
	}

	return "", bosherr.Errorf("Cannot find blob '%s' with SHA1 '%s'", blobID, sha1)
}

// Add adds file to cache and blobstore but does not guarantee
// that file have expected SHA1 when retrieved later.
func (c FSIndexBlobs) Add(name, path, sha1 string) (string, string, error) {
	dstPath, err := c.blobPath(sha1)
	if err != nil {
		return "", "", err
	}

	if !c.fs.FileExists(dstPath) {
		err := c.fs.CopyFile(path, dstPath)
		if err != nil {
			return "", "", bosherr.WrapErrorf(err, "Copying file '%s' with SHA1 '%s' into cache", path, sha1)
		}
	}

	if c.blobstore != nil {
		desc := fmt.Sprintf("sha1=%s", sha1)

		c.reporter.IndexEntryUploadStarted(name, desc)

		blobID, _, err := c.blobstore.Create(path)
		if err != nil {
			c.reporter.IndexEntryUploadFinished(name, desc, err)
			return "", "", bosherr.WrapErrorf(err, "Creating blob for path '%s'", path)
		}

		c.reporter.IndexEntryUploadFinished(name, desc, nil)

		return blobID, dstPath, nil
	}

	return "", dstPath, nil
}

func (c FSIndexBlobs) blobPath(sha1 string) (string, error) {
	absDirPath, err := c.fs.ExpandPath(c.dirPath)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Expanding cache directory")
	}

	err = c.fs.MkdirAll(absDirPath, os.ModePerm)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Creating cache directory")
	}

	return gopath.Join(absDirPath, sha1), nil
}
