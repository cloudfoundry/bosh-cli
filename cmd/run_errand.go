package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshssh "github.com/cloudfoundry/bosh-cli/v7/ssh"
	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
)

var safeLogPathRe = regexp.MustCompile(`^[a-zA-Z0-9._\-/*{},]+$`)

type ErrandContextFunc func() (context.Context, context.CancelFunc)

type RunErrandCmd struct {
	deployment      boshdir.Deployment
	downloader      Downloader
	ui              biui.UI
	nonIntSSHRunner boshssh.Runner
	taskReporter    boshdir.TaskReporter
	ctxFactory      ErrandContextFunc
}

func NewRunErrandCmd(
	deployment boshdir.Deployment,
	downloader Downloader,
	ui biui.UI,
	nonIntSSHRunner boshssh.Runner,
	taskReporter boshdir.TaskReporter,
	ctxFactory ErrandContextFunc,
) RunErrandCmd {
	if ctxFactory == nil {
		ctxFactory = func() (context.Context, context.CancelFunc) {
			return signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		}
	}
	return RunErrandCmd{
		deployment:      deployment,
		downloader:      downloader,
		ui:              ui,
		nonIntSSHRunner: nonIntSSHRunner,
		taskReporter:    taskReporter,
		ctxFactory:      ctxFactory,
	}
}

func (c RunErrandCmd) Run(opts RunErrandOpts) error {
	if opts.StreamLogs {
		return c.runWithStreaming(opts)
	}
	return c.runWithoutStreaming(opts)
}

func (c RunErrandCmd) runWithoutStreaming(opts RunErrandOpts) error {
	results, err := c.deployment.RunErrand(
		opts.Args.Name,
		opts.KeepAlive,
		opts.WhenChanged,
		opts.InstanceGroupOrInstanceSlugFlags.Slugs, //nolint:staticcheck
	)
	if err != nil {
		return err
	}

	return c.finishErrand(opts, results)
}

func (c RunErrandCmd) runWithStreaming(opts RunErrandOpts) error {
	if c.nonIntSSHRunner == nil {
		return bosherr.Errorf("SSH runner is required for --stream-logs")
	}

	tailCmd, err := BuildErrandTailCmd(opts.Args.Name, opts.StreamLogPath)
	if err != nil {
		return err
	}

	taskID, err := c.deployment.StartErrand(
		opts.Args.Name,
		opts.KeepAlive,
		opts.WhenChanged,
		opts.InstanceGroupOrInstanceSlugFlags.Slugs, //nolint:staticcheck
	)
	if err != nil {
		return err
	}

	c.ui.PrintLinef("Errand started as task %d, streaming logs...", taskID)

	sshOpts, connOpts, err := opts.GatewayFlags.AsSSHOpts() //nolint:staticcheck
	if err != nil {
		return err
	}

	// sshCtx is cancelled either when we call sshCancel (normal exit) or
	// when the process receives SIGINT/SIGTERM (Ctrl+C).  In both cases
	// ComboRunner.waitProcs sees ctx.Done() and terminates the ssh processes.
	sshCtx, sshCancel := c.ctxFactory()
	defer sshCancel()

	stopCh := make(chan struct{})
	var stopOnce sync.Once
	closeStop := func() { stopOnce.Do(func() { close(stopCh) }) }
	defer closeStop()

	watcher := NewErrandEventWatcher(c.deployment, taskID, 2*time.Second)
	if c.taskReporter != nil {
		watcher.WithTaskReporter(c.taskReporter)
	}
	slugCh := watcher.Watch(stopCh)

	var sessions []boshdir.AllOrInstanceGroupOrInstanceSlug
	var sshWg sync.WaitGroup

	// Consume slugs from the watcher, setting up SSH tails for each.
	// Also select on sshCtx.Done() so Ctrl+C breaks out immediately.
	for done := false; !done; {
		select {
		case slug, ok := <-slugCh:
			if !ok {
				done = true
				break
			}

			parts := strings.SplitN(slug, "/", 2)
			if len(parts) != 2 {
				continue
			}

			instanceSlug := boshdir.NewAllOrInstanceGroupOrInstanceSlug(parts[0], parts[1])

			result, setupErr := c.deployment.SetUpSSH(instanceSlug, sshOpts)
			if setupErr != nil {
				c.ui.PrintLinef("Warning: failed to set up SSH for %s: %s", slug, setupErr.Error())
				continue
			}

			sessions = append(sessions, instanceSlug)

			sshWg.Add(1)
			go func(s string) {
				defer sshWg.Done()
				runErr := c.nonIntSSHRunner.RunContext(sshCtx, connOpts, result, tailCmd)
				if runErr != nil && sshCtx.Err() == nil {
					c.ui.PrintLinef("Warning: SSH tail on %s exited: %s", s, runErr.Error())
				}
			}(slug)

		case <-sshCtx.Done():
			done = true
		}
	}

	// Check whether a signal fired before we cancel the context ourselves.
	interrupted := sshCtx.Err() != nil

	// Stop the event watcher so its goroutine exits.
	closeStop()

	// Cancel the SSH context to terminate all local ssh processes.
	sshCancel()
	sshWg.Wait()

	for _, slug := range sessions {
		if cleanupErr := c.deployment.CleanUpSSH(slug, sshOpts); cleanupErr != nil {
			c.ui.PrintLinef("Warning: failed to clean up SSH for %s: %s", slug, cleanupErr)
		}
	}

	if interrupted {
		c.ui.PrintLinef("\nStreaming interrupted. Errand task %d is still running on the director.", taskID)
		c.ui.PrintLinef("Use 'bosh task %d' to monitor or 'bosh cancel-task %d' to stop it.", taskID, taskID)
		return nil
	}

	results, err := c.deployment.WaitForErrandResult(taskID)
	if err != nil {
		return err
	}

	// Suppress stdout/stderr in the summary since we already streamed the logs.
	for i := range results {
		results[i].Stdout = ""
		results[i].Stderr = ""
	}

	return c.finishErrand(opts, results)
}

