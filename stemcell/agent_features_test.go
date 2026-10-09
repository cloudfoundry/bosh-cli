package stemcell_test

import (
	"github.com/cloudfoundry/bosh-cli/v7/stemcell"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

var _ = Describe("Agent features", func() {
	DescribeTable("detects supported features",
		func(metadata string, supports bool) {
			var manifest stemcell.Manifest
			Expect(yaml.Unmarshal([]byte(metadata), &manifest)).To(Succeed())
			Expect(manifest.SupportsAgentFeature("http-password-hmac-sha256")).To(Equal(supports))
		},
		Entry("legacy metadata without a feature list", "name: legacy\napi_version: 3\n", false),
		Entry("an unrelated feature", "agent_features: [unrelated-feature]\n", false),
		Entry("the HTTP password feature", "agent_features: [http-password-hmac-sha256]\n", true),
	)

	DescribeTable("preserves agent features when repacked",
		func(features []string) {
			manifest := stemcell.Manifest{AgentFeatures: features}
			packed, err := yaml.Marshal(manifest)
			Expect(err).NotTo(HaveOccurred())
			var unpacked stemcell.Manifest
			Expect(yaml.Unmarshal(packed, &unpacked)).To(Succeed())
			Expect(unpacked.AgentFeatures).To(Equal(features))
		},
		Entry("no feature list", []string(nil)),
		Entry("an unrelated feature", []string{"unrelated-feature"}),
		Entry("the HTTP password feature", []string{"http-password-hmac-sha256"}),
	)
})
