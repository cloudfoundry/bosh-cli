package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"
)

const (
	tagString = "Workspace"
)

type Workspace interface {
	Initialize(string) error
	Load(string) error
	BlobstorePath() string
	PackagesPath() string
	MicroBoshPath() string
}

type workspace struct {
	fileSystem    boshsys.FileSystem
	uuidGenerator boshuuid.Generator
	containingDir string
	uuid          string
	logger        boshlog.Logger
}

func NewWorkspace(
	fs boshsys.FileSystem,
	uuidGenerator boshuuid.Generator,
	containingDir string,
	logger boshlog.Logger,
) (Workspace, error) {

	ws := &workspace{
		fileSystem:    fs,
		containingDir: containingDir,
		uuidGenerator: uuidGenerator,
		logger:        logger,
	}

	return ws, nil
}

type DeploymentFile struct {
	UUID string
}

func (w *workspace) Initialize(manifestFile string) error {
	w.logger.Debug(tagString, "manifest file `%s'", manifestFile)

	manifestContainingPath := path.Dir(manifestFile)
	w.logger.Debug(tagString, "manifestContainingPath `%s'", manifestContainingPath)

	uuid, err := w.uuidGenerator.Generate()
	if err != nil {
		return bosherr.WrapError(err, "Generating UUID")
	}
	w.uuid = uuid
	w.logger.Debug(tagString, "Generated new UUID `%s'", uuid)

	deploymentFile := path.Join(manifestContainingPath, "deployment.json")
	deploymentJSON, err := json.MarshalIndent(DeploymentFile{UUID: uuid}, "", " ")
	if err != nil {
		return bosherr.WrapError(err, "Marshaling deployment file content")
	}

	w.logger.Debug(tagString, "Writing to file `%s'", deploymentFile)
	err = w.fileSystem.WriteFile(deploymentFile, deploymentJSON)
	if err != nil {
		return bosherr.WrapError(err, "Writing deployment file")
	}

	w.logger.Debug(tagString, "Making new blobstore directory `%s'", w.BlobstorePath())
	err = w.fileSystem.MkdirAll(w.BlobstorePath(), os.ModePerm)
	if err != nil {
		return bosherr.WrapError(err, "Creating blobs dir")
	}

	return nil
}

func (w *workspace) Load(manifestFile string) error {
	uuid, err := w.getUUID(manifestFile)
	if err != nil {
		return bosherr.WrapError(err, "Loading workspace")
	}
	w.uuid = uuid

	return nil
}

func (w *workspace) BlobstorePath() string {
	return path.Join(w.MicroBoshPath(), "blobs")
}

func (w *workspace) PackagesPath() string {
	return path.Join(w.MicroBoshPath(), "packages")
}

func (w *workspace) MicroBoshPath() string {
	return path.Join(w.containingDir, ".bosh_micro", w.uuid)
}

func (w *workspace) getUUID(manifestFile string) (string, error) {
	manifestContainingPath := path.Dir(manifestFile)

	deploymentFile := path.Join(manifestContainingPath, "deployment.json")
	w.logger.Debug(tagString, fmt.Sprintf("Getting UUID from `%s'", deploymentFile))

	deploymentRawContent, err := w.fileSystem.ReadFile(deploymentFile)
	if err != nil {
		return "", bosherr.WrapError(err, "Reading deployment file")
	}

	deploymentContent := &DeploymentFile{}
	err = json.Unmarshal(deploymentRawContent, deploymentContent)

	if err != nil {
		return "", bosherr.WrapError(err, "Unmarshalling deployment file")
	}
	return deploymentContent.UUID, nil
}
