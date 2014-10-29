package templatescompiler_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmerbrenderer "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/erbrenderer"
	fakebmrender "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/erbrenderer/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

var _ = Describe("JobRenderer", func() {
	var (
		jobRenderer      JobRenderer
		fakeERBRenderer  *fakebmrender.FakeERBRenderer
		job              bmrel.Job
		context          bmerbrenderer.TemplateEvaluationContext
		fs               *fakesys.FakeFileSystem
		renderProperties map[string]interface{}
		srcPath          string
		dstPath          string
	)

	BeforeEach(func() {
		srcPath = "fake-src-path"
		dstPath = "fake-dst-path"
		renderProperties = map[string]interface{}{
			"fake-property-key": "fake-property-value",
		}

		job = bmrel.Job{
			Templates: map[string]string{
				"director.yml.erb": "config/director.yml",
			},
			ExtractedPath: "fake-src-path",
		}

		logger := boshlog.NewLogger(boshlog.LevelNone)
		context = NewJobEvaluationContext(job, renderProperties, "fake-deployment-name", logger)

		fakeERBRenderer = fakebmrender.NewFakeERBRender()

		fs = fakesys.NewFakeFileSystem()
		jobRenderer = NewJobRenderer(fakeERBRenderer, fs, logger)

		fakeERBRenderer.SetRenderBehavior(
			filepath.Join(srcPath, "templates/director.yml.erb"),
			filepath.Join(dstPath, "config/director.yml"),
			context,
			nil,
		)

		fakeERBRenderer.SetRenderBehavior(
			filepath.Join(srcPath, "monit"),
			filepath.Join(dstPath, "monit"),
			context,
			nil,
		)
	})

	Describe("Render", func() {
		It("renders job templates", func() {
			err := jobRenderer.Render(srcPath, dstPath, job, renderProperties, "fake-deployment-name")
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeERBRenderer.RenderInputs).To(Equal([]fakebmrender.RenderInput{
				{
					SrcPath: filepath.Join(srcPath, "templates/director.yml.erb"),
					DstPath: filepath.Join(dstPath, "config/director.yml"),
					Context: context,
				},
				{
					SrcPath: filepath.Join(srcPath, "monit"),
					DstPath: filepath.Join(dstPath, "monit"),
					Context: context,
				},
			}))
		})

		Context("when rendering fails", func() {
			It("returns an error", func() {

			})
		})
	})
})
