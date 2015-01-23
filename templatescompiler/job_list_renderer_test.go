package templatescompiler_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.google.com/p/gomock/gomock"
	mock_template "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/mocks"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

var _ = Describe("JobListRenderer", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		logger boshlog.Logger

		mockJobRenderer *mock_template.MockJobRenderer

		releaseJobs    []bmrel.Job
		jobProperties  map[string]interface{}
		deploymentName string

		renderedJobs []*mock_template.MockRenderedJob

		jobListRenderer JobListRenderer

		expectRender0 *gomock.Call
		expectRender1 *gomock.Call
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		mockJobRenderer = mock_template.NewMockJobRenderer(mockCtrl)

		// release jobs are just passed through to JobRenderer.Render, so they do not need real contents
		releaseJobs = []bmrel.Job{
			{Name: "fake-release-job-name-0"},
			{Name: "fake-release-job-name-1"},
		}

		jobProperties = map[string]interface{}{
			"fake-key": "fake-value",
		}

		deploymentName = "fake-deployment-name"

		renderedJobs = []*mock_template.MockRenderedJob{
			mock_template.NewMockRenderedJob(mockCtrl),
			mock_template.NewMockRenderedJob(mockCtrl),
		}

		jobListRenderer = NewJobListRenderer(mockJobRenderer, logger)
	})

	JustBeforeEach(func() {
		expectRender0 = mockJobRenderer.EXPECT().Render(releaseJobs[0], jobProperties, deploymentName).Return(renderedJobs[0], nil)
		expectRender1 = mockJobRenderer.EXPECT().Render(releaseJobs[1], jobProperties, deploymentName).Return(renderedJobs[1], nil)
	})

	Describe("Render", func() {
		It("returns a new RenderedJobList with all the RenderedJobs", func() {
			renderedJobList, err := jobListRenderer.Render(releaseJobs, jobProperties, deploymentName)
			Expect(err).ToNot(HaveOccurred())
			Expect(renderedJobList.All()).To(Equal([]RenderedJob{
				renderedJobs[0],
				renderedJobs[1],
			}))
		})

		Context("when rendering a job fails", func() {
			JustBeforeEach(func() {
				expectRender1.Return(nil, bosherr.Error("fake-render-error"))
			})

			It("returns an error and cleans up any sucessfully rendered jobs", func() {
				renderedJobs[0].EXPECT().DeleteSilently()

				_, err := jobListRenderer.Render(releaseJobs, jobProperties, deploymentName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-render-error"))
			})
		})
	})

})
