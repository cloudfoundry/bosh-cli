package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshssh "github.com/cloudfoundry/bosh-cli/v7/ssh"
	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

type RunErrandCmd struct {
	deployment boshdir.Deployment
	downloader Downloader
	ui         biui.UI
	scpRunner  boshssh.SCPRunner
	fs         boshsys.FileSystem
	logger     boshlog.Logger
}

func NewRunErrandCmd(
	deployment boshdir.Deployment,
	downloader Downloader,
	ui biui.UI,
	scpRunner boshssh.SCPRunner,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) RunErrandCmd {
	return RunErrandCmd{
		deployment: deployment,
		downloader: downloader,
		ui:         ui,
		scpRunner:  scpRunner,
		fs:         fs,
		logger:     logger,
	}
}

func (c RunErrandCmd) Run(opts RunErrandOpts) error {
	if opts.StreamLogs != nil {
		return c.runWithStreaming(opts)
	}

	return c.runNormal(opts)
}

func (c RunErrandCmd) runNormal(opts RunErrandOpts) error {
	results, err := c.deployment.RunErrand(
		opts.Args.Name,
		opts.KeepAlive,
		opts.WhenChanged,
		opts.InstanceGroupOrInstanceSlugFlags.Slugs, //nolint:staticcheck
	)
	if err != nil {
		return err
	}

	errandErr := c.summarize(opts.Args.Name, results, false)
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

func (c RunErrandCmd) runWithStreaming(opts RunErrandOpts) error {
	errandName := opts.Args.Name
	logPath := fmt.Sprintf("/var/vcap/sys/log/%s/errand.log", errandName)

	sshOpts, connOpts, err := opts.GatewayFlags.AsSSHOpts() //nolint:staticcheck
	if err != nil {
		c.logger.Debug("RunErrandCmd", "Failed to generate SSH opts, falling back to non-streaming: %s", err.Error())
		return c.runNormal(opts)
	}

	connOpts.RawOpts = append(connOpts.RawOpts, "-o", "StrictHostKeyChecking=yes")

	instanceSlugs := opts.InstanceGroupOrInstanceSlugFlags.Slugs //nolint:staticcheck

	var sshSlug boshdir.AllOrInstanceGroupOrInstanceSlug

	if len(instanceSlugs) > 0 {
		sshSlug = boshdir.NewAllOrInstanceGroupOrInstanceSlug(instanceSlugs[0].Name(), instanceSlugs[0].IndexOrID())
	} else {
		instanceGroup := c.findInstanceGroupForJob(errandName)
		if instanceGroup == "" {
			c.logger.Debug("RunErrandCmd", "Could not find instance group for errand '%s' in manifest, falling back to non-streaming", errandName)
			return c.runNormal(opts)
		}
		sshSlug = boshdir.NewAllOrInstanceGroupOrInstanceSlug(instanceGroup, "")
	}

	sshResult, err := c.deployment.SetUpSSH(sshSlug, sshOpts)
	if err != nil {
		c.logger.Debug("RunErrandCmd", "Failed to set up SSH, falling back to non-streaming: %s", err.Error())
		return c.runNormal(opts)
	}

	defer func() {
		_ = c.deployment.CleanUpSSH(sshSlug, sshOpts) //nolint:errcheck
	}()

	baselines := c.captureBaselines(connOpts, sshResult, logPath)

	taskID, err := c.deployment.StartErrand(
		errandName,
		opts.KeepAlive,
		opts.WhenChanged,
		instanceSlugs,
	)
	if err != nil {
		return err
	}

	pollInterval := time.Duration(*opts.StreamLogs) * time.Second

	var hostSlugs []string
	for _, host := range sshResult.Hosts {
		hostSlugs = append(hostSlugs, fmt.Sprintf("  - %s/%s", host.Job, host.IndexOrID))
	}
	c.ui.PrintBlock([]byte(fmt.Sprintf("\nStreaming logs for errand '%s' (poll interval %ds) on:\n%s\n\n", errandName, *opts.StreamLogs, strings.Join(hostSlugs, "\n"))))

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	doneCh := make(chan struct{})
	var wg sync.WaitGroup
	var outputMu sync.Mutex

	for i, host := range sshResult.Hosts {
		wg.Add(1)
		go c.streamLogs(connOpts, sshResult, host, logPath, baselines, i, doneCh, &wg, &outputMu, pollInterval)
	}

	type errandResult struct {
		results []boshdir.ErrandResult
		err     error
	}
	resultCh := make(chan errandResult, 1)
	go func() {
		results, err := c.deployment.WaitForErrandSilently(taskID)
		resultCh <- errandResult{results, err}
	}()

	var results []boshdir.ErrandResult
	var waitErr error

	select {
	case r := <-resultCh:
		results = r.results
		waitErr = r.err
	case <-sigCh:
		c.ui.PrintBlock([]byte("\nInterrupted. Cleaning up...\n"))
		close(doneCh)
		wg.Wait()
		return bosherr.Error("Command interrupted by signal")
	}

	close(doneCh)
	wg.Wait()

	c.finalFlush(connOpts, sshResult, logPath, baselines, &outputMu)

	c.deployment.WaitForErrand(taskID) //nolint:errcheck

	if waitErr != nil {
		return waitErr
	}

	return c.summarize(errandName, results, true)
}

type manifestSchema struct {
	InstanceGroups []manifestInstanceGroup `yaml:"instance_groups"`
}

type manifestInstanceGroup struct {
	Name string             `yaml:"name"`
	Jobs []manifestGroupJob `yaml:"jobs"`
}

type manifestGroupJob struct {
	Name string `yaml:"name"`
}

func (c RunErrandCmd) findInstanceGroupForJob(jobName string) string {
	manifestYAML, err := c.deployment.Manifest()
	if err != nil {
		c.logger.Debug("RunErrandCmd", "Failed to fetch manifest: %s", err.Error())
		return ""
	}

	var m manifestSchema
	if err := yaml.Unmarshal([]byte(manifestYAML), &m); err != nil {
		c.logger.Debug("RunErrandCmd", "Failed to parse manifest: %s", err.Error())
		return ""
	}

	for _, ig := range m.InstanceGroups {
		for _, job := range ig.Jobs {
			if job.Name == jobName {
				return ig.Name
			}
		}
	}

	return ""
}

func (c RunErrandCmd) captureBaselines(connOpts boshssh.ConnectionOpts, sshResult boshdir.SSHResult, logPath string) []int64 {
	baselines := make([]int64, len(sshResult.Hosts))

	for i, host := range sshResult.Hosts {
		tmpFile, err := c.fs.TempFile("errand-log-baseline")
		if err != nil {
			c.logger.Debug("RunErrandCmd", "Failed to create temp file for baseline: %s", err.Error())
			continue
		}
		tmpPath := tmpFile.Name()
		tmpFile.Close() //nolint:errcheck

		hostSlug := fmt.Sprintf("%s/%s", host.Job, host.IndexOrID)
		scpArgs := boshssh.NewSCPArgs([]string{fmt.Sprintf("%s:%s", hostSlug, logPath), tmpPath}, false)

		singleHostResult := boshdir.SSHResult{
			Hosts:           []boshdir.Host{host},
			GatewayUsername: sshResult.GatewayUsername,
			GatewayHost:     sshResult.GatewayHost,
		}

		err = c.scpRunner.Run(connOpts, singleHostResult, scpArgs)
		if err != nil {
			c.logger.Debug("RunErrandCmd", "Baseline SCP failed for %s (file may not exist yet): %s", hostSlug, err.Error())
			baselines[i] = 0
			c.fs.RemoveAll(tmpPath) //nolint:errcheck
			continue
		}

		stat, err := c.fs.Stat(tmpPath)
		if err != nil {
			c.logger.Debug("RunErrandCmd", "Failed to stat baseline file: %s", err.Error())
			baselines[i] = 0
		} else {
			baselines[i] = stat.Size()
		}

		c.fs.RemoveAll(tmpPath) //nolint:errcheck
	}

	return baselines
}

func (c RunErrandCmd) streamLogs(
	connOpts boshssh.ConnectionOpts,
	sshResult boshdir.SSHResult,
	host boshdir.Host,
	logPath string,
	offsets []int64,
	hostIdx int,
	doneCh <-chan struct{},
	wg *sync.WaitGroup,
	outputMu *sync.Mutex,
	pollInterval time.Duration,
) {
	defer wg.Done()

	hostSlug := fmt.Sprintf("%s/%s", host.Job, host.IndexOrID)

	singleHostResult := boshdir.SSHResult{
		Hosts:           []boshdir.Host{host},
		GatewayUsername: sshResult.GatewayUsername,
		GatewayHost:     sshResult.GatewayHost,
	}

	poll := func() {
		newOffset := c.fetchAndPrint(connOpts, singleHostResult, hostSlug, logPath, offsets[hostIdx], outputMu)
		offsets[hostIdx] = newOffset
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	poll()

	for {
		select {
		case <-doneCh:
			return
		case <-ticker.C:
			poll()
		}
	}
}

func (c RunErrandCmd) fetchAndPrint(
	connOpts boshssh.ConnectionOpts,
	sshResult boshdir.SSHResult,
	hostSlug string,
	logPath string,
	lastOffset int64,
	outputMu *sync.Mutex,
) int64 {
	tmpFile, err := c.fs.TempFile("errand-log-stream")
	if err != nil {
		c.logger.Debug("RunErrandCmd", "Failed to create temp file for streaming: %s", err.Error())
		return lastOffset
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()               //nolint:errcheck
	defer c.fs.RemoveAll(tmpPath) //nolint:errcheck

	scpArgs := boshssh.NewSCPArgs([]string{fmt.Sprintf("%s:%s", hostSlug, logPath), tmpPath}, false)

	err = c.scpRunner.Run(connOpts, sshResult, scpArgs)
	if err != nil {
		c.logger.Debug("RunErrandCmd", "SCP failed for %s, will retry next cycle: %s", hostSlug, err.Error())
		return lastOffset
	}

	content, err := c.fs.ReadFile(tmpPath)
	if err != nil {
		c.logger.Debug("RunErrandCmd", "Failed to read downloaded log file: %s", err.Error())
		return lastOffset
	}

	fileSize := int64(len(content))
	if fileSize <= lastOffset {
		return lastOffset
	}

	newContent := string(content[lastOffset:])

	var buf strings.Builder

	for _, line := range strings.Split(strings.TrimRight(newContent, "\n"), "\n") {
		fmt.Fprintf(&buf, "[%s] %s\n", hostSlug, line)
	}

	outputMu.Lock()
	c.ui.PrintBlock([]byte(buf.String()))
	outputMu.Unlock()

	return fileSize
}

func (c RunErrandCmd) finalFlush(connOpts boshssh.ConnectionOpts, sshResult boshdir.SSHResult, logPath string, offsets []int64, outputMu *sync.Mutex) {
	for i, host := range sshResult.Hosts {
		hostSlug := fmt.Sprintf("%s/%s", host.Job, host.IndexOrID)

		singleHostResult := boshdir.SSHResult{
			Hosts:           []boshdir.Host{host},
			GatewayUsername: sshResult.GatewayUsername,
			GatewayHost:     sshResult.GatewayHost,
		}

		offsets[i] = c.fetchAndPrint(connOpts, singleHostResult, hostSlug, logPath, offsets[i], outputMu)
	}
}

func (c RunErrandCmd) summarize(errandName string, results []boshdir.ErrandResult, streamed bool) error {
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

		stdout := result.Stdout
		if streamed {
			stdout = "(errand output was streamed above)"
		}

		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(instance),
			boshtbl.NewValueInt(result.ExitCode),
			boshtbl.NewValueString(stdout),
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
