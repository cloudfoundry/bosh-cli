package logger

import (
	"fmt"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"os"
)

func NewSignalableLogger(level boshlog.LogLevel, signalChannel chan os.Signal) (boshlog.Logger, chan bool) {
	logger := boshlog.NewLogger(level)
	doneChannel := make(chan bool, 1)
	go func() {
		for {
			<-signalChannel
			fmt.Println("Received SIGHUP - toggling debug output")
			logger.ToggleForcedDebug()
			doneChannel <- true
		}
	}()
	return logger, doneChannel
}
