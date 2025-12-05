package erbrenderer_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/templatescompiler/erbrenderer"
	"github.com/cloudfoundry/bosh-cli/v7/templatescompiler/erbrenderer/erbrendererfakes"
)

type testTemplateEvaluationStruct struct {
	Index             int                    `json:"index"`
	ID                string                 `json:"id"`
	GlobalProperties  map[string]interface{} `json:"global_properties"`
	ClusterProperties map[string]interface{} `json:"cluster_properties"`
	DefaultProperties map[string]interface{} `json:"default_properties"`
}

type testTemplateEvaluationContext struct {
	testStruct testTemplateEvaluationStruct
}

func (t testTemplateEvaluationContext) MarshalJSON() ([]byte, error) {
	jsonBytes, err := json.Marshal(t.testStruct)
	if err != nil {
		return []byte{}, err
	}

	return jsonBytes, nil
}

var _ = Describe("ErbRenderer", func() {
	Describe("Render", func() {
		var (
			fs     boshsys.FileSystem
			runner boshsys.CmdRunner
			logger boshlog.Logger

			tmpDir                   string
			erbTemplateFilepath      string
			renderedTemplatePath     string
			erbTemplateContent       string
			expectedTemplateContents string

			context erbrenderer.TemplateEvaluationContext

			erbRenderer erbrenderer.ERBRenderer
		)

		BeforeEach(func() {
			logger = boshlog.NewLogger(boshlog.LevelNone)
			fs = boshsys.NewOsFileSystemWithStrictTempRoot(logger)
			runner = boshsys.NewExecCmdRunner(logger)

			tmpDir = GinkgoT().TempDir()
			Expect(fs.ChangeTempRoot(tmpDir)).To(Succeed())

			templateName := "test_template.yml"
			erbTemplateName := fmt.Sprintf("%s.erb", templateName)
			erbTemplateFilepath = filepath.Join(tmpDir, erbTemplateName)
			renderedTemplatePath = filepath.Join(tmpDir, templateName)

			context = &testTemplateEvaluationContext{
				testTemplateEvaluationStruct{
					Index: 867_5309,
					GlobalProperties: map[string]interface{}{
						"property1": "global_value1",
					},
					ClusterProperties: map[string]interface{}{
						"property2": "cluster_value1",
					},
					DefaultProperties: map[string]interface{}{
						"property1": "default_value1",
						"property2": "default_value2",
						"property3": "default_value3",
					},
				},
			}
		})

		Context("when actually executing `ruby`", func() {
			BeforeEach(func() {
				erbRenderer = erbrenderer.NewERBRenderer(fs, runner, logger)
			})

			Context("with valid ERB", func() {
				BeforeEach(func() {
					erbTemplateContent = `---
property1: <%= p('property1') %>
property2: <%= p('property2') %>
property3: <%= p('property3') %>
`
					expectedTemplateContents = `---
property1: global_value1
property2: cluster_value1
property3: default_value3
`
					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())
				})

				It("renders output", func() {
					err := erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())
					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())

					Expect(expectedTemplateContents).To(Equal(string(templateBytes)))
				})
			})

			Describe("error handling within Ruby", func() {
				var (
					// see template_evaluation_context.rb
					rubyExceptionPrefixTemplate = "Error filling in template '%s' "
				)

				Context("with invalid ERB", func() {
					BeforeEach(func() {
						erbTemplateContent = `<%= raise "test error" %>`

						err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns an error with a known ruby exception", func() {
						err := erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("RuntimeError: test error"))
						Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(rubyExceptionPrefixTemplate, erbTemplateFilepath)))
					})

					Context("with missing ERB template", func() {
						It("returns an error with a known ruby exception", func() {
							invalidErbPath := "invalid/template.erb"
							err := erbRenderer.Render(invalidErbPath, renderedTemplatePath, context)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("<Errno::ENOENT: No such file or directory @ rb_sysopen - %s>", invalidErbPath)))
							Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(rubyExceptionPrefixTemplate, invalidErbPath)))
						})
					})

					Context("with context JSON which does not have the expected elements", func() {
						It("returns an error with a known ruby exception", func() {
							err := erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, &testTemplateEvaluationContext{})
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("undefined method `recursive_merge!'"))
							Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(rubyExceptionPrefixTemplate, erbTemplateFilepath)))
						})
					})
				})
			})
		})

		Describe("interactions with FileSystem", func() {
			var (
				fakeFs     *erbrendererfakes.FakeFileSystem
				fakeRunner *erbrendererfakes.FakeCmdRunner
			)

			BeforeEach(func() {
				fakeFs = &erbrendererfakes.FakeFileSystem{}
				fakeRunner = &erbrendererfakes.FakeCmdRunner{}

				erbTemplateContent = ""
				err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
				Expect(err).ToNot(HaveOccurred())
				erbRenderer = erbrenderer.NewERBRenderer(fakeFs, fakeRunner, logger)
			})

			It("cleans up temporary directory", func() {
				err := erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
				Expect(err).ToNot(HaveOccurred())
				Expect(fs.FileExists("fake-temp-dir")).To(BeFalse())
			})

			Context("when creating temporary directory fails", func() {
				var tempDirErr error

				BeforeEach(func() {
					tempDirErr = errors.New("fake-temp-dir-err")
					fakeFs.TempDirReturns("", tempDirErr)
				})

				It("returns an error", func() {
					err := erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(tempDirErr.Error()))
				})
			})

			Context("when writing renderer script fails", func() {
				var writerErr error

				BeforeEach(func() {
					writerErr = errors.New("fake-writer-err")
					fakeFs.WriteFileStringReturns(writerErr)
				})

				It("returns an error", func() {
					err := erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Writing renderer script: fake-writer-err"))
				})
			})

			Context("when writing renderer context fails", func() {
				var writerErr error

				BeforeEach(func() {
					writerErr = errors.New("fake-writer-err")
					fakeFs.WriteFileStringReturnsOnCall(0, nil)
					fakeFs.WriteFileStringReturnsOnCall(1, writerErr)
				})

				It("returns an error", func() {
					err := erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Writing context: fake-writer-err"))
				})
			})
		})

		Describe("interactions with CmdRunner", func() {
			var (
				fakeRunner *erbrendererfakes.FakeCmdRunner
			)

			BeforeEach(func() {
				fakeRunner = &erbrendererfakes.FakeCmdRunner{}
				erbRenderer = erbrenderer.NewERBRenderer(fs, fakeRunner, logger)
			})

			It("constructs ruby erb rendering command", func() {
				err := erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeRunner.RunComplexCommandCallCount()).To(Equal(1))
				command := fakeRunner.RunComplexCommandArgsForCall(0)
				Expect(command.Name).To(Equal("ruby"))
				Expect(command.Args[0]).To(MatchRegexp("/erb_render\\.rb$"))
				Expect(command.Args[1]).To(MatchRegexp("/erb-context\\.json$"))
				Expect(command.Args[2]).To(Equal(erbTemplateFilepath))
				Expect(command.Args[3]).To(Equal(renderedTemplatePath))
			})

			Context("when running ruby command fails", func() {
				BeforeEach(func() {
					fakeRunner.RunComplexCommandReturns("", "", 1, errors.New("fake-cmd-error"))
				})

				It("returns an error", func() {
					err := erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-cmd-error"))
				})
			})
		})
	})
})
