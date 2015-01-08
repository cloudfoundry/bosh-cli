package manifest

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	bmkeystr "github.com/cloudfoundry/bosh-micro-cli/keystringifier"
)

type Manifest struct {
	Name          string
	Releases      []ReleaseRef
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

func (d Manifest) NetworksSpec(jobName string) (map[string]interface{}, error) {
	job, found := d.findJobByName(jobName)
	if !found {
		return map[string]interface{}{}, bosherr.Errorf("Could not find job with name: %s", jobName)
	}

	networksMap := d.networksToMap()

	result := map[string]interface{}{}
	var err error
	for _, jobNetwork := range job.Networks {
		network := networksMap[jobNetwork.Name]
		staticIPs := jobNetwork.StaticIPs
		if len(staticIPs) > 0 {
			network.IP = staticIPs[0]
		}
		result[jobNetwork.Name], err = network.Spec()
		if err != nil {
			return map[string]interface{}{}, bosherr.WrapError(err, "Building network spec")
		}
	}

	return result, nil
}

func (d Manifest) DiskPool(jobName string) (DiskPool, error) {
	job, found := d.findJobByName(jobName)
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

func (d Manifest) ReleasesByName() map[string]ReleaseRef {
	releasesByName := map[string]ReleaseRef{}
	for _, release := range d.Releases {
		releasesByName[release.Name] = release
	}
	return releasesByName
}

func (d Manifest) networksToMap() map[string]Network {
	result := map[string]Network{}
	for _, network := range d.Networks {
		result[network.Name] = network
	}
	return result
}

func (d Manifest) findJobByName(jobName string) (Job, bool) {
	for _, job := range d.Jobs {
		if job.Name == jobName {
			return job, true
		}
	}

	return Job{}, false
}
