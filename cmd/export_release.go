package cmd

import (
	"fmt"
	"sort"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
)

type ExportReleaseCmd struct {
	deployment boshdir.Deployment
	downloader Downloader
}

func NewExportReleaseCmd(deployment boshdir.Deployment, downloader Downloader) ExportReleaseCmd {
	return ExportReleaseCmd{deployment: deployment, downloader: downloader}
}

func (c ExportReleaseCmd) Run(opts ExportReleaseOpts) error {
	rel := opts.Args.ReleaseSlug
	os := opts.Args.OSVersionSlug
	jobs := opts.Jobs

	err := c.ensureJobsValidForRelease(jobs, rel)
	if err != nil {
		return err
	}

	result, err := c.deployment.ExportRelease(rel, os, jobs)
	if err != nil {
		return err
	}

	prefix := fmt.Sprintf("%s-%s-%s-%s", rel.Name(), rel.Version(), os.OS(), os.Version())
	if len(jobs) != 0 {
		sort.Strings(jobs)
		prefix = fmt.Sprintf("%s-%s", strings.Join(jobs, "-"), prefix)
	}

	err = c.downloader.Download(
		result.BlobstoreID,
		result.SHA1,
		prefix,
		opts.Directory.Path,
	)
	if err != nil {
		return bosherr.WrapError(err, "Downloading exported release")
	}

	return nil
}
func (c ExportReleaseCmd) ensureJobsValidForRelease(jobs []string, rel boshdir.ReleaseSlug) error {
	releases, err := c.deployment.Releases()
	if err != nil {
		return err
	}
	releaseJobs := make([]string, 0)
	for _, release := range releases {
		if release.Name() == rel.Name() &&
			release.Version().String() == rel.Version() {
			jobs, err := release.Jobs()
			if err != nil {
				return err
			}
			for _, job := range jobs {
				releaseJobs = append(releaseJobs, job.Name)
			}
		}
	}
	for _, job := range jobs {
		if !contains(releaseJobs, job) {
			return fmt.Errorf(
				"Job '%s' for release '%s' doesn't exist", job, rel)
		}
	}
	return nil
}

func contains(s []string, searchterm string) bool {
	sort.Strings(s)
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}
