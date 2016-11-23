package resource

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type ResourceImpl struct {
	name        string
	fingerprint string

	archivePath string
	archiveSHA1 string

	expectToExist bool
	archive       Archive
}

func NewResource(name, fp string, archive Archive) *ResourceImpl {
	return &ResourceImpl{
		name:        name,
		fingerprint: fp,
		archive:     archive,
	}
}

func NewExistingResource(name, fp, sha1 string) *ResourceImpl {
	return &ResourceImpl{
		name:          name,
		fingerprint:   fp,
		archiveSHA1:   sha1,
		expectToExist: true,
	}
}

func NewResourceWithBuiltArchive(name, fp, path, sha1 string) *ResourceImpl {
	return &ResourceImpl{
		name:          name,
		fingerprint:   fp,
		archivePath:   path,
		archiveSHA1:   sha1,
		expectToExist: true,
	}
}

func (r *ResourceImpl) Name() string        { return r.name }
func (r *ResourceImpl) Fingerprint() string { return r.fingerprint }

func (r *ResourceImpl) ArchivePath() string {
	if len(r.archivePath) == 0 {
		errMsg := "Internal inconsistency: Resource '%s/%s' must be found or built before getting its archive path"
		panic(fmt.Sprintf(errMsg, r.name, r.fingerprint))
	}
	return r.archivePath
}

func (r *ResourceImpl) ArchiveSHA1() string {
	if len(r.archiveSHA1) == 0 {
		errMsg := "Internal inconsistency: Resource '%s/%s' must be found or built before getting its archive SHA1"
		panic(fmt.Sprintf(errMsg, r.name, r.fingerprint))
	}
	return r.archiveSHA1
}

func (r *ResourceImpl) Build(devIndex, finalIndex ArchiveIndex) error {
	if r.hasArchive() {
		return nil
	}

	err := r.findAndAttach(devIndex, finalIndex, r.expectToExist)
	if err != nil {
		return err
	}

	if r.hasArchive() {
		return nil
	}

	path, sha1, err := r.archive.Build(r.fingerprint)
	if err != nil {
		return err
	}

	newDevPath, newDevSHA1, err := devIndex.Add(r.name, r.fingerprint, path, sha1)
	if err != nil {
		return err
	}

	r.attachArchive(newDevPath, newDevSHA1)

	return nil
}

func (r *ResourceImpl) Finalize(finalIndex ArchiveIndex) error {
	finalPath, finalSHA1, err := finalIndex.Find(r.name, r.fingerprint)
	if err != nil {
		return err
	} else if len(finalPath) > 0 {
		r.attachArchive(finalPath, finalSHA1)
		return nil
	}

	_, _, err = finalIndex.Add(r.name, r.fingerprint, r.ArchivePath(), r.ArchiveSHA1())

	return err
}

func (r *ResourceImpl) findAndAttach(devIndex, finalIndex ArchiveIndex, errIfNotFound bool) error {
	devPath, devSHA1, err := devIndex.Find(r.name, r.fingerprint)
	if err != nil {
		return err
	} else if len(devPath) > 0 {
		r.attachArchive(devPath, devSHA1)
		return nil
	}

	finalPath, finalSHA1, err := finalIndex.Find(r.name, r.fingerprint)
	if err != nil {
		return err
	} else if len(finalPath) > 0 {
		r.attachArchive(finalPath, finalSHA1)
		return nil
	}

	if errIfNotFound {
		return bosherr.Errorf("Expected to find '%s/%s'", r.name, r.fingerprint)
	}

	return nil
}

func (r *ResourceImpl) attachArchive(path, sha1 string) {
	r.archivePath = path
	r.archiveSHA1 = sha1
}

func (r *ResourceImpl) hasArchive() bool {
	return len(r.archivePath) > 0 && len(r.archiveSHA1) > 0
}
