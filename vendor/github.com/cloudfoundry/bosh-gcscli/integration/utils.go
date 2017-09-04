/*
 * Copyright 2017 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package integration

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os/exec"

	"time"

	"github.com/cloudfoundry/bosh-gcscli/config"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const alphanum = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// GenerateRandomString generates a random string of desired length (default: 25)
func GenerateRandomString(params ...int) string {
	size := 25
	if len(params) == 1 {
		size = params[0]
	}

	randBytes := make([]byte, size)
	for i := range randBytes {
		randBytes[i] = alphanum[rand.Intn(len(alphanum))]
	}
	return string(randBytes)
}

// MakeConfigFile creates a config file from a GCSCli config struct
func MakeConfigFile(cfg *config.GCSCli) string {
	cfgBytes, err := json.Marshal(cfg)
	Expect(err).ToNot(HaveOccurred())
	tmpFile, err := ioutil.TempFile("", "gcscli-test")
	Expect(err).ToNot(HaveOccurred())
	_, err = tmpFile.Write(cfgBytes)
	Expect(err).ToNot(HaveOccurred())
	err = tmpFile.Close()
	Expect(err).ToNot(HaveOccurred())
	return tmpFile.Name()
}

// MakeContentFile creates a temporary file with content to upload to GCS
func MakeContentFile(content string) string {
	tmpFile, err := ioutil.TempFile("", "gcscli-test-content")
	Expect(err).ToNot(HaveOccurred())
	_, err = tmpFile.Write([]byte(content))
	Expect(err).ToNot(HaveOccurred())
	err = tmpFile.Close()
	Expect(err).ToNot(HaveOccurred())
	return tmpFile.Name()
}

// RunGCSCLI run the gcscli and outputs the session
// after waiting for it to finish
func RunGCSCLI(gcsCLIPath, configPath, subcommand string,
	args ...string) (*gexec.Session, error) {

	cmdArgs := []string{
		"-c",
		configPath,
		subcommand,
	}
	cmdArgs = append(cmdArgs, args...)

	command := exec.Command(gcsCLIPath, cmdArgs...)
	gexecSession, err := gexec.Start(command,
		ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	if err != nil {
		return nil, err
	}
	gexecSession.Wait(1 * time.Minute)

	return gexecSession, nil
}
