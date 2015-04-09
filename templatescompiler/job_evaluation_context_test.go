package templatescompiler_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bmproperty "github.com/cloudfoundry/bosh-init/common/property"
	bmreljob "github.com/cloudfoundry/bosh-init/release/job"
	"github.com/cloudfoundry/bosh-init/templatescompiler/erbrenderer"

	. "github.com/cloudfoundry/bosh-init/templatescompiler"
)

var _ = Describe("JobEvaluationContext", func() {
	var (
		generatedContext RootContext

		releaseJob        bmreljob.Job
		clusterProperties bmproperty.Map
		globalProperties  bmproperty.Map
	)
	BeforeEach(func() {
		generatedContext = RootContext{}

		releaseJob = bmreljob.Job{
			Name: "fake-job-name",
			Properties: map[string]bmreljob.PropertyDefinition{
				"fake-default-property1.fake-default-property2": bmreljob.PropertyDefinition{
					Default: "value-from-job-defaults",
				},
			},
		}

		clusterProperties = bmproperty.Map{
			"fake-job-property1": bmproperty.Map{
				"fake-job-property2": "value-from-cluster-properties",
			},
		}

		globalProperties = bmproperty.Map{
			"fake-global-property1": bmproperty.Map{
				"fake-global-property2": "value-from-global-properties",
			},
		}
	})

	JustBeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)

		jobEvaluationContext := NewJobEvaluationContext(
			releaseJob,
			clusterProperties,
			globalProperties,
			"fake-deployment-name",
			logger,
		)

		generatedJSON, err := jobEvaluationContext.MarshalJSON()
		Expect(err).ToNot(HaveOccurred())

		err = json.Unmarshal(generatedJSON, &generatedContext)
		Expect(err).ToNot(HaveOccurred())
	})

	It("it has a network context section with empty IP", func() {
		Expect(generatedContext.NetworkContexts["default"].IP).To(Equal(""))
	})

	var erbRenderer erbrenderer.ERBRenderer
	getValueFor := func(key string) string {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs := boshsys.NewOsFileSystem(logger)
		commandRunner := boshsys.NewExecCmdRunner(logger)
		erbRenderer = erbrenderer.NewERBRenderer(fs, commandRunner, logger)

		srcFile, err := ioutil.TempFile("", "source.txt.erb")
		Expect(err).ToNot(HaveOccurred())
		defer os.Remove(srcFile.Name())

		erbContents := fmt.Sprintf("<%%= p('%s') %%>", key)
		_, err = srcFile.WriteString(erbContents)
		Expect(err).ToNot(HaveOccurred())

		destFile, err := fs.TempFile("dest.txt")
		Expect(err).ToNot(HaveOccurred())
		err = destFile.Close()
		Expect(err).ToNot(HaveOccurred())
		defer os.Remove(destFile.Name())

		jobEvaluationContext := NewJobEvaluationContext(
			releaseJob,
			clusterProperties,
			globalProperties,
			"fake-deployment-name",
			logger,
		)

		err = erbRenderer.Render(srcFile.Name(), destFile.Name(), jobEvaluationContext)
		Expect(err).ToNot(HaveOccurred())
		contents, err := ioutil.ReadFile(destFile.Name())
		Expect(err).ToNot(HaveOccurred())
		return (string)(contents)
	}

	Context("when a cluster property overrides a global property or default value", func() {
		BeforeEach(func() {
			releaseJob = bmreljob.Job{
				Name: "fake-job-name",
				Properties: map[string]bmreljob.PropertyDefinition{
					"fake-overridden-property1.fake-overridden-property2": bmreljob.PropertyDefinition{},
				},
			}

			globalProperties = bmproperty.Map{
				"fake-overridden-property1": bmproperty.Map{
					"fake-overridden-property2": "value-from-global-properties",
				},
			}

			clusterProperties = bmproperty.Map{
				"fake-overridden-property1": bmproperty.Map{
					"fake-overridden-property2": "value-from-cluster-properties",
				},
			}
		})

		It("prefers cluster values over global values", func() {
			Expect(getValueFor("fake-overridden-property1.fake-overridden-property2")).
				To(Equal("value-from-cluster-properties"))
		})
	})

	Context("when a global property overrides a default property", func() {
		BeforeEach(func() {
			releaseJob = bmreljob.Job{
				Name: "fake-job-name",
				Properties: map[string]bmreljob.PropertyDefinition{
					"fake-overridden-property1.fake-overridden-property2": bmreljob.PropertyDefinition{
						Default: "value-from-job-defaults",
					},
				},
			}

			globalProperties = bmproperty.Map{
				"fake-overridden-property1": bmproperty.Map{
					"fake-overridden-property2": "value-from-global-properties",
				},
			}

			clusterProperties = bmproperty.Map{}
		})

		It("prefers global values over default values", func() {
			Expect(getValueFor("fake-overridden-property1.fake-overridden-property2")).
				To(Equal("value-from-global-properties"))
		})
	})

	Context("when a cluster property overrides a default property", func() {
		BeforeEach(func() {
			releaseJob = bmreljob.Job{
				Name: "fake-job-name",
				Properties: map[string]bmreljob.PropertyDefinition{
					"fake-overridden-property1.fake-overridden-property2": bmreljob.PropertyDefinition{
						Default: "value-from-job-defaults",
					},
				},
			}

			globalProperties = bmproperty.Map{}

			clusterProperties = bmproperty.Map{
				"fake-overridden-property1": bmproperty.Map{
					"fake-overridden-property2": "value-from-cluster-properties",
				},
			}
		})

		It("prefers cluster values over default values", func() {
			Expect(getValueFor("fake-overridden-property1.fake-overridden-property2")).
				To(Equal("value-from-cluster-properties"))
		})
	})

	Context("when a property is not specified in cluster or global properties", func() {
		BeforeEach(func() {
			releaseJob = bmreljob.Job{
				Name: "fake-job-name",
				Properties: map[string]bmreljob.PropertyDefinition{
					"fake-overridden-property1.fake-overridden-property2": bmreljob.PropertyDefinition{
						Default: "value-from-job-defaults",
					},
				},
			}

			globalProperties = bmproperty.Map{}

			clusterProperties = bmproperty.Map{}
		})

		It("uses the property's default value", func() {
			Expect(getValueFor("fake-overridden-property1.fake-overridden-property2")).
				To(Equal("value-from-job-defaults"))
		})
	})
})
