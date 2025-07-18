package opts

import (
	"time"

	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"github.com/cppforlife/go-patch/patch"

	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
)

type BoshOpts struct {
	// -----> Global options

	VersionOpt func() error `long:"version" short:"v" description:"Show CLI version"`

	ConfigPathOpt string `long:"config" description:"Config file path" env:"BOSH_CONFIG" default:"~/.bosh/config"`

	EnvironmentOpt string    `long:"environment" short:"e" description:"Director environment name or URL" env:"BOSH_ENVIRONMENT"`
	CACertOpt      CACertArg `long:"ca-cert"               description:"Director CA certificate path or value" env:"BOSH_CA_CERT"`
	Sha2           bool      `long:"sha2"                  description:"Use SHA256 checksums" env:"BOSH_SHA2"`
	Parallel       int       `long:"parallel" description:"The max number of parallel operations" default:"5"`

	// Specify client credentials
	ClientOpt       string `long:"client"        description:"Override username or UAA client"        env:"BOSH_CLIENT"`
	ClientSecretOpt string `long:"client-secret" description:"Override password or UAA client secret" env:"BOSH_CLIENT_SECRET"`

	DeploymentOpt string `long:"deployment" short:"d" description:"Deployment name" env:"BOSH_DEPLOYMENT"`

	// Output formatting
	ColumnOpt         []ColumnOpt `long:"column"                    description:"Filter to show only given column(s), use the --column flag for each column you wish to include"`
	JSONOpt           bool        `long:"json"                      description:"Output as JSON"`
	TTYOpt            bool        `long:"tty"                       description:"Force TTY-like output" env:"BOSH_TTY"`
	NoColorOpt        bool        `long:"no-color"                  description:"Toggle colorized output"`
	NonInteractiveOpt bool        `long:"non-interactive" short:"n" description:"Don't ask for user input" env:"BOSH_NON_INTERACTIVE"`

	Help       HelpOpts `command:"help" description:"Show this help message"`
	Completion NoOpts   `command:"completion" description:"Generate the autocompletion script for bosh for the specified shell."`

	// -----> Director management

	// Environments
	Environment  EnvironmentOpts  `command:"environment"  alias:"env"  description:"Show environment"`
	Environments EnvironmentsOpts `command:"environments" alias:"envs" description:"List environments"`
	CreateEnv    CreateEnvOpts    `command:"create-env"                description:"Create or update BOSH environment"`
	DeleteEnv    DeleteEnvOpts    `command:"delete-env"                description:"Delete BOSH environment"`
	StopEnv      StopEnvOpts      `command:"stop-env"                  description:"Stop BOSH environment"`
	StartEnv     StartEnvOpts     `command:"start-env"                 description:"Start BOSH environment"`
	AliasEnv     AliasEnvOpts     `command:"alias-env"                 description:"Alias environment to save URL and CA certificate"`
	UnaliasEnv   UnaliasEnvOpts   `command:"unalias-env"               description:"Remove an aliased environment"`

	// Authentication
	LogIn  LogInOpts  `command:"log-in"  alias:"l" alias:"login"  description:"Log in"` //nolint:staticcheck
	LogOut LogOutOpts `command:"log-out"           alias:"logout" description:"Log out"`

	// Tasks
	Task        TaskOpts        `command:"task"         alias:"t"   description:"Show task status and start tracking its output"`
	Tasks       TasksOpts       `command:"tasks"        alias:"ts"  description:"List running or recent tasks"`
	CancelTask  CancelTaskOpts  `command:"cancel-task"  alias:"ct"  description:"Cancel task at its next checkpoint"`
	CancelTasks CancelTasksOpts `command:"cancel-tasks" alias:"cts" description:"Cancel tasks at their next checkpoints"`

	// Misc
	Locks   LocksOpts   `command:"locks"    description:"List current locks"`
	CleanUp CleanUpOpts `command:"clean-up" description:"Clean up old unused resources except orphaned disks"`
	Curl    CurlOpts    `command:"curl"     description:"Make an HTTP request to the Director"`

	// Config
	Config       ConfigOpts       `command:"config" alias:"c" description:"Show current config for either ID or both type and name"`
	Configs      ConfigsOpts      `command:"configs" alias:"cs" description:"List configs"`
	UpdateConfig UpdateConfigOpts `command:"update-config" alias:"uc" description:"Update config"`
	DeleteConfig DeleteConfigOpts `command:"delete-config" alias:"dc" description:"Delete config"`
	DiffConfig   DiffConfigOpts   `command:"diff-config" description:"Diff two configs by ID or content"`

	// Cloud config
	CloudConfig       CloudConfigOpts       `command:"cloud-config"        alias:"cc"  description:"Show current cloud config"`
	UpdateCloudConfig UpdateCloudConfigOpts `command:"update-cloud-config" alias:"ucc" description:"Update current cloud config"`

	// CPI Config
	CPIConfig       CPIConfigOpts       `command:"cpi-config"        description:"Show current CPI config"`
	UpdateCPIConfig UpdateCPIConfigOpts `command:"update-cpi-config" description:"Update current CPI config"`

	// Runtime config
	RuntimeConfig       RuntimeConfigOpts       `command:"runtime-config"        alias:"rc"  description:"Show current runtime config"`
	UpdateRuntimeConfig UpdateRuntimeConfigOpts `command:"update-runtime-config" alias:"urc" description:"Update current runtime config"`

	// Deployments
	Deployment       DeploymentOpts       `command:"deployment"        alias:"dep"             description:"Show deployment information"`
	Deployments      DeploymentsOpts      `command:"deployments"       alias:"ds" alias:"deps" description:"List deployments"` //nolint:staticcheck
	DeleteDeployment DeleteDeploymentOpts `command:"delete-deployment" alias:"deld"            description:"Delete deployment"`

	Deploy   DeployOpts   `command:"deploy"   alias:"d"   description:"Update deployment"`
	Manifest ManifestOpts `command:"manifest" alias:"man" description:"Show deployment manifest"`

	Interpolate InterpolateOpts `command:"interpolate" alias:"int" description:"Interpolates variables into a manifest"`

	// Events
	Events EventsOpts `command:"events" description:"List events"`
	Event  EventOpts  `command:"event" description:"Show event details"`

	// Stemcells
	Stemcells            StemcellsOpts              `command:"stemcells"       alias:"ss"   description:"List stemcells"`
	InspectLocalStemcell InspectStemcellTarballOpts `command:"inspect-local-stemcell"     description:"Display information from stemcell metadata"`
	UploadStemcell       UploadStemcellOpts         `command:"upload-stemcell" alias:"us"   description:"Upload stemcell"`
	DeleteStemcell       DeleteStemcellOpts         `command:"delete-stemcell" alias:"dels" description:"Delete stemcell"`
	RepackStemcell       RepackStemcellOpts         `command:"repack-stemcell"              description:"Repack stemcell"`

	// Releases
	Releases            ReleasesOpts            `command:"releases"        alias:"rs"   description:"List releases"`
	UploadRelease       UploadReleaseOpts       `command:"upload-release"  alias:"ur"   description:"Upload release"`
	ExportRelease       ExportReleaseOpts       `command:"export-release"               description:"Export the compiled release to a tarball"`
	InspectRelease      InspectReleaseOpts      `command:"inspect-release"              description:"List release contents such as jobs"`
	InspectLocalRelease InspectLocalReleaseOpts `command:"inspect-local-release"     description:"Display information from release metadata"`
	DeleteRelease       DeleteReleaseOpts       `command:"delete-release"  alias:"delr" description:"Delete release"`

	// Errands
	Errands   ErrandsOpts   `command:"errands"    alias:"es" description:"List errands"`
	RunErrand RunErrandOpts `command:"run-errand"            description:"Run errand"`

	// Disks
	Disks      DisksOpts      `command:"disks"       description:"List disks"`
	AttachDisk AttachDiskOpts `command:"attach-disk" description:"Attach disk to an instance"`
	DeleteDisk DeleteDiskOpts `command:"delete-disk" description:"Delete disk"`
	OrphanDisk OrphanDiskOpts `command:"orphan-disk" description:"Orphan disk"`

	// Networks
	Networks      NetworksOpts      `command:"networks"       description:"List networks"`
	DeleteNetwork DeleteNetworkOpts `command:"delete-network" description:"Delete network"`

	// Snapshots
	Snapshots       SnapshotsOpts       `command:"snapshots"        description:"List snapshots"`
	TakeSnapshot    TakeSnapshotOpts    `command:"take-snapshot"    description:"Take snapshot"`
	DeleteSnapshot  DeleteSnapshotOpts  `command:"delete-snapshot"  description:"Delete snapshot"`
	DeleteSnapshots DeleteSnapshotsOpts `command:"delete-snapshots" description:"Delete all snapshots in a deployment"`

	// Instances
	Instances          InstancesOpts          `command:"instances"       alias:"is"                     description:"List all instances in a deployment"`
	VMs                VMsOpts                `command:"vms"                                            description:"List all VMs in all deployments"`
	UpdateResurrection UpdateResurrectionOpts `command:"update-resurrection"                            description:"Enable/disable resurrection"`
	Ignore             IgnoreOpts             `command:"ignore"                                         description:"Ignore an instance"`
	Unignore           UnignoreOpts           `command:"unignore"                                       description:"Unignore an instance"`
	CloudCheck         CloudCheckOpts         `command:"cloud-check"     alias:"cck" alias:"cloudcheck" description:"Cloud consistency check and interactive repair"` //nolint:staticcheck
	CreateRecoveryPlan CreateRecoveryPlanOpts `command:"create-recovery-plan"                           description:"Interactively generate a recovery plan for disaster repair"`
	Recover            RecoverOpts            `command:"recover"                           description:"Apply a recovery plan for disaster repair"`
	OrphanedVMs        OrphanedVMsOpts        `command:"orphaned-vms"                                   description:"List all the orphaned VMs in all deployments"`

	// Instance management
	Logs     LogsOpts     `command:"logs"      description:"Fetch logs from instance(s)"`
	Start    StartOpts    `command:"start"     description:"Start instance(s)"`
	Stop     StopOpts     `command:"stop"      description:"Stop instance(s)"`
	Restart  RestartOpts  `command:"restart"   description:"Restart instance(s)"`
	Recreate RecreateOpts `command:"recreate"  description:"Recreate instance(s)"`
	DeleteVM DeleteVMOpts `command:"delete-vm" description:"Delete VM"`
	Pcap     PcapOpts     `command:"pcap"      description:"Capture network packets on instance(s)"`

	// SSH instance
	SSH SSHOpts `command:"ssh" description:"SSH into instance(s)"`
	SCP SCPOpts `command:"scp" description:"SCP to/from instance(s)"`

	// -----> Release authoring

	// Release creation
	InitRelease     InitReleaseOpts     `command:"init-release"                description:"Initialize release"`
	ResetRelease    ResetReleaseOpts    `command:"reset-release"               description:"Reset release"`
	GenerateJob     GenerateJobOpts     `command:"generate-job"                description:"Generate job"`
	GeneratePackage GeneratePackageOpts `command:"generate-package"            description:"Generate package"`
	CreateRelease   CreateReleaseOpts   `command:"create-release"   alias:"cr" description:"Create release"`
	VendorPackage   VendorPackageOpts   `command:"vendor-package"              description:"Vendor package"`

	Sha1ifyRelease Sha1ifyReleaseOpts `command:"sha1ify-release"  description:"Convert release tarball to use SHA1"`
	Sha2ifyRelease Sha2ifyReleaseOpts `command:"sha2ify-release"  description:"Convert release tarball to use SHA256"`

	FinalizeRelease FinalizeReleaseOpts `command:"finalize-release"               description:"Create final release from dev release tarball"`

	// Blob management
	Blobs       BlobsOpts       `command:"blobs"        description:"List blobs"`
	AddBlob     AddBlobOpts     `command:"add-blob"     description:"Add blob"`
	RemoveBlob  RemoveBlobOpts  `command:"remove-blob"  description:"Remove blob"`
	SyncBlobs   SyncBlobsOpts   `command:"sync-blobs"   description:"Sync blobs"`
	UploadBlobs UploadBlobsOpts `command:"upload-blobs" description:"Upload blobs"`

	Variables VariablesOpts `command:"variables" alias:"vars" description:"List variables"`
}

