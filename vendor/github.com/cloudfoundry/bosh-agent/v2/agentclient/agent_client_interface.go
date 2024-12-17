package agentclient

import "github.com/cloudfoundry/bosh-agent/v2/agentclient/applyspec"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fakes/fake_agent_client.go . AgentClient

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
	SetUpSSH(username string, publicKey string) (SSHResult, error)
	CleanUpSSH(username string) (SSHResult, error)
	BundleLogs(owningUser string, logType string, filters []string) (BundleLogsResult, error)
	RemoveFile(path string) error
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

type SSHResult struct {
	Command       string
	Status        string
	Ip            string
	HostPublicKey string
}

type BundleLogsResult struct {
	LogsTarPath  string
	SHA512Digest string
}
