package ui_test

import (
	"bytes"
	"io"
	"time"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/ui"
	. "github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("LockingUI", func() {
	var (
		logOutBuffer             *bytes.Buffer
		uiOutBuffer, uiErrBuffer *bytes.Buffer
		uiOut, uiErr             io.Writer
		logger                   boshlog.Logger
		ui                       UI
	)

	BeforeEach(func() {
		uiOutBuffer = bytes.NewBufferString("")
		uiOut = uiOutBuffer
		uiErrBuffer = bytes.NewBufferString("")
		uiErr = uiErrBuffer

		logOutBuffer = bytes.NewBufferString("")
		logger = boshlog.NewWriterLogger(boshlog.LevelDebug, logOutBuffer)
	})

	JustBeforeEach(func() {
		ui = NewLockingUI(NewWriterUI(uiOut, uiErr, logger))
	})

	Describe("ErrorLinef in parellel", func() {
		It("prints to errWriter with a trailing newline", func() {
			go ui.ErrorLinef("fake-error-line")
			go ui.ErrorLinef("fake-error-line")
			time.Sleep(100 * time.Millisecond)

			Expect(uiOutBuffer.String()).To(Equal(""))
			Expect(uiErrBuffer.String()).To(ContainSubstring("fake-error-line\nfake-error-line\n"))
		})

		Context("when writing fails", func() {
			BeforeEach(func() {
				reader, writer := io.Pipe()
				uiErr = writer
				reader.Close()
			})

			It("logs an error", func() {
				go ui.ErrorLinef("fake-error-line")
				go ui.ErrorLinef("fake-error-line")
				time.Sleep(100 * time.Millisecond)

				Expect(uiOutBuffer.String()).To(Equal(""))
				Expect(uiErrBuffer.String()).To(Equal(""))
				Expect(logOutBuffer.String()).To(MatchRegexp(".*UI.ErrorLinef failed \\(message='fake-error-line'\\).*\n.*UI.ErrorLinef failed \\(message='fake-error-line'\\)"))
			})
		})
	})

	Describe("PrintLinef in parellel", func() {
		It("prints to outWriter with a trailing newline", func() {
			go ui.PrintLinef("fake-line")
			go ui.PrintLinef("fake-line")
			time.Sleep(100 * time.Millisecond)

			Expect(uiOutBuffer.String()).To(ContainSubstring("fake-line\nfake-line\n"))
		})

		Context("when writing fails", func() {
			BeforeEach(func() {
				reader, writer := io.Pipe()
				uiOut = writer
				reader.Close()
			})

			It("logs an error", func() {
				go ui.PrintLinef("fake-start")
				go ui.PrintLinef("fake-start")
				time.Sleep(100 * time.Millisecond)

				Expect(uiOutBuffer.String()).To(Equal(""))
				Expect(uiErrBuffer.String()).To(Equal(""))
				Expect(logOutBuffer.String()).To(MatchRegexp(".*UI.PrintLinef failed \\(message='fake-start'\\).*\n.*UI.PrintLinef failed \\(message='fake-start'\\)"))
			})
		})
	})

	Describe("BeginLinef in parallel", func() {
		It("prints to outWriter", func() {
			go ui.BeginLinef("fake-start")
			go ui.BeginLinef("fake-start")
			time.Sleep(100 * time.Millisecond)
			Expect(uiOutBuffer.String()).To(ContainSubstring("fake-startfake-start"))
		})

		Context("when writing fails", func() {
			BeforeEach(func() {
				reader, writer := io.Pipe()
				uiOut = writer
				reader.Close()
			})

			It("logs an error", func() {
				go ui.BeginLinef("fake-start")
				go ui.BeginLinef("fake-start")
				time.Sleep(100 * time.Millisecond)

				Expect(uiOutBuffer.String()).To(Equal(""))
				Expect(uiErrBuffer.String()).To(Equal(""))
				Expect(logOutBuffer.String()).To(MatchRegexp(".*UI.BeginLinef failed \\(message='fake-start'\\).*\n.*UI.BeginLinef failed \\(message='fake-start'\\)"))
			})
		})
	})

	Describe("EndLinef in parallel", func() {
		It("prints to outWriter with a trailing newline", func() {
			go ui.EndLinef("fake-end")
			go ui.EndLinef("fake-end")
			time.Sleep(100 * time.Millisecond)
			Expect(uiOutBuffer.String()).To(ContainSubstring("fake-end\nfake-end\n"))
		})

		Context("when writing fails", func() {
			BeforeEach(func() {
				reader, writer := io.Pipe()
				uiOut = writer
				reader.Close()
			})

			It("logs an error", func() {
				go ui.EndLinef("fake-start")
				go ui.EndLinef("fake-start")
				time.Sleep(100 * time.Millisecond)

				Expect(uiOutBuffer.String()).To(Equal(""))
				Expect(uiErrBuffer.String()).To(Equal(""))
				Expect(logOutBuffer.String()).To(MatchRegexp(".*UI.EndLinef failed \\(message='fake-start'\\).*\n.*UI.EndLinef failed \\(message='fake-start'\\)"))
			})
		})
	})

	Describe("PrintBlock in parallel", func() {
		It("prints to outWriter as is", func() {
			go ui.PrintBlock([]byte("block"))
			go ui.PrintBlock([]byte("block"))
			time.Sleep(100 * time.Millisecond)
			Expect(uiOutBuffer.String()).To(Equal("blockblock"))
			Expect(uiErrBuffer.String()).To(Equal(""))
		})

		Context("when writing fails", func() {
			BeforeEach(func() {
				reader, writer := io.Pipe()
				uiOut = writer
				reader.Close()
			})

			It("logs an error", func() {
				go ui.PrintBlock([]byte("block"))
				go ui.PrintBlock([]byte("block"))
				time.Sleep(100 * time.Millisecond)
				Expect(uiOutBuffer.String()).To(Equal(""))
				Expect(uiErrBuffer.String()).To(Equal(""))
				Expect(logOutBuffer.String()).To(MatchRegexp(".*UI.PrintBlock failed \\(message='block'\\).*\n.*UI.PrintBlock failed \\(message='block'\\)"))
			})
		})
	})

	Describe("PrintErrorBlock in parallel", func() {
		It("prints to outWriter as is", func() {
			go ui.PrintErrorBlock("block")
			go ui.PrintErrorBlock("block")
			time.Sleep(100 * time.Millisecond)
			Expect(uiOutBuffer.String()).To(Equal("blockblock"))
			Expect(uiErrBuffer.String()).To(Equal(""))
		})

		Context("when writing fails", func() {
			BeforeEach(func() {
				reader, writer := io.Pipe()
				uiOut = writer
				reader.Close()
			})

			It("logs an error", func() {
				go ui.PrintErrorBlock("block")
				go ui.PrintErrorBlock("block")
				time.Sleep(100 * time.Millisecond)
				Expect(uiOutBuffer.String()).To(Equal(""))
				Expect(uiErrBuffer.String()).To(Equal(""))
				Expect(logOutBuffer.String()).To(MatchRegexp(".*UI.PrintErrorBlock failed \\(message='block'\\).*\n.*UI.PrintErrorBlock failed \\(message='block'\\)"))
			})
		})
	})

	Describe("PrintTable in parallel", func() {
		It("prints 2 tables in parallel and does not mix the output", func() {
			table := Table{
				Title:   "Title",
				Content: "things",
				Header:  []Header{NewHeader("Header1"), NewHeader("Header2")},

				Rows: [][]Value{
					{ValueString{S: "r1c1"}, ValueString{S: "r1c2"}},
					{ValueString{S: "r2c1"}, ValueString{S: "r2c2"}},
				},

				Notes:         []string{"note1", "note2"},
				BackgroundStr: ".",
				BorderStr:     "|",
			}
			go ui.PrintTable(table)
			go ui.PrintTable(table)
			time.Sleep(100 * time.Millisecond)
			Expect("\n" + uiOutBuffer.String()).To(Equal(`
Title

Header1|Header2|
r1c1...|r1c2|
r2c1...|r2c2|

note1
note2

2 things
Title

Header1|Header2|
r1c1...|r1c2|
r2c1...|r2c2|

note1
note2

2 things
`))
		})
	})

	Describe("IsInteractive", func() {
		It("returns true", func() {
			Expect(ui.IsInteractive()).To(BeTrue())
		})
	})

	Describe("Flush", func() {
		It("does nothing", func() {
			Expect(func() { ui.Flush() }).ToNot(Panic())
		})
	})
})
