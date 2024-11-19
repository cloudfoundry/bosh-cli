package ssh

import (
	"strconv"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
)

// You only need **one** of these per package!
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . HostBuilder

type HostBuilder interface {
	BuildHost(slug boshdir.AllOrInstanceGroupOrInstanceSlug, username string, deploymentFetcher DeploymentFetcher) (boshdir.Host, error)
}

type hostBuilder struct{}

func NewHostBuilder() HostBuilder {
	return &hostBuilder{}
}

type DeploymentFetcher func() (boshdir.Deployment, error)

func (h *hostBuilder) BuildHost(slug boshdir.AllOrInstanceGroupOrInstanceSlug, username string, deploymentFetcher DeploymentFetcher) (boshdir.Host, error) {

	var targetHost string
	if slug.IP() == "" {
		deployment, err := deploymentFetcher()
		if err != nil {
			return boshdir.Host{}, err
		}

		vms, err := deployment.VMInfos()
		if err != nil {
			return boshdir.Host{}, bosherr.WrapErrorf(err, "Finding VM for %s", slug)
		}
		var targetVM boshdir.VMInfo
		indexOrId := slug.IndexOrID()
		for _, vm := range vms {
			if vm.Active == nil || !*vm.Active {
				continue
			}
			if vm.JobName == slug.Name() {
				if indexOrId == "" {
					if targetVM.JobName != "" {
						return boshdir.Host{}, bosherr.Errorf("Instance %s refers to more than 1 VM", slug)
					} else {
						targetVM = vm
					}
				} else if index, err := strconv.Atoi(indexOrId); err == nil && index == *vm.Index {
					targetVM = vm
					break
				} else if indexOrId == vm.ID {
					targetVM = vm
					break
				}
			}
		}
		if targetVM.JobName == "" {
			return boshdir.Host{}, bosherr.Errorf("Instance %s has no active VM", slug)
		}
		if len(targetVM.IPs) == 0 {
			return boshdir.Host{}, bosherr.Errorf("VM %s has no IP address", targetVM.VMID)
		}
		targetHost = targetVM.IPs[0]
	} else {
		targetHost = slug.IP()
	}

	return boshdir.Host{
		Job:       slug.Name(),
		IndexOrID: slug.IndexOrID(),
		Username:  username,
		Host:      targetHost,
	}, nil
}