type HelpOpts struct {
	cmd
}
type NoOpts struct{}

// Original bosh-init

type CreateEnvOpts struct {
	Args CreateEnvArgs `positional-args:"true" required:"true"`
	VarFlags
	OpsFlags
	SkipDrain               bool   `long:"skip-drain" description:"Skip running drain and pre-stop scripts"`
	StatePath               string `long:"state" value-name:"PATH" description:"State file path"`
	Recreate                bool   `long:"recreate" description:"Recreate VM in deployment"`
	RecreatePersistentDisks bool   `long:"recreate-persistent-disks" description:"Recreate persistent disks in the deployment"`
	PackageDir              string `long:"package-dir" value-name:"DIR" description:"Package cache location override"`
	cmd
}

type CreateEnvArgs struct {
	Manifest FileBytesWithPathArg `positional-arg-name:"PATH" description:"Path to a manifest file"`
}

type DeleteEnvOpts struct {
	Args DeleteEnvArgs `positional-args:"true" required:"true"`
	VarFlags
	OpsFlags
	SkipDrain  bool   `long:"skip-drain" description:"Skip running drain and pre-stop scripts"`
	StatePath  string `long:"state" value-name:"PATH" description:"State file path"`
	PackageDir string `long:"package-dir" value-name:"DIR" description:"Package cache location override"`
	cmd
}

