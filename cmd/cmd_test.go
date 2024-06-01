package cmd_test

import (
	"errors"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

var _ = Describe("Cmd", func() {
	var (
		ui      *fakeui.FakeUI
		confUI  *boshui.ConfUI
		fs      *fakesys.FakeFileSystem
		boshCmd cmd.Cmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		logger := boshlog.NewLogger(boshlog.LevelNone)
		confUI = boshui.NewWrappingConfUI(ui, logger)

		fs = fakesys.NewFakeFileSystem()

		deps := cmd.NewBasicDeps(confUI, logger)
		deps.FS = fs

		boshCmd = cmd.NewCmd(opts.BoshOpts{}, nil, deps)
	})

	Describe("Execute", func() {
		It("succeeds executing at least one command", func() {
			boshCmd.Opts = &opts.InterpolateOpts{}

			err := boshCmd.Execute()
			Expect(err).ToNot(HaveOccurred())

			Expect(ui.Blocks).To(Equal([]string{"null\n"}))
		})

		It("prints message if specified", func() {
			boshCmd.Opts = &opts.MessageOpts{Message: "output"}

			err := boshCmd.Execute()
			Expect(err).ToNot(HaveOccurred())

			Expect(ui.Blocks).To(Equal([]string{"output"}))
		})

		It("allows to enable json output", func() {
			boshCmd.BoshOpts = opts.BoshOpts{JSONOpt: true}
			boshCmd.Opts = &opts.InterpolateOpts{}

			err := boshCmd.Execute()
			Expect(err).ToNot(HaveOccurred())

			confUI.Flush()

			Expect(ui.Blocks[0]).To(ContainSubstring(`Blocks": [`))
		})

		Describe("color", func() {
			executeCmdAndPrintTable := func() {
				err := boshCmd.Execute()
				Expect(err).ToNot(HaveOccurred())

				// Tables have emboldened header values
				confUI.PrintTable(boshtbl.Table{Header: []boshtbl.Header{boshtbl.NewHeader("State")}})
			}

			It("has color in the output enabled by default", func() {
				boshCmd.BoshOpts = opts.BoshOpts{}
				boshCmd.Opts = &opts.InterpolateOpts{}

				executeCmdAndPrintTable()

				// Expect that header values are bold
				Expect(ui.Tables[0].HeaderFormatFunc).ToNot(BeNil())
			})

			It("allows to disable color in the output", func() {
				boshCmd.BoshOpts = opts.BoshOpts{NoColorOpt: true}
				boshCmd.Opts = &opts.InterpolateOpts{}

				executeCmdAndPrintTable()

				// Expect that header values are empty because they were not emboldened
				Expect(ui.Tables[0].HeaderFormatFunc).To(BeNil())
			})
		})

		It("returns error if changing tmp root fails", func() {
			fs.ChangeTempRootErr = errors.New("fake-err")

			err := boshCmd.Execute()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("fake-err"))
		})

		It("returns error for unknown commands", func() {
			err := boshCmd.Execute()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Unhandled command: <nil>"))
		})
	})
})
