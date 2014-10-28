package instanceupdater

import (
	"crypto/sha1"
	"fmt"
	"hash"
	"io"
	"os"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/agentclient"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type applySpecCreator struct {
	fs boshsys.FileSystem
}

type ApplySpecCreator interface {
	Create(
		bmstemcell.ApplySpec,
		string,
		string,
		map[string]interface{},
		string,
		string,
		string,
	) (bmagentclient.ApplySpec, error)
}

func NewApplySpecCreator(fs boshsys.FileSystem) ApplySpecCreator {
	return &applySpecCreator{
		fs: fs,
	}
}

func (c *applySpecCreator) Create(
	stemcellApplySpec bmstemcell.ApplySpec,
	deploymentName string,
	jobName string,
	networksSpec map[string]interface{},
	archivedTemplatesBlobID string,
	archivedTemplatesPath string,
	templatesDir string,
) (bmagentclient.ApplySpec, error) {
	archivedTemplatesSha1, err := c.archivedTemplatesSha1(archivedTemplatesPath)
	if err != nil {
		return bmagentclient.ApplySpec{}, bosherr.WrapError(err, "Calculating archived templates SHA1")
	}

	templatesDirSha1, err := c.templatesDirSha1(templatesDir)
	if err != nil {
		return bmagentclient.ApplySpec{}, bosherr.WrapError(err, "Calculating templates dir SHA1")
	}

	applySpec := bmagentclient.ApplySpec{
		Deployment: deploymentName,
		Index:      0,
		Packages:   c.packagesSpec(stemcellApplySpec.Packages),
		Job:        c.jobSpec(stemcellApplySpec.Job.Templates, jobName),
		Networks:   networksSpec,
		RenderedTemplatesArchive: bmagentclient.RenderedTemplatesArchiveSpec{
			BlobstoreID: archivedTemplatesBlobID,
			SHA1:        archivedTemplatesSha1,
		},
		ConfigurationHash: templatesDirSha1,
	}
	return applySpec, nil
}

func (c *applySpecCreator) packagesSpec(stemcellPackages map[string]bmstemcell.Blob) map[string]bmagentclient.Blob {
	result := map[string]bmagentclient.Blob{}
	for packageName, packageBlob := range stemcellPackages {
		result[packageName] = bmagentclient.Blob{
			Name:        packageBlob.Name,
			Version:     packageBlob.Version,
			SHA1:        packageBlob.SHA1,
			BlobstoreID: packageBlob.BlobstoreID,
		}
	}

	return result
}

func (c *applySpecCreator) jobSpec(stemcellTemplates []bmstemcell.Blob, jobName string) bmagentclient.Job {
	templates := []bmagentclient.Blob{}
	for _, templateBlob := range stemcellTemplates {
		templates = append(templates, bmagentclient.Blob{
			Name:        templateBlob.Name,
			Version:     templateBlob.Version,
			SHA1:        templateBlob.SHA1,
			BlobstoreID: templateBlob.BlobstoreID,
		})
	}

	return bmagentclient.Job{
		Name:      jobName,
		Templates: templates,
	}
}

func (c *applySpecCreator) templatesDirSha1(templatesDir string) (string, error) {
	h := sha1.New()

	c.fs.Walk(templatesDir+"/", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			err := c.populateSha1(path, h)
			if err != nil {
				return bosherr.WrapError(err, "Calculating SHA1 for %s", path)
			}
		}
		return nil
	})

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (c *applySpecCreator) archivedTemplatesSha1(templatesPath string) (string, error) {
	h := sha1.New()
	err := c.populateSha1(templatesPath, h)
	if err != nil {
		return "", bosherr.WrapError(err, "Calculating SHA1 for %s", templatesPath)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (c *applySpecCreator) populateSha1(filePath string, hash hash.Hash) error {
	file, err := c.fs.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		return bosherr.WrapError(err, "Opening file %s for sha1 calculation", filePath)
	}
	defer file.Close()

	_, err = io.Copy(hash, file)
	if err != nil {
		return bosherr.WrapError(err, "Copying file for sha1 calculation")
	}

	return nil
}