type StopEnvOpts struct {
	Args StartStopEnvArgs `positional-args:"true" required:"true"`
	VarFlags
	OpsFlags
	SkipDrain bool   `long:"skip-drain" description:"Skip running drain and pre-stop scripts"`
	StatePath string `long:"state" value-name:"PATH" description:"State file path"`
	cmd
}

type StartEnvOpts struct {
	Args StartStopEnvArgs `positional-args:"true" required:"true"`
	VarFlags
	OpsFlags
	StatePath string `long:"state" value-name:"PATH" description:"State file path"`
	cmd
}

type DeleteEnvArgs struct {
	Manifest FileBytesWithPathArg `positional-arg-name:"PATH" description:"Path to a manifest file"`
}

type StartStopEnvArgs struct {
	Manifest FileBytesWithPathArg `positional-arg-name:"PATH" description:"Path to a manifest file"`
}

// Environment

type EnvironmentOpts struct {
	Details bool `long:"details" description:"Show director's certificates details"`
	cmd
}

type EnvironmentsOpts struct {
	cmd
}

type AliasEnvOpts struct {
	Args AliasEnvArgs `positional-args:"true" required:"true"`

	URL    string
	CACert CACertArg

	cmd
}

type AliasEnvArgs struct {
	Alias string `positional-arg-name:"ALIAS" description:"Environment alias"`
}

type UnaliasEnvOpts struct {
	Args UnaliasEnvArgs `positional-args:"true" required:"true"`

	cmd
}

type UnaliasEnvArgs struct {
	Alias string `positional-arg-name:"ALIAS" description:"Environment alias"`
}

type LogInOpts struct {
	cmd
}

type LogOutOpts struct {
	cmd
}

// Tasks

type TaskOpts struct {
	Args TaskArgs `positional-args:"true"`

	Event  bool `long:"event"  description:"Track event log"`
	CPI    bool `long:"cpi"    description:"Track CPI log"`
	Debug  bool `long:"debug"  description:"Track debug log"`
	Result bool `long:"result" description:"Track result log"`

	All        bool `long:"all" short:"a" description:"Include all task types (ssh, logs, vms, etc)"`
	Deployment string

	cmd
}

type TaskArgs struct {
	ID int `positional-arg-name:"ID"`
}

type TasksOpts struct {
	Recent     *int `long:"recent" short:"r" description:"Show 30 recent tasks. Use '=' to specify the number of tasks to show" optional:"true" optional-value:"30"`
	All        bool `long:"all" short:"a" description:"Include all task types (ssh, logs, vms, etc)"`
	Deployment string

	cmd
}

type CancelTaskOpts struct {
	Args TaskArgs `positional-args:"true" required:"true"`
	cmd
}

type CancelTasksOpts struct {
	Types      []string `long:"type"  short:"t" description:"task types to cancel (cck_scan_and_fix, cck_apply, update_release, update_deployment, vms, etc) (default is all types)" optional:"true"`
	States     []string `long:"state" short:"s" description:"task states to cancel (queued, processing) (default: queued)" optional:"true"`
	Deployment string

	cmd
}

// Misc

type LocksOpts struct {
	cmd
}

type CleanUpOpts struct {
	All               bool `long:"all" description:"Clean up all unused resources including all orphaned disks"`
	DryRun            bool `long:"dry-run" description:"Print out the resources that will be deleted but does not delete anything"`
	KeepOrphanedDisks bool `long:"keep-orphaned-disks" description:"Keep orphaned disks even with '--all'"`

	cmd
}

type AttachDiskOpts struct {
	Args AttachDiskArgs `positional-args:"true" required:"true"`

	DiskProperties string `long:"disk-properties" description:"Disk properties to use for the new disk. Use 'copy' to copy the properties from the currently attached disk" optional:"true"`

	cmd
}

