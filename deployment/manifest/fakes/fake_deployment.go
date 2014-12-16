package fakes

import (
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
)

func NewFakeReleaseJobRef() bmmanifest.ReleaseJobRef {
	return bmmanifest.ReleaseJobRef{
		Name:    "fake-release-job-ref-name",
		Release: "fake-release-job-ref-release",
	}
}

func NewFakeJob() bmmanifest.Job {
	return bmmanifest.Job{
		Name:      "fake-deployment-job",
		Instances: 1,
		Templates: []bmmanifest.ReleaseJobRef{NewFakeReleaseJobRef()},
	}
}

func NewFakeDeployment() bmmanifest.Manifest {
	return bmmanifest.Manifest{
		Name:          "fake-deployment-name",
		RawProperties: map[interface{}]interface{}{},
		Jobs:          []bmmanifest.Job{NewFakeJob()},
	}
}
