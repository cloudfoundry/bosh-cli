package manifestfakes

import (
	biproperty "github.com/cloudfoundry/bosh-utils/property"

	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
)

func NewFakeReleaseJobRef() bideplmanifest.ReleaseJobRef {
	return bideplmanifest.ReleaseJobRef{
		Name:    "fake-release-job-ref-name",
		Release: "fake-release-job-ref-release",
	}
}

func NewFakeJob() bideplmanifest.Job {
	return bideplmanifest.Job{
		Name:      "fake-deployment-job",
		Instances: 1,
		Templates: []bideplmanifest.ReleaseJobRef{NewFakeReleaseJobRef()},
	}
}

func NewFakeDeployment() bideplmanifest.Manifest {
	return bideplmanifest.Manifest{
		Name:       "fake-deployment-name",
		Properties: biproperty.Map{},
		Jobs:       []bideplmanifest.Job{NewFakeJob()},
	}
}
