package manifest

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	bmkeystr "github.com/cloudfoundry/bosh-micro-cli/keystringifier"
)

type Manifest struct {
	Name          string
	RawProperties map[interface{}]interface{}
	Jobs          []Job
	Networks      []Network
	DiskPools     []DiskPool
	ResourcePools []ResourcePool
	Update        Update
}

type Update struct {
	UpdateWatchTime WatchTime
}

func (d Manifest) Properties() (map[string]interface{}, error) {
	return bmkeystr.NewKeyStringifier().ConvertMap(d.RawProperties)
}

// NetworkInterfaces returns a map of network names to network interfaces.
// We can't use map[string]NetworkInterface, because it's impossible to down-cast to what the cloud client requires.
func (d Manifest) NetworkInterfaces(jobName string) (map[string]map[string]interface{}, error) {
	job, found := d.FindJobByName(jobName)
	if !found {
		return map[string]map[string]interface{}{}, bosherr.Errorf("Could not find job with name: %s", jobName)
	}

	networkMap := d.networkMap()

	ifaceMap := map[string]map[string]interface{}{}
	var err error
	for _, jobNetwork := range job.Networks {
		network := networkMap[jobNetwork.Name]
		staticIPs := jobNetwork.StaticIPs
		if len(staticIPs) > 0 {
			network.IP = staticIPs[0]
		}
		ifaceMap[jobNetwork.Name], err = network.Interface()
		if err != nil {
			return map[string]map[string]interface{}{}, bosherr.WrapError(err, "Building network spec")
		}
	}

	return ifaceMap, nil
}

func (d Manifest) DiskPool(jobName string) (DiskPool, error) {
	job, found := d.FindJobByName(jobName)
	if !found {
		return DiskPool{}, bosherr.Errorf("Could not find job with name: %s", jobName)
	}

	if job.PersistentDiskPool != "" {
		for _, diskPool := range d.DiskPools {
			if diskPool.Name == job.PersistentDiskPool {
				return diskPool, nil
			}
		}
		err := bosherr.Errorf("Could not find persistent disk pool '%s' for job '%s'", job.PersistentDiskPool, jobName)
		return DiskPool{}, err
	}

	if job.PersistentDisk > 0 {
		diskPool := DiskPool{
			DiskSize:           job.PersistentDisk,
			RawCloudProperties: map[interface{}]interface{}{},
		}
		return diskPool, nil
	}

	return DiskPool{}, nil
}

func (d Manifest) networkMap() map[string]Network {
	result := map[string]Network{}
	for _, network := range d.Networks {
		result[network.Name] = network
	}
	return result
}

func (d Manifest) FindJobByName(jobName string) (Job, bool) {
	for _, job := range d.Jobs {
		if job.Name == jobName {
			return job, true
		}
	}

	return Job{}, false
}
