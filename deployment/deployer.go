package deployment

import (
	"net/http"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	biblobstore "github.com/cloudfoundry/bosh-cli/v7/blobstore"
	bicloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	bidisk "github.com/cloudfoundry/bosh-cli/v7/deployment/disk"
	biinstance "github.com/cloudfoundry/bosh-cli/v7/deployment/instance"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	bivm "github.com/cloudfoundry/bosh-cli/v7/deployment/vm"
	bistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell"
	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type Deployer interface {
	Deploy(
		cloud bicloud.Cloud,
		deploymentManifest bideplmanifest.Manifest,
		cloudStemcell bistemcell.CloudStemcell,
		vmManager bivm.Manager,
		blobstoreFactory biblobstore.Factory,
		blobstoreHTTPClient *http.Client,
		skipDrain bool,
		diskCIDs []string,
		deployStage biui.Stage,
	) (Deployment, error)
}

type deployer struct {
	vmManagerFactory       bivm.ManagerFactory
	instanceManagerFactory biinstance.ManagerFactory
	deploymentFactory      Factory
	logger                 boshlog.Logger
	logTag                 string
}

func NewDeployer(
	vmManagerFactory bivm.ManagerFactory,
	instanceManagerFactory biinstance.ManagerFactory,
	deploymentFactory Factory,
	logger boshlog.Logger,
) Deployer {
	return &deployer{
		vmManagerFactory:       vmManagerFactory,
		instanceManagerFactory: instanceManagerFactory,
		deploymentFactory:      deploymentFactory,
		logger:                 logger,
		logTag:                 "deployer",
	}
}

func (d *deployer) Deploy(
	cloud bicloud.Cloud,
	deploymentManifest bideplmanifest.Manifest,
	cloudStemcell bistemcell.CloudStemcell,
	vmManager bivm.Manager,
	blobstoreFactory biblobstore.Factory,
	blobstoreHTTPClient *http.Client,
	skipDrain bool,
	diskCIDs []string,
	deployStage biui.Stage,
) (Deployment, error) {
	if len(deploymentManifest.Jobs) != 1 {
		return nil, bosherr.Errorf("There must only be one job, found %d", len(deploymentManifest.Jobs))
	}

	instanceManager := d.instanceManagerFactory.NewManager(cloud, vmManager, blobstoreFactory, blobstoreHTTPClient)

	pingTimeout := 10 * time.Second
	pingDelay := 500 * time.Millisecond

	// Snapshot currently running instances before making any changes.
	existingInstances, err := instanceManager.FindCurrent()
	if err != nil {
		return nil, bosherr.WrapError(err, "Finding current instances")
	}

	// Rolling update: for each desired instance, stop and delete the matching
	// existing instance first (if any), then create and configure the new one.
	// This ensures at most one instance is down at a time.
	var instances []biinstance.Instance
	var disks []bidisk.Disk

	for _, jobSpec := range deploymentManifest.Jobs {
		// Track how many instances have been assigned to each AZ so far during
		// this deployment pass (used for round-robin placement of new instances).
		azCounts := make(map[string]int)
		for _, az := range jobSpec.AZs {
			azCounts[az] = 0
		}
		// Seed counts from existing instances that will be kept (sticky AZ).
		for _, existing := range existingInstances {
			if existing.JobName() == jobSpec.Name {
				if az := existing.AZ(); az != "" {
					azCounts[az]++
				}
			}
		}

		for instanceID := 0; instanceID < jobSpec.Instances; instanceID++ {
			// Determine AZ: sticky for existing instances, round-robin for new.
			az := ""
			if len(jobSpec.AZs) > 0 {
				if old := findExistingInstance(existingInstances, jobSpec.Name, instanceID); old != nil && old.AZ() != "" {
					az = old.AZ()
				} else {
					az = pickLeastLoadedAZ(jobSpec.AZs, azCounts)
					azCounts[az]++
				}
			}

			if old := findExistingInstance(existingInstances, jobSpec.Name, instanceID); old != nil {
				if err := old.Delete(pingTimeout, pingDelay, skipDrain, deployStage); err != nil {
					return nil, bosherr.WrapErrorf(err, "Deleting existing instance '%s/%d'", jobSpec.Name, instanceID)
				}
			}

			instance, instanceDisks, err := instanceManager.Create(jobSpec.Name, instanceID, az, deploymentManifest, cloudStemcell, diskCIDs, deployStage)
			if err != nil {
				return nil, bosherr.WrapErrorf(err, "Creating instance '%s/%d'", jobSpec.Name, instanceID)
			}
			instances = append(instances, instance)
			disks = append(disks, instanceDisks...)

			if err := instance.UpdateJobs(deploymentManifest, deployStage); err != nil {
				return nil, err
			}
		}
	}

	// Delete any instances that were running before but are no longer in the manifest.
	for _, old := range existingInstances {
		if !instanceIsDesired(deploymentManifest, old) {
			if err := old.Delete(pingTimeout, pingDelay, skipDrain, deployStage); err != nil {
				return nil, bosherr.WrapErrorf(err, "Deleting stale instance '%s/%d'", old.JobName(), old.ID())
			}
		}
	}

	stemcells := []bistemcell.CloudStemcell{cloudStemcell}
	return d.deploymentFactory.NewDeployment(instances, disks, stemcells), nil
}

// findExistingInstance returns the running instance matching the given job name
// and instance ID, or nil if none is found.
func findExistingInstance(instances []biinstance.Instance, jobName string, instanceID int) biinstance.Instance {
	for _, inst := range instances {
		if inst.JobName() == jobName && inst.ID() == instanceID {
			return inst
		}
	}
	return nil
}

// instanceIsDesired reports whether the manifest still includes the given instance.
func instanceIsDesired(manifest bideplmanifest.Manifest, inst biinstance.Instance) bool {
	for _, job := range manifest.Jobs {
		if job.Name == inst.JobName() && inst.ID() < job.Instances {
			return true
		}
	}
	return false
}

// pickLeastLoadedAZ returns the AZ from azs with the lowest count in counts.
// Ties are broken by the order the AZs appear in azs (manifest order), giving
// round-robin behaviour for fresh deployments.
func pickLeastLoadedAZ(azs []string, counts map[string]int) string {
	best := azs[0]
	for _, az := range azs[1:] {
		if counts[az] < counts[best] {
			best = az
		}
	}
	return best
}
