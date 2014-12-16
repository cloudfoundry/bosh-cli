package deployer

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bminstance "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type Deployer interface {
	Deploy(
		bmcloud.Cloud,
		bmmanifest.Manifest,
		bmstemcell.ExtractedStemcell,
		bmmanifest.Registry,
		bmmanifest.SSHTunnel,
		string,
	) error
}

type deployer struct {
	stemcellManagerFactory bmstemcell.ManagerFactory
	vmManagerFactory       bmvm.ManagerFactory
	sshTunnelFactory       bmsshtunnel.Factory
	diskDeployer           bminstance.DiskDeployer
	eventLoggerStage       bmeventlog.Stage
	logger                 boshlog.Logger
	logTag                 string
}

func NewDeployer(
	stemcellManagerFactory bmstemcell.ManagerFactory,
	vmManagerFactory bmvm.ManagerFactory,
	sshTunnelFactory bmsshtunnel.Factory,
	diskDeployer bminstance.DiskDeployer,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger,
) *deployer {
	eventLoggerStage := eventLogger.NewStage("deploying")

	return &deployer{
		stemcellManagerFactory: stemcellManagerFactory,
		vmManagerFactory:       vmManagerFactory,
		sshTunnelFactory:       sshTunnelFactory,
		diskDeployer:           diskDeployer,
		eventLoggerStage:       eventLoggerStage,
		logger:                 logger,
		logTag:                 "deployer",
	}
}

func (m *deployer) Deploy(
	cloud bmcloud.Cloud,
	deploymentManifest bmmanifest.Manifest,
	extractedStemcell bmstemcell.ExtractedStemcell,
	registryConfig bmmanifest.Registry,
	sshTunnelConfig bmmanifest.SSHTunnel,
	mbusURL string,
) error {
	stemcellManager := m.stemcellManagerFactory.NewManager(cloud)
	cloudStemcell, err := stemcellManager.Upload(extractedStemcell)
	if err != nil {
		return bosherr.WrapError(err, "Uploading stemcell")
	}

	m.eventLoggerStage.Start()

	vmManager := m.vmManagerFactory.NewManager(cloud, mbusURL)
	instanceManager := bminstance.NewManager(cloud, vmManager, m.sshTunnelFactory, m.diskDeployer, m.logger)

	pingTimeout := 10 * time.Second
	pingDelay := 500 * time.Millisecond
	if err = instanceManager.DeleteAll(pingTimeout, pingDelay, m.eventLoggerStage); err != nil {
		return err
	}

	if err = m.createAllInstances(deploymentManifest, instanceManager, extractedStemcell, cloudStemcell, registryConfig, sshTunnelConfig); err != nil {
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
	deploymentManifest bmmanifest.Manifest,
	instanceManager bminstance.Manager,
	extractedStemcell bmstemcell.ExtractedStemcell,
	cloudStemcell bmstemcell.CloudStemcell,
	registryConfig bmmanifest.Registry,
	sshTunnelConfig bmmanifest.SSHTunnel,
) error {
	if len(deploymentManifest.Jobs) != 1 {
		return bosherr.Errorf("There must only be one job, found %d", len(deploymentManifest.Jobs))
	}

	for _, jobSpec := range deploymentManifest.Jobs {
		if jobSpec.Instances != 1 {
			return bosherr.Errorf("Job '%s' must have only one instance, found %d", jobSpec.Name, jobSpec.Instances)
		}
		for instanceID := 0; instanceID < jobSpec.Instances; instanceID++ {
			_, err := instanceManager.Create(jobSpec.Name, instanceID,
				deploymentManifest, extractedStemcell, cloudStemcell,
				registryConfig, sshTunnelConfig, m.eventLoggerStage)
			if err != nil {
				return bosherr.WrapErrorf(err, "Creating instance '%s/%d'", jobSpec.Name, instanceID)
			}
		}
	}
	return nil
}
