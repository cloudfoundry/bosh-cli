package deployment_test

import (
	// "errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

var _ = Describe("Deployment", func() {
	Describe("NetworksSpec", func() {
		var (
			deployment Deployment
		)

		Context("when the deployment has networks", func() {
			BeforeEach(func() {
				deployment = Deployment{
					Networks: []Network{
						{
							Name: "fake-network-name",
							Type: "dynamic",
						},
						{
							Name: "fake-manual-network-name",
							Type: "manual",
						},
						{
							Name: "vip",
							Type: "vip",
						},
						{
							Name: "fake",
							Type: "dynamic",
						},
					},
					Jobs: []Job{
						{
							Name: "bosh",
							Networks: []JobNetwork{
								{
									Name:      "vip",
									StaticIPs: []string{"1.2.3.4"},
								},
								{
									Name:      "fake-network-name",
									StaticIPs: []string{"5.6.7.8"},
								},
								{
									Name:      "fake-manual-network-name",
									StaticIPs: []string{"5.6.7.9"},
								},
							},
						},
					},
				}
			})

			It("is a map of the networks in spec form", func() {
				Expect(deployment.NetworksSpec("bosh")).To(Equal(map[string]interface{}{
					"fake-network-name": map[string]interface{}{
						"type":             "dynamic",
						"ip":               "5.6.7.8",
						"cloud_properties": map[string]interface{}{},
					},
					"fake-manual-network-name": map[string]interface{}{
						"type":             "manual",
						"ip":               "5.6.7.9",
						"cloud_properties": map[string]interface{}{},
					},
					"vip": map[string]interface{}{
						"type":             "vip",
						"ip":               "1.2.3.4",
						"cloud_properties": map[string]interface{}{},
					},
				}))
			})
		})

		Context("when the deployment does not have networks", func() {
			BeforeEach(func() {
				deployment = Deployment{
					Jobs: []Job{
						{
							Name: "bosh",
						},
					},
					Networks: []Network{},
				}
			})

			It("is an empty map", func() {
				Expect(deployment.NetworksSpec("bosh")).To(Equal(map[string]interface{}{}))
			})
		})

		Context("when the deployment does not have a job with requested name", func() {
			BeforeEach(func() {
				deployment = Deployment{}
			})

			It("returns an error", func() {
				networksSpec, err := deployment.NetworksSpec("bosh")
				Expect(networksSpec).To(Equal(map[string]interface{}{}))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Could not find job with name: bosh"))
			})
		})
	})
})
