package deployer

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bminstance "github.com/cloudfoundry/bosh-micro-cli/deployer/instance"
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
	vmManagerFactory       bmvm.ManagerFactory
	sshTunnelFactory       bmsshtunnel.Factory
	diskDeployer           bminstance.DiskDeployer
	registryServerFactory  bmregistry.ServerFactory
	eventLoggerStage       bmeventlog.Stage
	logger                 boshlog.Logger
	logTag                 string
}

func NewDeployer(
	stemcellManagerFactory bmstemcell.ManagerFactory,
	vmManagerFactory bmvm.ManagerFactory,
	sshTunnelFactory bmsshtunnel.Factory,
	diskDeployer bminstance.DiskDeployer,
	registryServerFactory bmregistry.ServerFactory,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger,
) *deployer {
	eventLoggerStage := eventLogger.NewStage("deploying")

	return &deployer{
		stemcellManagerFactory: stemcellManagerFactory,
		vmManagerFactory:       vmManagerFactory,
		sshTunnelFactory:       sshTunnelFactory,
		diskDeployer:           diskDeployer,
		registryServerFactory:  registryServerFactory,
		eventLoggerStage:       eventLoggerStage,
		logger:                 logger,
		logTag:                 "deployer",
	}
}

func (m *deployer) Deploy(
	cloud bmcloud.Cloud,
	deployment bmdepl.Deployment,
	extractedStemcell bmstemcell.ExtractedStemcell,
	registrySpec bmdepl.Registry,
	sshTunnelSpec bmdepl.SSHTunnel,
	mbusURL string,
) error {
	stemcellManager := m.stemcellManagerFactory.NewManager(cloud)
	cloudStemcell, err := stemcellManager.Upload(extractedStemcell)
	if err != nil {
		return bosherr.WrapError(err, "Uploading stemcell")
	}

	m.eventLoggerStage.Start()

	vmManager := m.vmManagerFactory.NewManager(cloud, mbusURL)
	instanceManager := bminstance.NewManager(cloud, vmManager, m.registryServerFactory, m.sshTunnelFactory, m.diskDeployer, m.logger)

	pingTimeout := 10 * time.Second
	pingDelay := 500 * time.Millisecond
	if err = instanceManager.DeleteAll(pingTimeout, pingDelay, m.eventLoggerStage); err != nil {
		return err
	}

	if err = m.createAllInstances(deployment, instanceManager, extractedStemcell, cloudStemcell, registrySpec, sshTunnelSpec); err != nil {
		return err
	}

	// TODO: cleanup unused disks?

	if err = stemcellManager.DeleteUnused(m.eventLoggerStage); err != nil {
		return err
	}

	m.eventLoggerStage.Finish()
	return nil
}

func (m *deployer) createAllInstances(
	deployment bmdepl.Deployment,
	instanceManager bminstance.Manager,
	extractedStemcell bmstemcell.ExtractedStemcell,
	cloudStemcell bmstemcell.CloudStemcell,
	registrySpec bmdepl.Registry,
	sshTunnelSpec bmdepl.SSHTunnel,
) error {
	if len(deployment.Jobs) != 1 {
		return bosherr.Errorf("There must only be one job, found %d", len(deployment.Jobs))
	}

	for _, jobSpec := range deployment.Jobs {
		if jobSpec.Instances != 1 {
			return bosherr.Errorf("Job '%s' must have only one instance, found %d", jobSpec.Name, jobSpec.Instances)
		}
		for instanceID := 0; instanceID < jobSpec.Instances; instanceID++ {
			_, err := instanceManager.Create(jobSpec.Name, instanceID,
				deployment, extractedStemcell, cloudStemcell,
				registrySpec, sshTunnelSpec, m.eventLoggerStage)
			if err != nil {
				return bosherr.WrapErrorf(err, "Creating instance '%s/%d'", jobSpec.Name, instanceID)
			}
		}
	}
	return nil
}
