package cmd

import (
	// goflags "github.com/jessevdk/go-flags" // use goflags.Filename

	boshdir "github.com/cloudfoundry/bosh-init/director"
)

type BoshOpts struct {
	// -----> Global options

	VersionOpt func() error `long:"version" short:"v" description:"Show CLI version"`

	ConfigPathOpt string `long:"config" description:"Config file path" env:"BOSH_CONFIG" default:"~/.bosh/config"`

	EnvironmentOpt string `long:"environment" short:"e" description:"Director environment name or URL"`
	CACertOpt      string `long:"ca-cert"               description:"Director CA certificate path or value"`

	// Specify basic credentaials
	UsernameOpt string `long:"user"     description:"Override username" env:"BOSH_USER"`
	PasswordOpt string `long:"password" description:"Override password" env:"BOSH_PASSWORD"`

	// Specify UAA client credentaials
	UAAClientOpt       string `long:"uaa-client"        description:"Override UAA client"        env:"BOSH_CLIENT"`
	UAAClientSecretOpt string `long:"uaa-client-secret" description:"Override UAA client secret" env:"BOSH_CLIENT_SECRET"`

	DeploymentOpt string `long:"deployment" short:"d" description:"Deployment name or manifest path"`

	// Output formatting
	JSONOpt           bool `long:"json"                      description:"Output as JSON"`
	TTYOpt            bool `long:"tty"                       description:"Force TTY-like output"`
	NoColorOpt        bool `long:"no-color"                  description:"Toggle colorized output"`
	NonInteractiveOpt bool `long:"non-interactive" short:"n" description:"Don't ask for user input"`

	// -----> Director management

	// Original bosh-init
	CreateEnv CreateEnvOpts `command:"create-env" description:"Create or update BOSH environment"`
	DeleteEnv DeleteEnvOpts `command:"delete-env" description:"Delete BOSH environment"`

	// Targets
	Environment  EnvironmentOpts  `command:"env"  alias:"target"  alias:"tg"  description:"Set or show current environment"`
	Environments EnvironmentsOpts `command:"envs" alias:"targets" alias:"tgs" description:"List environments"`

	// Authentication
	LogIn  LogInOpts  `command:"log-in"  alias:"login" alias:"l" description:"Log in"`
	LogOut LogOutOpts `command:"log-out" alias:"logout"          description:"Forget saved credentials for Director in the current environment"`

	// Tasks
	Task       TaskOpts       `command:"task"        alias:"t"  description:"Show task status and start tracking its output"`
	Tasks      TasksOpts      `command:"tasks"       alias:"ts" description:"List running or recent tasks"`
	CancelTask CancelTaskOpts `command:"cancel-task" alias:"ct" description:"Cancel task at its next checkpoint"`

	// Misc
	Locks         LocksOpts         `command:"locks"    alias:"ls"                 description:"List current locks"`
	CleanUp       CleanUpOpts       `command:"clean-up" alias:"cl" alias:"cleanup" description:"Clean up releases, stemcells, disks, etc."`
	BackUp        BackUpOpts        `command:"back-up"  alias:"bu" alias:"backup"  description:"Backup the Director to a tarball"`
	BuildManifest BuildManifestOpts `command:"build-manifest"  alias:"bm" hidden:"yes" description:"Interpolates variables into a manifest template."`

	// Cloud config
	CloudConfig       CloudConfigOpts       `command:"cloud-config"        alias:"cc"  description:"Show current cloud config"`
	UpdateCloudConfig UpdateCloudConfigOpts `command:"update-cloud-config" alias:"ucc" description:"Update current cloud config"`

	// Runtime config
	RuntimeConfig       RuntimeConfigOpts       `command:"runtime-config"        alias:"rc"  description:"Show current runtime config"`
	UpdateRuntimeConfig UpdateRuntimeConfigOpts `command:"update-runtime-config" alias:"urc" description:"Update current runtime config"`

	// Deployments
	Deployment       DeploymentOpts       `command:"deployment"        alias:"dep"             description:"Set or show current deployment"`
	Deployments      DeploymentsOpts      `command:"deployments"       alias:"ds" alias:"deps" description:"List deployments"`
	DeleteDeployment DeleteDeploymentOpts `command:"delete-deployment" alias:"deld"            description:"Delete deployment"`

	Deploy   DeployOpts   `command:"deploy"   alias:"d"                                       description:"Deploy according to the currently selected deployment manifest"`
	Manifest ManifestOpts `command:"manifest" alias:"m" alias:"man" alias:"download-manifest" description:"Download deployment manifest locally"`

	// Stemcells
	Stemcells      StemcellsOpts      `command:"stemcells"       alias:"ss" alias:"stems" description:"List stemcells"`
	UploadStemcell UploadStemcellOpts `command:"upload-stemcell" alias:"us"               description:"Upload stemcell"`
	DeleteStemcell DeleteStemcellOpts `command:"delete-stemcell" alias:"dels"             description:"Delete stemcell"`

	// Releases
	Releases       ReleasesOpts       `command:"releases"        alias:"rs" alias:"rels" description:"List releases"`
	UploadRelease  UploadReleaseOpts  `command:"upload-release"  alias:"ur"              description:"Upload release"`
	ExportRelease  ExportReleaseOpts  `command:"export-release"  alias:"expr"            description:"Export the compiled release to a tarball"`
	InspectRelease InspectReleaseOpts `command:"inspect-release" alias:"insr"            description:"List all jobs, packages, and compiled packages associated with a release"`
	DeleteRelease  DeleteReleaseOpts  `command:"delete-release"  alias:"delr"            description:"Delete release"`

	// Errands
	Errands   ErrandsOpts   `command:"errands"    alias:"es" alias:"errs" description:"List errands"`
	RunErrand RunErrandOpts `command:"run-errand" alias:"re"              description:"Run errand"`

	// Disks
	Disks      DisksOpts      `command:"disks"       description:"List disks"`
	DeleteDisk DeleteDiskOpts `command:"delete-disk" description:"Delete disk"`

	// Snapshots
	Snapshots       SnapshotsOpts       `command:"snapshots"        alias:"snaps"    description:"List snapshots"`
	TakeSnapshot    TakeSnapshotOpts    `command:"take-snapshot"    alias:"tsnap"    description:"Take snapshot"`
	DeleteSnapshot  DeleteSnapshotOpts  `command:"delete-snapshot"  alias:"delsnap"  description:"Delete snapshot"`
	DeleteSnapshots DeleteSnapshotsOpts `command:"delete-snapshots" alias:"delsnaps" description:"Delete all snapshots in a deployment"`

	// Instances
	Instances      InstancesOpts      `command:"instances"       alias:"is" alias:"ins"         description:"List all instances in a deployment"`
	VMs            VMsOpts            `command:"vms"                                            description:"List all VMs in all deployments"`
	VMResurrection VMResurrectionOpts `command:"vm-resurrection" alias:"vmr"                    description:"Enable/disable resurrection for a given VM"`
	CloudCheck     CloudCheckOpts     `command:"cloud-check"     alias:"cck" alias:"cloudcheck" description:"Cloud consistency check and interactive repair"`

	// Instance management
	Logs     LogsOpts     `command:"logs"     description:"Fetch logs from instance(s)"`
	Start    StartOpts    `command:"start"    description:"Start instance(s)"`
	Stop     StopOpts     `command:"stop"     description:"Stop instance(s)"`
	Restart  RestartOpts  `command:"restart"  description:"Restart instance(s)"`
	Recreate RecreateOpts `command:"recreate" description:"Recreate instance(s)"`

	// SSH instance
	SSH SSHOpts `command:"ssh" description:"SSH into instance(s)"`
	SCP SCPOpts `command:"scp" description:"SCP to/from instance(s)"`

	// -----> Release authoring

	// Release creation
	InitRelease     InitReleaseOpts     `command:"init-release"                  description:"Initialize release"`
	ResetRelease    ResetReleaseOpts    `command:"reset-release"                 description:"Reset release"`
	GenerateJob     GenerateJobOpts     `command:"generate-job"     alias:"genj" description:"Generate job"`
	GeneratePackage GeneratePackageOpts `command:"generate-package" alias:"genp" description:"Generate package"`
	CreateRelease   CreateReleaseOpts   `command:"create-release"   alias:"cr"   description:"Create release"`
	FinalizeRelease FinalizeReleaseOpts `command:"finalize-release" alias:"finr" description:"Create final release from dev release tarball"`

	// Blob management
	Blobs       BlobsOpts       `command:"blobs"        alias:"bls"  description:"List blobs"`
	AddBlob     AddBlobOpts     `command:"add-blob"     alias:"abl"  description:"Add blob"`
	RemoveBlob  RemoveBlobOpts  `command:"remove-blob"  alias:"rmbl" description:"Remove blob"`
	SyncBlobs   SyncBlobsOpts   `command:"sync-blobs"   alias:"sbls" description:"Sync blobs"`
	UploadBlobs UploadBlobsOpts `command:"upload-blobs" alias:"ubls" description:"Upload blobs"`
}

