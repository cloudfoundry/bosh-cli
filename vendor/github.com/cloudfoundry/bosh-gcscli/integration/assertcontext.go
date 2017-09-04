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
	"fmt"
	"os"

	"github.com/cloudfoundry/bosh-gcscli/config"

	"io/ioutil"

	. "github.com/onsi/gomega"
)

// GoogleAppCredentialsEnv is the environment variable
// expected to be populated with a path to a Service Account File for testing
// Application Default Credentials
const GoogleAppCredentialsEnv = "GOOGLE_APPLICATION_CREDENTIALS"

// ServiceAccountFileEnv is the environment variable
// expected to be populated with a Service Account File for testing
const ServiceAccountFileEnv = "GOOGLE_SERVICE_ACCOUNT"

// ServiceAccountFileMsg is the template used when ServiceAccountFileEnv
// has not been populated
const ServiceAccountFileMsg = "environment variable %s expected to contain a valid Service Account File but was empty"

// AssertContext contains the generated content to be used within tests.
//
// This allows Assertions to not have to worry about setup and teardown.
type AssertContext struct {
	// Config is the configuration used to
	Config *config.GCSCli
	// ConfigPath is the path to the file containing the
	// serialized content of Config.
	ConfigPath string

	// GCSFileName is the name of whatever blob is generated in an
	// assertion. It is the assert's responsibility to remove the blob.
	GCSFileName string
	// ExpectedString is the generated content used in an assertion.
	ExpectedString string
	// ContentFile is the path of the file containing ExpectedString
	ContentFile string

	// serviceAccountFile is the contents of a Service Account File
	// This is used in various contexts to authenticate
	serviceAccountFile string

	// serviceAccountPath is the path to a Service Account File created
	// to allow the use of Application Default Credentials
	serviceAccountPath string

	// options are the AssertContextConfigOption which are used to modify
	// the configuration whenever AddConfig is called.
	options []AssertContextConfigOption
}

// NewAssertContext returns an AssertContext with all fields
// which can be generated filled in.
func NewAssertContext(options ...AssertContextConfigOption) AssertContext {
	expectedString := GenerateRandomString()

	serviceAccountFile := os.Getenv(ServiceAccountFileEnv)
	Expect(serviceAccountFile).ToNot(BeEmpty(),
		fmt.Sprintf(ServiceAccountFileMsg, ServiceAccountFileEnv))

	return AssertContext{
		ExpectedString:     expectedString,
		ContentFile:        MakeContentFile(expectedString),
		GCSFileName:        GenerateRandomString(),
		serviceAccountFile: serviceAccountFile,
		options:            options,
	}
}

// AddConfig includes the config.GCSCli required for AssertContext
//
// Configuration is typically not available immediately before a test
// can be run, hence the need to add it later.
func (ctx *AssertContext) AddConfig(config *config.GCSCli) {
	ctx.Config = config
	for _, opt := range ctx.options {
		opt(ctx)
	}
	ctx.ConfigPath = MakeConfigFile(ctx.Config)
}

// AssertContextConfigOption is an option used for configuring an
// AssertContext's handling of the config.
//
// The behavior for an AssertContextAuthOption is applied when config is added
type AssertContextConfigOption func(ctx *AssertContext)

// AsReadOnlyCredentials configures the AssertContext to be used soley for
// public, read-only operations.
func AsReadOnlyCredentials(ctx *AssertContext) {
	conf := ctx.Config
	Expect(conf).ToNot(BeNil(),
		"cannot set read-only AssertContext without config")

	conf.CredentialsSource = config.NoneCredentialsSource
}

// AsStaticCredentials configures the AssertContext to authenticate explicitly
// using a Service Account File
func AsStaticCredentials(ctx *AssertContext) {
	conf := ctx.Config
	Expect(conf).ToNot(BeNil(),
		"cannot set static AssertContext without config")

	conf.ServiceAccountFile = ctx.serviceAccountFile
	conf.CredentialsSource = config.ServiceAccountFileCredentialsSource
}

// AsDefaultCredentials configures the AssertContext to authenticate using
// Application Default Credentials populated using the
// testing service account file.
func AsDefaultCredentials(ctx *AssertContext) {
	conf := ctx.Config
	Expect(conf).ToNot(BeNil(),
		"cannot set static AssertContext without config")

	tempFile, err := ioutil.TempFile("", "bosh-gcscli-service-account-file")
	Expect(err).ToNot(HaveOccurred())
	defer tempFile.Close()

	tempFile.WriteString(ctx.serviceAccountFile)

	ctx.serviceAccountPath = tempFile.Name()
	os.Setenv(GoogleAppCredentialsEnv, ctx.serviceAccountPath)

	conf.CredentialsSource = config.ApplicationDefaultCredentialsSource
}

// Clone returns a new AssertContext configured using the provided options.
// This overwrites the previous options of the context.
//
// This is useful in assertions where initial setup must be done under one
// form of authentication and the actual assertion is done under another.
//
// Note: The returned AssertContext is a distinct AssertContext, Cleanup must
// be called to remove testing files from the filesystem.
func (ctx *AssertContext) Clone(options ...AssertContextConfigOption) AssertContext {
	conf := *ctx.Config

	newContext := *ctx
	newContext.options = options
	newContext.AddConfig(&conf)

	return newContext
}

// Cleanup removes artifacts generated by the AssertContext.
func (ctx *AssertContext) Cleanup() {
	os.Remove(ctx.ConfigPath)
	os.Remove(ctx.ContentFile)

	if ctx.serviceAccountPath != "" {
		os.Remove(ctx.serviceAccountPath)
	}
	os.Unsetenv(GoogleAppCredentialsEnv)
}
