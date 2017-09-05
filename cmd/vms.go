package cmd

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/workpool"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type VMsCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewVMsCmd(ui boshui.UI, director boshdir.Director) VMsCmd {
	return VMsCmd{ui: ui, director: director}
}

func (c VMsCmd) Run(opts VMsOpts) error {
	instTable := InstanceTable{
		// VMs command should always show VM specifics
		VMDetails: true,

		Details: false,
		DNS:     opts.DNS,
		Vitals:  opts.Vitals,
	}

	if len(opts.Deployment) > 0 {
		dep, err := c.director.FindDeployment(opts.Deployment)
		if err != nil {
			return err
		}

		vmInfos, err := dep.VMInfos()
		if err != nil {
			return err
		}

		return c.printDeployment(dep, instTable, vmInfos)
	}

	return c.printDeployments(instTable, opts.ParallelOpt)
}

func (c VMsCmd) printDeployments(instTable InstanceTable, parallel int) error {
	deployments, err := c.director.Deployments()
	if err != nil {
		return err
	}

	vmInfos, err := parallelVMInfos(deployments, parallel)
	if err != nil {
		return err
	}

	for _, dep := range deployments {
		err := c.printDeployment(dep, instTable, vmInfos[dep.Name()])
		if err != nil {
			return err
		}
	}

	return nil
}

type deploymentInfo struct {
	depName string
	vmInfos []boshdir.VMInfo
}

func parallelVMInfos(deployments []boshdir.Deployment, parallel int) (map[string][]boshdir.VMInfo, error) {
	if parallel == 0 {
		parallel = 5
	}
	workSize := len(deployments)
	resultc := make(chan deploymentInfo, workSize)
	errorc := make(chan error, workSize)
	defer close(resultc)
	defer close(errorc)
	works := make([]func(), workSize)

	for i, dep := range deployments {
		dep := dep
		works[i] = func() {
			vmInfos, err := dep.VMInfos()
			errorc <- err
			resultc <- deploymentInfo{dep.Name(), vmInfos}
		}
	}

	throttler, err := workpool.NewThrottler(parallel, works)
	if err != nil {
		return nil, err
	}
	throttler.Work()
	vms := make(map[string][]boshdir.VMInfo, workSize)
	var vmInfoErrors []string
	for i := 0; i < workSize; i++ {
		errc := <-errorc
		result := <-resultc
		if errc != nil {
			vmInfoErrors = append(vmInfoErrors, errc.Error())
		}
		vms[result.depName] = result.vmInfos
	}
	if len(vmInfoErrors) > 0 {
		err := fmt.Errorf("%s", strings.Join(vmInfoErrors, "\n"))
		return nil, err
	}
	return vms, nil
}

func (c VMsCmd) printDeployment(dep boshdir.Deployment, instTable InstanceTable, vmInfos []boshdir.VMInfo) error {
	table := boshtbl.Table{
		Title: fmt.Sprintf("Deployment '%s'", dep.Name()),

		Content: "vms",

		Header: instTable.Headers(),

		SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},
	}

	for _, info := range vmInfos {
		row := instTable.AsValues(instTable.ForVMInfo(info))

		table.Rows = append(table.Rows, row)
	}

	c.ui.PrintTable(table)

	return nil
}
