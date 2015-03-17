package acceptance

import (
	"fmt"
	"path/filepath"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type Environment interface {
	Home() string
	Path(string) string
	Copy(string, string) error
	WriteContent(string, []byte) error
	RemoteDownload(string, string) error
	DownloadOrCopy(string, string, string) error
}

type remoteTestEnvironment struct {
	vmUsername     string
	vmIP           string
	vmPort         string
	privateKeyPath string
	cmdRunner      boshsys.CmdRunner
	fileSystem     boshsys.FileSystem
}

func NewRemoteTestEnvironment(
	vmUsername string,
	vmIP string,
	vmPort string,
	privateKeyPath string,
	fileSystem boshsys.FileSystem,
	logger boshlog.Logger,
) Environment {
	return remoteTestEnvironment{
		vmUsername:     vmUsername,
		vmIP:           vmIP,
		vmPort:         vmPort,
		privateKeyPath: privateKeyPath,
		cmdRunner:      boshsys.NewExecCmdRunner(logger),
		fileSystem:     fileSystem,
	}
}

func (e remoteTestEnvironment) Home() string {
	return filepath.Join("/", "home", e.vmUsername)
}

func (e remoteTestEnvironment) Path(name string) string {
	return filepath.Join(e.Home(), name)
}

func (e remoteTestEnvironment) Copy(destName, srcPath string) error {
	if srcPath == "" {
		return fmt.Errorf("Cannot use an empty file for '%s'", destName)
	}

	_, _, exitCode, err := e.cmdRunner.RunCommand(
		"scp",
		"-o", "StrictHostKeyChecking=no",
		"-i", e.privateKeyPath,
		"-P", e.vmPort,
		srcPath,
		fmt.Sprintf("%s@%s:%s", e.vmUsername, e.vmIP, e.Path(destName)),
	)
	if exitCode != 0 {
		return fmt.Errorf("scp of '%s' to '%s' failed", srcPath, destName)
	}
	return err
}

func (e remoteTestEnvironment) DownloadOrCopy(destName, srcPath, srcURL string) error {
	if srcPath != "" {
		return e.Copy(destName, srcPath)
	}
	return e.RemoteDownload(destName, srcURL)
}

func (e remoteTestEnvironment) RemoteDownload(destName, srcURL string) error {
	if srcURL == "" {
		return fmt.Errorf("Cannot use an empty file for '%s'", destName)
	}

	_, _, exitCode, err := e.cmdRunner.RunCommand(
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-i", e.privateKeyPath,
		"-p", e.vmPort,
		fmt.Sprintf("%s@%s", e.vmUsername, e.vmIP),
		fmt.Sprintf("wget -q -O %s %s", destName, srcURL),
	)
	if exitCode != 0 {
		return fmt.Errorf("download of '%s' to '%s' failed", srcURL, destName)
	}
	return err
}

func (e remoteTestEnvironment) WriteContent(destName string, contents []byte) error {
	tmpFile, err := e.fileSystem.TempFile("bosh-micro-cli-acceptance")
	if err != nil {
		return err
	}
	defer e.fileSystem.RemoveAll(tmpFile.Name())
	_, err = tmpFile.Write(contents)
	if err != nil {
		return err
	}
	err = tmpFile.Close()
	if err != nil {
		return err
	}

	return e.Copy(destName, tmpFile.Name())
}
