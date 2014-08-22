package workspace

import (
	"os"
	"path"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

const (
	tagString = "Workspace"
)

type Workspace interface {
	Initialize(string) error
}

type workspace struct {
	fileSystem    boshsys.FileSystem
	containingDir string
	uuid          string
	logger        boshlog.Logger
}

func NewWorkspace(
	fs boshsys.FileSystem,
	containingDir string,
	logger boshlog.Logger,
) (Workspace, error) {

	ws := &workspace{
		fileSystem:    fs,
		containingDir: containingDir,
		logger:        logger,
	}

	return ws, nil
}

func (w *workspace) Initialize(uuid string) error {
	blobstoreDir := path.Join(w.containingDir, ".bosh_micro", uuid, "blobs")
	w.logger.Debug(tagString, "Making new blobstore directory `%s'", blobstoreDir)
	err := w.fileSystem.MkdirAll(blobstoreDir, os.ModePerm)
	if err != nil {
		return bosherr.WrapError(err, "Creating blobs dir")
	}

	return nil
}