// Original bosh-init
type CreateEnvOpts struct {
	Args CreateEnvArgs `positional-args:"true" required:"true"`

	VarFlags

	call func() error
}

type CreateEnvArgs struct {
	Manifest FileBytesArg `positional-arg-name:"PATH" description:"Path to a manifest file"`
}

type DeleteEnvOpts struct {
	Args DeleteEnvArgs `positional-args:"true" required:"true"`

	VarFlags

	call func() error
}

type DeleteEnvArgs struct {
	Manifest FileBytesArg `positional-arg-name:"PATH" description:"Path to a manifest file"`
}

func (o CreateEnvOpts) Execute(_ []string) error { return o.call() }
func (o DeleteEnvOpts) Execute(_ []string) error { return o.call() }

// Target
type EnvironmentOpts struct {
	Args EnvironmentArgs `positional-args:"true"`

	CACert FileBytesArg `long:"ca-cert" short:"c" description:"CA certificate path to verify SSL connection to the Director and UAA"`

	call func() error
}

type EnvironmentArgs struct {
	URL   string `positional-arg-name:"URL"   description:"Director URL (e.g.: https://192.168.50.4:25555 or 192.168.50.4)"`
	Alias string `positional-arg-name:"ALIAS" description:"Target alias"`
}

type EnvironmentsOpts struct {
	call func() error
}

