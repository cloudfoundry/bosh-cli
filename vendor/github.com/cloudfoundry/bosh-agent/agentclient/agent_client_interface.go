package agentclient

import "github.com/cloudfoundry/bosh-agent/agentclient/applyspec"

//go:generate counterfeiter -o fakes/fake_agent_client.go agent_client_interface.go AgentClient

type AgentClient interface {
	Ping() (string, error)
	Stop() error
	Drain(string) (int64, error)
	Apply(applyspec.ApplySpec) error
	Start() error
	GetState() (AgentState, error)
	AddPersistentDisk(string, interface{}) error
	RemovePersistentDisk(string) error
	MountDisk(string) error
	UnmountDisk(string) error
	ListDisk() ([]string, error)
	MigrateDisk() error
	CompilePackage(packageSource BlobRef, compiledPackageDependencies []BlobRef) (compiledPackageRef BlobRef, err error)
	DeleteARPEntries(ips []string) error
	SyncDNS(blobID, sha1 string, version uint64) (string, error)
	RunScript(scriptName string, options map[string]interface{}) error
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
