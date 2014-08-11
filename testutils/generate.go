package testutils

import (
	"io/ioutil"
	"os"
)

func GenerateDeploymentManifest(deploymentManifestFilePath string) error {
	return ioutil.WriteFile(deploymentManifestFilePath, []byte(""), os.ModePerm)
}
