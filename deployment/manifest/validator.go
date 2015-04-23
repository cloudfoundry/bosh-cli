package manifest

import (
	"net"
	"regexp"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	birel "github.com/cloudfoundry/bosh-init/release"
	birelsetmanifest "github.com/cloudfoundry/bosh-init/release/set/manifest"
)

type Validator interface {
	Validate(Manifest, birelsetmanifest.Manifest) error
	ValidateReleaseJobs(Manifest, birel.Manager) error
}

type validator struct {
	logger boshlog.Logger
}

func NewValidator(logger boshlog.Logger) Validator {
	return &validator{
		logger: logger,
	}
}

func (v *validator) Validate(deploymentManifest Manifest, releaseSetManifest birelsetmanifest.Manifest) error {
	errs := []error{}
	if v.isBlank(deploymentManifest.Name) {
		errs = append(errs, bosherr.Error("name must be provided"))
	}

	for idx, network := range deploymentManifest.Networks {
		if v.isBlank(network.Name) {
			errs = append(errs, bosherr.Errorf("networks[%d].name must be provided", idx))
		}
		if network.Type != Dynamic && network.Type != Manual && network.Type != VIP {
			errs = append(errs, bosherr.Errorf("networks[%d].type must be 'manual', 'dynamic', or 'vip'", idx))
		}
		if network.Type == Manual {
			if len(network.Subnets) != 1 {
				errs = append(errs, bosherr.Errorf("networks[%d].subnets must be of size 1", idx))
			} else {
				if v.isBlank(network.Subnets[0].Range) {
					errs = append(errs, bosherr.Errorf("networks[%d].subnets[0].range must be provided", idx))
				}
				if v.isBlank(network.Subnets[0].Gateway) {
					errs = append(errs, bosherr.Errorf("networks[%d].subnets[0].gateway must be provided", idx))
				}
			}
		}
	}

	if len(deploymentManifest.ResourcePools) != 1 {
		errs = append(errs, bosherr.Error("resource_pools must be of size 1"))
	}

	for idx, resourcePool := range deploymentManifest.ResourcePools {
		if v.isBlank(resourcePool.Name) {
			errs = append(errs, bosherr.Errorf("resource_pools[%d].name must be provided", idx))
		}
		if v.isBlank(resourcePool.Network) {
			errs = append(errs, bosherr.Errorf("resource_pools[%d].network must be provided", idx))
		} else if _, ok := v.networkNames(deploymentManifest)[resourcePool.Network]; !ok {
			errs = append(errs, bosherr.Errorf("resource_pools[%d].network must be the name of a network", idx))
		}

		if v.isBlank(resourcePool.Stemcell.URL) {
			errs = append(errs, bosherr.Errorf("resource_pools[%d].stemcell.url must be provided", idx))
		}

		matched, err := regexp.MatchString("^(file|http|https)://", resourcePool.Stemcell.URL)
		if err != nil || !matched {
			errs = append(errs, bosherr.Errorf("resource_pools[%d].stemcell.url must be a valid URL (file:// or http(s)://)", idx))
		}

		if strings.HasPrefix(resourcePool.Stemcell.URL, "http") && v.isBlank(resourcePool.Stemcell.SHA1) {
			errs = append(errs, bosherr.Errorf("resource_pools[%d].stemcell.sha1 must be provided for http URL", idx))
		}
	}

	for idx, diskPool := range deploymentManifest.DiskPools {
		if v.isBlank(diskPool.Name) {
			errs = append(errs, bosherr.Errorf("disk_pools[%d].name must be provided", idx))
		}
		if diskPool.DiskSize <= 0 {
			errs = append(errs, bosherr.Errorf("disk_pools[%d].disk_size must be > 0", idx))
		}
	}

	if len(deploymentManifest.Jobs) > 1 {
		errs = append(errs, bosherr.Error("jobs must be of size 1"))
	}

	for idx, job := range deploymentManifest.Jobs {
		if v.isBlank(job.Name) {
			errs = append(errs, bosherr.Errorf("jobs[%d].name must be provided", idx))
		}
		if job.PersistentDisk < 0 {
			errs = append(errs, bosherr.Errorf("jobs[%d].persistent_disk must be >= 0", idx))
		}
		if job.PersistentDiskPool != "" {
			if _, ok := v.diskPoolNames(deploymentManifest)[job.PersistentDiskPool]; !ok {
				errs = append(errs, bosherr.Errorf("jobs[%d].persistent_disk_pool must be the name of a disk pool", idx))
			}
		}
		if job.Instances < 0 {
			errs = append(errs, bosherr.Errorf("jobs[%d].instances must be >= 0", idx))
		}
		if len(job.Networks) == 0 {
			errs = append(errs, bosherr.Errorf("jobs[%d].networks must be a non-empty array", idx))
		}
		if v.isBlank(job.ResourcePool) {
			errs = append(errs, bosherr.Errorf("jobs[%d].resource_pool must be provided", idx))
		} else {
			if _, ok := v.resourcePoolNames(deploymentManifest)[job.ResourcePool]; !ok {
				errs = append(errs, bosherr.Errorf("jobs[%d].resource_pool must be the name of a resource pool", idx))
			}
		}
		for networkIdx, jobNetwork := range job.Networks {
			if v.isBlank(jobNetwork.Name) {
				errs = append(errs, bosherr.Errorf("jobs[%d].networks[%d].name must be provided", idx, networkIdx))
			}

			for ipIdx, ip := range jobNetwork.StaticIPs {
				if !v.isValidIP(ip) {
					errs = append(errs, bosherr.Errorf("jobs[%d].networks[%d].static_ips[%d] must be a valid IP", idx, networkIdx, ipIdx))
				}
			}

			for defaultIdx, value := range jobNetwork.Default {
				if value != NetworkDefaultDNS && value != NetworkDefaultGateway {
					errs = append(errs, bosherr.Errorf("jobs[%d].networks[%d].default[%d] must be 'dns' or 'gateway'", idx, networkIdx, defaultIdx))
				}
			}
		}

		if job.Lifecycle != "" && job.Lifecycle != JobLifecycleService {
			errs = append(errs, bosherr.Errorf("jobs[%d].lifecycle must be 'service' ('%s' not supported)", idx, job.Lifecycle))
		}

		templateNames := map[string]struct{}{}
		for templateIdx, template := range job.Templates {
			if v.isBlank(template.Name) {
				errs = append(errs, bosherr.Errorf("jobs[%d].templates[%d].name must be provided", idx, templateIdx))
			}
			if _, found := templateNames[template.Name]; found {
				errs = append(errs, bosherr.Errorf("jobs[%d].templates[%d].name '%s' must be unique", idx, templateIdx, template.Name))
			}
			templateNames[template.Name] = struct{}{}

			if v.isBlank(template.Release) {
				errs = append(errs, bosherr.Errorf("jobs[%d].templates[%d].release must be provided", idx, templateIdx))
			} else {
				_, found := releaseSetManifest.FindByName(template.Release)
				if !found {
					errs = append(errs, bosherr.Errorf("jobs[%d].templates[%d].release '%s' must refer to release in releases", idx, templateIdx, template.Release))
				}
			}
		}
	}

	if len(errs) > 0 {
		return bosherr.NewMultiError(errs...)
	}

	return nil
}

