package validator

import (
	"net"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmerr "github.com/cloudfoundry/bosh-micro-cli/release/errors"
)

type boshDeploymentValidator struct {
}

func NewBoshDeploymentValidator() DeploymentValidator {
	return &boshDeploymentValidator{}
}

func (v *boshDeploymentValidator) Validate(deployment bmdepl.Deployment) error {
	errs := []error{}
	if v.isBlank(deployment.Name) {
		errs = append(errs, bosherr.New("name must not be empty or blank"))
	}

	for idx, network := range deployment.Networks {
		if v.isBlank(network.Name) {
			errs = append(errs, bosherr.New("networks[%d].name must not be empty or blank", idx))
		}
		if network.Type != bmdepl.Dynamic && network.Type != bmdepl.Manual && network.Type != bmdepl.VIP {
			errs = append(errs, bosherr.New("networks[%d].type must be 'manual', 'dynamic', or 'vip'", idx))
		}
		if _, err := network.CloudProperties(); err != nil {
			errs = append(errs, bosherr.New("networks[%d].cloud_properties must have only string keys", idx))
		}
	}

	for idx, resourcePool := range deployment.ResourcePools {
		if v.isBlank(resourcePool.Name) {
			errs = append(errs, bosherr.New("resource_pools[%d].name must not be empty or blank", idx))
		}
		if v.isBlank(resourcePool.Network) {
			errs = append(errs, bosherr.New("resource_pools[%d].network must not be empty or blank", idx))
		} else if _, ok := v.networkNames(deployment)[resourcePool.Network]; !ok {
			errs = append(errs, bosherr.New("resource_pools[%d].network must be the name of a network", idx))
		}
		if _, err := resourcePool.CloudProperties(); err != nil {
			errs = append(errs, bosherr.New("resource_pools[%d].cloud_properties must have only string keys", idx))
		}
		if _, err := resourcePool.Env(); err != nil {
			errs = append(errs, bosherr.New("resource_pools[%d].env must have only string keys", idx))
		}
	}

	for idx, diskPool := range deployment.DiskPools {
		if v.isBlank(diskPool.Name) {
			errs = append(errs, bosherr.New("disk_pools[%d].name must not be empty or blank", idx))
		}
		if diskPool.DiskSize <= 0 {
			errs = append(errs, bosherr.New("disk_pools[%d].disk_size must be > 0", idx))
		}
		if _, err := diskPool.CloudProperties(); err != nil {
			errs = append(errs, bosherr.New("disk_pools[%d].cloud_properties must have only string keys", idx))
		}
	}

	if len(deployment.Jobs) > 1 {
		errs = append(errs, bosherr.New("jobs must be of size 1"))
	}

	for idx, job := range deployment.Jobs {
		if v.isBlank(job.Name) {
			errs = append(errs, bosherr.New("jobs[%d].name must not be empty or blank", idx))
		}
		if job.PersistentDisk < 0 {
			errs = append(errs, bosherr.New("jobs[%d].persistent_disk must be >= 0", idx))
		}
		if job.PersistentDiskPool != "" {
			if _, ok := v.diskPoolNames(deployment)[job.PersistentDiskPool]; !ok {
				errs = append(errs, bosherr.New("jobs[%d].persistent_disk_pool must be the name of a disk pool", idx))
			}
		}
		if job.Instances < 0 {
			errs = append(errs, bosherr.New("jobs[%d].instances must be >= 0", idx))
		}
		if len(job.Networks) == 0 {
			errs = append(errs, bosherr.New("jobs[%d].networks must be a non-empty array", idx))
		}
		for networkIdx, jobNetwork := range job.Networks {
			if v.isBlank(jobNetwork.Name) {
				errs = append(errs, bosherr.New("jobs[%d].networks[%d].name must not be empty or blank", idx, networkIdx))
			}

			for ipIdx, ip := range jobNetwork.StaticIPs {
				if !v.isValidIP(ip) {
					errs = append(errs, bosherr.New("jobs[%d].networks[%d].static_ips[%d] must be a valid IP", idx, networkIdx, ipIdx))
				}
			}

			for defaultIdx, value := range jobNetwork.Default {
				if value != bmdepl.NetworkDefaultDNS && value != bmdepl.NetworkDefaultGateway {
					errs = append(errs, bosherr.New("jobs[%d].networks[%d].default[%d] must be 'dns' or 'gateway'", idx, networkIdx, defaultIdx))
				}
			}
		}

		if job.Lifecycle != "" && job.Lifecycle != bmdepl.JobLifecycleService {
			errs = append(errs, bosherr.New("jobs[%d].lifecycle must be 'service' ('%s' not supported)", idx, job.Lifecycle))
		}

		if _, err := job.Properties(); err != nil {
			errs = append(errs, bosherr.New("jobs[%d].properties must have only string keys", idx))
		}
	}

	if _, err := deployment.Properties(); err != nil {
		errs = append(errs, bosherr.New("properties must have only string keys"))
	}

	if len(errs) > 0 {
		return bmerr.NewExplainableError(errs)
	}

	return nil
}

func (v *boshDeploymentValidator) isBlank(str string) bool {
	return str == "" || strings.TrimSpace(str) == ""
}

func (v *boshDeploymentValidator) networkNames(deployment bmdepl.Deployment) map[string]struct{} {
	names := make(map[string]struct{})
	for _, network := range deployment.Networks {
		names[network.Name] = struct{}{}
	}
	return names
}

func (v *boshDeploymentValidator) diskPoolNames(deployment bmdepl.Deployment) map[string]struct{} {
	names := make(map[string]struct{})
	for _, diskPool := range deployment.DiskPools {
		names[diskPool.Name] = struct{}{}
	}
	return names
}

func (v *boshDeploymentValidator) isValidIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil
}
