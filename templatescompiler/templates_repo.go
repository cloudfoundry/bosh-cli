package templatescompiler

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmindex "github.com/cloudfoundry/bosh-init/index"
	bmreljob "github.com/cloudfoundry/bosh-init/release/job"
)

type TemplateRecord struct {
	BlobID   string
	BlobSHA1 string
}

type TemplatesRepo interface {
	Save(bmreljob.Job, TemplateRecord) error
	Find(bmreljob.Job) (TemplateRecord, bool, error)
}

type templatesRepo struct {
	index bmindex.Index
}

func NewTemplatesRepo(index bmindex.Index) TemplatesRepo {
	return templatesRepo{index: index}
}

func (tr templatesRepo) Save(job bmreljob.Job, record TemplateRecord) error {
	err := tr.index.Save(tr.jobKey(job), record)

	if err != nil {
		return bosherr.WrapError(err, "Saving job templates")
	}

	return nil
}

func (tr templatesRepo) Find(job bmreljob.Job) (TemplateRecord, bool, error) {
	var record TemplateRecord

	err := tr.index.Find(tr.jobKey(job), &record)
	if err != nil {
		if err == bmindex.ErrNotFound {
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

func (tr templatesRepo) jobKey(job bmreljob.Job) jobTemplateKey {
	return jobTemplateKey{
		JobName:        job.Name,
		JobFingerprint: job.Fingerprint,
	}
}