type LogInOpts struct {
	call func() error
}

type LogOutOpts struct {
	call func() error
}

func (o EnvironmentOpts) Execute(_ []string) error  { return o.call() }
func (o EnvironmentsOpts) Execute(_ []string) error { return o.call() }
func (o LogInOpts) Execute(_ []string) error        { return o.call() }
func (o LogOutOpts) Execute(_ []string) error       { return o.call() }

// Tasks
type TaskOpts struct {
	Args TaskArgs `positional-args:"true"`

	Event  bool `long:"event"  short:"e" description:"Track event log"`
	CPI    bool `long:"cpi"    short:"c" description:"Track CPI log"`
	Debug  bool `long:"debug"  short:"d" description:"Track debug log"`
	Result bool `long:"result" short:"r" description:"Track result log"`
	Raw    bool `long:"raw"              description:"Track raw log"`

	All bool `long:"all" short:"a" description:"Include all task types (ssh, logs, vms, etc)"`

	call func() error
}

type TaskArgs struct {
	ID int `positional-arg-name:"ID"`
}

type TasksOpts struct {
	Recent *int `long:"recent" short:"r" description:"Number of tasks to show" optional:"true" optional-value:"30"`

	All bool `long:"all" short:"a" description:"Include all task types (ssh, logs, vms, etc)"`

	call func() error
}

