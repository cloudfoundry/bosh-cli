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
	"github.com/cloudfoundry/bosh-gcscli/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Integration", func() {
	Context("general (Default Applicaton Credentials) configuration", func() {
		var ctx AssertContext
		BeforeEach(func() {
			ctx = NewAssertContext(AsDefaultCredentials)
		})
		AfterEach(func() {
			ctx.Cleanup()
		})

		baseConfigs, baseConfigErr := getBaseConfigs()
		encryptedConfigs, encryptedConfigErr := getEncryptedConfigs()
		configurations := append(baseConfigs, encryptedConfigs...)
		It("fetches configurations", func() {
			Expect(baseConfigErr).To(BeNil(), "failed to get configurations")
			Expect(encryptedConfigErr).To(BeNil(), "failed to get configurations")
		})

		DescribeTable("Blobstore lifecycle works",
			func(config *config.GCSCli) {
				ctx.AddConfig(config)
				AssertLifecycleWorks(gcsCLIPath, ctx)
			},
			configurations...)

		DescribeTable("Invalid Delete works",
			func(config *config.GCSCli) {
				ctx.AddConfig(config)
				AssertDeleteNonexistentWorks(gcsCLIPath, ctx)
			},
			configurations...)

		DescribeTable("Multipart Put works",
			func(config *config.GCSCli) {
				ctx.AddConfig(config)
				AssertMultipartPutWorks(gcsCLIPath, ctx)
			},
			configurations...)

		DescribeTable("Invalid Put should fail",
			func(config *config.GCSCli) {
				ctx.AddConfig(config)
				AssertBrokenSourcePutFails(gcsCLIPath, ctx)
			},
			configurations...)

		DescribeTable("Invalid Get should fail",
			func(config *config.GCSCli) {
				ctx.AddConfig(config)
				AssertGetNonexistentFails(gcsCLIPath, ctx)
			},
			configurations...)
	})
})
