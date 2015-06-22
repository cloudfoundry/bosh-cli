package infrastructure

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	boshplat "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/platform"
	boshsettings "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/settings"
	bosherr "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/logger"
)

type httpMetadataService struct {
	metadataHost string
	resolver     DNSResolver
	platform     boshplat.Platform
	logTag       string
	logger       boshlog.Logger
}

func NewHTTPMetadataService(
	metadataHost string,
	resolver DNSResolver,
	platform boshplat.Platform,
	logger boshlog.Logger,
) MetadataService {
	return httpMetadataService{
		metadataHost: metadataHost,
		resolver:     resolver,
		platform:     platform,
		logTag:       "httpMetadataService",
		logger:       logger,
	}
}

func (ms httpMetadataService) Load() error {
	return nil
}

func (ms httpMetadataService) GetPublicKey() (string, error) {
	err := ms.ensureMinimalNetworkSetup()
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/latest/meta-data/public-keys/0/openssh-key", ms.metadataHost)
	resp, err := http.Get(url)
	if err != nil {
		return "", bosherr.WrapError(err, "Getting open ssh key")
	}

	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", bosherr.WrapError(err, "Reading ssh key response body")
	}

	return string(bytes), nil
}

func (ms httpMetadataService) GetInstanceID() (string, error) {
	err := ms.ensureMinimalNetworkSetup()
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/latest/meta-data/instance-id", ms.metadataHost)
	resp, err := http.Get(url)
	if err != nil {
		return "", bosherr.WrapError(err, "Getting instance id from url")
	}

	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", bosherr.WrapError(err, "Reading instance id response body")
	}

	return string(bytes), nil
}

func (ms httpMetadataService) GetServerName() (string, error) {
	userData, err := ms.getUserData()
	if err != nil {
		return "", bosherr.WrapError(err, "Getting user data")
	}

	serverName := userData.Server.Name

	if len(serverName) == 0 {
		return "", bosherr.Error("Empty server name")
	}

	return serverName, nil
}

func (ms httpMetadataService) GetRegistryEndpoint() (string, error) {
	userData, err := ms.getUserData()
	if err != nil {
		return "", bosherr.WrapError(err, "Getting user data")
	}

	endpoint := userData.Registry.Endpoint
	nameServers := userData.DNS.Nameserver

	if len(nameServers) > 0 {
		endpoint, err = ms.resolver.LookupHost(nameServers, endpoint)
		if err != nil {
			return "", bosherr.WrapError(err, "Resolving registry endpoint")
		}
	}

	return endpoint, nil
}

func (ms httpMetadataService) GetNetworks() (boshsettings.Networks, error) {
	return nil, nil
}

func (ms httpMetadataService) IsAvailable() bool { return true }

func (ms httpMetadataService) getUserData() (UserDataContentsType, error) {
	var userData UserDataContentsType

	err := ms.ensureMinimalNetworkSetup()
	if err != nil {
		return userData, err
	}

	userDataURL := fmt.Sprintf("%s/latest/user-data", ms.metadataHost)
	userDataResp, err := http.Get(userDataURL)
	if err != nil {
		return userData, bosherr.WrapError(err, "Getting user data from url")
	}

	defer userDataResp.Body.Close()

	userDataBytes, err := ioutil.ReadAll(userDataResp.Body)
	if err != nil {
		return userData, bosherr.WrapError(err, "Reading user data response body")
	}

	err = json.Unmarshal(userDataBytes, &userData)
	if err != nil {
		return userData, bosherr.WrapError(err, "Unmarshalling user data")
	}

	return userData, nil
}

func (ms httpMetadataService) ensureMinimalNetworkSetup() error {
	// We check for configuration presence instead of verifying
	// that network is reachable because we want to preserve
	// network configuration that was passed to agent.
	configuredInterfaces, err := ms.platform.GetConfiguredNetworkInterfaces()
	if err != nil {
		return bosherr.WrapError(err, "Getting configured network interfaces")
	}

	if len(configuredInterfaces) == 0 {
		ms.logger.Debug(ms.logTag, "No configured networks found, setting up DHCP network")
		err = ms.platform.SetupNetworking(boshsettings.Networks{
			"eth0": {
				Type: boshsettings.NetworkTypeDynamic,
			},
		})
		if err != nil {
			return bosherr.WrapError(err, "Setting up initial DHCP network")
		}
	}

	return nil
}
