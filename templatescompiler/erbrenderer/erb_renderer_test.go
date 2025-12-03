package erbrenderer_test

import (
	"encoding/json"
	"errors"
	"path/filepath"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/templatescompiler/erbrenderer"
)

type testTemplateEvaluationStruct struct {
	Index          int                    `json:"index"`
	ID             string                 `json:"id"`
	TestProperties map[string]interface{} `json:"test_properties"`
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
			fs          *fakesys.FakeFileSystem
			runner      *fakesys.FakeCmdRunner
			erbRenderer erbrenderer.ERBRenderer
			context     erbrenderer.TemplateEvaluationContext
		)

		BeforeEach(func() {
			logger := boshlog.NewLogger(boshlog.LevelNone)
			fs = fakesys.NewFakeFileSystem()
			runner = fakesys.NewFakeCmdRunner()
			context = &testTemplateEvaluationContext{}

			erbRenderer = erbrenderer.NewERBRenderer(fs, runner, logger)
			fs.TempDirDir = "fake-temp-dir"
		})

		It("constructs ruby erb rendering command", func() {
			err := erbRenderer.Render("fake-src-path", "fake-dst-path", context)
			Expect(err).ToNot(HaveOccurred())
			Expect(runner.RunComplexCommands).To(Equal([]boshsys.Command{
				{
					Name: "ruby",
					Args: []string{
						filepath.Join("fake-temp-dir", "erb-render.rb"),
						filepath.Join("fake-temp-dir", "erb-context.json"),
						"fake-src-path",
						"fake-dst-path",
					},
				},
			}))
		})

		It("cleans up temporary directory", func() {
			err := erbRenderer.Render("fake-src-path", "fake-dst-path", context)
			Expect(err).ToNot(HaveOccurred())
			Expect(fs.FileExists("fake-temp-dir")).To(BeFalse())
		})

		Context("when creating temporary directory fails", func() {
			It("returns an error", func() {
				fs.TempDirError = errors.New("fake-temp-dir-error")
				err := erbRenderer.Render("fake-src-path", "fake-dst-path", context)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-temp-dir-error"))
			})
		})

		Context("when writing renderer script fails", func() {
			It("returns an error", func() {
				fs.WriteFileError = errors.New("fake-write-error")
				err := erbRenderer.Render("src-path", "dst-path", context)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-write-error"))
			})
		})

		Context("when running ruby command fails", func() {
			BeforeEach(func() {
				runner.AddCmdResult(
					"ruby fake-temp-dir/erb-render.rb fake-temp-dir/erb-context.json fake-src-path fake-dst-path",
					fakesys.FakeCmdResult{
						Error: errors.New("fake-cmd-error"),
					})
			})

			It("returns an error", func() {
				err := erbRenderer.Render("fake-src-path", "fake-dst-path", context)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-cmd-error"))
			})
		})
	})
})
