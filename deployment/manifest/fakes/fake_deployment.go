package fakes

import (
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
)

func NewFakeReleaseJobRef() bmdeplmanifest.ReleaseJobRef {
	return bmdeplmanifest.ReleaseJobRef{
		Name:    "fake-release-job-ref-name",
		Release: "fake-release-job-ref-release",
	}
}

func NewFakeJob() bmdeplmanifest.Job {
	return bmdeplmanifest.Job{
		Name:      "fake-deployment-job",
		Instances: 1,
		Templates: []bmdeplmanifest.ReleaseJobRef{NewFakeReleaseJobRef()},
	}
}

func NewFakeDeployment() bmdeplmanifest.Manifest {
	return bmdeplmanifest.Manifest{
		Name:          "fake-deployment-name",
		RawProperties: map[interface{}]interface{}{},
		Jobs:          []bmdeplmanifest.Job{NewFakeJob()},
	}
}
