package stemcell_test

import (
	"github.com/cloudfoundry/bosh-cli/v7/stemcell"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

var _ = Describe("Agent features", func() {
	It("detects supported features and preserves them when repacked", func() {
		for _, tc := range []struct {
			metadata string
			supports bool
		}{
			{"name: legacy\napi_version: 3\n", false},
			{"agent_features: [unrelated-feature]\n", false},
			{"agent_features: [http-password-hmac-sha256]\n", true},
		} {
			var manifest stemcell.Manifest
			Expect(yaml.Unmarshal([]byte(tc.metadata), &manifest)).To(Succeed())
			Expect(manifest.SupportsAgentFeature("http-password-hmac-sha256")).To(Equal(tc.supports))
			packed, err := yaml.Marshal(manifest)
			Expect(err).NotTo(HaveOccurred())
			var unpacked stemcell.Manifest
			Expect(yaml.Unmarshal(packed, &unpacked)).To(Succeed())
			Expect(unpacked.SupportsAgentFeature("http-password-hmac-sha256")).To(Equal(tc.supports))
		}
	})
})
