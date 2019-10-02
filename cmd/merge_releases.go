package cmd

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-cli/cmd/opts"
	"github.com/cloudfoundry/bosh-cli/release"
	boshjob "github.com/cloudfoundry/bosh-cli/release/job"
	boshpkg "github.com/cloudfoundry/bosh-cli/release/pkg"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type MergeReleasesCmd struct {
	releaseReader release.Reader
	releaseWriter release.Writer
	fs            boshsys.FileSystem
}

func NewMergeReleasesCmd(releaseReader release.Reader, releaseWriter release.Writer, fs boshsys.FileSystem) MergeReleasesCmd {
	return MergeReleasesCmd{
		releaseReader: releaseReader,
		releaseWriter: releaseWriter,
		fs:            fs,
	}
}

func (c MergeReleasesCmd) Run(opts opts.MergeReleasesOpts) error {
	args := opts.Args

	release1, err := c.releaseReader.Read(args.ReleasePath1)
	if err != nil {
		return err
	}

	release2, err := c.releaseReader.Read(args.ReleasePath2)
	if err != nil {
		return err
	}

	defer release1.CleanUp()
	defer release2.CleanUp()

	if err := validateReleases(release1, release2); err != nil {
		return err
	}

	mergedJobs, err := mergeJobs(release1, release2)
	if err != nil {
		return err
	}

	mergedCompiledPackages, err := mergeCompiledPackages(release1, release2)
	if err != nil {
		return err
	}

	mergedRelease := release.NewRelease(
		release1.Name(),
		release1.Version(),
		release1.CommitHashWithMark(""),
		false,
		mergedJobs,
		[]*boshpkg.Package{},
		mergedCompiledPackages,
		release1.License(),
		"",
		c.fs,
	)

	path, err := c.releaseWriter.Write(mergedRelease, []string{})
	if err != nil {
		return err
	}

	return c.fs.Rename(path, args.TargetPath)
}

func validateReleases(release1, release2 release.Release) error {
	if release1.Name() != release2.Name() {
		return fmt.Errorf("Releases have conflicting names: %s, %s", release1.Name(), release2.Name())
	}

	if release1.Version() != release2.Version() {
		return fmt.Errorf("Releases have conflicting versions: %s, %s", release1.Version(), release2.Version())
	}

	if len(release1.Packages())+len(release2.Packages()) > 0 {
		return errors.New("Only compiled releases can be specified")
	}

	return nil
}

func mergeJobs(release1, release2 release.Release) ([]*boshjob.Job, error) {
	knownJobs := map[string]string{}

	totalJobs := append(release1.Jobs(), release2.Jobs()...)

	result := []*boshjob.Job{}
	for _, job := range totalJobs {
		if fingerprint, ok := knownJobs[job.Name()]; ok {
			if fingerprint != job.Fingerprint() {
				return nil, fmt.Errorf("Job %s has conflicting fingerprints (%s, %s)", job.Name(), fingerprint, job.Fingerprint())
			}
		} else {
			result = append(result, job)
			knownJobs[job.Name()] = job.Fingerprint()
		}
	}

	return result, nil
}

func mergeCompiledPackages(release1, release2 release.Release) ([]*boshpkg.CompiledPackage, error) {
	knownCompiledPackages := map[string]string{}

	totalCompiledPackages := append(release1.CompiledPackages(), release2.CompiledPackages()...)

	result := []*boshpkg.CompiledPackage{}
	for _, pkg := range totalCompiledPackages {
		if fingerprint, ok := knownCompiledPackages[pkg.Name()]; ok {
			if fingerprint != pkg.Fingerprint() {
				return nil, fmt.Errorf("Compiled package %s has conflicting fingerprints (%s, %s)", pkg.Name(), fingerprint, pkg.Fingerprint())
			}
		} else {
			result = append(result, pkg)
			knownCompiledPackages[pkg.Name()] = pkg.Fingerprint()
		}
	}

	return result, nil
}
