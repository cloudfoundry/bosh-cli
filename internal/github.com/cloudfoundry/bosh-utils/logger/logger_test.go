package logger_test

import (
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/logger"
)

func expectedLogFormat(tag, msg string) string {
	return fmt.Sprintf("\\[%s\\] [0-9]{4}/[0-9]{2}/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} %s\n", tag, msg)
}

func captureOutputs(f func()) (stdout, stderr []byte) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, err := os.Pipe()
	Expect(err).ToNot(HaveOccurred())

	rErr, wErr, err := os.Pipe()
	Expect(err).ToNot(HaveOccurred())

	os.Stdout = wOut
	os.Stderr = wErr

	f()

	outC := make(chan []byte)
	errC := make(chan []byte)

	go func() {
		bytes, _ := ioutil.ReadAll(rOut)
		outC <- bytes

		bytes, _ = ioutil.ReadAll(rErr)
		errC <- bytes
	}()

	err = wOut.Close()
	Expect(err).ToNot(HaveOccurred())

	err = wErr.Close()
	Expect(err).ToNot(HaveOccurred())

	stdout = <-outC
	stderr = <-errC

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return
}

var _ = Describe("Levelify", func() {
	It("converts strings into LogLevel constants", func() {
		level, err := Levelify("NONE")
		Expect(err).ToNot(HaveOccurred())
		Expect(level).To(Equal(LevelNone))

		level, err = Levelify("none")
		Expect(err).ToNot(HaveOccurred())
		Expect(level).To(Equal(LevelNone))

		level, err = Levelify("DEBUG")
		Expect(err).ToNot(HaveOccurred())
		Expect(level).To(Equal(LevelDebug))

		level, err = Levelify("debug")
		Expect(err).ToNot(HaveOccurred())
		Expect(level).To(Equal(LevelDebug))

		level, err = Levelify("INFO")
		Expect(err).ToNot(HaveOccurred())
		Expect(level).To(Equal(LevelInfo))

		level, err = Levelify("info")
		Expect(err).ToNot(HaveOccurred())
		Expect(level).To(Equal(LevelInfo))

		level, err = Levelify("WARN")
		Expect(err).ToNot(HaveOccurred())
		Expect(level).To(Equal(LevelWarn))

		level, err = Levelify("warn")
		Expect(err).ToNot(HaveOccurred())
		Expect(level).To(Equal(LevelWarn))

		level, err = Levelify("ERROR")
		Expect(err).ToNot(HaveOccurred())
		Expect(level).To(Equal(LevelError))

		level, err = Levelify("error")
		Expect(err).ToNot(HaveOccurred())
		Expect(level).To(Equal(LevelError))
	})

	It("errors on unknown input", func() {
		_, err := Levelify("unknown")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Unknown LogLevel string 'unknown', expected one of [DEBUG, INFO, WARN, ERROR, NONE]"))

		_, err = Levelify("")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Unknown LogLevel string '', expected one of [DEBUG, INFO, WARN, ERROR, NONE]"))
	})
})

