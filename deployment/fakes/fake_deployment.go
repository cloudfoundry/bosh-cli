package fakes

import (
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

func NewFakeReleaseJobRef() bmdepl.ReleaseJobRef {
	return bmdepl.ReleaseJobRef{
		Name:    "fake-release-job-ref-name",
		Release: "fake-release-job-ref-release",
	}
}

func NewFakeJob() bmdepl.Job {
	return bmdepl.Job{
		Name:      "fake-deployment-job",
		Instances: 1,
		Templates: []bmdepl.ReleaseJobRef{NewFakeReleaseJobRef()},
	}
}

func NewFakeDeployment() bmdepl.Deployment {
	return bmdepl.Deployment{
		Name:       "fake-deployment-name",
		Properties: map[string]interface{}{},
		Jobs:       []bmdepl.Job{NewFakeJob()},
	}
}