type CancelTaskOpts struct {
	Args TaskArgs `positional-args:"true" required:"true"`
	call func() error
}

func (o TaskOpts) Execute(_ []string) error       { return o.call() }
func (o TasksOpts) Execute(_ []string) error      { return o.call() }
func (o CancelTaskOpts) Execute(_ []string) error { return o.call() }

// Misc
type LocksOpts struct {
	call func() error
}

type CleanUpOpts struct {
	All  bool `long:"all" description:"Remove all unused releases, stemcells, etc.; otherwise most recent resources will be kept"`
	call func() error
}

type BackUpOpts struct {
	Args BackUpArgs `positional-args:"true" required:"true"`

	Force bool `long:"force" description:"Overwrite if the backup file already exists"`

	call func() error
}

type BackUpArgs struct {
	Path string `positional-arg-name:"PATH"`
}

type BuildManifestOpts struct {
	Args BuildManifestArgs `positional-args:"true" required:"true"`

	VarFlags

	call func() error
}

type BuildManifestArgs struct {
	Manifest FileBytesArg `positional-arg-name:"PATH" description:"Path to a template that will be interpolated"`
}

func (o LocksOpts) Execute(_ []string) error         { return o.call() }
func (o CleanUpOpts) Execute(_ []string) error       { return o.call() }
func (o BackUpOpts) Execute(_ []string) error        { return o.call() }
func (o BuildManifestOpts) Execute(_ []string) error { return o.call() }

// Cloud config
type CloudConfigOpts struct {
	call func() error
}

type UpdateCloudConfigOpts struct {
	Args UpdateCloudConfigArgs `positional-args:"true" required:"true"`

	VarFlags

	call func() error
}

type UpdateCloudConfigArgs struct {
	CloudConfig FileBytesArg `positional-arg-name:"PATH" description:"Path to a cloud config file"`
}

// Runtime config
type RuntimeConfigOpts struct {
	call func() error
}

type UpdateRuntimeConfigOpts struct {
	Args UpdateRuntimeConfigArgs `positional-args:"true" required:"true"`

	VarFlags

	call func() error
}

type UpdateRuntimeConfigArgs struct {
	RuntimeConfig FileBytesArg `positional-arg-name:"PATH" description:"Path to a runtime config file"`
}

// Deployments
type DeploymentOpts struct {
	Args DeploymentArgs `positional-args:"true"`
	call func() error
}

type DeploymentsOpts struct {
	call func() error
}

type DeployOpts struct {
	Args DeployArgs `positional-args:"true" required:"true"`

	VarFlags

	Recreate   bool `long:"recreate"    description:"Recreate all VMs in deployment"`
	RedactDiff bool `long:"redact-diff" description:"Redact manifest value changes in deployment"`

	SkipDrain boshdir.SkipDrain `long:"skip-drain" description:"Skip running drain scripts"`

	call func() error
}

type DeployArgs struct {
	Manifest FileBytesArg `positional-arg-name:"PATH" description:"Path to a manifest file"`
}

type ManifestOpts struct {
	call func() error
}

type DeleteDeploymentOpts struct {
	Force bool `long:"force" description:"Ignore errors"`
	call  func() error
}

type DeploymentArgs struct {
	NameOrPath string `positional-arg-name:"NAME"`
}

type DeploymentNameArgs struct {
	Name string `positional-arg-name:"NAME"`
}

func (o CloudConfigOpts) Execute(_ []string) error         { return o.call() }
func (o UpdateCloudConfigOpts) Execute(_ []string) error   { return o.call() }
func (o RuntimeConfigOpts) Execute(_ []string) error       { return o.call() }
func (o UpdateRuntimeConfigOpts) Execute(_ []string) error { return o.call() }
func (o DeploymentOpts) Execute(_ []string) error          { return o.call() }
func (o DeploymentsOpts) Execute(_ []string) error         { return o.call() }
func (o DeployOpts) Execute(_ []string) error              { return o.call() }
func (o ManifestOpts) Execute(_ []string) error            { return o.call() }
func (o DeleteDeploymentOpts) Execute(_ []string) error    { return o.call() }

