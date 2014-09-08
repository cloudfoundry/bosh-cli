package testutils

import (
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

func GenerateDeploymentManifest(deploymentManifestFilePath string, fs boshsys.FileSystem) error {
	manifestContents := `
---
name: fake-deployment
cloud_provider:
  properties:
    fake_cpi_specified_property:
      second_level: fake_specified_property_value
`
	return fs.WriteFileString(deploymentManifestFilePath, manifestContents)
}
