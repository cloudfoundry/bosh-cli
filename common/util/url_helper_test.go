package util_test

import (
	urlhelper "github.com/cloudfoundry/bosh-cli/common/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RedactBasicAuth", func() {
	for _, scheme := range []string{"http", "https"} {
		Context("When "+scheme+" URL contains basic auth credentials", func() {
			It("redacts user and password", func() {
				redactedUrl := urlhelper.RedactBasicAuth(scheme + "://me:secret@artifacts.com")
				Expect(redactedUrl).To(Equal(scheme + "://<redacted>:<redacted>@artifacts.com"))
			})
		})

		Context("When "+scheme+" URL contains no basic auth credentials", func() {
			It("returns input URL", func() {
				redactedUrl := urlhelper.RedactBasicAuth(scheme + "://artifacts.com")
				Expect(redactedUrl).To(Equal(scheme + "://artifacts.com"))
			})
		})
	}

})
