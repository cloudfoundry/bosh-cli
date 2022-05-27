package ssh

import (
	"io"

	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
)

// You only need **one** of these per package!
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Runner

type Runner interface {
	Run(ConnectionOpts, boshdir.SSHResult, []string) error
}

//counterfeiter:generate . SCPRunner

type SCPRunner interface {
	Run(ConnectionOpts, boshdir.SSHResult, SCPArgs) error
}

type ConnectionOpts struct {
	PrivateKey string

	GatewayDisable bool

	GatewayUsername       string
	GatewayHost           string
	GatewayPrivateKeyPath string

	SOCKS5Proxy string

	RawOpts []string
}

//counterfeiter:generate . Session

type Session interface {
	Start() (SSHArgs, error)
	Finish() error
}

type Writer interface {
	ForInstance(jobName, indexOrID string) InstanceWriter
	Flush()
}

type InstanceWriter interface {
	Stdout() io.Writer
	Stderr() io.Writer
	End(exitStatus int, err error)
}
