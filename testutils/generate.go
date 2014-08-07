package testutils

import (
	"io/ioutil"
	"os"

	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

func GenerateDeploymentManifest(deploymentManifestFilePath string) error {
	return ioutil.WriteFile(deploymentManifestFilePath, []byte(""), os.ModePerm)
}

func GenerateCPIRelease(fs boshsys.FileSystem, cpiFilePath string) error {
	contents := []TarFileContent{
		{
			Name: "release.MF",
			Body: `---
name: fake-release
version: fake-version
`,
		},
	}

	return GenerateTarfile(fs, contents, cpiFilePath)
}
