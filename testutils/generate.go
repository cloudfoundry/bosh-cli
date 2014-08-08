package testutils

import (
	"bytes"
	"io/ioutil"
	"os"

	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

func GenerateDeploymentManifest(deploymentManifestFilePath string) error {
	return ioutil.WriteFile(deploymentManifestFilePath, []byte(""), os.ModePerm)
}

func GenerateCPIRelease(fs boshsys.FileSystem, cpiFilePath string) error {
	jobContents := []TarFileContent{
		{
			Name: "job.MF",
			Body: `---
name: fake-job
templates:
  fake-template: fake-file
packages:
- fake-package
`,
		},
		{
			Name: "templates/fake-template",
			Body: "",
		},
	}
	var jobTarData bytes.Buffer
	GenerateTar(&jobTarData, jobContents)
	jobTarBytes := jobTarData.Bytes()

	packageContents := []TarFileContent{}
	var packageTarData bytes.Buffer
	GenerateTar(&packageTarData, packageContents)
	packageTarBytes := packageTarData.Bytes()

	contents := []TarFileContent{
		{
			Name: "release.MF",
			Body: `---
name: fake-release
version: fake-version
jobs:
- name: fake-job
  version: fake-version
  fingerprint: fake-fingerprint
  sha1: fake-sha1
packages:
- name: fake-package
  version: fake-version
  fingerprint: fake-fingerpritn
  sha1: fake-sha1
  dependencies: []
`,
		},
		{
			Name: "jobs/fake-job.tgz",
			Body: string(jobTarBytes),
		},
		{
			Name: "packages/fake-package.tgz",
			Body: string(packageTarBytes),
		},
	}

	return GenerateTarfile(fs, contents, cpiFilePath)
}
