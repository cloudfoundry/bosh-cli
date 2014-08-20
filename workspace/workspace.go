package workspace

import (
	"encoding/json"
	"os"
	"path"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
)

type Workspace interface {
	BlobstorePath() string
	PackagesPath() string
	MicroBoshPath() string
}

type workspace struct {
	fileSystem    boshsys.FileSystem
	config        bmconfig.Config
	uuidGenerator boshuuid.Generator
	containingDir string
	uuid          string
}

func NewWorkspace(
	fs boshsys.FileSystem,
	config bmconfig.Config,
	uuidGenerator boshuuid.Generator,
	containingDir string,
) (Workspace, error) {

	ws := &workspace{
		fileSystem:    fs,
		config:        config,
		containingDir: containingDir,
		uuidGenerator: uuidGenerator,
	}

	err := ws.initialize()
	if err != nil {
		return &workspace{}, err
	}
	return ws, nil
}

type DeploymentFile struct {
	UUID string
}

func (w *workspace) initialize() error {
	manifestFile := w.config.Deployment
	manifestContainingPath := path.Dir(manifestFile)

	uuid, err := w.getUUID()
	if err != nil {
		return bosherr.WrapError(err, "Getting UUID")
	}
	w.uuid = uuid

	hasDeployment := manifestFile != ""

	if hasDeployment {
		deploymentFile := path.Join(manifestContainingPath, "deployment.json")
		deploymentFileExists := w.fileSystem.FileExists(deploymentFile)

		if !deploymentFileExists {
			uuid, err := w.uuidGenerator.Generate()
			if err != nil {
				return bosherr.WrapError(err, "Generating UUID")
			}
			w.uuid = uuid

			deploymentJSON, err := json.MarshalIndent(DeploymentFile{UUID: uuid}, "", " ")
			if err != nil {
				return bosherr.WrapError(err, "Marshaling deployment file content")
			}

			err = w.fileSystem.WriteFile(deploymentFile, deploymentJSON)
			if err != nil {
				return bosherr.WrapError(err, "Writing deployment file")
			}
			err = w.fileSystem.MkdirAll(w.BlobstorePath(), os.ModePerm)
			if err != nil {
				return bosherr.WrapError(err, "Creating blobs dir")
			}
		}
	}

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

func (w *workspace) getUUID() (string, error) {
	manifestContainingPath := path.Dir(w.config.Deployment)

	deploymentFile := path.Join(manifestContainingPath, "deployment.json")
	deploymentFileExists := w.fileSystem.FileExists(deploymentFile)

	if deploymentFileExists {
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

	return "", nil
}
