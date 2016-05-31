package cmd

import (
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"github.com/pivotal-golang/clock"

	bicrypto "github.com/cloudfoundry/bosh-init/crypto"
	boshui "github.com/cloudfoundry/bosh-init/ui"
)

type BasicDeps struct {
	FS     boshsys.FileSystem
	UI     *boshui.ConfUI
	Logger boshlog.Logger

	UUIDGen    boshuuid.Generator
	CmdRunner  boshsys.CmdRunner
	Compressor boshcmd.Compressor
	SHA1Calc   bicrypto.SHA1Calculator

	Time clock.Clock
}

func NewBasicDeps(ui *boshui.ConfUI, logger boshlog.Logger) BasicDeps {
	fs := boshsys.NewOsFileSystemWithStrictTempRoot(logger)

	cmdRunner := boshsys.NewExecCmdRunner(logger)

	return BasicDeps{
		FS:     fs,
		UI:     ui,
		Logger: logger,

		UUIDGen:    boshuuid.NewGenerator(),
		CmdRunner:  cmdRunner,
		Compressor: boshcmd.NewTarballCompressor(cmdRunner, fs),
		SHA1Calc:   bicrypto.NewSha1Calculator(fs),

		Time: clock.NewClock(),
	}
}
