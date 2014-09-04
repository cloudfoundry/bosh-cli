package testutils

import (
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

func GenerateDeploymentManifest(deploymentManifestFilePath string, fs boshsys.FileSystem) error {
	manifestContents := `
---
name: fake-deployment
cloud_provider:
  properties: {}
`
	return fs.WriteFileString(deploymentManifestFilePath, manifestContents)
}
