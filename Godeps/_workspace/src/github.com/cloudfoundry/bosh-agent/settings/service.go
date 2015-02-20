package settings

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type Service interface {
	LoadSettings() error

	// GetSettings does not return error because without settings Agent cannot start.
	GetSettings() Settings

	PublicSSHKeyForUsername(string) (string, error)

	InvalidateSettings() error
}

const settingsServiceLogTag = "settingsService"

type settingsService struct {
	fs                     boshsys.FileSystem
	settingsPath           string
	settings               Settings
	settingsSource         Source
	defaultNetworkDelegate DefaultNetworkDelegate
	logger                 boshlog.Logger
}

func NewService(
	fs boshsys.FileSystem,
	settingsPath string,
	settingsSource Source,
	defaultNetworkDelegate DefaultNetworkDelegate,
	logger boshlog.Logger,
) (service Service) {
	return &settingsService{
		fs:                     fs,
		settingsPath:           settingsPath,
		settings:               Settings{},
		settingsSource:         settingsSource,
		defaultNetworkDelegate: defaultNetworkDelegate,
		logger:                 logger,
	}
}

func (s *settingsService) PublicSSHKeyForUsername(username string) (string, error) {
	return s.settingsSource.PublicSSHKeyForUsername(username)
}

func (s *settingsService) LoadSettings() error {
	s.logger.Debug(settingsServiceLogTag, "Loading settings from fetcher")

	newSettings, fetchErr := s.settingsSource.Settings()
	if fetchErr != nil {
		s.logger.Error(settingsServiceLogTag, "Failed loading settings via fetcher: %v", fetchErr)

		existingSettingsJSON, readError := s.fs.ReadFile(s.settingsPath)
		if readError != nil {
			s.logger.Error(settingsServiceLogTag, "Failed reading settings from file %s", readError.Error())
			return bosherr.WrapError(fetchErr, "Invoking settings fetcher")
		}

		s.logger.Debug(settingsServiceLogTag, "Successfully read settings from file")

		err := json.Unmarshal(existingSettingsJSON, &s.settings)
		if err != nil {
			s.logger.Error(settingsServiceLogTag, "Failed unmarshalling settings from file %s", err.Error())
			return bosherr.WrapError(fetchErr, "Invoking settings fetcher")
		}

		err = s.checkAtMostOneDynamicNetwork(s.settings)
		if err != nil {
			return err
		}

		return nil
	}

	s.logger.Debug(settingsServiceLogTag, "Successfully received settings from fetcher")

	err := s.checkAtMostOneDynamicNetwork(newSettings)
	if err != nil {
		return err
	}

	s.settings = newSettings

	newSettingsJSON, err := json.Marshal(newSettings)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling settings json")
	}

	err = s.fs.WriteFile(s.settingsPath, newSettingsJSON)
	if err != nil {
		return bosherr.WrapError(err, "Writing setting json")
	}

	return nil
}

func (s settingsService) checkAtMostOneDynamicNetwork(settings Settings) error {
	var foundOneDynamicNetwork bool

	for _, network := range settings.Networks {
		// Currently proper support for multiple dynamic networks is not possible
		// because CPIs (e.g. AWS and OpenStack) do not include MAC address
		// for dynamic networks and that is the only way to reliably determine
		// network to interface to IP mapping
		if network.IsDynamic() {
			if foundOneDynamicNetwork {
				return bosherr.Error("Multiple dynamic networks are not supported")
			}
			foundOneDynamicNetwork = true
		}
	}

	return nil
}

// GetSettings returns setting even if it fails to resolve IPs for dynamic networks.
func (s *settingsService) GetSettings() Settings {
	for networkName, network := range s.settings.Networks {
		if !network.IsDynamic() {
			continue
		}

		// Ideally this would be GetNetworkByMACAddress(mac string)
		resolvedNetwork, err := s.defaultNetworkDelegate.GetDefaultNetwork()
		if err != nil {
			s.logger.Error(settingsServiceLogTag, "Failed retrieving default network %s", err.Error())
			break
		}

		// resolvedNetwork does not have all information for a network
		network.IP = resolvedNetwork.IP
		network.Netmask = resolvedNetwork.Netmask
		network.Gateway = resolvedNetwork.Gateway

		s.settings.Networks[networkName] = network
	}

	return s.settings
}

func (s *settingsService) InvalidateSettings() error {
	err := s.fs.RemoveAll(s.settingsPath)
	if err != nil {
		return bosherr.WrapError(err, "Removing settings file")
	}

	return nil
}
