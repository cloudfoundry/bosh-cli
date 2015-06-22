package platform

import (
	boshdpresolv "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/infrastructure/devicepathresolver"
	"github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/platform/cert"
	boshvitals "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/platform/vitals"
	boshsettings "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/settings"
	boshdir "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/settings/directories"
	boshcmd "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/fileutil"
	boshsys "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/system"
)

type Platform interface {
	GetFs() boshsys.FileSystem
	GetRunner() boshsys.CmdRunner
	GetCompressor() boshcmd.Compressor
	GetCopier() boshcmd.Copier
	GetDirProvider() boshdir.Provider
	GetVitalsService() boshvitals.Service

	GetDevicePathResolver() (devicePathResolver boshdpresolv.DevicePathResolver)

	// User management
	CreateUser(username, password, basePath string) (err error)
	AddUserToGroups(username string, groups []string) (err error)
	DeleteEphemeralUsersMatching(regex string) (err error)

	// Bootstrap functionality
	SetupSSH(publicKey, username string) (err error)
	SetUserPassword(user, encryptedPwd string) (err error)
	SetupHostname(hostname string) (err error)
	SetupNetworking(networks boshsettings.Networks) (err error)
	SetupLogrotate(groupName, basePath, size string) (err error)
	SetTimeWithNtpServers(servers []string) (err error)
	SetupEphemeralDiskWithPath(devicePath string) (err error)
	SetupDataDir() (err error)
	SetupTmpDir() (err error)
	SetupMonitUser() (err error)
	StartMonit() (err error)
	SetupRuntimeConfiguration() (err error)

	// Disk management
	MountPersistentDisk(diskSettings boshsettings.DiskSettings, mountPoint string) error
	UnmountPersistentDisk(diskSettings boshsettings.DiskSettings) (didUnmount bool, err error)
	MigratePersistentDisk(fromMountPoint, toMountPoint string) (err error)
	GetEphemeralDiskPath(diskSettings boshsettings.DiskSettings) string
	IsMountPoint(path string) (result bool, err error)
	IsPersistentDiskMounted(diskSettings boshsettings.DiskSettings) (result bool, err error)

	GetFileContentsFromCDROM(filePath string) (contents []byte, err error)
	GetFilesContentsFromDisk(diskPath string, fileNames []string) (contents [][]byte, err error)

	// Network misc
	GetDefaultNetwork() (boshsettings.Network, error)
	GetConfiguredNetworkInterfaces() ([]string, error)
	PrepareForNetworkingChange() error

	// Additional monit management
	GetMonitCredentials() (username, password string, err error)

	GetCertManager() cert.Manager
}