// Stemcells
type StemcellsOpts struct {
	call func() error
}

type UploadStemcellOpts struct {
	Args UploadStemcellArgs `positional-args:"true" required:"true"`

	Fix bool `long:"fix" description:"Replaces the stemcell if already exists"`

	Name    string     `long:"name"     description:"Name used in existence check (is not used with local stemcell file)"`
	Version VersionArg `long:"version"  description:"Version used in existence check (is not used with local stemcell file)"`

	SHA1 string `long:"sha1" description:"SHA1 of the remote stemcell (is not used with local files)"`

	call func() error
}

type UploadStemcellArgs struct {
	URL URLArg `positional-arg-name:"URL" description:"Path to a local file or URL"`
}

type DeleteStemcellOpts struct {
	Args DeleteStemcellArgs `positional-args:"true" required:"true"`

	Force bool `long:"force" description:"Ignore errors"`

	call func() error
}

type DeleteStemcellArgs struct {
	Slug boshdir.StemcellSlug `positional-arg-name:"NAME/VERSION"`
}

func (o StemcellsOpts) Execute(_ []string) error      { return o.call() }
func (o UploadStemcellOpts) Execute(_ []string) error { return o.call() }
func (o DeleteStemcellOpts) Execute(_ []string) error { return o.call() }

// Releases
type ReleasesOpts struct {
	call func() error
}

type UploadReleaseOpts struct {
	Args UploadReleaseArgs `positional-args:"true"`

	Directory DirOrCWDArg `long:"dir" description:"zzzRelease directory path if not current working directory" default:"."`

	Rebase bool `long:"rebase" description:"Rebases this release onto the latest version known by the Director"`

	Fix bool `long:"fix" description:"Replaces corrupt and missing jobs and packages"`

	Name    string     `long:"name"     description:"Name used in existence check (is not used with local release file)"`
	Version VersionArg `long:"version"  description:"Version used in existence check (is not used with local release file)"`

	SHA1 string `long:"sha1" description:"SHA1 of the remote release (is not used with local files)"`

	call func(DirOrCWDArg) error
}

type UploadReleaseArgs struct {
	URL URLArg `positional-args-name:"URL" description:"Path to a local file or URL"`
}

type DeleteReleaseOpts struct {
	Args DeleteReleaseArgs `positional-args:"true" required:"true"`

	Force bool `long:"force" description:"Ignore errors"`

	call func() error
}

type DeleteReleaseArgs struct {
	Slug boshdir.ReleaseOrSeriesSlug `positional-arg-name:"NAME[/VERSION]"`
}

type ExportReleaseOpts struct {
	Args ExportReleaseArgs `positional-args:"true" required:"true"`

	Directory DirOrCWDArg `long:"dir" description:"Destination directory" default:"."`

	call func() error
}

type ExportReleaseArgs struct {
	ReleaseSlug   boshdir.ReleaseSlug   `positional-arg-name:"NAME/VERSION"`
	OSVersionSlug boshdir.OSVersionSlug `positional-arg-name:"OS/VERSION"`
}

type InspectReleaseOpts struct {
	Args InspectReleaseArgs `positional-args:"true" required:"true"`
	call func() error
}

type InspectReleaseArgs struct {
	Slug boshdir.ReleaseSlug `positional-arg-name:"NAME/VERSION"`
}

func (o ReleasesOpts) Execute(_ []string) error       { return o.call() }
func (o UploadReleaseOpts) Execute(_ []string) error  { return o.call(o.Directory) }
func (o DeleteReleaseOpts) Execute(_ []string) error  { return o.call() }
func (o ExportReleaseOpts) Execute(_ []string) error  { return o.call() }
func (o InspectReleaseOpts) Execute(_ []string) error { return o.call() }

// Errands
type ErrandsOpts struct {
	call func() error
}

