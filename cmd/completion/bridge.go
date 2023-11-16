package completion

import (
	boshcmd "github.com/cloudfoundry/bosh-cli/v7/cmd"
	cmdconf "github.com/cloudfoundry/bosh-cli/v7/cmd/config"
)

type CmdBridge struct {
	cmd  boshcmd.Cmd
	deps boshcmd.BasicDeps
}

func NewCmdBridge(cmd boshcmd.Cmd, deps boshcmd.BasicDeps) *CmdBridge {
	return &CmdBridge{cmd: cmd, deps: deps}
}

func (c CmdBridge) config() cmdconf.Config {
	config, err := cmdconf.NewFSConfigFromPath(c.cmd.BoshOpts.ConfigPathOpt, c.deps.FS)
	if err != nil {
		panic(err)
	}
	return config
}

func (c CmdBridge) Session() boshcmd.Session {
	return boshcmd.NewSessionFromOpts(c.cmd.BoshOpts, c.config(), c.deps.UI, true, true, c.deps.FS, c.deps.Logger)
}
