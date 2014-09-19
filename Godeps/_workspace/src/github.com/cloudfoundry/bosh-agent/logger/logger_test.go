package logger_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-agent/logger"
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

		It("includes the message from an error", func() {
			stdout, stderr := captureOutputs(func() {
				logger := NewLogger(LevelDebug)
				logger.Debug("TAG", "some %s info to log", "awesome", errors.New("some error message"))
			})

			expectedContent := expectedLogFormat("TAG", "DEBUG - some awesome info to log - some error message")
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

		It("includes the message from an error", func() {
			stdout, stderr := captureOutputs(func() {
				logger := NewLogger(LevelInfo)
				logger.Info("TAG", "some %s info to log", "awesome", errors.New("some error message"))
			})

			expectedContent := expectedLogFormat("TAG", "INFO - some awesome info to log - some error message")
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

		It("includes the message from an error", func() {
			stdout, stderr := captureOutputs(func() {
				logger := NewLogger(LevelWarn)
				logger.Warn("TAG", "some %s info to log", "awesome", errors.New("some error message"))
			})

			expectedContent := expectedLogFormat("TAG", "WARN - some awesome info to log - some error message")
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

		It("includes the message from an error", func() {
			stdout, stderr := captureOutputs(func() {
				logger := NewLogger(LevelError)
				logger.Error("TAG", "some %s info to log", "awesome", errors.New("some error message"))
			})

			expectedContent := expectedLogFormat("TAG", "ERROR - some awesome info to log - some error message")
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
})
