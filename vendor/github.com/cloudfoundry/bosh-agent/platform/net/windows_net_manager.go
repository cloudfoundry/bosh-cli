package net

import (
	"fmt"
	"strings"
	"time"

	"github.com/pivotal-golang/clock"

	boshsettings "github.com/cloudfoundry/bosh-agent/settings"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type WindowsNetManager struct {
	scriptRunner boshsys.ScriptRunner

	clock clock.Clock

	logTag string
	logger boshlog.Logger
}

func NewWindowsNetManager(scriptRunner boshsys.ScriptRunner, logger boshlog.Logger, clock clock.Clock) Manager {
	return WindowsNetManager{
		scriptRunner: scriptRunner,

		clock: clock,

		logTag: "WindowsNetManager",
		logger: logger,
	}
}

const (
	SetDNSTemplate = `
[array]$interfaces = Get-DNSClientServerAddress
$dns = @("%s")
foreach($interface in $interfaces) {
	Set-DnsClientServerAddress -InterfaceAlias $interface.InterfaceAlias -ServerAddresses ($dns -join ",")
}
`

	ResetDNSTemplate = `
[array]$interfaces = Get-DNSClientServerAddress
foreach($interface in $interfaces) {
	Set-DnsClientServerAddress -InterfaceAlias $interface.InterfaceAlias -ResetServerAddresses
}
`

	NicSettingsTemplate = `
$connectionName=(get-wmiobject win32_networkadapter | where-object {$_.MacAddress -eq '%s'}).netconnectionid
netsh interface ip set address $connectionName static %s %s %s
`
)

func (net WindowsNetManager) SetupNetworking(networks boshsettings.Networks, errCh chan error) error {

	nonVipNetworks := boshsettings.Networks{}

	for networkName, networkSettings := range networks {
		if networkSettings.IsVIP() {
			continue
		}
		nonVipNetworks[networkName] = networkSettings
	}

	err := net.setupInterfaces(nonVipNetworks)
	if err != nil {
		return err
	}

	dnsNetwork, _ := nonVipNetworks.DefaultNetworkFor("dns")
	dns := net.setupDNS(dnsNetwork)
	net.clock.Sleep(5 * time.Second)
	return dns
}

func (net WindowsNetManager) setupInterfaces(networks boshsettings.Networks) error {
	for _, network := range networks {
		var gateway string

		if network.IsDefaultFor("gateway") || len(networks) == 1 {
			gateway = network.Gateway
		}
		_, _, err := net.scriptRunner.Run(fmt.Sprintf(NicSettingsTemplate, network.Mac, network.IP, network.Netmask, gateway))
		if err != nil {
			return bosherr.WrapError(err, "Configuring interface")
		}
	}
	return nil
}

func (net WindowsNetManager) setupDNS(dnsNetwork boshsettings.Network) error {
	if len(dnsNetwork.DNS) > 0 {
		_, _, err := net.scriptRunner.Run(fmt.Sprintf(SetDNSTemplate, strings.Join(dnsNetwork.DNS, `","`)))
		if err != nil {
			return bosherr.WrapError(err, "Configuring DNS servers")
		}
	} else {
		_, _, err := net.scriptRunner.Run(ResetDNSTemplate)
		if err != nil {
			return bosherr.WrapError(err, "Resetting DNS servers")
		}
	}
	return nil
}

func (net WindowsNetManager) GetConfiguredNetworkInterfaces() ([]string, error) {
	panic("Not implemented")
}
