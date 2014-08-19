package settings

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type ServiceProvider interface {
	NewService(
		boshsys.FileSystem,
		string,
		Fetcher,
		DefaultNetworkDelegate,
		boshlog.Logger,
	) Service
}

type Service interface {
	LoadSettings() error

	// GetSettings does not return error
	// because without settings Agent cannot start.
	GetSettings() Settings

	InvalidateSettings() error
}
