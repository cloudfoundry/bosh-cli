package platform

import (
	"encoding/json"
	"path/filepath"

	boshdpresolv "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/infrastructure/devicepathresolver"
	boshcert "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/platform/cert"
	boshstats "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/platform/stats"
	boshvitals "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/platform/vitals"
	boshsettings "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/settings"
	boshdir "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/settings/directories"
	boshdirs "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/settings/directories"
	bosherr "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/errors"
	boshcmd "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/system"
)

type dummyPlatform struct {
	collector          boshstats.Collector
	fs                 boshsys.FileSystem
	cmdRunner          boshsys.CmdRunner
	compressor         boshcmd.Compressor
	copier             boshcmd.Copier
	dirProvider        boshdirs.Provider
	vitalsService      boshvitals.Service
	devicePathResolver boshdpresolv.DevicePathResolver
	logger             boshlog.Logger
	certManager        boshcert.Manager
}

func NewDummyPlatform(
	collector boshstats.Collector,
	fs boshsys.FileSystem,
	cmdRunner boshsys.CmdRunner,
	dirProvider boshdirs.Provider,
	devicePathResolver boshdpresolv.DevicePathResolver,
	logger boshlog.Logger,
) Platform {
	return &dummyPlatform{
		fs:                 fs,
		cmdRunner:          cmdRunner,
		collector:          collector,
		compressor:         boshcmd.NewTarballCompressor(cmdRunner, fs),
		copier:             boshcmd.NewCpCopier(cmdRunner, fs, logger),
		dirProvider:        dirProvider,
		devicePathResolver: devicePathResolver,
		vitalsService:      boshvitals.NewService(collector, dirProvider),
		certManager:        boshcert.NewDummyCertManager(fs, cmdRunner, logger),
	}
}

func (p dummyPlatform) GetFs() (fs boshsys.FileSystem) {
	return p.fs
}

func (p dummyPlatform) GetRunner() (runner boshsys.CmdRunner) {
	return p.cmdRunner
}

func (p dummyPlatform) GetCompressor() (compressor boshcmd.Compressor) {
	return p.compressor
}

func (p dummyPlatform) GetCopier() (copier boshcmd.Copier) {
	return p.copier
}

func (p dummyPlatform) GetDirProvider() (dirProvider boshdir.Provider) {
	return p.dirProvider
}

func (p dummyPlatform) GetVitalsService() (service boshvitals.Service) {
	return p.vitalsService
}

func (p dummyPlatform) GetDevicePathResolver() (devicePathResolver boshdpresolv.DevicePathResolver) {
	return p.devicePathResolver
}

func (p dummyPlatform) SetupRuntimeConfiguration() (err error) {
	return
}

func (p dummyPlatform) CreateUser(username, password, basePath string) (err error) {
	return
}

func (p dummyPlatform) AddUserToGroups(username string, groups []string) (err error) {
	return
}

func (p dummyPlatform) DeleteEphemeralUsersMatching(regex string) (err error) {
	return
}

func (p dummyPlatform) SetupSSH(publicKey, username string) (err error) {
	return
}

func (p dummyPlatform) SetUserPassword(user, encryptedPwd string) (err error) {
	return
}

func (p dummyPlatform) SetupHostname(hostname string) (err error) {
	return
}

func (p dummyPlatform) SetupNetworking(networks boshsettings.Networks) (err error) {
	return
}

func (p dummyPlatform) GetConfiguredNetworkInterfaces() (interfaces []string, err error) {
	return
}

func (p dummyPlatform) GetCertManager() (certManager boshcert.Manager) {
	return p.certManager
}

func (p dummyPlatform) SetupLogrotate(groupName, basePath, size string) (err error) {
	return
}

func (p dummyPlatform) SetTimeWithNtpServers(servers []string) (err error) {
	return
}

func (p dummyPlatform) SetupEphemeralDiskWithPath(devicePath string) (err error) {
	return
}

func (p dummyPlatform) SetupDataDir() error {
	return nil
}

func (p dummyPlatform) SetupTmpDir() error {
	return nil
}

func (p dummyPlatform) MountPersistentDisk(diskSettings boshsettings.DiskSettings, mountPoint string) (err error) {
	return
}

func (p dummyPlatform) UnmountPersistentDisk(diskSettings boshsettings.DiskSettings) (didUnmount bool, err error) {
	return
}

func (p dummyPlatform) GetEphemeralDiskPath(diskSettings boshsettings.DiskSettings) string {
	return "/dev/sdb"
}

func (p dummyPlatform) GetFileContentsFromCDROM(filePath string) (contents []byte, err error) {
	return
}

func (p dummyPlatform) GetFilesContentsFromDisk(diskPath string, fileNames []string) (contents [][]byte, err error) {
	return
}

func (p dummyPlatform) MigratePersistentDisk(fromMountPoint, toMountPoint string) (err error) {
	return
}

func (p dummyPlatform) IsMountPoint(path string) (result bool, err error) {
	return
}

func (p dummyPlatform) IsPersistentDiskMounted(diskSettings boshsettings.DiskSettings) (bool, error) {
	return true, nil
}

func (p dummyPlatform) StartMonit() (err error) {
	return
}

func (p dummyPlatform) SetupMonitUser() (err error) {
	return
}

func (p dummyPlatform) GetMonitCredentials() (username, password string, err error) {
	return
}

func (p dummyPlatform) PrepareForNetworkingChange() error {
	return nil
}

func (p dummyPlatform) GetDefaultNetwork() (boshsettings.Network, error) {
	var network boshsettings.Network

	networkPath := filepath.Join(p.dirProvider.BoshDir(), "dummy-default-network-settings.json")
	contents, err := p.fs.ReadFile(networkPath)
	if err != nil {
		return network, nil
	}

	err = json.Unmarshal([]byte(contents), &network)
	if err != nil {
		return network, bosherr.WrapError(err, "Unmarshal json settings")
	}

	return network, nil
}
