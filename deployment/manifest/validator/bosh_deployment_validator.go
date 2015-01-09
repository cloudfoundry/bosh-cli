package validator

import (
	"net"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmerr "github.com/cloudfoundry/bosh-micro-cli/release/errors"
)

type boshDeploymentValidator struct {
	releaseManager bmrel.Manager
}

func NewBoshDeploymentValidator(releaseManager bmrel.Manager) DeploymentValidator {
	return &boshDeploymentValidator{
		releaseManager: releaseManager,
	}
}

func (v *boshDeploymentValidator) Validate(deploymentManifest bmdeplmanifest.Manifest) error {
	errs := []error{}
	if v.isBlank(deploymentManifest.Name) {
		errs = append(errs, bosherr.Error("name must be provided"))
	}

	releaseNames := map[string]struct{}{}
	for releaseIdx, release := range deploymentManifest.Releases {
		if v.isBlank(release.Name) {
			errs = append(errs, bosherr.Errorf("releases[%d].name must be provided", releaseIdx))
		}

		if v.isBlank(release.Version) {
			errs = append(errs, bosherr.Errorf("releases[%d].version must be provided", releaseIdx))
		}

		if !release.IsLatest() {
			if _, err := release.VersionConstraints(); err != nil {
				errs = append(errs, bosherr.WrapErrorf(err, "releases[%d].version must be a semantic version", releaseIdx))
			}
		}

		if _, found := releaseNames[release.Name]; found {
			errs = append(errs, bosherr.Errorf("releases[%d].name '%s' must be unique", releaseIdx, release.Name))
		}
		releaseNames[release.Name] = struct{}{}
	}

	for idx, network := range deploymentManifest.Networks {
		if v.isBlank(network.Name) {
			errs = append(errs, bosherr.Errorf("networks[%d].name must be provided", idx))
		}
		if network.Type != bmdeplmanifest.Dynamic && network.Type != bmdeplmanifest.Manual && network.Type != bmdeplmanifest.VIP {
			errs = append(errs, bosherr.Errorf("networks[%d].type must be 'manual', 'dynamic', or 'vip'", idx))
		}
		if _, err := network.CloudProperties(); err != nil {
			errs = append(errs, bosherr.Errorf("networks[%d].cloud_properties must have only string keys", idx))
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
		if _, err := resourcePool.CloudProperties(); err != nil {
			errs = append(errs, bosherr.Errorf("resource_pools[%d].cloud_properties must have only string keys", idx))
		}
		if _, err := resourcePool.Env(); err != nil {
			errs = append(errs, bosherr.Errorf("resource_pools[%d].env must have only string keys", idx))
		}
	}

	for idx, diskPool := range deploymentManifest.DiskPools {
		if v.isBlank(diskPool.Name) {
			errs = append(errs, bosherr.Errorf("disk_pools[%d].name must be provided", idx))
		}
		if diskPool.DiskSize <= 0 {
			errs = append(errs, bosherr.Errorf("disk_pools[%d].disk_size must be > 0", idx))
		}
		if _, err := diskPool.CloudProperties(); err != nil {
			errs = append(errs, bosherr.Errorf("disk_pools[%d].cloud_properties must have only string keys", idx))
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
				if value != bmdeplmanifest.NetworkDefaultDNS && value != bmdeplmanifest.NetworkDefaultGateway {
					errs = append(errs, bosherr.Errorf("jobs[%d].networks[%d].default[%d] must be 'dns' or 'gateway'", idx, networkIdx, defaultIdx))
				}
			}
		}

		if job.Lifecycle != "" && job.Lifecycle != bmdeplmanifest.JobLifecycleService {
			errs = append(errs, bosherr.Errorf("jobs[%d].lifecycle must be 'service' ('%s' not supported)", idx, job.Lifecycle))
		}

		if _, err := job.Properties(); err != nil {
			errs = append(errs, bosherr.Errorf("jobs[%d].properties must have only string keys", idx))
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
			}
		}
	}

	if _, err := deploymentManifest.Properties(); err != nil {
		errs = append(errs, bosherr.Error("properties must have only string keys"))
	}

	if len(errs) > 0 {
		return bmerr.NewExplainableError(errs)
	}

	return nil
}

func (v *boshDeploymentValidator) isBlank(str string) bool {
	return str == "" || strings.TrimSpace(str) == ""
}

func (v *boshDeploymentValidator) networkNames(deploymentManifest bmdeplmanifest.Manifest) map[string]struct{} {
	names := make(map[string]struct{})
	for _, network := range deploymentManifest.Networks {
		names[network.Name] = struct{}{}
	}
	return names
}

func (v *boshDeploymentValidator) diskPoolNames(deploymentManifest bmdeplmanifest.Manifest) map[string]struct{} {
	names := make(map[string]struct{})
	for _, diskPool := range deploymentManifest.DiskPools {
		names[diskPool.Name] = struct{}{}
	}
	return names
}

func (v *boshDeploymentValidator) isValidIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil
}