type AttachDiskArgs struct {
	Slug    boshdir.InstanceSlug `positional-arg-name:"INSTANCE-GROUP/INSTANCE-ID"`
	DiskCID string               `positional-arg-name:"DISK-CID"`
}

type InterpolateOpts struct {
	Args InterpolateArgs `positional-args:"true" required:"true"`

	VarFlags
	OpsFlags

	Path            patch.Pointer `long:"path" value-name:"OP-PATH" description:"Extract value out of template (e.g.: /private_key)"`
	VarErrors       bool          `long:"var-errs"                  description:"Expect all variables to be found, otherwise error"`
	VarErrorsUnused bool          `long:"var-errs-unused"           description:"Expect all variables to be used, otherwise error"`

	cmd
}

type InterpolateArgs struct {
	Manifest FileBytesArg `positional-arg-name:"PATH" description:"Path to a template that will be interpolated"`
}

// Config

type ConfigOpts struct {
	Args ConfigArgs `positional-args:"true"`
	Name string     `long:"name" description:"Config name"`
	Type string     `long:"type" description:"Config type"`

	cmd
}

type ConfigArgs struct {
	ID string `positional-arg-name:"ID" description:"Config ID"`
}

type ConfigsOpts struct {
	Name   string `long:"name" description:"Config name" optional:"true"`
	Type   string `long:"type" description:"Config type" optional:"true"`
	Recent int    `long:"recent" short:"r" description:"Number of configs to show" optional:"true" optional-value:"1" default:"1"`

	cmd
}

type DiffConfigOpts struct {
	FromID      string       `long:"from-id" description:"ID of first config to compare"`
	ToID        string       `long:"to-id" description:"ID of second config to compare"`
	FromContent FileBytesArg `long:"from-content" description:"Path to first config file to compare"`
	ToContent   FileBytesArg `long:"to-content" description:"Path to second config file to compare"`
	cmd
}

type UpdateConfigOpts struct {
	Args             UpdateConfigArgs `positional-args:"true" required:"true"`
	Type             string           `long:"type" required:"true" description:"Config type, e.g. 'cloud', 'runtime', or 'cpi'"`
	Name             string           `long:"name" required:"true" description:"Config name"`
	ExpectedLatestId string           `long:"expected-latest-id" description:"Expected ID of latest config"`
	VarFlags
	OpsFlags
	cmd
}

type UpdateConfigArgs struct {
	Config FileBytesArg `positional-arg-name:"PATH" description:"Path to a YAML config file"`
}

type DeleteConfigOpts struct {
	Args DeleteConfigArgs `positional-args:"true"`
	Type string           `long:"type" description:"Config type, e.g. 'cloud', 'runtime', or 'cpi'"`
	Name string           `long:"name" description:"Config name"`

	cmd
}

type DeleteConfigArgs struct {
	ID string `positional-arg-name:"ID" description:"Config ID"`
}

// Cloud config

type CloudConfigOpts struct {
	Name string `long:"name" description:"Cloud-Config name (default: '')" default:""`
	cmd
}

type UpdateCloudConfigOpts struct {
	Args UpdateCloudConfigArgs `positional-args:"true" required:"true"`
	VarFlags
	OpsFlags

	Name string `long:"name" description:"Cloud-Config name (default: '')" default:""`

	cmd
}

type UpdateCloudConfigArgs struct {
	CloudConfig FileBytesArg `positional-arg-name:"PATH" description:"Path to a cloud config file"`
}

type CPIConfigOpts struct {
	cmd
}

type UpdateCPIConfigOpts struct {
	Args UpdateCPIConfigArgs `positional-args:"true" required:"true"`
	VarFlags
	OpsFlags

	NoRedact bool `long:"no-redact" description:"Show non-redacted manifest diff"`

	cmd
}

type UpdateCPIConfigArgs struct {
	CPIConfig FileBytesArg `positional-arg-name:"PATH" description:"Path to a CPI config file"`
}

// Runtime config

type RuntimeConfigOpts struct {
	Name string `long:"name" description:"Runtime-Config name (default: '')" default:""`
	cmd
}

type UpdateRuntimeConfigOpts struct {
	Args UpdateRuntimeConfigArgs `positional-args:"true" required:"true"`
	VarFlags
	OpsFlags

	NoRedact    bool   `long:"no-redact" description:"Show non-redacted manifest diff"`
	Name        string `long:"name" description:"Runtime-Config name (default: '')" default:""`
	FixReleases bool   `long:"fix-releases" description:"Reupload releases in config and replace corrupt or missing jobs/packages"`

	cmd
}

type UpdateRuntimeConfigArgs struct {
	RuntimeConfig FileBytesArg `positional-arg-name:"PATH" description:"Path to a runtime config file"`
}

// Deployments

type DeploymentOpts struct {
	cmd
}

type DeploymentsOpts struct {
	cmd
}

type DeployOpts struct {
	Args DeployArgs `positional-args:"true" required:"true"`

	VarFlags
	OpsFlags

	NoRedact bool `long:"no-redact" description:"Show non-redacted manifest diff"`

	Recreate                bool                `long:"recreate"                                description:"Recreate all VMs in deployment"`
	RecreatePersistentDisks bool                `long:"recreate-persistent-disks"               description:"Recreate all persistent disks in deployment"`
	Fix                     bool                `long:"fix"                                     description:"Recreate an instance with an unresponsive agent instead of erroring"`
	FixReleases             bool                `long:"fix-releases"                            description:"Reupload releases in manifest and replace corrupt or missing jobs/packages"`
	SkipDrain               []boshdir.SkipDrain `long:"skip-drain" value-name:"[INSTANCE-GROUP[/INSTANCE-ID]]"  description:"Skip running drain and pre-stop scripts for specific instance groups" optional:"true" optional-value:"*"`
	SkipUploadReleases      bool                `long:"skip-upload-releases"                  description:"Skips the upload procedure for releases"`

	Canaries    string `long:"canaries" description:"Override manifest values for canaries"`
	MaxInFlight string `long:"max-in-flight" description:"Override manifest values for max_in_flight"`

	DryRun               bool `long:"dry-run" description:"Renders job templates without altering deployment"`
	ForceLatestVariables bool `long:"force-latest-variables" description:"Retrieve the latest variable values from the config server regardless of their update strategy"`

	cmd
}

