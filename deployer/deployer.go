package deployer

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/deployer/registry"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type Deployer interface {
	Deploy(
		bmcloud.Cloud,
		bmdepl.Deployment,
		bmstemcell.ExtractedStemcell,
		bmdepl.Registry,
		bmdepl.SSHTunnel,
		string,
	) error
}

type deployer struct {
	stemcellManagerFactory bmstemcell.ManagerFactory
	vmDeployer             VMDeployer
	diskDeployer           DiskDeployer
	registryServer         bmregistry.Server
	eventLoggerStage       bmeventlog.Stage
	logger                 boshlog.Logger
	logTag                 string
}

func NewDeployer(
	stemcellManagerFactory bmstemcell.ManagerFactory,
	vmDeployer VMDeployer,
	diskDeployer DiskDeployer,
	registryServer bmregistry.Server,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger,
) *deployer {
	eventLoggerStage := eventLogger.NewStage("deploying")

	return &deployer{
		stemcellManagerFactory: stemcellManagerFactory,
		vmDeployer:             vmDeployer,
		diskDeployer:           diskDeployer,
		registryServer:         registryServer,
		eventLoggerStage:       eventLoggerStage,
		logger:                 logger,
		logTag:                 "deployer",
	}
}

func (m *deployer) Deploy(
	cloud bmcloud.Cloud,
	deployment bmdepl.Deployment,
	extractedStemcell bmstemcell.ExtractedStemcell,
	registry bmdepl.Registry,
	sshTunnelConfig bmdepl.SSHTunnel,
	mbusURL string,
) error {
	stemcellManager := m.stemcellManagerFactory.NewManager(cloud)
	cloudStemcell, err := stemcellManager.Upload(extractedStemcell)
	if err != nil {
		return bosherr.WrapError(err, "Uploading stemcell")
	}

	m.eventLoggerStage.Start()

	if !registry.IsEmpty() {
		registryReadyErrCh := make(chan error)
		go m.startRegistry(registry, registryReadyErrCh)
		defer m.registryServer.Stop()

		err = <-registryReadyErrCh
		if err != nil {
			return bosherr.WrapError(err, "Starting registry")
		}
	}

	vm, err := m.vmDeployer.Deploy(cloud, deployment, cloudStemcell, mbusURL, m.eventLoggerStage)
	if err != nil {
		return bosherr.WrapError(err, "Deploying VM")
	}

	sshTunnelOptions := bmsshtunnel.Options{
		Host:              sshTunnelConfig.Host,
		Port:              sshTunnelConfig.Port,
		User:              sshTunnelConfig.User,
		Password:          sshTunnelConfig.Password,
		PrivateKey:        sshTunnelConfig.PrivateKey,
		LocalForwardPort:  registry.Port,
		RemoteForwardPort: registry.Port,
	}

	err = m.deleteUnusedStemcells(stemcellManager)
	if err != nil {
		return err
	}

	err = m.vmDeployer.WaitUntilReady(vm, sshTunnelOptions, m.eventLoggerStage)
	if err != nil {
		return bosherr.WrapError(err, "Waiting until VM is ready")
	}

	jobName := deployment.Jobs[0].Name

	diskPool, err := deployment.DiskPool(jobName)
	if err != nil {
		return bosherr.WrapError(err, "Getting disk pool")
	}

	err = m.diskDeployer.Deploy(diskPool, cloud, vm, m.eventLoggerStage)
	if err != nil {
		return bosherr.WrapError(err, "Deploying disk")
	}

	err = m.startVM(vm, extractedStemcell.ApplySpec(), deployment, jobName)
	if err != nil {
		return err
	}

	err = m.waitUntilRunning(vm, deployment.Update.UpdateWatchTime, jobName)
	if err != nil {
		return err
	}

	m.eventLoggerStage.Finish()
	return nil
}

func (m *deployer) startRegistry(registry bmdepl.Registry, readyErrCh chan error) {
	err := m.registryServer.Start(registry.Username, registry.Password, registry.Host, registry.Port, readyErrCh)
	if err != nil {
		m.logger.Debug(m.logTag, "Registry error occurred: %s", err.Error())
	}
}

func (m *deployer) startVM(vm bmvm.VM, stemcellApplySpec bmstemcell.ApplySpec, deployment bmdepl.Deployment, jobName string) error {
	err := m.eventLoggerStage.PerformStep(fmt.Sprintf("Starting '%s'", jobName), func() error {
		err := vm.Apply(stemcellApplySpec, deployment)
		if err != nil {
			return bosherr.WrapError(err, "Updating the vm")
		}

		err = vm.Start()
		if err != nil {
			return bosherr.WrapError(err, "Starting vm")
		}

		return nil
	})

	return err
}

func (m *deployer) waitUntilRunning(vm bmvm.VM, updateWatchTime bmdepl.WatchTime, jobName string) error {
	err := m.eventLoggerStage.PerformStep(fmt.Sprintf("Waiting for '%s'", jobName), func() error {
		time.Sleep(time.Duration(updateWatchTime.Start) * time.Millisecond)
		numAttempts := int((updateWatchTime.End - updateWatchTime.Start) / 1000)

		if err := vm.WaitToBeRunning(numAttempts, 1*time.Second); err != nil {
			return bosherr.WrapError(err, fmt.Sprintf("Waiting for '%s'", jobName))
		}

		return nil
	})

	return err
}

func (m *deployer) deleteUnusedStemcells(stemcellManager bmstemcell.Manager) error {
	stemcells, err := stemcellManager.FindUnused()
	if err != nil {
		return bosherr.WrapError(err, "Finding unused stemcells")
	}

	for _, stemcell := range stemcells {
		err = m.eventLoggerStage.PerformStep(fmt.Sprintf("Deleting unused stemcell '%s'", stemcell.CID()), func() error {
			if err = stemcell.Delete(); err != nil {
				return bosherr.WrapErrorf(err, "Deleting unused stemcell '%s'", stemcell.CID())
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}
