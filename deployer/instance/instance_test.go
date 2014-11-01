package instance_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployer/agentclient/fakes"
	fakebmas "github.com/cloudfoundry/bosh-micro-cli/deployer/applyspec/fakes"
	fakebmins "github.com/cloudfoundry/bosh-micro-cli/deployer/instance/fakes"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployer/agentclient"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployer/applyspec"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer/instance"
)

var _ = Describe("Instance", func() {
	var (
		fakeAgentClient            *fakebmagentclient.FakeAgentClient
		instance                   Instance
		applySpec                  bmstemcell.ApplySpec
		fakeTemplatesSpecGenerator *fakebmins.FakeTemplatesSpecGenerator
		fakeApplySpecFactory       *fakebmas.FakeApplySpecFactory
		deployment                 bmdepl.Deployment
		deploymentJob              bmdepl.Job
		stemcellJob                bmstemcell.Job
		fs                         *fakesys.FakeFileSystem
		logger                     boshlog.Logger
	)

	BeforeEach(func() {
		fakeTemplatesSpecGenerator = fakebmins.NewFakeTemplatesSpecGenerator()
		fakeTemplatesSpecGenerator.SetCreateBehavior(TemplatesSpec{
			BlobID:            "fake-blob-id",
			ArchiveSha1:       "fake-archive-sha1",
			ConfigurationHash: "fake-configuration-hash",
		}, nil)

		fakeAgentClient = fakebmagentclient.NewFakeAgentClient()
		stemcellJob = bmstemcell.Job{
			Name: "fake-job-name",
			Templates: []bmstemcell.Blob{
				{
					Name:        "first-job-name",
					Version:     "first-job-version",
					SHA1:        "first-job-sha1",
					BlobstoreID: "first-job-blobstore-id",
				},
				{
					Name:        "second-job-name",
					Version:     "second-job-version",
					SHA1:        "second-job-sha1",
					BlobstoreID: "second-job-blobstore-id",
				},
				{
					Name:        "third-job-name",
					Version:     "third-job-version",
					SHA1:        "third-job-sha1",
					BlobstoreID: "third-job-blobstore-id",
				},
			},
		}
		applySpec = bmstemcell.ApplySpec{
			Packages: map[string]bmstemcell.Blob{
				"first-package-name": bmstemcell.Blob{
					Name:        "first-package-name",
					Version:     "first-package-version",
					SHA1:        "first-package-sha1",
					BlobstoreID: "first-package-blobstore-id",
				},
				"second-package-name": bmstemcell.Blob{
					Name:        "second-package-name",
					Version:     "second-package-version",
					SHA1:        "second-package-sha1",
					BlobstoreID: "second-package-blobstore-id",
				},
			},
			Job: stemcellJob,
		}

		deploymentJob = bmdepl.Job{
			Name: "fake-manifest-job-name",
			Templates: []bmdepl.ReleaseJobRef{
				{Name: "first-job-name"},
				{Name: "third-job-name"},
			},
			RawProperties: map[interface{}]interface{}{
				"fake-property-key": "fake-property-value",
			},
			Networks: []bmdepl.JobNetwork{
				{
					Name:      "fake-network-name",
					StaticIPs: []string{"fake-network-ip"},
				},
			},
		}
		deployment = bmdepl.Deployment{
			Name: "fake-deployment-name",
			Jobs: []bmdepl.Job{
				deploymentJob,
			},
			Networks: []bmdepl.Network{
				{
					Name: "fake-network-name",
					Type: "fake-network-type",
				},
			},
		}

		fakeApplySpecFactory = fakebmas.NewFakeApplySpecFactory()

		logger = boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		instance = NewInstance(
			fakeAgentClient,
			fakeTemplatesSpecGenerator,
			fakeApplySpecFactory,
			"fake-mbus-url",
			fs,
			logger,
		)
	})

	Describe("Apply", func() {
		It("stops the agent", func() {
			err := instance.Apply(applySpec, deployment)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeAgentClient.StopCalled).To(BeTrue())
		})

		It("generates templates spec", func() {
			err := instance.Apply(applySpec, deployment)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeTemplatesSpecGenerator.CreateInputs).To(ContainElement(fakebmins.CreateInput{
				DeploymentJob:  deploymentJob,
				StemcellJob:    stemcellJob,
				DeploymentName: "fake-deployment-name",
				Properties: map[string]interface{}{
					"fake-property-key": "fake-property-value",
				},
				MbusURL: "fake-mbus-url",
			}))
		})

		It("creates apply spec", func() {
			err := instance.Apply(applySpec, deployment)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeApplySpecFactory.CreateInput).To(Equal(
				fakebmas.CreateInput{
					ApplySpec:      applySpec,
					DeploymentName: "fake-deployment-name",
					JobName:        "fake-manifest-job-name",
					NetworksSpec: map[string]interface{}{
						"fake-network-name": map[string]interface{}{
							"type":             "fake-network-type",
							"ip":               "fake-network-ip",
							"cloud_properties": map[string]interface{}{},
						},
					},
					ArchivedTemplatesBlobID: "fake-blob-id",
					ArchivedTemplatesSha1:   "fake-archive-sha1",
					TemplatesDirSha1:        "fake-configuration-hash",
				},
			))
		})

		It("sends apply spec to the agent", func() {
			agentApplySpec := bmas.ApplySpec{
				Deployment: "fake-deployment-name",
			}
			fakeApplySpecFactory.CreateApplySpec = agentApplySpec
			err := instance.Apply(applySpec, deployment)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeAgentClient.ApplyApplySpec).To(Equal(agentApplySpec))
		})

		Context("when creating templates spec fails", func() {
			BeforeEach(func() {
				fakeTemplatesSpecGenerator.CreateErr = errors.New("fake-template-err")
			})

			It("returns an error", func() {
				err := instance.Apply(applySpec, deployment)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-template-err"))
			})
		})

		Context("when sending apply spec to the agent fails", func() {
			BeforeEach(func() {
				fakeAgentClient.ApplyErr = errors.New("fake-agent-apply-err")
			})

			It("returns an error", func() {
				err := instance.Apply(applySpec, deployment)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-agent-apply-err"))
			})
		})

		Context("when stopping an agent fails", func() {
			BeforeEach(func() {
				fakeAgentClient.SetStopBehavior(errors.New("fake-stop-error"))
			})

			It("returns an error", func() {
				err := instance.Apply(applySpec, deployment)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-stop-error"))
			})
		})
	})

	Describe("Start", func() {
		It("starts agent services", func() {
			err := instance.Start()
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeAgentClient.StartCalled).To(BeTrue())
		})

		Context("when starting an agent fails", func() {
			BeforeEach(func() {
				fakeAgentClient.SetStartBehavior(errors.New("fake-start-error"))
			})

			It("returns an error", func() {
				err := instance.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-start-error"))
			})
		})
	})

	Describe("WaitToBeRunning", func() {
		BeforeEach(func() {
			fakeAgentClient.SetGetStateBehavior(bmagentclient.State{JobState: "pending"}, nil)
			fakeAgentClient.SetGetStateBehavior(bmagentclient.State{JobState: "pending"}, nil)
			fakeAgentClient.SetGetStateBehavior(bmagentclient.State{JobState: "running"}, nil)
		})

		It("waits until agent reports state as running", func() {
			err := instance.WaitToBeRunning(5, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeAgentClient.GetStateCalledTimes).To(Equal(3))
		})
	})
})