type DeployArgs struct {
	Manifest FileBytesArg `positional-arg-name:"PATH" description:"Path to a manifest file"`
}

type ManifestOpts struct {
	cmd
}

type DeleteDeploymentOpts struct {
	Force bool `long:"force" description:"Ignore errors"`
	cmd
}

// Events

type EventsOpts struct {
	BeforeID   string `long:"before-id"    description:"Show events with ID less than the given ID"`
	Before     string `long:"before"       description:"Show events before the given timestamp (ex: 2016-05-08 17:26:32)"`
	After      string `long:"after"        description:"Show events after the given timestamp (ex: 2016-05-08 17:26:32)"`
	Deployment string
	Task       string `long:"task"         description:"Show events with the given task ID"`
	Instance   string `long:"instance"     description:"Show events with given instance"`
	User       string `long:"event-user"   description:"Show events with given user"`
	Action     string `long:"action"       description:"Show events with given action"`
	ObjectType string `long:"object-type"  description:"Show events with given object type"`
	ObjectName string `long:"object-name"  description:"Show events with given object name"`

	cmd
}

type EventOpts struct {
	Args EventArgs `positional-args:"true" required:"true"`

	cmd
}

type EventArgs struct {
	ID string `positional-arg-name:"ID"`
}

// Stemcells

type StemcellsOpts struct {
	cmd
}

type UploadStemcellOpts struct {
	Args UploadStemcellArgs `positional-args:"true" required:"true"`

	Fix bool `long:"fix" description:"Replaces the stemcell if already exists"`

	Name    string     `long:"name"     description:"Name used in existence check (is not used with local stemcell file)"`
	Version VersionArg `long:"version"  description:"Version used in existence check (is not used with local stemcell file)"`

	SHA1 string `long:"sha1" description:"SHA1 of the remote stemcell (is not used with local files)"`

	cmd
}

type UploadStemcellArgs struct {
	URL URLArg `positional-arg-name:"URL" description:"Path to a local file or URL"`
}

type DeleteStemcellOpts struct {
	Args DeleteStemcellArgs `positional-args:"true" required:"true"`

	Force bool `long:"force" description:"Ignore errors"`

	cmd
}

type DeleteStemcellArgs struct {
	Slug boshdir.StemcellSlug `positional-arg-name:"NAME/VERSION"`
}

type RepackStemcellOpts struct {
	Args            RepackStemcellArgs `positional-args:"true" required:"true"`
	Name            string             `long:"name" description:"Repacked stemcell name"`
	CloudProperties string             `long:"cloud-properties" description:"Repacked stemcell cloud properties"`
	EmptyImage      bool               `long:"empty-image" description:"Pack zero byte file instead of image"`
	Format          []string           `long:"format" description:"Repacked stemcell formats. Can be used multiple times. Overrides existing formats."`
	Version         string             `long:"version" description:"Repacked stemcell version"`

	cmd
}

type RepackStemcellArgs struct {
	PathToStemcell string  `positional-arg-name:"PATH-TO-STEMCELL" description:"Path to stemcell"`
	PathToResult   FileArg `positional-arg-name:"PATH-TO-RESULT" description:"Path to repacked stemcell"`
}

type InspectStemcellTarballOpts struct {
	Args InspectStemcellTarballArgs `positional-args:"true" required:"true"`
	cmd
}

type InspectStemcellTarballArgs struct {
	PathToStemcell string `positional-arg-name:"PATH-TO-STEMCELL" description:"Path to stemcell"`
}

// Releases

type ReleasesOpts struct {
	cmd
}

