package cmd

import (
	"fmt"
	"strings"

	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

type InstanceTableValues struct {
	Name    boshtbl.Value
	Process boshtbl.Value

	ProcessState boshtbl.Value
	State        boshtbl.Value
	AZ           boshtbl.Value
	VMType       boshtbl.Value
	Active       boshtbl.Value
	IPs          boshtbl.Value
	IPs_cidr     boshtbl.Value
	Deployment   boshtbl.Value

	// Details
	VMCID           boshtbl.Value
	DiskCIDs        boshtbl.Value
	AgentID         boshtbl.Value
	Index           boshtbl.Value
	Bootstrap       boshtbl.Value
	Ignore          boshtbl.Value
	VMCreatedAt     boshtbl.Value
	CloudProperties boshtbl.Value

	Stemcell boshtbl.Value

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
	Name:    boshtbl.NewValueString("Instance"),
	Process: boshtbl.NewValueString("Process"),

	ProcessState: boshtbl.NewValueString("Process State"),
	AZ:           boshtbl.NewValueString("AZ"),
	VMType:       boshtbl.NewValueString("VM Type"),
	Active:       boshtbl.NewValueString("Active"),
	IPs:          boshtbl.NewValueString("IPs"),
	IPs_cidr:     boshtbl.NewValueString("IPs"),
	Deployment:   boshtbl.NewValueString("Deployment"),

	// Details
	State:           boshtbl.NewValueString("State"),
	VMCID:           boshtbl.NewValueString("VM CID"),
	DiskCIDs:        boshtbl.NewValueString("Disk CIDs"),
	AgentID:         boshtbl.NewValueString("Agent ID"),
	Index:           boshtbl.NewValueString("Index"),
	Bootstrap:       boshtbl.NewValueString("Bootstrap"),
	Ignore:          boshtbl.NewValueString("Ignore"),
	VMCreatedAt:     boshtbl.NewValueString("VM Created At"),
	CloudProperties: boshtbl.NewValueString("Cloud Properties"),

	Stemcell: boshtbl.NewValueString("Stemcell"),

	// Vitals
	Uptime: boshtbl.NewValueString("Uptime"), // only for Process
	Load:   boshtbl.NewValueString("Load\n(1m, 5m, 15m)"),

	CPUTotal: boshtbl.NewValueString("CPU\nTotal"), // only for Process
	CPUUser:  boshtbl.NewValueString("CPU\nUser"),
	CPUSys:   boshtbl.NewValueString("CPU\nSys"),
	CPUWait:  boshtbl.NewValueString("CPU\nWait"),

	Memory: boshtbl.NewValueString("Memory\nUsage"),
	Swap:   boshtbl.NewValueString("Swap\nUsage"),

	SystemDisk:     boshtbl.NewValueString("System\nDisk Usage"),
	EphemeralDisk:  boshtbl.NewValueString("Ephemeral\nDisk Usage"),
	PersistentDisk: boshtbl.NewValueString("Persistent\nDisk Usage"),
}

type InstanceTable struct {
	Processes, VMDetails, DeploymentDetails, Details, Stemcell, Vitals, CloudProperties, Cidr bool
}

func (t InstanceTable) Headers() []boshtbl.Header {
	headerVals := t.AsValues(InstanceTableHeader)
	var headers []boshtbl.Header
	for _, val := range headerVals {
		headers = append(headers, boshtbl.NewHeader(val.String()))
	}
	return headers
}

func (t InstanceTable) ForVMInfo(i boshdir.VMInfo) InstanceTableValues {

	var vmInfoIndex boshtbl.ValueInt

	if i.Index != nil {
		vmInfoIndex = boshtbl.NewValueInt(*i.Index)
	}

	activeStatus := "-"
	if i.Active != nil {
		activeStatus = fmt.Sprintf("%t", *i.Active)
	}

	stemcell := "-"
	if i.Stemcell.Name != "" {
		stemcell = fmt.Sprintf("%s/%s", i.Stemcell.Name, i.Stemcell.Version)
	}

	vals := InstanceTableValues{
		Name:    t.buildName(i),
		Process: boshtbl.ValueString{},

		ProcessState: boshtbl.ValueFmt{
			V:     boshtbl.NewValueString(i.InstanceState()),
			Error: !i.IsRunning(),
		},

		AZ:         boshtbl.NewValueString(i.AZ),
		VMType:     boshtbl.NewValueString(i.VMType),
		Active:     boshtbl.NewValueString(activeStatus),
		IPs:        boshtbl.NewValueStrings(i.IPs),
		IPs_cidr:   boshtbl.NewValueStrings(i.IPs_cidr),
		Deployment: boshtbl.NewValueString(i.Deployment),

		// Details
		State:           boshtbl.NewValueString(i.State),
		VMCID:           boshtbl.NewValueString(i.VMID),
		DiskCIDs:        boshtbl.NewValueStrings(i.DiskIDs),
		AgentID:         boshtbl.NewValueString(i.AgentID),
		Index:           vmInfoIndex,
		Bootstrap:       boshtbl.NewValueBool(i.Bootstrap),
		Ignore:          boshtbl.NewValueBool(i.Ignore),
		VMCreatedAt:     boshtbl.NewValueTime(i.VMCreatedAt.UTC()),
		CloudProperties: boshtbl.NewValueInterface(i.CloudProperties),

		Stemcell: boshtbl.NewValueString(stemcell),

		// Vitals
		Uptime: ValueUptime{i.Vitals.Uptime.Seconds},
		Load:   boshtbl.NewValueString(strings.Join(i.Vitals.Load, ", ")),

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
		vals.VMType = boshtbl.NewValueString(i.ResourcePool)
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
	}

	return boshtbl.NewValueString(name)
}

func (t InstanceTable) ForProcess(p boshdir.VMInfoProcess) InstanceTableValues {
	return InstanceTableValues{
		Name:    boshtbl.ValueString{},
		Process: boshtbl.NewValueString(p.Name),

		ProcessState: boshtbl.ValueFmt{
			V:     boshtbl.NewValueString(p.State),
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
	if t.Cidr {
		result = append(result, []boshtbl.Value{v.ProcessState, v.AZ, v.IPs_cidr}...)
	} else {
		result = append(result, []boshtbl.Value{v.ProcessState, v.AZ, v.IPs}...)
	}

	if t.DeploymentDetails {
		result = append(result, v.Deployment)
	}

	if t.Details {
		result = append(result, []boshtbl.Value{v.State, v.VMCID, v.VMType, v.DiskCIDs, v.AgentID, v.Index, v.Bootstrap, v.Ignore}...)
	} else if t.VMDetails {
		result = append(result, []boshtbl.Value{v.VMCID, v.VMType, v.Active}...)
	}

	if t.Stemcell {
		result = append(result, v.Stemcell)
	}

	if t.CloudProperties {
		result = append(result, v.CloudProperties)
	}

	if t.Vitals {
		result = append(result, []boshtbl.Value{v.VMCreatedAt, v.Uptime, v.Load}...)
		result = append(result, []boshtbl.Value{v.CPUTotal, v.CPUUser, v.CPUSys, v.CPUWait}...)
		result = append(result, []boshtbl.Value{v.Memory, v.Swap}...)
		result = append(result, []boshtbl.Value{v.SystemDisk, v.EphemeralDisk, v.PersistentDisk}...)
	}

	return result
}
