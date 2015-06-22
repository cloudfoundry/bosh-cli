package agentclient

import (
	"github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/agentclient/applyspec"
	"github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/settings"
)

//go:generate mockgen -source=agent_client_interface.go -package=mocks -destination=mocks/mocks.go -imports=.=github.com/cloudfoundry/bosh-agent/agentclient

type AgentClient interface {
	Ping() (string, error)
	Stop() error
	Apply(applyspec.ApplySpec) error
	Start() error
	GetState() (AgentState, error)
	MountDisk(string) error
	UnmountDisk(string) error
	ListDisk() ([]string, error)
	MigrateDisk() error
	CompilePackage(packageSource BlobRef, compiledPackageDependencies []BlobRef) (compiledPackageRef BlobRef, err error)
	UpdateSettings(settings settings.Settings) error
}

type AgentState struct {
	JobState string
}

type BlobRef struct {
	Name        string
	Version     string
	BlobstoreID string
	SHA1        string
}
