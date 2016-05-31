package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-init/director"
	boshui "github.com/cloudfoundry/bosh-init/ui"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
)

type SnapshotsCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewSnapshotsCmd(ui boshui.UI, deployment boshdir.Deployment) SnapshotsCmd {
	return SnapshotsCmd{ui: ui, deployment: deployment}
}

func (c SnapshotsCmd) Run(opts SnapshotsOpts) error {
	snapshots, err := c.deployment.Snapshots()
	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Content: "snapshots",
		Header:  []string{"Instance", "CID", "Created At", "Clean"},
	}

	for _, s := range snapshots {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.ValueString{s.InstanceDesc()},
			boshtbl.ValueString{s.CID},
			boshtbl.ValueTime{s.CreatedAt},
			boshtbl.ValueBool{s.Clean},
		})
	}

	c.ui.PrintTable(table)

	return nil
}