type RunErrandOpts struct {
	Args RunErrandArgs `positional-args:"true" required:"true"`

	KeepAlive bool `long:"keep-alive" description:"Use existing VM to run an errand and keep it after completion"`

	DownloadLogs  bool        `long:"download-logs" description:"Download logs"`
	LogsDirectory DirOrCWDArg `long:"logs-dir" description:"Destination directory for logs" default:"."`

	call func() error
}

type RunErrandArgs struct {
	Name string `positional-arg-name:"NAME"`
}

func (o ErrandsOpts) Execute(_ []string) error   { return o.call() }
func (o RunErrandOpts) Execute(_ []string) error { return o.call() }

// Disks
type DisksOpts struct {
	Orphaned bool `long:"orphaned" short:"o" description:"List orphaned disks"`
	call     func() error
}

type DeleteDiskOpts struct {
	Args DeleteDiskArgs `positional-args:"true" required:"true"`
	call func() error
}

type DeleteDiskArgs struct {
	CID string `positional-arg-name:"CID"`
}

func (o DisksOpts) Execute(_ []string) error      { return o.call() }
func (o DeleteDiskOpts) Execute(_ []string) error { return o.call() }

// Snapshots
type SnapshotsOpts struct {
	Args InstanceSlugArgs `positional-args:"true"`
	call func() error
}

type TakeSnapshotOpts struct {
	Args InstanceSlugArgs `positional-args:"true"`
	call func() error
}

type DeleteSnapshotOpts struct {
	Args DeleteSnapshotArgs `positional-args:"true" required:"true"`
	call func() error
}

type DeleteSnapshotArgs struct {
	CID string `positional-arg-name:"CID"`
}

type DeleteSnapshotsOpts struct {
	call func() error
}

type InstanceSlugArgs struct {
	Slug boshdir.InstanceSlug `positional-arg-name:"POOL/ID"`
}

func (o SnapshotsOpts) Execute(_ []string) error       { return o.call() }
func (o TakeSnapshotOpts) Execute(_ []string) error    { return o.call() }
func (o DeleteSnapshotOpts) Execute(_ []string) error  { return o.call() }
func (o DeleteSnapshotsOpts) Execute(_ []string) error { return o.call() }

// Instances
type InstancesOpts struct {
	Details   bool `long:"details" short:"i" description:"Show details including VM CID, persistent disk CID, etc."`
	DNS       bool `long:"dns"               description:"Show DNS A records"`
	Vitals    bool `long:"vitals"  short:"v" description:"Show vitals"`
	Processes bool `long:"ps"      short:"p" description:"Show processes"`
	Failing   bool `long:"failing" short:"f" description:"Only show failing instances"`

	call func() error
}

type VMsOpts struct {
	Details bool `long:"details" short:"i" description:"Show details including VM CID, persistent disk CID, etc."`
	DNS     bool `long:"dns"               description:"Show DNS A records"`
	Vitals  bool `long:"vitals"  short:"v" description:"Show vitals"`

	call func() error
}

type CloudCheckOpts struct {
	Auto   bool `long:"auto"   short:"a" description:"Resolve problems automatically"`
	Report bool `long:"report" short:"r" description:"Only generate report; don't attempt to resolve problems"`

	call func() error
}

func (o InstancesOpts) Execute(_ []string) error      { return o.call() }
func (o VMsOpts) Execute(_ []string) error            { return o.call() }
func (o VMResurrectionOpts) Execute(_ []string) error { return o.call() }
func (o CloudCheckOpts) Execute(_ []string) error     { return o.call() }

// Instance management
type VMResurrectionOpts struct {
	Args VMResurrectionArgs `positional-args:"true" required:"true"`
	call func() error
}

type VMResurrectionArgs struct {
	Enabled BoolArg `positional-arg-name:"on|off"`
}

