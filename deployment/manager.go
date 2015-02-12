package deployment

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bminstance "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type Manager interface {
	FindCurrent() (deployment Deployment, found bool, err error)
	Cleanup(bmui.Stage) error
}

type manager struct {
	instanceManager   bminstance.Manager
	diskManager       bmdisk.Manager
	stemcellManager   bmstemcell.Manager
	deploymentFactory Factory
}

func NewManager(
	instanceManager bminstance.Manager,
	diskManager bmdisk.Manager,
	stemcellManager bmstemcell.Manager,
	deploymentFactory Factory,
) Manager {
	return &manager{
		instanceManager:   instanceManager,
		diskManager:       diskManager,
		stemcellManager:   stemcellManager,
		deploymentFactory: deploymentFactory,
	}
}

func (m *manager) FindCurrent() (deployment Deployment, found bool, err error) {
	instances, err := m.instanceManager.FindCurrent()
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Finding current deployment instances")
	}

	disks, err := m.diskManager.FindCurrent()
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Finding current deployment disks")
	}

	stemcells, err := m.stemcellManager.FindCurrent()
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Finding current deployment stemcells")
	}

	if len(instances) == 0 && len(disks) == 0 && len(stemcells) == 0 {
		return nil, false, nil
	}

	return m.deploymentFactory.NewDeployment(instances, disks, stemcells), true, nil
}

func (m *manager) Cleanup(stage bmui.Stage) error {
	if err := m.diskManager.DeleteUnused(stage); err != nil {
		return err
	}

	if err := m.stemcellManager.DeleteUnused(stage); err != nil {
		return err
	}

	return nil
}
