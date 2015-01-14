package deployment

import (
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
)

type Deployment interface {
}

type deployment struct {
	manifest bmdeplmanifest.Manifest
}

func NewDeployment(manifest bmdeplmanifest.Manifest) Deployment {
	return &deployment{
		manifest: manifest,
	}
}
