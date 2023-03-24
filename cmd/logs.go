package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"code.cloudfoundry.org/clock"

	biagentclient "github.com/cloudfoundry/bosh-agent/agentclient"
	bihttpagent "github.com/cloudfoundry/bosh-agent/agentclient/http"
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshfu "github.com/cloudfoundry/bosh-utils/fileutil"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshssh "github.com/cloudfoundry/bosh-cli/v7/ssh"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
)

type LogsCmd struct {
	deployment      boshdir.Deployment
	downloader      Downloader
	uuidGen         boshuuid.Generator
	nonIntSSHRunner boshssh.Runner
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

	if (opts.System && !opts.Agent) || opts.All {
		// Globbing makes it very difficult to do fine-grained exclusions of file types. Because we do not want to tail compressed files,
		// rotated files, or the files in the `sysstat` directory (some of which are binary), we're using find to get fine-grained
		// control.
		// If folks complain that we're tailing something that makes their terminal sad, feel free to add additional filtering here.
		// Also note that this string will eventually get executed as "sudo bash -c '$TAIL_COMMAND'", so we need to backslash-escape
		// globbing characters, rather than wrapping them in single-quotes.
		tail = append(tail, "$(find /var/log -type f -not -name \\*.gz -and -not -name \\*.xz -and -not -name \\*.\\[1-9] -and -not -path /var/log/sysstat/\\* -and -not -wholename /var/log/wtmp -and -not -wholename /var/log/lastlog)")
	}

	if (opts.Agent && !opts.System) || opts.All {
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
	} else if (!opts.Agent && !opts.System) || opts.All {
		// includes only directory and its subdirectories
		tail = append(tail, fmt.Sprintf("%s/**/*.log", logsDir))
		tail = append(tail, fmt.Sprintf("$(if [ -f %s/*.log ]; then echo %s/*.log ; fi)", logsDir, logsDir))
	}

	// append combined tail command
	cmd = append(cmd, "'"+strings.Join(tail, " ")+"'")
	return cmd
}

func buildLogTypeArgument(opts LogsOpts) string {
	var logTypes []string
	if opts.All {
		logTypes = append(logTypes, "agent")
		logTypes = append(logTypes, "job")
		logTypes = append(logTypes, "system")
	} else {
		if opts.Agent {
			logTypes = append(logTypes, "agent")
		} else if opts.System {
			logTypes = append(logTypes, "system")
		} else {
			logTypes = append(logTypes, "job")
		}
	}
	logType := strings.Join(logTypes, ",")
	return logType
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

	logType := buildLogTypeArgument(opts)
	result, err := c.deployment.FetchLogs(slug, opts.Filters, logType)
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

type EnvLogsCmd struct {
	agentClientFactory bihttpagent.AgentClientFactory
	nonIntSSHRunner    boshssh.Runner
	scpRunner          boshssh.SCPRunner
	fs                 boshsys.FileSystem
	timeService        clock.Clock
	ui                 boshui.UI
}

func NewEnvLogsCmd(
	agentClientFactory bihttpagent.AgentClientFactory,
	nonIntSSHRunner boshssh.Runner,
	scpRunner boshssh.SCPRunner,
	fs boshsys.FileSystem,
	timeService clock.Clock,
	ui boshui.UI,
) EnvLogsCmd {
	return EnvLogsCmd{
		agentClientFactory: agentClientFactory,
		nonIntSSHRunner:    nonIntSSHRunner,
		scpRunner:          scpRunner,
		fs:                 fs,
		timeService:        timeService,
		ui:                 ui,
	}
}

func (c EnvLogsCmd) Run(opts LogsOpts) error {
	if opts.Endpoint == "" || opts.Certificate == "" {
		return errors.New("the --director flag requires both the --agent-endpoint and --agent-certificate flags to be set")
	}

	agentClient, err := c.agentClientFactory.NewAgentClient("bosh-cli", opts.Endpoint, opts.Certificate)
	if err != nil {
		return err
	}

	sshOpts, connOpts, err := opts.GatewayFlags.AsSSHOpts()
	if err != nil {
		return err
	}

	agentResult, err := agentClient.SetUpSSH(sshOpts.Username, sshOpts.PublicKey)
	if err != nil {
		return err
	}
	sshResult := boshdir.SSHResult{
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
		_, _ = agentClient.CleanUpSSH(sshOpts.Username)
	}()

	if opts.Follow || opts.Num > 0 {
		return c.tail(opts, connOpts, sshResult)
	}

	return c.fetch(opts, connOpts, sshResult, agentClient)
}

func (c EnvLogsCmd) tail(opts LogsOpts, connOpts boshssh.ConnectionOpts, sshResult boshdir.SSHResult) error {
	err := c.nonIntSSHRunner.Run(connOpts, sshResult, buildTailCmd(opts))
	if err != nil {
		return bosherr.WrapErrorf(err, "Running follow over non-interactive SSH")
	}

	return nil
}

func (c EnvLogsCmd) fetch(opts LogsOpts, connOpts boshssh.ConnectionOpts, sshResult boshdir.SSHResult, agentClient biagentclient.AgentClient) error {
	logType := buildLogTypeArgument(opts)

	bundleLogsResult, err := agentClient.BundleLogs(sshResult.Hosts[0].Username, logType, opts.Filters)
	if err != nil {
		return err
	}
	defer func() {
		_ = agentClient.RemoveFile(bundleLogsResult.LogsTarPath)
	}()

	// This is section is lifted from Downloader.Download, an opportunity for refactor in the future?
	tsSuffix := strings.Replace(c.timeService.Now().Format("20060102-150405.999999999"), ".", "-", -1)
	dstFileName := fmt.Sprintf("%s-%s.tgz", "create-env-vm-logs", tsSuffix)
	dstFilePath := filepath.Join(opts.Directory.Path, dstFileName)

	tmpFile, err := c.fs.TempFile("bosh-cli-scp-download")
	if err != nil {
		return err
	}
	defer tmpFile.Close()                //nolint:errcheck
	defer c.fs.RemoveAll(tmpFile.Name()) //nolint:errcheck

	c.ui.PrintLinef("Downloading create-env-vm/0 logs to '%s'...", dstFilePath)
	scpArgs := boshssh.NewSCPArgs([]string{fmt.Sprintf("%s:%s", "create-env-vm/0", bundleLogsResult.LogsTarPath), tmpFile.Name()}, false)
	err = c.scpRunner.Run(connOpts, sshResult, scpArgs)
	if err != nil {
		return bosherr.WrapErrorf(err, "Running SCP")
	}

	expectedMultipleDigest, err := boshcrypto.ParseMultipleDigest(bundleLogsResult.SHA512Digest)
	if err != nil {
		return err
	}

	err = expectedMultipleDigest.VerifyFilePath(tmpFile.Name(), c.fs)
	if err != nil {
		return err
	}

	err = tmpFile.Close()
	if err != nil {
		return err
	}

	err = boshfu.NewFileMover(c.fs).Move(tmpFile.Name(), dstFilePath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Moving to final destination")
	}

	return nil
}