type LogsOpts struct {
	Args AllOrPoolOrInstanceSlugArgs `positional-args:"true"`

	Directory DirOrCWDArg `long:"dir" description:"Destination directory" default:"."`

	Follow bool `long:"follow" short:"f" description:"Follow logs via SSH"`
	Num    int  `long:"num"    short:"n" description:"Last number of lines"`
	Quiet  bool `long:"quiet"  short:"q" description:"Suppresses printing of headers when multiple files are being examined."`

	Jobs    []string `long:"job"   description:"Limit to only specific jobs"`
	Filters []string `long:"only"  description:"Filter logs (comma-separated)"`
	Agent   bool     `long:"agent" description:"Include only agent logs"`

	GatewayFlags

	call func() error
}

type StartOpts struct {
	Args AllOrPoolOrInstanceSlugArgs `positional-args:"true"`

	Force bool `long:"force" description:"No-op for backwards compatibility"`

	call func() error
}

type StopOpts struct {
	Args AllOrPoolOrInstanceSlugArgs `positional-args:"true"`

	Soft bool `long:"soft" description:"Stop process only (default)"`
	Hard bool `long:"hard" description:"Delete VM (but keep persistent disk)"`

	SkipDrain boshdir.SkipDrain `long:"skip-drain" description:"Skip running drain scripts"`
	Force     bool              `long:"force"      description:"No-op for backwards compatibility"`

	call func() error
}

type RestartOpts struct {
	Args AllOrPoolOrInstanceSlugArgs `positional-args:"true"`

	SkipDrain boshdir.SkipDrain `long:"skip-drain" description:"Skip running drain scripts"`
	Force     bool              `long:"force"      description:"No-op for backwards compatibility"`

	call func() error
}

type RecreateOpts struct {
	Args AllOrPoolOrInstanceSlugArgs `positional-args:"true"`

	SkipDrain boshdir.SkipDrain `long:"skip-drain" description:"Skip running drain scripts"`
	Force     bool              `long:"force"      description:"No-op for backwards compatibility"`

	call func() error
}

type AllOrPoolOrInstanceSlugArgs struct {
	Slug boshdir.AllOrPoolOrInstanceSlug `positional-arg-name:"[POOL[/ID]]"`
}

func (o LogsOpts) Execute(_ []string) error     { return o.call() }
func (o StartOpts) Execute(_ []string) error    { return o.call() }
func (o StopOpts) Execute(_ []string) error     { return o.call() }
func (o RestartOpts) Execute(_ []string) error  { return o.call() }
func (o RecreateOpts) Execute(_ []string) error { return o.call() }

// SSH instance
type SSHOpts struct {
	Args AllOrPoolOrInstanceSlugArgs `positional-args:"true"`

	Command []string         `long:"command" short:"c" description:"Command"`
	RawOpts TrimmedSpaceArgs `long:"opts"              description:"Options to pass through to SSH"`

	Results bool `long:"results" short:"r" description:"Collect results into a table instead of streaming"`

	GatewayFlags

	call func() error
}

type SCPOpts struct {
	Args SCPArgs `positional-args:"true" required:"true"`

	Recursive bool `long:"recursive" short:"r" description:"Recursively copy entire directories. Note that symbolic links encountered are followed in the tree traversal"`

	GatewayFlags

	call func() error
}

type SCPArgs struct {
	Paths []string `positional-arg-name:"PATH"`
}

type GatewayFlags struct {
	Disable bool `long:"gw-disable" description:"Disable usage of gateway connection"`

	Username       string `long:"gw-user"        description:"Username for gateway connection"`
	Host           string `long:"gw-host"        description:"Host for gateway connection"`
	PrivateKeyPath string `long:"gw-private-key" description:"Private key path for gateway connection"` // todo private file?
}

func (o *SSHOpts) Execute(rest []string) error {
	if len(o.Command) == 0 {
		o.Command = rest
	}
	return o.call()
}

func (o SCPOpts) Execute(_ []string) error { return o.call() }

// Release creation
type InitReleaseOpts struct {
	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`

	Git bool `long:"git" description:"Initialize git repository"`

	call func(DirOrCWDArg) error
}

type ResetReleaseOpts struct {
	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`
	call      func(DirOrCWDArg) error
}

