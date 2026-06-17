package templatescompiler_test

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	boshreljob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	. "github.com/cloudfoundry/bosh-cli/v7/release/resource"
	. "github.com/cloudfoundry/bosh-cli/v7/templatescompiler"
	"github.com/cloudfoundry/bosh-cli/v7/templatescompiler/erbrenderer"
)

var _ = Describe("JobEvaluationContext", func() {
	var (
		releaseJob              *boshreljob.Job
		jobProperties           *biproperty.Map
		instanceGroupProperties biproperty.Map
		deploymentProperties    biproperty.Map
		erbRenderer             erbrenderer.ERBRenderer
		jobEvaluationContext    erbrenderer.TemplateEvaluationContext
		uuidGen                 *fakeuuid.FakeGenerator
		instanceSpec            InstanceSpec
	)

	BeforeEach(func() {
		releaseJob = boshreljob.NewJob(NewResource("fake-job-name", "fake-job-fp", nil))
		releaseJob.Properties = map[string]boshreljob.PropertyDefinition{
			"property1.subproperty1": boshreljob.PropertyDefinition{
				Default: "spec-default",
			},
			"property2.subproperty2": boshreljob.PropertyDefinition{
				Default: "spec-default",
			},
		}

		deploymentProperties = biproperty.Map{}
		instanceGroupProperties = biproperty.Map{}
		uuidGen = fakeuuid.NewFakeGenerator()
		jobProperties = nil

		instanceSpec = InstanceSpec{
			Name:           "fake-instance-group",
			Index:          2,
			AZ:             "z1",
			Bootstrap:      false,
			Address:        "1.2.3.4",
			PersistentDisk: 10240,
			Networks: map[string]NetworkSpecContext{
				"default": {IP: "1.2.3.4", Netmask: "255.255.255.0", Gateway: "1.2.3.1"},
			},
			ReleaseNamesByJob: map[string]string{
				"fake-job-name": "fake-release",
			},
		}
	})

	JustBeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)

		jobEvaluationContext = NewJobEvaluationContext(
			*releaseJob,
			jobProperties,
			instanceGroupProperties,
			deploymentProperties,
			"fake-deployment-name",
			instanceSpec,
			uuidGen,
			logger,
		)
	})

	act := func() RootContext {
		generatedJSON, err := jobEvaluationContext.MarshalJSON()
		Expect(err).ToNot(HaveOccurred())

		generatedContext := RootContext{}

		err = json.Unmarshal(generatedJSON, &generatedContext)
		Expect(err).ToNot(HaveOccurred())

		return generatedContext
	}

	It("exposes spec.name as the instance group name", func() {
		generatedContext := act()
		Expect(generatedContext.Name).To(Equal("fake-instance-group"))
	})

	It("exposes spec.index as the instance index", func() {
		generatedContext := act()
		Expect(generatedContext.Index).To(Equal(2))
	})

	It("exposes spec.az as the availability zone", func() {
		generatedContext := act()
		Expect(generatedContext.AZ).To(Equal("z1"))
	})

	It("exposes spec.bootstrap correctly", func() {
		generatedContext := act()
		Expect(generatedContext.Bootstrap).To(BeFalse())
	})

	Context("when instance is the first (index 0)", func() {
		BeforeEach(func() {
			instanceSpec.Index = 0
			instanceSpec.Bootstrap = true
		})

		It("sets spec.bootstrap to true", func() {
			generatedContext := act()
			Expect(generatedContext.Bootstrap).To(BeTrue())
		})
	})

	It("exposes spec.address", func() {
		generatedContext := act()
		Expect(generatedContext.Address).To(Equal("1.2.3.4"))
	})

	It("exposes spec.ip from the default network address", func() {
		generatedContext := act()
		Expect(generatedContext.IP).To(Equal("1.2.3.4"))
	})

	It("exposes spec.networks with real network data", func() {
		generatedContext := act()
		Expect(generatedContext.NetworkContexts["default"].IP).To(Equal("1.2.3.4"))
		Expect(generatedContext.NetworkContexts["default"].Netmask).To(Equal("255.255.255.0"))
		Expect(generatedContext.NetworkContexts["default"].Gateway).To(Equal("1.2.3.1"))
	})

	It("exposes spec.persistent_disk", func() {
		generatedContext := act()
		Expect(generatedContext.PersistentDisk).To(Equal(10240))
	})

	It("exposes spec.dns_domain_name as 'bosh'", func() {
		generatedContext := act()
		Expect(generatedContext.DnsDomainName).To(Equal("bosh"))
	})

	It("exposes spec.release.name from the release names map", func() {
		generatedContext := act()
		Expect(generatedContext.ReleaseContext.Name).To(Equal("fake-release"))
	})

	It("exposes spec.release.version as the job fingerprint", func() {
		generatedContext := act()
		Expect(generatedContext.ReleaseContext.Version).To(Equal("fake-job-fp"))
	})

	It("exposes spec.job.name as the instance group name", func() {
		generatedContext := act()
		Expect(generatedContext.JobContext.Name).To(Equal("fake-instance-group"))
	})

	It("it has id available in the spec", func() {
		uuidGen.GeneratedUUID = "fake-uuid"
		generatedContext := act()
		Expect(generatedContext.ID).To(Equal("fake-uuid"))
	})

	Context("when the UUID generator raises an error", func() {
		It("it raises an error", func() {
			uuidGen.GenerateError = errors.Error("boom")
			_, err := jobEvaluationContext.MarshalJSON()
			Expect(err).To(HaveOccurred())
			Ω(err.Error()).Should(ContainSubstring("Setting job eval context's ID to UUID"))
		})
	})

	Context("when no networks are set in InstanceSpec", func() {
		BeforeEach(func() {
			instanceSpec.Networks = nil
		})

		It("returns an empty map for spec.networks", func() {
			generatedContext := act()
			Expect(generatedContext.NetworkContexts).To(Equal(map[string]NetworkSpecContext{}))
		})
	})

	getValueFor := func(key string) string {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs := boshsys.NewOsFileSystem(logger)
		commandRunner := boshsys.NewExecCmdRunner(logger)
		erbRenderer = erbrenderer.NewERBRenderer(fs, commandRunner, logger)

		srcFile, err := os.CreateTemp("", "source.txt.erb")
		Expect(err).ToNot(HaveOccurred())
		defer os.Remove(srcFile.Name()) //nolint:errcheck

		erbContents := fmt.Sprintf("<%%= p('%s') %%>", key)
		_, err = srcFile.WriteString(erbContents)
		Expect(err).ToNot(HaveOccurred())

		destFile, err := fs.TempFile("dest.txt")
		Expect(err).ToNot(HaveOccurred())
		err = destFile.Close()
		Expect(err).ToNot(HaveOccurred())
		defer os.Remove(destFile.Name()) //nolint:errcheck

		jobEvaluationContext := NewJobEvaluationContext(
			*releaseJob,
			jobProperties,
			instanceGroupProperties,
			deploymentProperties,
			"fake-deployment-name",
			InstanceSpec{Address: "1.2.3.4"},
			uuidGen,
			logger,
		)

		err = erbRenderer.Render(srcFile.Name(), destFile.Name(), jobEvaluationContext)
		Expect(err).ToNot(HaveOccurred())
		contents, err := os.ReadFile(destFile.Name())
		Expect(err).ToNot(HaveOccurred())
		return (string)(contents)
	}

	Context("when a deployment and instance group set a property", func() {
		BeforeEach(func() {
			deploymentProperties = biproperty.Map{
				"property1": biproperty.Map{
					"subproperty1": "value-from-global-properties",
				},
			}

			instanceGroupProperties = biproperty.Map{
				"property1": biproperty.Map{
					"subproperty1": "value-from-cluster-properties",
				},
			}
		})

		It("gives precedence to the instance group value", func() {
			Expect(getValueFor("property1.subproperty1")).To(Equal("value-from-cluster-properties"))
		})
	})

	Context("when a deployment sets a property", func() {
		BeforeEach(func() {
			deploymentProperties = biproperty.Map{
				"property1": biproperty.Map{
					"subproperty1": "value-from-global-properties",
				},
			}
		})

		It("uses the value", func() {
			Expect(getValueFor("property1.subproperty1")).To(Equal("value-from-global-properties"))
		})
	})

	Context("when an instance group sets a property", func() {
		BeforeEach(func() {
			instanceGroupProperties = biproperty.Map{
				"property1": biproperty.Map{
					"subproperty1": "value-from-cluster-properties",
				},
			}
		})

		It("uses the value", func() {
			Expect(getValueFor("property1.subproperty1")).To(Equal("value-from-cluster-properties"))
		})
	})

	Context("when a property is not set", func() {
		It("uses the release's default value", func() {
			Expect(getValueFor("property1.subproperty1")).To(Equal("spec-default"))
		})
	})

	Context("when a job sets a property", func() {
		BeforeEach(func() {
			jobProperties = &biproperty.Map{
				"property1": biproperty.Map{
					"subproperty1": "job-property",
				},
			}
		})

		It("uses the value", func() {
			Expect(getValueFor("property1.subproperty1")).To(Equal("job-property"))
		})

		Context("when the instance group also sets a property", func() {
			instanceGroupProperties = biproperty.Map{
				"property2": biproperty.Map{
					"subproperty2": "instance-group-property",
				},
			}

			It("is not used", func() {
				Expect(getValueFor("property2.subproperty2")).To(Equal("spec-default"))
			})
		})
	})

	Context("when the job sets a property to an empty hash ({})", func() {
		BeforeEach(func() {
			jobProperties = &biproperty.Map{}
		})

		Context("when an instance group sets a property", func() {
			BeforeEach(func() {
				instanceGroupProperties = biproperty.Map{
					"property1": biproperty.Map{
						"subproperty1": "value-from-instance-group-properties",
					},
				}
			})

			It("does not use the instance group value", func() {
				Expect(getValueFor("property1.subproperty1")).To(Equal("spec-default"))
			})
		})

		Context("when an deployment sets a property", func() {
			BeforeEach(func() {
				deploymentProperties = biproperty.Map{
					"property1": biproperty.Map{
						"subproperty1": "value-from-global-properties",
					},
				}
			})

			It("does not use the instance group value", func() {
				Expect(getValueFor("property1.subproperty1")).To(Equal("spec-default"))
			})
		})
	})

	Describe("link support in RootContext JSON", func() {
		var linkSpec LinkSpec

		BeforeEach(func() {
			linkSpec = LinkSpec{
				DeploymentName: "fake-deployment-name",
				Domain:         "bosh",
				InstanceGroup:  "fake-instance-group",
				DefaultNetwork: "default",
				GroupName:      "mysql.bosh.fake-deployment-name.bosh",
				Instances: []LinkInstanceSpec{
					{Name: "fake-instance-group", ID: "inst-0", Index: 0, Bootstrap: true, AZ: "z1", Address: "10.0.0.1"},
				},
				Properties:           map[string]interface{}{"port": float64(13306)},
				UseLinkDNSNames:      false,
				UseShortDNSAddresses: false,
			}

			instanceSpec.Links = map[string]map[string]LinkSpec{
				"fake-job-name": {
					"mysql": linkSpec,
				},
			}
		})

		It("includes job_template_name in the rendered JSON", func() {
			generatedContext := act()
			Expect(generatedContext.JobTemplateName).To(Equal("fake-job-name"))
		})

		It("includes the resolved links in the rendered JSON", func() {
			generatedContext := act()
			Expect(generatedContext.Links).To(HaveKey("fake-job-name"))
			jobLinks := generatedContext.Links["fake-job-name"]
			Expect(jobLinks).To(HaveKey("mysql"))
			mysql := jobLinks["mysql"]
			Expect(mysql.DeploymentName).To(Equal("fake-deployment-name"))
			Expect(mysql.InstanceGroup).To(Equal("fake-instance-group"))
			Expect(mysql.Instances).To(HaveLen(1))
			Expect(mysql.Instances[0].Address).To(Equal("10.0.0.1"))
			Expect(mysql.Instances[0].Bootstrap).To(BeTrue())
		})

		It("renders link() in ERB templates using the resolved links", func() {
			logger := boshlog.NewLogger(boshlog.LevelNone)
			fs := boshsys.NewOsFileSystem(logger)
			commandRunner := boshsys.NewExecCmdRunner(logger)
			erbRenderer = erbrenderer.NewERBRenderer(fs, commandRunner, logger)

			srcFile, err := os.CreateTemp("", "source.txt.erb")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(srcFile.Name()) //nolint:errcheck

			_, err = srcFile.WriteString(`<%= link("mysql").instances.map(&:address).join(",") %>`)
			Expect(err).ToNot(HaveOccurred())

			destFile, err := fs.TempFile("dest.txt")
			Expect(err).ToNot(HaveOccurred())
			err = destFile.Close()
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(destFile.Name()) //nolint:errcheck

			ctx := NewJobEvaluationContext(
				*releaseJob,
				jobProperties,
				instanceGroupProperties,
				deploymentProperties,
				"fake-deployment-name",
				instanceSpec,
				uuidGen,
				logger,
			)

			err = erbRenderer.Render(srcFile.Name(), destFile.Name(), ctx)
			Expect(err).ToNot(HaveOccurred())
			contents, err := os.ReadFile(destFile.Name())
			Expect(err).ToNot(HaveOccurred())
			Expect(string(contents)).To(Equal("10.0.0.1"))
		})

		Context("when InstanceSpec.Links is nil", func() {
			BeforeEach(func() {
				instanceSpec.Links = nil
			})

			It("returns an empty links map", func() {
				generatedContext := act()
				Expect(generatedContext.Links).To(Equal(map[string]map[string]LinkSpec{}))
			})
		})
	})
})
