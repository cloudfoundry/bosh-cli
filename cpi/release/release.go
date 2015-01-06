package release

import (
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

const (
	ReleaseJobName    = "cpi"
	ReleaseBinaryName = "bin/cpi"
)

func IsCPIRelease(release bmrel.Release) bool {
	_, found := release.FindJobByName(ReleaseJobName)
	return found
}

func FindCPIRelease(releases []bmrel.Release) (cpiRelease bmrel.Release, found bool) {
	for _, release := range releases {
		if IsCPIRelease(release) {
			return release, true
		}
	}
	return nil, false
}