type GenerateJobOpts struct {
	Args GenerateJobArgs `positional-args:"true" required:"true"`

	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`

	call func(DirOrCWDArg) error
}

type GenerateJobArgs struct {
	Name string `positional-arg-name:"NAME"`
}

type GeneratePackageOpts struct {
	Args GeneratePackageArgs `positional-args:"true" required:"true"`

	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`

	call func(DirOrCWDArg) error
}

type GeneratePackageArgs struct {
	Name string `positional-arg-name:"NAME"`
}

type CreateReleaseOpts struct {
	Args CreateReleaseArgs `positional-args:"true"`

	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`

	Name             string     `long:"name"               description:"Custom release name"`
	Version          VersionArg `long:"version"            description:"Custom release version (e.g.: 1.0.0, 1.0-beta.2+dev.10)"`
	TimestampVersion bool       `long:"timestamp-version"  description:"Create release with the timestamp as the dev version (e.g.: 1+dev.TIMESTAMP)"`

	Final bool `long:"final" description:"Make it a final release"`

	WithTarball bool `long:"with-tarball" description:"Create release tarball"`

	Force bool `long:"force" description:"Ignore Git dirty state check"`

	call func(DirOrCWDArg) error
}

type CreateReleaseArgs struct {
	Manifest FileBytesArg `positional-arg-name:"PATH"`
}

type FinalizeReleaseOpts struct {
	Args FinalizeReleaseArgs `positional-args:"true" required:"true"`

	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`

	Name    string     `long:"name"    description:"Custom release name"`
	Version VersionArg `long:"version" description:"Custom release version (e.g.: 1.0.0, 1.0-beta.2+dev.10)"`

	Force bool `long:"force" description:"Ignore Git dirty state check"`

	call func(DirOrCWDArg) error
}

type FinalizeReleaseArgs struct {
	Path string `positional-arg-name:"PATH"`
}

func (o InitReleaseOpts) Execute(_ []string) error     { return o.call(o.Directory) }
func (o GenerateJobOpts) Execute(_ []string) error     { return o.call(o.Directory) }
func (o GeneratePackageOpts) Execute(_ []string) error { return o.call(o.Directory) }
func (o CreateReleaseOpts) Execute(_ []string) error   { return o.call(o.Directory) }
func (o FinalizeReleaseOpts) Execute(_ []string) error { return o.call(o.Directory) }

// Blobs
type BlobsOpts struct {
	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`
	call      func(DirOrCWDArg) error
}

type AddBlobOpts struct {
	Args AddBlobArgs `positional-args:"true" required:"true"`

	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`

	call func(DirOrCWDArg) error
}

type AddBlobArgs struct {
	Path      string `positional-arg-name:"PATH"`
	BlobsPath string `positional-arg-name:"BLOBS-PATH"`
}

type RemoveBlobOpts struct {
	Args RemoveBlobArgs `positional-args:"true" required:"true"`

	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`

	call func(DirOrCWDArg) error
}

type RemoveBlobArgs struct {
	BlobsPath string `positional-arg-name:"BLOBS-PATH"`
}

type SyncBlobsOpts struct {
	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`
	call      func(DirOrCWDArg) error
}

type UploadBlobsOpts struct {
	Directory DirOrCWDArg `long:"dir" description:"Release directory path if not current working directory" default:"."`
	call      func(DirOrCWDArg) error
}

func (o BlobsOpts) Execute(_ []string) error       { return o.call(o.Directory) }
func (o AddBlobOpts) Execute(_ []string) error     { return o.call(o.Directory) }
func (o RemoveBlobOpts) Execute(_ []string) error  { return o.call(o.Directory) }
func (o SyncBlobsOpts) Execute(_ []string) error   { return o.call(o.Directory) }
func (o UploadBlobsOpts) Execute(_ []string) error { return o.call(o.Directory) }