func (c RunErrandCmd) finishErrand(opts RunErrandOpts, results []boshdir.ErrandResult) error {
	errandErr := c.summarize(opts.Args.Name, results)

	for _, result := range results {
		if opts.DownloadLogs && len(result.LogsBlobstoreID) > 0 {
			err := c.downloader.Download(
				result.LogsBlobstoreID,
				result.LogsSHA1,
				opts.Args.Name,
				opts.LogsDirectory.Path,
			)
			if err != nil {
				return bosherr.WrapError(err, "Downloading errand logs")
			}
		}
	}

	return errandErr
}

func BuildErrandTailCmd(errandName, customPath string) ([]string, error) {
	var logPath string
	if customPath != "" {
		if !safeLogPathRe.MatchString(customPath) {
			return nil, bosherr.Errorf("--stream-log-path contains invalid characters: %q", customPath)
		}
		logPath = fmt.Sprintf("/var/vcap/sys/log/%s", customPath)
	} else {
		if !safeLogPathRe.MatchString(errandName) {
			return nil, bosherr.Errorf("errand name contains invalid characters: %q", errandName)
		}
		logPath = fmt.Sprintf("/var/vcap/sys/log/%[1]s/%[1]s.{stdout,stderr}.log", errandName)
	}

	tailScript := fmt.Sprintf(
		`until ls %[1]s >/dev/null 2>&1;do sleep 1; done && exec tail -n 0 -F %[1]s`,
		logPath,
	)

	// The script must be single-quoted because ssh concatenates all trailing
	// arguments with spaces before passing them to the remote shell.  Without
	// quotes, "bash -c <script>" is interpreted as "bash -c for" with the
	// remaining tokens becoming positional parameters.
	return []string{"sudo", "bash", "-c", "'" + tailScript + "'"}, nil
}

func (c RunErrandCmd) summarize(errandName string, results []boshdir.ErrandResult) error {
	table := boshtbl.Table{
		Content: "errand(s)",

		Header: []boshtbl.Header{
			boshtbl.NewHeader("Instance"),
			boshtbl.NewHeader("Exit Code"),
			boshtbl.NewHeader("Stdout"),
			boshtbl.NewHeader("Stderr"),
		},

		SortBy: []boshtbl.ColumnSort{
			{Column: 0, Asc: true},
		},

		Notes: []string{},

		FillFirstColumn: true,

		Transpose: true,
	}

	var errandErr error
	for _, result := range results {
		instance := ""
		if result.InstanceGroup != "" {
			instance = boshdir.NewInstanceGroupOrInstanceSlug(result.InstanceGroup, result.InstanceID).String()
		}

		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(instance),
			boshtbl.NewValueInt(result.ExitCode),
			boshtbl.NewValueString(result.Stdout),
			boshtbl.NewValueString(result.Stderr),
		})

		prefix := fmt.Sprintf("Errand '%s'", errandName)
		suffix := fmt.Sprintf("(exit code %d)", result.ExitCode)

		switch {
		case result.ExitCode == 0:
		case result.ExitCode > 128:
			errandErr = bosherr.Errorf("%s was canceled %s", prefix, suffix)
		default:
			errandErr = bosherr.Errorf("%s completed with error %s", prefix, suffix)
		}
	}
	c.ui.PrintTable(table)

	return errandErr
}
