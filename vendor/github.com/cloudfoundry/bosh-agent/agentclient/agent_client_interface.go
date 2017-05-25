package agentclient

import (
	"github.com/cloudfoundry/bosh-agent/agentclient/applyspec"
	"github.com/cloudfoundry/bosh-agent/settings"
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
	DeleteARPEntries(ips []string) error
	SyncDNS(blobID, sha1 string, version uint64) (string, error)
	UpdateSettings(settings.UpdateSettings) error
	RunScript(scriptName string, options map[string]interface{}) error
	SSH(cmd string, params SSHParams) error
}

type SSHParams struct {
	UserRegex string `json:"user_regex"`
	User      string
	PublicKey string `json:"public_key"`
}

type SSHResult struct {
	Command       string `json:"command"`
	Status        string `json:"status"`
	IP            string `json:"ip,omitempty"`
	HostPublicKey string `json:"host_public_key,omitempty"`
}

type AgentState struct {
	JobState     string
	NetworkSpecs map[string]NetworkSpec
}

type NetworkSpec struct {
	IP string `json:"ip"`
}

type BlobRef struct {
	Name        string
	Version     string
	BlobstoreID string
	SHA1        string
}