var _ = Describe("Logger", func() {
	Describe("Debug", func() {
		It("logs the formatted message to Logger.out at the debug level", func() {
			stdout, stderr := captureOutputs(func() {
				logger := NewLogger(LevelDebug)
				logger.Debug("TAG", "some %s info to log", "awesome")
			})

			expectedContent := expectedLogFormat("TAG", "DEBUG - some awesome info to log")
			Expect(stdout).To(MatchRegexp(expectedContent))
			Expect(stderr).ToNot(MatchRegexp(expectedContent))
		})
	})

	Describe("DebugWithDetails", func() {
		It("logs the message to Logger.out at the debug level with specially formatted arguments", func() {
			stdout, stderr := captureOutputs(func() {
				logger := NewLogger(LevelDebug)
				logger.DebugWithDetails("TAG", "some info to log", "awesome")
			})

			expectedContent := expectedLogFormat("TAG", "DEBUG - some info to log")
			Expect(stdout).To(MatchRegexp(expectedContent))
			Expect(stderr).ToNot(MatchRegexp(expectedContent))

			expectedDetails := "\n********************\nawesome\n********************"
			Expect(stdout).To(ContainSubstring(expectedDetails))
			Expect(stderr).ToNot(ContainSubstring(expectedDetails))
		})
	})

	Describe("Info", func() {
		It("logs the formatted message to Logger.out at the info level", func() {
			stdout, stderr := captureOutputs(func() {
				logger := NewLogger(LevelInfo)
				logger.Info("TAG", "some %s info to log", "awesome")
			})

			expectedContent := expectedLogFormat("TAG", "INFO - some awesome info to log")
			Expect(stdout).To(MatchRegexp(expectedContent))
			Expect(stderr).ToNot(MatchRegexp(expectedContent))
		})
	})

	Describe("Warn", func() {
		It("logs the formatted message to Logger.err at the warn level", func() {
			stdout, stderr := captureOutputs(func() {
				logger := NewLogger(LevelWarn)
				logger.Warn("TAG", "some %s info to log", "awesome")
			})

			expectedContent := expectedLogFormat("TAG", "WARN - some awesome info to log")
			Expect(stdout).ToNot(MatchRegexp(expectedContent))
			Expect(stderr).To(MatchRegexp(expectedContent))
		})
	})

	Describe("Error", func() {
		It("logs the formatted message to Logger.err at the error level", func() {
			stdout, stderr := captureOutputs(func() {
				logger := NewLogger(LevelError)
				logger.Error("TAG", "some %s info to log", "awesome")
			})

			expectedContent := expectedLogFormat("TAG", "ERROR - some awesome info to log")
			Expect(stdout).ToNot(MatchRegexp(expectedContent))
			Expect(stderr).To(MatchRegexp(expectedContent))
		})
	})

	Describe("ErrorWithDetails", func() {
		It("logs the message to Logger.err at the error level with specially formatted arguments", func() {
			stdout, stderr := captureOutputs(func() {
				logger := NewLogger(LevelError)
				logger.ErrorWithDetails("TAG", "some error to log", "awesome")
			})

			expectedContent := expectedLogFormat("TAG", "ERROR - some error to log")
			Expect(stdout).ToNot(MatchRegexp(expectedContent))
			Expect(stderr).To(MatchRegexp(expectedContent))

			expectedDetails := "\n********************\nawesome\n********************"
			Expect(stdout).ToNot(ContainSubstring(expectedDetails))
			Expect(stderr).To(ContainSubstring(expectedDetails))
		})
	})

	It("log level debug", func() {
		stdout, stderr := captureOutputs(func() {
			logger := NewLogger(LevelDebug)
			logger.Debug("DEBUG", "some debug log")
			logger.Info("INFO", "some info log")
			logger.Warn("WARN", "some warn log")
			logger.Error("ERROR", "some error log")
		})

		Expect(stdout).To(ContainSubstring("DEBUG"))
		Expect(stdout).To(ContainSubstring("INFO"))
		Expect(stderr).To(ContainSubstring("WARN"))
		Expect(stderr).To(ContainSubstring("ERROR"))
	})

	It("log level info", func() {
		stdout, stderr := captureOutputs(func() {
			logger := NewLogger(LevelInfo)
			logger.Debug("DEBUG", "some debug log")
			logger.Info("INFO", "some info log")
			logger.Warn("WARN", "some warn log")
			logger.Error("ERROR", "some error log")
		})

		Expect(stdout).ToNot(ContainSubstring("DEBUG"))
		Expect(stdout).To(ContainSubstring("INFO"))
		Expect(stderr).To(ContainSubstring("WARN"))
		Expect(stderr).To(ContainSubstring("ERROR"))
	})

	It("log level warn", func() {
		stdout, stderr := captureOutputs(func() {
			logger := NewLogger(LevelWarn)
			logger.Debug("DEBUG", "some debug log")
			logger.Info("INFO", "some info log")
			logger.Warn("WARN", "some warn log")
			logger.Error("ERROR", "some error log")
		})

		Expect(stdout).ToNot(ContainSubstring("DEBUG"))
		Expect(stdout).ToNot(ContainSubstring("INFO"))
		Expect(stderr).To(ContainSubstring("WARN"))
		Expect(stderr).To(ContainSubstring("ERROR"))
	})

	It("log level error", func() {
		stdout, stderr := captureOutputs(func() {
			logger := NewLogger(LevelError)
			logger.Debug("DEBUG", "some debug log")
			logger.Info("INFO", "some info log")
			logger.Warn("WARN", "some warn log")
			logger.Error("ERROR", "some error log")
		})

		Expect(stdout).ToNot(ContainSubstring("DEBUG"))
		Expect(stdout).ToNot(ContainSubstring("INFO"))
		Expect(stderr).ToNot(ContainSubstring("WARN"))
		Expect(stderr).To(ContainSubstring("ERROR"))
	})

	Describe("Toggling forced debug", func() {
		Describe("when the log level is error", func() {
			It("outputs at debug level", func() {
				stdout, stderr := captureOutputs(func() {
					logger := NewLogger(LevelError)
					logger.ToggleForcedDebug()
					logger.Debug("TOGGLED_DEBUG", "some debug log")
					logger.Info("TOGGLED_INFO", "some info log")
					logger.Warn("TOGGLED_WARN", "some warn log")
					logger.Error("TOGGLED_ERROR", "some error log")
				})

				Expect(stdout).To(ContainSubstring("TOGGLED_DEBUG"))
				Expect(stdout).To(ContainSubstring("TOGGLED_INFO"))
				Expect(stderr).To(ContainSubstring("TOGGLED_WARN"))
				Expect(stderr).To(ContainSubstring("TOGGLED_ERROR"))
			})

			It("outputs at error level when toggled back", func() {
				stdout, stderr := captureOutputs(func() {
					logger := NewLogger(LevelError)
					logger.ToggleForcedDebug()
					logger.ToggleForcedDebug()
					logger.Debug("STANDARD_DEBUG", "some debug log")
					logger.Info("STANDARD_INFO", "some info log")
					logger.Warn("STANDARD_WARN", "some warn log")
					logger.Error("STANDARD_ERROR", "some error log")
				})

				Expect(stdout).ToNot(ContainSubstring("STANDARD_DEBUG"))
				Expect(stdout).ToNot(ContainSubstring("STANDARD_INFO"))
				Expect(stderr).ToNot(ContainSubstring("STANDARD_WARN"))
				Expect(stderr).To(ContainSubstring("STANDARD_ERROR"))
			})
		})
	})
})
