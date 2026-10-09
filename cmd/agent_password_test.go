package cmd

import (
	"github.com/cloudfoundry/bosh-agent/v2/agentpassword"
	"github.com/cloudfoundry/bosh-utils/property"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("hashCPIAgentMbus", func() {
	It("hashes agent.mbus in place when present", func() {
		raw := "https://vcap:original-password@0.0.0.0:6868"
		properties := property.Map{
			"agent": property.Map{
				"mbus": raw,
			},
			"unrelated": "preserved",
		}

		err := hashCPIAgentMbus(properties)
		Expect(err).NotTo(HaveOccurred())

		hashed := properties["agent"].(property.Map)["mbus"].(string)
		Expect(hashed).To(HavePrefix("https://vcap:bosh-hmac-sha256$"))
		Expect(properties["unrelated"]).To(Equal("preserved"))

		verifierStr := "bosh-hmac-sha256$" + hashed[len("https://vcap:bosh-hmac-sha256$"):len(hashed)-len("@0.0.0.0:6868")]
		verifier, err := agentpassword.ParseVerifier(verifierStr)
		Expect(err).NotTo(HaveOccurred())
		Expect(verifier.Matches("original-password")).To(BeTrue())
	})

	It("is a no-op when props is nil or empty", func() {
		Expect(hashCPIAgentMbus(nil)).To(Succeed())
		Expect(hashCPIAgentMbus(property.Map{})).To(Succeed())
	})

	It("is a no-op when agent is absent or not a property.Map", func() {
		props := property.Map{"unrelated": "preserved"}
		Expect(hashCPIAgentMbus(props)).To(Succeed())
		Expect(props["unrelated"]).To(Equal("preserved"))

		props = property.Map{"agent": "not-a-map"}
		Expect(hashCPIAgentMbus(props)).To(Succeed())
		Expect(props["agent"]).To(Equal("not-a-map"))
	})

	It("is a no-op when agent.mbus is absent or not a string", func() {
		props := property.Map{"agent": property.Map{"other": "preserved"}}
		Expect(hashCPIAgentMbus(props)).To(Succeed())
		Expect(props["agent"].(property.Map)["other"]).To(Equal("preserved"))

		props = property.Map{"agent": property.Map{"mbus": 123}}
		Expect(hashCPIAgentMbus(props)).To(Succeed())
		Expect(props["agent"].(property.Map)["mbus"]).To(Equal(123))
	})

	It("returns error if HashURL fails", func() {
		props := property.Map{"agent": property.Map{"mbus": "invalid-url-%%%"}}
		Expect(hashCPIAgentMbus(props)).To(HaveOccurred())
	})
})

var _ = Describe("hashResourcePoolMbusURLs", func() {
	It("hashes urls entries in place and preserves unrelated keys", func() {
		raw := "https://vcap:original-password@0.0.0.0:6868"
		env := property.Map{
			"bosh": property.Map{
				"mbus": property.Map{
					"urls": property.List{raw},
					"cert": property.Map{"certificate": "preserved"},
				},
			},
			"unrelated": "preserved",
		}

		err := hashResourcePoolMbusURLs(env)
		Expect(err).NotTo(HaveOccurred())

		boshMap := env["bosh"].(property.Map)
		mbusMap := boshMap["mbus"].(property.Map)
		hashedURLs := mbusMap["urls"].(property.List)
		Expect(hashedURLs).To(HaveLen(1))

		hashed := hashedURLs[0].(string)
		Expect(hashed).To(HavePrefix("https://vcap:bosh-hmac-sha256$"))
		Expect(mbusMap["cert"].(property.Map)["certificate"]).To(Equal("preserved"))
		Expect(env["unrelated"]).To(Equal("preserved"))

		verifierStr := "bosh-hmac-sha256$" + hashed[len("https://vcap:bosh-hmac-sha256$"):len(hashed)-len("@0.0.0.0:6868")]
		verifier, err := agentpassword.ParseVerifier(verifierStr)
		Expect(err).NotTo(HaveOccurred())
		Expect(verifier.Matches("original-password")).To(BeTrue())
	})

	It("is a no-op when env is nil or empty", func() {
		Expect(hashResourcePoolMbusURLs(nil)).To(Succeed())
		Expect(hashResourcePoolMbusURLs(property.Map{})).To(Succeed())
	})

	It("is a no-op when missing a key at any level", func() {
		env := property.Map{"unrelated": "val"}
		Expect(hashResourcePoolMbusURLs(env)).To(Succeed())

		env = property.Map{"bosh": property.Map{"other": "val"}}
		Expect(hashResourcePoolMbusURLs(env)).To(Succeed())

		env = property.Map{"bosh": property.Map{"mbus": property.Map{"other": "val"}}}
		Expect(hashResourcePoolMbusURLs(env)).To(Succeed())
	})

	It("returns an error without URLs when bosh is not a map", func() {
		env := property.Map{"bosh": "not-a-map"}
		err := hashResourcePoolMbusURLs(env)
		Expect(err).To(MatchError("env.bosh must be a map"))
	})

	It("returns an error without URLs when mbus is not a map", func() {
		env := property.Map{"bosh": property.Map{"mbus": "not-a-map"}}
		err := hashResourcePoolMbusURLs(env)
		Expect(err).To(MatchError("env.bosh.mbus must be a map"))
	})

	It("returns an error without URLs when urls is not a list", func() {
		env := property.Map{"bosh": property.Map{"mbus": property.Map{"urls": "not-a-list"}}}
		err := hashResourcePoolMbusURLs(env)
		Expect(err).To(MatchError("env.bosh.mbus.urls must be a list"))
	})

	It("returns an error without URLs when any entry in urls is not a string", func() {
		env := property.Map{"bosh": property.Map{"mbus": property.Map{"urls": property.List{123}}}}
		err := hashResourcePoolMbusURLs(env)
		Expect(err).To(MatchError("env.bosh.mbus.urls entry must be a string"))
	})

	It("returns error if HashURL fails without leaking URL", func() {
		env := property.Map{"bosh": property.Map{"mbus": property.Map{"urls": property.List{"invalid-url-%%%"}}}}
		err := hashResourcePoolMbusURLs(env)
		Expect(err).To(HaveOccurred())
	})
})
