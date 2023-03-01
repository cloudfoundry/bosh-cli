package cmd

import (
	"fmt"
	"strconv"
	"strings"

	biagentclient "github.com/cloudfoundry/bosh-agent/agentclient"
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshssh "github.com/cloudfoundry/bosh-cli/v7/ssh"
)

type LogsCmd struct {
	deployment      boshdir.Deployment
	downloader      Downloader
	uuidGen         boshuuid.Generator
	nonIntSSHRunner boshssh.Runner
}

type EnvLogsCmd struct {
	agentClient     biagentclient.AgentClient
	nonIntSSHRunner boshssh.Runner
	scpRunner       boshssh.SCPRunner
	fs              boshsys.FileSystem
}

func NewEnvLogsCmd(
	agentClient biagentclient.AgentClient,
	nonIntSSHRunner boshssh.Runner,
	scpRunner boshssh.SCPRunner,
	fs boshsys.FileSystem,
) EnvLogsCmd {
	return EnvLogsCmd{
		agentClient:     agentClient,
		nonIntSSHRunner: nonIntSSHRunner,
		scpRunner:       scpRunner,
		fs:              fs,
	}
}

func (c EnvLogsCmd) Run(opts LogsOpts) error {
	if opts.Follow || opts.Num > 0 {
		return c.tail(opts)
	}
	return c.fetch(opts)
}

func (c EnvLogsCmd) tail(opts LogsOpts) error {
	sshOpts, connOpts, err := opts.GatewayFlags.AsSSHOpts()
	if err != nil {
		return err
	}

	var agentResult biagentclient.SSHResult
	agentResult, err = c.agentClient.SetUpSSH(sshOpts.Username, sshOpts.PublicKey)
	if err != nil {
		return err
	}
	result := boshdir.SSHResult{
		Hosts: []boshdir.Host{
			{
				Username:      sshOpts.Username,
				Host:          agentResult.Ip,
				HostPublicKey: agentResult.HostPublicKey,
				Job:           "create-env-vm",
				IndexOrID:     "0",
			},
		},
	}

	defer func() {
		_, _ = c.agentClient.CleanUpSSH(sshOpts.Username)
	}()

	err = c.nonIntSSHRunner.Run(connOpts, result, buildTailCmd(opts))
	if err != nil {
		return bosherr.WrapErrorf(err, "Running follow over non-interactive SSH")
	}

	return nil
}

func (c EnvLogsCmd) fetch(opts LogsOpts) error {
	sshOpts, connOpts, err := opts.GatewayFlags.AsSSHOpts()
	if err != nil {
		return err
	}

	var agentResult biagentclient.SSHResult
	agentResult, err = c.agentClient.SetUpSSH(sshOpts.Username, sshOpts.PublicKey)
	if err != nil {
		return err
	}
	result := boshdir.SSHResult{
		Hosts: []boshdir.Host{
			{
				Username:      sshOpts.Username,
				Host:          agentResult.Ip,
				HostPublicKey: agentResult.HostPublicKey,
				Job:           "create-env-vm",
				IndexOrID:     "0",
			},
		},
	}

	defer func() {
		_, _ = c.agentClient.CleanUpSSH(sshOpts.Username)
	}()

	logType := "job"
	if opts.Agent {
		logType = "agent"
	}
	bundleLogsResult, err := c.agentClient.BundleLogs(result.Hosts[0].Username, logType, opts.Filters)
	if err != nil {
		return err
	}
	defer func() {
		_ = c.agentClient.RemoveFile(bundleLogsResult.LogsTarPath)
	}()

	scpArgs := boshssh.NewSCPArgs([]string{fmt.Sprintf("%s:%s", "create-env-vm", bundleLogsResult.LogsTarPath), "./logs.tgz"}, false)

	err = c.scpRunner.Run(connOpts, result, scpArgs)
	if err != nil {
		return bosherr.WrapErrorf(err, "Running SCP")
	}

	expectedMultipleDigest, err := boshcrypto.ParseMultipleDigest(bundleLogsResult.SHA512Digest)
	if err != nil {
		return err
	}

	err = expectedMultipleDigest.VerifyFilePath("./logs.tgz", c.fs)
	if err != nil {
		return err
	}

	return nil
}

func NewLogsCmd(
	deployment boshdir.Deployment,
	downloader Downloader,
	uuidGen boshuuid.Generator,
	nonIntSSHRunner boshssh.Runner,
) LogsCmd {
	return LogsCmd{
		deployment:      deployment,
		downloader:      downloader,
		uuidGen:         uuidGen,
		nonIntSSHRunner: nonIntSSHRunner,
	}
}

func (c LogsCmd) Run(opts LogsOpts) error {
	if opts.Follow || opts.Num > 0 {
		return c.tail(opts)
	}
	return c.fetch(opts)
}

func (c LogsCmd) tail(opts LogsOpts) error {
	sshOpts, connOpts, err := opts.GatewayFlags.AsSSHOpts()
	if err != nil {
		return err
	}

	result, err := c.deployment.SetUpSSH(opts.Args.Slug, sshOpts)
	if err != nil {
		return err
	}

	defer func() {
		_ = c.deployment.CleanUpSSH(opts.Args.Slug, sshOpts)
	}()

	err = c.nonIntSSHRunner.Run(connOpts, result, buildTailCmd(opts))
	if err != nil {
		return bosherr.WrapErrorf(err, "Running follow over non-interactive SSH")
	}

	return nil
}

func buildTailCmd(opts LogsOpts) []string {
	cmd := []string{"sudo", "bash", "-c"}
	tail := []string{"exec", "tail"}

	if opts.Follow {
		// -F for continuing to follow after renames
		tail = append(tail, "-F")
	}

	if opts.Num > 0 {
		tail = append(tail, "-n", strconv.Itoa(opts.Num))
	}

	if opts.Quiet {
		tail = append(tail, "-q")
	}

	var logsDir string

	if opts.Agent {
		tail = append(tail, "/var/vcap/bosh/log/current")
	}

	logsDir = "/var/vcap/sys/log"

	if len(opts.Jobs) > 0 {
		for _, job := range opts.Jobs {
			tail = append(tail, fmt.Sprintf("%s/%s/*.log", logsDir, job))
		}
	} else if len(opts.Filters) > 0 {
		for _, filter := range opts.Filters {
			tail = append(tail, fmt.Sprintf("%s/%s", logsDir, filter))
		}
	} else if !opts.Agent {
		// includes only directory and its subdirectories
		tail = append(tail, fmt.Sprintf("%s/**/*.log", logsDir))
		tail = append(tail, fmt.Sprintf("$(if [ -f %s/*.log ]; then echo %s/*.log ; fi)", logsDir, logsDir))
	}

	// append combined tail command
	cmd = append(cmd, "'"+strings.Join(tail, " ")+"'")
	return cmd
}

func (c LogsCmd) fetch(opts LogsOpts) error {
	slug := opts.Args.Slug
	name := c.deployment.Name()

	if len(slug.Name()) > 0 {
		name += "." + slug.Name()
	}

	if len(slug.IndexOrID()) > 0 {
		name += "." + slug.IndexOrID()
	}

	result, err := c.deployment.FetchLogs(slug, opts.Filters, opts.Agent)
	if err != nil {
		return err
	}

	err = c.downloader.Download(
		result.BlobstoreID,
		result.SHA1,
		name,
		opts.Directory.Path,
	)
	if err != nil {
		return bosherr.WrapError(err, "Downloading logs")
	}

	return nil
}
