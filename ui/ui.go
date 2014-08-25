package ui

import (
	"fmt"
	"io"
)

type UI interface {
	Say(string)
	Sayln(string)
	Error(string)
}

type ui struct {
	stdOut io.Writer
	stdErr io.Writer
}

func (u *ui) Say(message string) {
	u.stdOut.Write([]byte(fmt.Sprint(message)))
}

func (u *ui) Sayln(message string) {
	u.stdOut.Write([]byte(fmt.Sprintln(message)))
}

func (u *ui) Error(message string) {
	u.stdErr.Write([]byte(fmt.Sprintln(message)))
}

func NewUI(stdOut, stdErr io.Writer) UI {
	return &ui{
		stdOut: stdOut,
		stdErr: stdErr,
	}
}
