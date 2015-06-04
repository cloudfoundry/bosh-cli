package templatescompiler

import (
	biindex "github.com/cloudfoundry/bosh-init/index"
	bosherr "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/errors"
	bireljob "github.com/cloudfoundry/bosh-init/release/job"
)

type TemplateRecord struct {
	BlobID   string
	BlobSHA1 string
}

type TemplatesRepo interface {
	Save(bireljob.Job, TemplateRecord) error
	Find(bireljob.Job) (TemplateRecord, bool, error)
}

type templatesRepo struct {
	index biindex.Index
}

func NewTemplatesRepo(index biindex.Index) TemplatesRepo {
	return templatesRepo{index: index}
}

func (tr templatesRepo) Save(job bireljob.Job, record TemplateRecord) error {
	err := tr.index.Save(tr.jobKey(job), record)

	if err != nil {
		return bosherr.WrapError(err, "Saving job templates")
	}

	return nil
}

func (tr templatesRepo) Find(job bireljob.Job) (TemplateRecord, bool, error) {
	var record TemplateRecord

	err := tr.index.Find(tr.jobKey(job), &record)
	if err != nil {
		if err == biindex.ErrNotFound {
			return record, false, nil
		}

		return record, false, bosherr.WrapError(err, "Finding job templates")
	}

	return record, true, nil
}

type jobTemplateKey struct {
	JobName        string
	JobFingerprint string
}

func (tr templatesRepo) jobKey(job bireljob.Job) jobTemplateKey {
	return jobTemplateKey{
		JobName:        job.Name,
		JobFingerprint: job.Fingerprint,
	}
}