type UploadReleaseOpts struct {
	Args UploadReleaseArgs `positional-args:"true"`

	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`

	Rebase bool `long:"rebase" description:"Rebases this release onto the latest version known by the Director"`

	Fix bool `long:"fix" description:"Replaces corrupt and missing jobs and packages"`

	Name    string     `long:"name"     description:"Name used in existence check (is not used with local release file)"`
	Version VersionArg `long:"version"  description:"Version used in existence check (is not used with local release file)"`

	SHA1 string `long:"sha1" description:"SHA1 of the remote release (is not used with local files)"`

	Stemcell boshdir.OSVersionSlug `long:"stemcell" value-name:"OS/VERSION" description:"Stemcell that the release is compiled against (applies to remote releases)"`

	Release boshrel.Release

	cmd
}

type UploadReleaseArgs struct {
	URL URLArg `positional-arg-name:"URL" description:"Path to a local file or URL"`
}

type DeleteReleaseOpts struct {
	Args DeleteReleaseArgs `positional-args:"true" required:"true"`

	Force bool `long:"force" description:"Ignore errors"`

	cmd
}

type DeleteReleaseArgs struct {
	Slug boshdir.ReleaseOrSeriesSlug `positional-arg-name:"NAME[/VERSION]"`
}

type ExportReleaseOpts struct {
	Args ExportReleaseArgs `positional-args:"true" required:"true"`

	Directory DirOrCWDArg `long:"dir" description:"Destination directory" default:"."`

	Jobs []string `long:"job" description:"Name of job to export"`
	cmd
}

type ExportReleaseArgs struct {
	ReleaseSlug   boshdir.ReleaseSlug   `positional-arg-name:"NAME/VERSION"`
	OSVersionSlug boshdir.OSVersionSlug `positional-arg-name:"OS/VERSION"`
}

type InspectReleaseOpts struct {
	Args InspectReleaseArgs `positional-args:"true" required:"true"`
	cmd
}

type InspectReleaseArgs struct {
	Slug boshdir.ReleaseSlug `positional-arg-name:"NAME/VERSION"`
}

type InspectLocalReleaseOpts struct {
	Args InspectLocalReleaseArgs `positional-args:"true" required:"true"`
	cmd
}

type InspectLocalReleaseArgs struct {
	PathToRelease string `positional-arg-name:"PATH-TO-RELEASE" description:"Path to release"`
}

// Errands

type ErrandsOpts struct {
	cmd
}

type InstanceGroupOrInstanceSlugFlags struct {
	Slugs []boshdir.InstanceGroupOrInstanceSlug `long:"instance" value-name:"INSTANCE-GROUP[/INSTANCE-ID]" description:"Instance or group the errand should run on (must specify errand by release job name)"`
}

type RunErrandOpts struct {
	Args RunErrandArgs `positional-args:"true" required:"true"`

	InstanceGroupOrInstanceSlugFlags

	KeepAlive   bool `long:"keep-alive" description:"Use existing VM to run an errand and keep it after completion"`
	WhenChanged bool `long:"when-changed" description:"Run errand only if errand configuration has changed or if the previous run was unsuccessful"`

	DownloadLogs  bool        `long:"download-logs" description:"Download logs"`
	LogsDirectory DirOrCWDArg `long:"logs-dir" description:"Destination directory for logs" default:"."`

	cmd
}

type RunErrandArgs struct {
	Name string `positional-arg-name:"NAME"`
}

// Networks

type NetworksOpts struct {
	Orphaned bool `long:"orphaned" short:"o" description:"List orphaned networks"`
	cmd
}

type DeleteNetworkOpts struct {
	Args DeleteNetworkArgs `positional-args:"true" required:"true"`
	cmd
}

type DeleteNetworkArgs struct {
	Name string `positional-arg-name:"name"`
}

// Disks

type DisksOpts struct {
	Orphaned bool `long:"orphaned" short:"o" description:"List orphaned disks"`
	cmd
}

type DeleteDiskOpts struct {
	Args DeleteDiskArgs `positional-args:"true" required:"true"`
	cmd
}

type DeleteDiskArgs struct {
	CID string `positional-arg-name:"CID"`
}

type OrphanDiskOpts struct {
	Args OrphanDiskArgs `positional-args:"true" required:"true"`
	cmd
}
type OrphanDiskArgs struct {
	CID string `positional-arg-name:"CID"`
}

// Snapshots

type SnapshotsOpts struct {
	Args InstanceSlugArgs `positional-args:"true"`
	cmd
}

type TakeSnapshotOpts struct {
	Args InstanceSlugArgs `positional-args:"true"`
	cmd
}

type DeleteSnapshotOpts struct {
	Args DeleteSnapshotArgs `positional-args:"true" required:"true"`
	cmd
}

type DeleteSnapshotArgs struct {
	CID string `positional-arg-name:"CID"`
}

type DeleteVMOpts struct {
	Args DeleteVMArgs `positional-args:"true" required:"true"`
	cmd
}

type DeleteVMArgs struct {
	CID string `positional-arg-name:"CID"`
}

type DeleteSnapshotsOpts struct {
	cmd
}

type InstanceSlugArgs struct {
	Slug boshdir.InstanceSlug `positional-arg-name:"INSTANCE-GROUP/INSTANCE-ID"`
}

// Instances

type InstancesOpts struct {
	Details    bool `long:"details" short:"i" description:"Show details including VM CID, persistent disk CID, etc."`
	Vitals     bool `long:"vitals"            description:"Show vitals"`
	Processes  bool `long:"ps"      short:"p" description:"Show processes"`
	Failing    bool `long:"failing" short:"f" description:"Only show failing instances"`
	Deployment string
	cmd
}

type VMsOpts struct {
	Vitals          bool `long:"vitals"            description:"Show vitals"`
	CloudProperties bool `long:"cloud-properties"  description:"Show cloud properties"`
	Deployment      string
	cmd
}

type CloudCheckOpts struct {
	Auto        bool     `long:"auto"       short:"a" description:"Resolve problems automatically"`
	Resolutions []string `long:"resolution"           description:"Apply resolution of given type (e.g.: 'recreate_vm'). Can be used multiple times."`
	Report      bool     `long:"report"     short:"r" description:"Only generate report; don't attempt to resolve problems"`
	cmd
}

type CreateRecoveryPlanOpts struct {
	Args CreateRecoveryPlanArgs `positional-args:"true" required:"true"`
	cmd
}

type CreateRecoveryPlanArgs struct {
	RecoveryPlan FileArg `positional-arg-name:"PATH" description:"Create recovery plan file at path"`
}

type RecoverOpts struct {
	Args RecoverArgs `positional-args:"true" required:"true"`
	cmd
}

type RecoverArgs struct {
	RecoveryPlan FileArg `positional-arg-name:"PATH" description:"Path to a recovery plan file"`
}

type OrphanedVMsOpts struct {
	cmd
}

// Instance management

type UpdateResurrectionOpts struct {
	Args UpdateResurrectionArgs `positional-args:"true" required:"true"`
	cmd
}

type UpdateResurrectionArgs struct {
	Enabled BoolArg `positional-arg-name:"on|off"`
}

type IgnoreOpts struct {
	Args InstanceSlugArgs `positional-args:"true" required:"true"`
	cmd
}

type UnignoreOpts struct {
	Args InstanceSlugArgs `positional-args:"true" required:"true"`
	cmd
}

type LogsOpts struct {
	Args AllOrInstanceGroupOrInstanceSlugArgs `positional-args:"true"`

	Directory DirOrCWDArg `long:"dir" description:"Destination directory" default:"."`

	Follow bool `long:"follow" short:"f" description:"Follow logs via SSH"`
	Num    int  `long:"num"              description:"Last number of lines"`
	Quiet  bool `long:"quiet"  short:"q" description:"Suppresses printing of headers when multiple files are being examined"`

	Jobs    []string `long:"job"   description:"Limit to only specific jobs (can only be used in combination with --follow)"`
	Filters []string `long:"only"  description:"Filter logs to specific jobs (comma-separated). Like --jobs, but for when --follow is not being used"`
	Agent   bool     `long:"agent" description:"Include only agent logs"`
	System  bool     `long:"system" description:"Include only system logs"`
	All     bool     `long:"all-logs" description:"Include all logs (agent, system, and job logs)"`

	GatewayFlags

	CreateEnvAuthFlags

	cmd
}

type CreateEnvAuthFlags struct {
	TargetDirector bool   `long:"director"             description:"Target the command at the BOSH director (or other type of VM deployed via create-env)"`
	Endpoint       string `long:"agent-endpoint"       description:"Address to connect to the agent's HTTPS endpoint (used with --director)"      env:"BOSH_AGENT_ENDPOINT"`
	Certificate    string `long:"agent-certificate"    description:"CA certificate to validate the agent's HTTPS endpoint (used with --director)" env:"BOSH_AGENT_CERTIFICATE"`
}

type StartOpts struct {
	Args AllOrInstanceGroupOrInstanceSlugArgs `positional-args:"true"`

	Canaries    string `long:"canaries" description:"Override manifest values for canaries"`
	MaxInFlight string `long:"max-in-flight" description:"Override manifest values for max_in_flight"`
	Converge    bool   `long:"converge" description:"Converge the deployment state before running action (default)"`
	NoConverge  bool   `long:"no-converge" description:"Act only on specified instance"`

	cmd
}

type StopOpts struct {
	Args AllOrInstanceGroupOrInstanceSlugArgs `positional-args:"true"`

	Soft bool `long:"soft" description:"Stop process only (default)"`
	Hard bool `long:"hard" description:"Delete VM (but keep persistent disk)"`

	SkipDrain bool `long:"skip-drain" description:"Skip running drain and pre-stop scripts"`

	Canaries    string `long:"canaries" description:"Override manifest values for canaries"`
	MaxInFlight string `long:"max-in-flight" description:"Override manifest values for max_in_flight"`

	Converge   bool `long:"converge" description:"Converge the deployment state before running action (default)"`
	NoConverge bool `long:"no-converge" description:"Act only on specified instance"`

	cmd
}

type RestartOpts struct {
	Args AllOrInstanceGroupOrInstanceSlugArgs `positional-args:"true"`

	SkipDrain bool `long:"skip-drain" description:"Skip running drain and pre-stop scripts"`

	Canaries    string `long:"canaries" description:"Override manifest values for canaries"`
	MaxInFlight string `long:"max-in-flight" description:"Override manifest values for max_in_flight"`

	Converge   bool `long:"converge" description:"Converge the deployment state before running action (default)"`
	NoConverge bool `long:"no-converge" description:"Act only on specified instance"`

	cmd
}

type RecreateOpts struct {
	Args AllOrInstanceGroupOrInstanceSlugArgs `positional-args:"true"`

	SkipDrain bool `long:"skip-drain" description:"Skip running drain and pre-stop scripts"`
	Fix       bool `long:"fix"        description:"Recreate an instance with an unresponsive agent instead of erroring"`

	Canaries    string `long:"canaries" description:"Override manifest values for canaries"`
	MaxInFlight string `long:"max-in-flight" description:"Override manifest values for max_in_flight"`

	DryRun bool `long:"dry-run" description:"Renders job templates without altering deployment"`

	Converge   bool `long:"converge" description:"Converge the deployment state before running action (default)"`
	NoConverge bool `long:"no-converge" description:"Act only on specified instance"`

	cmd
}

type PcapOpts struct {
	Args MultiAllOrInstanceGroupOrInstanceSlugArgs `positional-args:"true"`

	Interface   string        `long:"interface" short:"i" description:"Specifies the network interface to listen on." default:"eth0" required:"false"`
	Filter      string        `long:"filter" short:"f" description:"Filter to apply when running tcpdump."`
	SnapLength  uint32        `long:"snaplen" short:"s" description:"Snarf snaplen bytes of data from each packet rather than the default of 65535 bytes." default:"65535"`
	Output      string        `long:"output" short:"o" description:"File to write pcap to." required:"true"`
	StopTimeout time.Duration `long:"stop-timeout" description:"Timeout to wait for data to flush before session stop." default:"5s"`

	GatewayFlags

	cmd
}

type MultiAllOrInstanceGroupOrInstanceSlugArgs struct {
	Slugs []boshdir.AllOrInstanceGroupOrInstanceSlug `positional-arg-name:"INSTANCE-GROUP[/INSTANCE-ID]"`
}

type AllOrInstanceGroupOrInstanceSlugArgs struct {
	Slug boshdir.AllOrInstanceGroupOrInstanceSlug `positional-arg-name:"INSTANCE-GROUP[/INSTANCE-ID]"`
}

// SSH instance

type SshSlugArgs struct {
	Slug boshdir.AllOrInstanceGroupOrInstanceSlug `positional-arg-name:"INSTANCE-GROUP[/INSTANCE-ID] | IP"`
}

type SSHOpts struct {
	Args SshSlugArgs `positional-args:"true"`

	Command []string         `long:"command" short:"c" description:"Command"`
	RawOpts TrimmedSpaceArgs `long:"opts"              description:"Options to pass through to SSH"`

	Results bool `long:"results" short:"r" description:"Collect results into a table instead of streaming"`

	PrivateKey FileBytesWithPathArg `long:"private-key" short:"i" description:"SSH using authorized key"`

	Username string `long:"username" short:"l" description:"Login name for authorized key" default:"vcap"`

	GatewayFlags

	CreateEnvAuthFlags

	cmd
}

type SCPOpts struct {
	Args SCPArgs `positional-args:"true" required:"true"`

	Recursive bool `long:"recursive" short:"r" description:"Recursively copy entire directories. Note that symbolic links encountered are followed in the tree traversal"`

	PrivateKey FileBytesWithPathArg `long:"private-key" short:"i" description:"SSH using authorized key"`

	Username string `long:"username" short:"l" description:"Login name for authorized key" default:"vcap"`

	GatewayFlags

	CreateEnvAuthFlags

	cmd
}

type SCPArgs struct {
	Paths []string `positional-arg-name:"PATH"`
}

type GatewayFlags struct {
	UUIDGen boshuuid.Generator

	Disable bool `long:"gw-disable" description:"Disable usage of gateway connection" env:"BOSH_GW_DISABLE"`

	Username       string `long:"gw-user"        description:"Username for gateway connection" env:"BOSH_GW_USER"`
	Host           string `long:"gw-host"        description:"Host for gateway connection" env:"BOSH_GW_HOST"`
	PrivateKeyPath string `long:"gw-private-key" description:"Private key path for gateway connection" env:"BOSH_GW_PRIVATE_KEY"` // todo private file?

	SOCKS5Proxy string `long:"gw-socks5" description:"SOCKS5 URL" env:"BOSH_ALL_PROXY"`
}

// Release creation

type InitReleaseOpts struct {
	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`

	Git bool `long:"git" description:"Initialize git repository"`

	cmd
}