func (v *validator) ValidateReleaseJobs(deploymentManifest Manifest, releaseManager birel.Manager) error {
	errs := []error{}

	for idx, job := range deploymentManifest.Jobs {
		for templateIdx, template := range job.Templates {
			release, found := releaseManager.Find(template.Release)
			if !found {
				errs = append(errs, bosherr.Errorf("jobs[%d].templates[%d].release '%s' must refer to release in releases", idx, templateIdx, template.Release))
			} else {
				_, found := release.FindJobByName(template.Name)
				if !found {
					errs = append(errs, bosherr.Errorf("jobs[%d].templates[%d] must refer to a job in '%s', but there is no job named '%s'", idx, templateIdx, release.Name(), template.Name))
				}
			}
		}
	}

	if len(errs) > 0 {
		return bosherr.NewMultiError(errs...)
	}

	return nil
}

func (v *validator) isBlank(str string) bool {
	return str == "" || strings.TrimSpace(str) == ""
}

func (v *validator) networkNames(deploymentManifest Manifest) map[string]struct{} {
	names := make(map[string]struct{})
	for _, network := range deploymentManifest.Networks {
		names[network.Name] = struct{}{}
	}
	return names
}

func (v *validator) diskPoolNames(deploymentManifest Manifest) map[string]struct{} {
	names := make(map[string]struct{})
	for _, diskPool := range deploymentManifest.DiskPools {
		names[diskPool.Name] = struct{}{}
	}
	return names
}

func (v *validator) resourcePoolNames(deploymentManifest Manifest) map[string]struct{} {
	names := make(map[string]struct{})
	for _, resourcePool := range deploymentManifest.ResourcePools {
		names[resourcePool.Name] = struct{}{}
	}
	return names
}

func (v *validator) isValidIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil
}
