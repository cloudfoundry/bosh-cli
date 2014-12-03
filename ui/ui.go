package ui

import (
	"fmt"
	"io"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

type UI interface {
	Say(string)
	Sayln(string)
	Error(string)
}

type ui struct {
	stdOut io.Writer
	stdErr io.Writer
	logger boshlog.Logger
}

const logTag = "ui"

func NewUI(stdOut, stdErr io.Writer, logger boshlog.Logger) UI {
	return &ui{
		stdOut: stdOut,
		stdErr: stdErr,
		logger: logger,
	}
}

func (u *ui) Say(message string) {
	_, err := fmt.Fprint(u.stdOut, message)
	if err != nil {
		u.logger.Error(logTag, bosherr.WrapErrorf(err, "Writing to STDOUT: %s", message).Error())
	}
}

func (u *ui) Sayln(message string) {
	_, err := fmt.Fprintln(u.stdOut, message)
	if err != nil {
		u.logger.Error(logTag, bosherr.WrapErrorf(err, "Writing to STDOUT (with newline): %s", message).Error())
	}
}

func (u *ui) Error(message string) {
	_, err := fmt.Fprintln(u.stdErr, message)
	if err != nil {
		u.logger.Error(logTag, bosherr.WrapErrorf(err, "Writing to STDERR: %s", message).Error())
	}
}
