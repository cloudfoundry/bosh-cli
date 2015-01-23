package templatescompiler

import (
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmcrypto "github.com/cloudfoundry/bosh-micro-cli/crypto"
)

type RenderedJobListCompressor interface {
	Compress(RenderedJobList) (RenderedJobListArchive, error)
}

type renderedJobListCompressor struct {
	fs             boshsys.FileSystem
	compressor     boshcmd.Compressor
	sha1Calculator bmcrypto.SHA1Calculator
	logger         boshlog.Logger
	logTag         string
}

func NewRenderedJobListCompressor(
	fs boshsys.FileSystem,
	compressor boshcmd.Compressor,
	sha1Calculator bmcrypto.SHA1Calculator,
	logger boshlog.Logger,
) RenderedJobListCompressor {
	return &renderedJobListCompressor{
		fs:             fs,
		compressor:     compressor,
		sha1Calculator: sha1Calculator,
		logger:         logger,
		logTag:         "renderedJobListCompressor",
	}
}

func (c *renderedJobListCompressor) Compress(list RenderedJobList) (RenderedJobListArchive, error) {
	c.logger.Debug(c.logTag, "Compressing rendered job list")

	renderedJobListDir, err := c.fs.TempDir("rendered-job-list-archive")
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating rendered job directory")
	}
	defer func() {
		err := c.fs.RemoveAll(renderedJobListDir)
		if err != nil {
			c.logger.Error(c.logTag, "Failed to delete rendered job list dir: %s", err.Error())
		}
	}()

	// copy rendered job templates into a sub-dir
	for _, renderedJob := range list.All() {
		err = c.fs.CopyDir(renderedJob.Path(), filepath.Join(renderedJobListDir, renderedJob.Job().Name))
		if err != nil {
			return nil, bosherr.WrapError(err, "Creating rendered job directory")
		}
	}

	fingerprint, err := c.sha1Calculator.Calculate(renderedJobListDir)
	if err != nil {
		return nil, bosherr.WrapError(err, "Calculating templates dir SHA1")
	}

	archivePath, err := c.compressor.CompressFilesInDir(renderedJobListDir)
	if err != nil {
		return nil, bosherr.WrapError(err, "Compressing rendered job templates")
	}

	archiveSHA1, err := c.sha1Calculator.Calculate(archivePath)
	if err != nil {
		return nil, bosherr.WrapError(err, "Calculating archived templates SHA1")
	}

	return NewRenderedJobListArchive(list, archivePath, fingerprint, archiveSHA1, c.fs, c.logger), nil
}
