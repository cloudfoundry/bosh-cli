package cmd

import (
	"fmt"
	"strconv"
	"strings"

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

	err = c.nonIntSSHRunner.Run(connOpts, result, c.buildTailCmd(opts))
	if err != nil {
		return bosherr.WrapErrorf(err, "Running follow over non-interactive SSH")
	}

	return nil
}

func (c LogsCmd) buildTailCmd(opts LogsOpts) []string {
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

func (c LogsCmd) fetch(opts LogsOpts) error {
	slug := opts.Args.Slug
	name := c.deployment.Name()

	if len(slug.Name()) > 0 {
		name += "." + slug.Name()
	}

	if len(slug.IndexOrID()) > 0 {
		name += "." + slug.IndexOrID()
	}

	result, err := c.deployment.FetchLogs(slug, opts.Filters, opts.Agent, opts.System, opts.All)
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
