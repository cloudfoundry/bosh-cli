package templatescompiler_test

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	boshreljob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	. "github.com/cloudfoundry/bosh-cli/v7/release/resource"
	. "github.com/cloudfoundry/bosh-cli/v7/templatescompiler"
	"github.com/cloudfoundry/bosh-cli/v7/templatescompiler/templatescompilerfakes"
)

var _ = Describe("JobListRenderer", func() {
	var (
		logger boshlog.Logger

		mockJobRenderer *templatescompilerfakes.FakeJobRenderer

		releaseJobs          []boshreljob.Job
		releaseJobProperties map[string]*biproperty.Map
		jobProperties        biproperty.Map
		globalProperties     biproperty.Map
		deploymentName       string
		address              string

		renderedJobs []*templatescompilerfakes.FakeRenderedJob

		jobListRenderer JobListRenderer
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		mockJobRenderer = &templatescompilerfakes.FakeJobRenderer{}

		// release jobs are just passed through to JobRenderer.Render, so they do not need real contents
		releaseJobs = []boshreljob.Job{
			*boshreljob.NewJob(NewResource("fake-release-job-name-0", "", nil)),
			*boshreljob.NewJob(NewResource("fake-release-job-name-1", "", nil)),
		}

		releaseJobProperties = map[string]*biproperty.Map{
			"fake-release-job-name-0": &biproperty.Map{
				"fake-template-property": "fake-template-property-value",
			},
			"fake-release-job-name-1": &biproperty.Map{},
		}

		jobProperties = biproperty.Map{
			"fake-key": "fake-job-value",
		}

		globalProperties = biproperty.Map{
			"fake-key": "fake-global-value",
		}

		deploymentName = "fake-deployment-name"
		address = "1.2.3.4"

		renderedJobs = []*templatescompilerfakes.FakeRenderedJob{
			{},
			{},
		}

		jobListRenderer = NewJobListRenderer(mockJobRenderer, logger)
	})

	JustBeforeEach(func() {
		mockJobRenderer.RenderReturnsOnCall(0, renderedJobs[0], nil)
		mockJobRenderer.RenderReturnsOnCall(1, renderedJobs[1], nil)
	})

	Describe("Render", func() {
		It("returns a new RenderedJobList with all the RenderedJobs", func() {
			renderedJobList, err := jobListRenderer.Render(releaseJobs, releaseJobProperties, jobProperties, globalProperties, deploymentName, address)
			Expect(err).ToNot(HaveOccurred())
			Expect(renderedJobList.All()).To(Equal([]RenderedJob{
				renderedJobs[0],
				renderedJobs[1],
			}))

			Expect(mockJobRenderer.RenderCallCount()).To(Equal(2))

			actualJob0, actualProps0, actualJobProperties0, actualGlobalProperties0, actualDeploymentName0, actualAddress0 := mockJobRenderer.RenderArgsForCall(0)
			Expect(actualJob0).To(Equal(releaseJobs[0]))
			Expect(actualProps0).To(Equal(releaseJobProperties[releaseJobs[0].Name()]))
			Expect(actualJobProperties0).To(Equal(jobProperties))
			Expect(actualGlobalProperties0).To(Equal(globalProperties))
			Expect(actualDeploymentName0).To(Equal(deploymentName))
			Expect(actualAddress0).To(Equal(address))

			actualJob1, actualProps1, actualJobProperties1, actualGlobalProperties1, actualDeploymentName1, actualAddress1 := mockJobRenderer.RenderArgsForCall(1)
			Expect(actualJob1).To(Equal(releaseJobs[1]))
			Expect(actualProps1).To(Equal(releaseJobProperties[releaseJobs[1].Name()]))
			Expect(actualJobProperties1).To(Equal(jobProperties))
			Expect(actualGlobalProperties1).To(Equal(globalProperties))
			Expect(actualDeploymentName1).To(Equal(deploymentName))
			Expect(actualAddress1).To(Equal(address))
		})

		Context("when rendering a job fails", func() {
			JustBeforeEach(func() {
				mockJobRenderer.RenderReturnsOnCall(1, nil, bosherr.Error("fake-render-error"))
			})

			It("returns an error and cleans up any sucessfully rendered jobs", func() {
				_, err := jobListRenderer.Render(releaseJobs, releaseJobProperties, jobProperties, globalProperties, deploymentName, address)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-render-error"))
				Expect(renderedJobs[0].DeleteSilentlyCallCount()).To(Equal(1))
			})
		})
	})

})
