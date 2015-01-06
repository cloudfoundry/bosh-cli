package deployer

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bminstance "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
)

type Deployer interface {
	Deploy(
		bmcloud.Cloud,
		bmdeplmanifest.Manifest,
		bmstemcell.ExtractedStemcell,
		bminstallmanifest.Registry,
		bminstallmanifest.SSHTunnel,
		bmvm.Manager,
	) error
}

type deployer struct {
	stemcellManagerFactory bmstemcell.ManagerFactory
	vmManagerFactory       bmvm.ManagerFactory
	sshTunnelFactory       bmsshtunnel.Factory
	eventLoggerStage       bmeventlog.Stage
	logger                 boshlog.Logger
	logTag                 string
}

func NewDeployer(
	stemcellManagerFactory bmstemcell.ManagerFactory,
	vmManagerFactory bmvm.ManagerFactory,
	sshTunnelFactory bmsshtunnel.Factory,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger,
) *deployer {
	eventLoggerStage := eventLogger.NewStage("deploying")

	return &deployer{
		stemcellManagerFactory: stemcellManagerFactory,
		vmManagerFactory:       vmManagerFactory,
		sshTunnelFactory:       sshTunnelFactory,
		eventLoggerStage:       eventLoggerStage,
		logger:                 logger,
		logTag:                 "deployer",
	}
}

func (m *deployer) Deploy(
	cloud bmcloud.Cloud,
	deploymentManifest bmdeplmanifest.Manifest,
	extractedStemcell bmstemcell.ExtractedStemcell,
	registryConfig bminstallmanifest.Registry,
	sshTunnelConfig bminstallmanifest.SSHTunnel,
	vmManager bmvm.Manager,
) error {
	stemcellManager := m.stemcellManagerFactory.NewManager(cloud)
	cloudStemcell, err := stemcellManager.Upload(extractedStemcell)
	if err != nil {
		return bosherr.WrapError(err, "Uploading stemcell")
	}

	m.eventLoggerStage.Start()

	instanceManager := bminstance.NewManager(cloud, vmManager, m.sshTunnelFactory, m.logger)

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
	deploymentManifest bmdeplmanifest.Manifest,
	instanceManager bminstance.Manager,
	extractedStemcell bmstemcell.ExtractedStemcell,
	cloudStemcell bmstemcell.CloudStemcell,
	registryConfig bminstallmanifest.Registry,
	sshTunnelConfig bminstallmanifest.SSHTunnel,
) error {
	if len(deploymentManifest.Jobs) != 1 {
		return bosherr.Errorf("There must only be one job, found %d", len(deploymentManifest.Jobs))
	}

	for _, jobSpec := range deploymentManifest.Jobs {
		if jobSpec.Instances != 1 {
			return bosherr.Errorf("Job '%s' must have only one instance, found %d", jobSpec.Name, jobSpec.Instances)
		}
		for instanceID := 0; instanceID < jobSpec.Instances; instanceID++ {
			instance, err := instanceManager.Create(jobSpec.Name, instanceID, deploymentManifest, cloudStemcell, registryConfig, sshTunnelConfig, m.eventLoggerStage)
			if err != nil {
				return bosherr.WrapErrorf(err, "Creating instance '%s/%d'", jobSpec.Name, instanceID)
			}

			err = instance.StartJobs(extractedStemcell.ApplySpec(), deploymentManifest, m.eventLoggerStage)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
