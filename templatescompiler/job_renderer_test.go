package templatescompiler_test

import (
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	boshreljob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	. "github.com/cloudfoundry/bosh-cli/v7/release/resource"
	. "github.com/cloudfoundry/bosh-cli/v7/templatescompiler"
	bierbrenderer "github.com/cloudfoundry/bosh-cli/v7/templatescompiler/erbrenderer"
	fakebirender "github.com/cloudfoundry/bosh-cli/v7/templatescompiler/erbrenderer/fakes"
)

var _ = Describe("JobRenderer", func() {
	var (
		jobRenderer          JobRenderer
		fakeERBRenderer      *fakebirender.FakeERBRenderer
		job                  *boshreljob.Job
		context              bierbrenderer.TemplateEvaluationContext
		fs                   *fakesys.FakeFileSystem
		releaseJobProperties biproperty.Map
		jobProperties        biproperty.Map
		globalProperties     biproperty.Map
		srcPath              string
		dstPath              string
	)

	BeforeEach(func() {
		srcPath = "fake-src-path"
		dstPath = "fake-dst-path"

		releaseJobProperties = biproperty.Map{
			"fake-release-job-name": biproperty.Map{
				"fake-template-property": "fake-template-property-value",
			},
		}

		jobProperties = biproperty.Map{
			"fake-property-key": "fake-job-property-value",
		}

		globalProperties = biproperty.Map{
			"fake-property-key": "fake-global-property-value",
		}

		job = boshreljob.NewExtractedJob(NewResourceWithBuiltArchive("cpi", "job-fp", "path", "sha1"), srcPath, fs)
		job.Templates = map[string]string{
			"director.yml.erb": "config/director.yml",
		}

		logger := boshlog.NewLogger(boshlog.LevelNone)

		context = NewJobEvaluationContext(*job, &releaseJobProperties, jobProperties, globalProperties, "fake-deployment-name", "1.2.3.4", nil, logger)

		fakeERBRenderer = fakebirender.NewFakeERBRender()

		fs = fakesys.NewFakeFileSystem()
		jobRenderer = NewJobRenderer(fakeERBRenderer, fs, nil, logger)

		_ = fakeERBRenderer.SetRenderBehavior(
			filepath.Join(srcPath, "templates/director.yml.erb"),
			filepath.Join(dstPath, "config/director.yml"),
			context,
			nil,
		)

		_ = fakeERBRenderer.SetRenderBehavior(
			filepath.Join(srcPath, "monit"),
			filepath.Join(dstPath, "monit"),
			context,
			nil,
		)

		fs.TempDirDir = dstPath
	})

	AfterEach(func() {
		err := fs.RemoveAll(dstPath)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Render", func() {
		It("renders job templates", func() {
			renderedjob, err := jobRenderer.Render(*job, &releaseJobProperties, jobProperties, globalProperties, "fake-deployment-name", "1.2.3.4")
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeERBRenderer.RenderInputs).To(Equal([]fakebirender.RenderInput{
				{
					SrcPath: filepath.Join(srcPath, "templates/director.yml.erb"),
					DstPath: filepath.Join(renderedjob.Path(), "config/director.yml"),
					Context: context,
				},
				{
					SrcPath: filepath.Join(srcPath, "monit"),
					DstPath: filepath.Join(renderedjob.Path(), "monit"),
					Context: context,
				},
			}))
		})

		Context("when rendering fails", func() {
			BeforeEach(func() {
				_ = fakeERBRenderer.SetRenderBehavior(
					filepath.Join(srcPath, "templates/director.yml.erb"),
					filepath.Join(dstPath, "config/director.yml"),
					context,
					bosherr.Error("fake-template-render-error"),
				)
			})

			It("returns an error", func() {
				_, err := jobRenderer.Render(*job, &releaseJobProperties, jobProperties, globalProperties, "fake-deployment-name", "1.2.3.4")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-template-render-error"))
			})
		})
	})
})
