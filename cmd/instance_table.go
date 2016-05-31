package cmd

import (
	"fmt"
	"strings"

	boshdir "github.com/cloudfoundry/bosh-init/director"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
)

type InstanceTableValues struct {
	Name    boshtbl.Value
	Process boshtbl.Value

	State  boshtbl.Value
	AZ     boshtbl.Value
	VMType boshtbl.Value
	IPs    boshtbl.Value

	// Details
	VMCID        boshtbl.Value
	DiskCID      boshtbl.Value
	AgentID      boshtbl.Value
	Resurrection boshtbl.Value

	// DNS
	DNS boshtbl.Value

	// Vitals
	Uptime boshtbl.Value // only for Process
	Load   boshtbl.Value

	CPUTotal boshtbl.Value // only for Process
	CPUUser  boshtbl.Value
	CPUSys   boshtbl.Value
	CPUWait  boshtbl.Value

	Memory boshtbl.Value
	Swap   boshtbl.Value

	SystemDisk     boshtbl.Value
	EphemeralDisk  boshtbl.Value
	PersistentDisk boshtbl.Value
}

var InstanceTableHeader = InstanceTableValues{
	Name:    boshtbl.ValueString{"Instance"},
	Process: boshtbl.ValueString{"Process"},

	State:  boshtbl.ValueString{"State"},
	AZ:     boshtbl.ValueString{"AZ"},
	VMType: boshtbl.ValueString{"VM Type"},
	IPs:    boshtbl.ValueString{"IPs"},

	// Details
	VMCID:        boshtbl.ValueString{"VM CID"},
	DiskCID:      boshtbl.ValueString{"Disk CID"},
	AgentID:      boshtbl.ValueString{"Agent ID"},
	Resurrection: boshtbl.ValueString{"Resurrection\nPaused"},

	// DNS
	DNS: boshtbl.ValueString{"DNS A Records"},

	// Vitals
	Uptime: boshtbl.ValueString{"Uptime"}, // only for Process
	Load:   boshtbl.ValueString{"Load\n(1m, 5m, 15m)"},

	CPUTotal: boshtbl.ValueString{"CPU\nTotal"}, // only for Process
	CPUUser:  boshtbl.ValueString{"CPU\nUser"},
	CPUSys:   boshtbl.ValueString{"CPU\nSys"},
	CPUWait:  boshtbl.ValueString{"CPU\nWait"},

	Memory: boshtbl.ValueString{"Memory\nUsage"},
	Swap:   boshtbl.ValueString{"Swap\nUsage"},

	SystemDisk:     boshtbl.ValueString{"System\nDisk Usage"},
	EphemeralDisk:  boshtbl.ValueString{"Ephemeral\nDisk Usage"},
	PersistentDisk: boshtbl.ValueString{"Persistent\nDisk Usage"},
}

type InstanceTable struct {
	Processes, VMDetails, Details, DNS, Vitals bool
}

func (t InstanceTable) Header() InstanceTableValues {
	return InstanceTableHeader
}

func (t InstanceTable) ForVMInfo(i boshdir.VMInfo) InstanceTableValues {
	vals := InstanceTableValues{
		Name:    t.buildName(i),
		Process: boshtbl.ValueString{},

		State: boshtbl.ValueFmt{
			V:     boshtbl.ValueString{i.State},
			Error: !i.IsRunning(),
		},

		AZ:     boshtbl.ValueString{i.AZ},
		VMType: boshtbl.ValueString{i.VMType},
		IPs:    boshtbl.ValueStrings{i.IPs},

		// Details
		VMCID:        boshtbl.ValueString{i.VMID},
		DiskCID:      boshtbl.ValueString{i.DiskID},
		AgentID:      boshtbl.ValueString{i.AgentID},
		Resurrection: boshtbl.ValueBool{i.ResurrectionPaused},

		// DNS
		DNS: boshtbl.ValueStrings{i.DNS},

		// Vitals
		Uptime: ValueUptime{i.Vitals.Uptime.Seconds},
		Load:   boshtbl.ValueString{strings.Join(i.Vitals.Load, ", ")},

		CPUTotal: ValueCPUTotal{i.Vitals.CPU.Total},
		CPUUser:  NewValueStringPercent(i.Vitals.CPU.User),
		CPUSys:   NewValueStringPercent(i.Vitals.CPU.Sys),
		CPUWait:  NewValueStringPercent(i.Vitals.CPU.Wait),

		Memory: ValueMemSize{i.Vitals.Mem},
		Swap:   ValueMemSize{i.Vitals.Swap},

		SystemDisk:     ValueDiskSize{i.Vitals.SystemDisk()},
		EphemeralDisk:  ValueDiskSize{i.Vitals.EphemeralDisk()},
		PersistentDisk: ValueDiskSize{i.Vitals.PersistentDisk()},
	}

	if len(i.VMType) == 0 {
		vals.VMType = boshtbl.ValueString{i.ResourcePool}
	}

	return vals
}

func (t InstanceTable) buildName(i boshdir.VMInfo) boshtbl.ValueString {
	name := "?"

	if len(i.JobName) > 0 {
		name = i.JobName
	}

	if len(i.ID) > 0 {
		name += "/" + i.ID

		if i.Bootstrap {
			name += "*"
		}

		if i.Index != nil {
			name += fmt.Sprintf(" (%d)", *i.Index)
		}
	} else {
		if i.Index == nil {
			name += "/?"
		} else {
			name += fmt.Sprintf("/%d", *i.Index)
		}

		if i.Bootstrap {
			name += "*"
		}
	}

	return boshtbl.ValueString{name}
}

func (t InstanceTable) ForProcess(p boshdir.VMInfoProcess) InstanceTableValues {
	return InstanceTableValues{
		Name:    boshtbl.ValueString{},
		Process: boshtbl.ValueString{p.Name},

		State: boshtbl.ValueFmt{
			V:     boshtbl.ValueString{p.State},
			Error: !p.IsRunning(),
		},

		// Vitals
		Uptime:   ValueUptime{p.Uptime.Seconds},
		Memory:   ValueMemIntSize{p.Mem},
		CPUTotal: ValueCPUTotal{p.CPU.Total},
	}
}

// AsValues is public instead of being private to aid ease of accessing vals in tests
func (t InstanceTable) AsValues(v InstanceTableValues) []boshtbl.Value {
	result := []boshtbl.Value{v.Name}

	if t.Processes {
		result = append(result, v.Process)
	}

	result = append(result, []boshtbl.Value{v.State, v.AZ, v.IPs}...)

	if t.Details {
		result = append(result, []boshtbl.Value{v.VMCID, v.VMType, v.DiskCID, v.AgentID, v.Resurrection}...)
	} else if t.VMDetails {
		result = append(result, []boshtbl.Value{v.VMCID, v.VMType}...)
	}

	if t.DNS {
		result = append(result, v.DNS)
	}

	if t.Vitals {
		result = append(result, []boshtbl.Value{v.Uptime, v.Load}...)
		result = append(result, []boshtbl.Value{v.CPUTotal, v.CPUUser, v.CPUSys, v.CPUWait}...)
		result = append(result, []boshtbl.Value{v.Memory, v.Swap}...)
		result = append(result, []boshtbl.Value{v.SystemDisk, v.EphemeralDisk, v.PersistentDisk}...)
	}

	return result
}