type ResetReleaseOpts struct {
	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`

	cmd
}

type GenerateJobOpts struct {
	Args GenerateJobArgs `positional-args:"true" required:"true"`

	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`

	cmd
}

type GenerateJobArgs struct {
	Name string `positional-arg-name:"NAME"`
}

type GeneratePackageOpts struct {
	Args GeneratePackageArgs `positional-args:"true" required:"true"`

	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`

	cmd
}

type GeneratePackageArgs struct {
	Name string `positional-arg-name:"NAME"`
}

type VendorPackageOpts struct {
	Args VendorPackageArgs `positional-args:"true" required:"true"`

	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`
	Prefix    string      `long:"prefix" description:"Prefix to add to the package name" default:""`

	cmd
}

type VendorPackageArgs struct {
	PackageName string      `positional-arg-name:"PACKAGE"`
	URL         DirOrCWDArg `positional-arg-name:"SRC-DIR" default:"."`
}

type Sha1ifyReleaseOpts struct {
	Args RedigestReleaseArgs `positional-args:"true"`

	cmd
}

type Sha2ifyReleaseOpts struct {
	Args RedigestReleaseArgs `positional-args:"true"`

	cmd
}

type RedigestReleaseArgs struct {
	Path        string  `positional-arg-name:"PATH"`
	Destination FileArg `positional-arg-name:"DESTINATION"`
}

type CreateReleaseOpts struct {
	Args CreateReleaseArgs `positional-args:"true"`

	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`

	Name             string     `long:"name"               description:"Custom release name"`
	Version          VersionArg `long:"version"            description:"Custom release version (e.g.: 1.0.0, 1.0-beta.2+dev.10)"`
	TimestampVersion bool       `long:"timestamp-version"  description:"Create release with the timestamp as the dev version (e.g.: 1+dev.TIMESTAMP)"`

	Final   bool    `long:"final"   description:"Make it a final release"`
	Tarball FileArg `long:"tarball" description:"Create release tarball at path (e.g. /tmp/release.tgz)"`
	Force   bool    `long:"force"   description:"Ignore Git dirty state check"`

	cmd
}

type CreateReleaseArgs struct {
	Manifest FileBytesWithPathArg `positional-arg-name:"PATH"`
}

type FinalizeReleaseOpts struct {
	Args FinalizeReleaseArgs `positional-args:"true" required:"true"`

	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`

	Name    string     `long:"name"    description:"Custom release name"`
	Version VersionArg `long:"version" description:"Custom release version (e.g.: 1.0.0, 1.0-beta.2+dev.10)"`

	Force bool `long:"force" description:"Ignore Git dirty state check"`

	cmd
}

type FinalizeReleaseArgs struct {
	Path string `positional-arg-name:"PATH"`
}

// Blobs

type BlobsOpts struct {
	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`
	cmd
}

type AddBlobOpts struct {
	Args AddBlobArgs `positional-args:"true" required:"true"`

	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`

	cmd
}

type AddBlobArgs struct {
	Path      string `positional-arg-name:"PATH"`
	BlobsPath string `positional-arg-name:"BLOBS-PATH"`
}

type RemoveBlobOpts struct {
	Args RemoveBlobArgs `positional-args:"true" required:"true"`

	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`

	cmd
}

type RemoveBlobArgs struct {
	BlobsPath string `positional-arg-name:"BLOBS-PATH"`
}

type SyncBlobsOpts struct {
	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`
	cmd
}

type UploadBlobsOpts struct {
	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`
	cmd
}

type CurlOpts struct {
	Args CurlArgs `positional-args:"true" required:"true"`

	Method  string       `long:"method" short:"X" description:"HTTP method" default:"GET"`
	Headers []CurlHeader `long:"header" short:"H" description:"HTTP header in 'name: value' format"`
	Body    FileBytesArg `long:"body"             description:"HTTP request body (path)"`

	ShowHeaders bool `long:"show-headers" short:"i"   description:"Show HTTP headers"`

	cmd
}

type CurlArgs struct {
	Path string `positional-arg-name:"PATH" description:"URL path which can include query string"`
}

// MessageOpts is used for version and help flags
type MessageOpts struct {
	Message string
}

type VariablesOpts struct {
	Deployment string
	cmd
}

type cmd struct{}

// Execute is necessary for each command to be goflags.Commander
func (c cmd) Execute(_ []string) error {
	panic("Unreachable")
}
