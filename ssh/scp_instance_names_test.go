package ssh

import (
	"testing"

	"github.com/cloudfoundry/bosh-cli/v7/director"
)

func TestSCPInstanceNamesBeforeCommandConstruction(t *testing.T) {
	args := NewSCPArgs([]string{"host:source", "output-((instance_id))"}, false)
	for _, id := range []string{"../../outside", `..\outside`, "..", "-oProxyCommand=cmd", "bad name", ""} {
		host := director.Host{IndexOrID: id}
		if err := (SCPRunnerImpl{}).Run(ConnectionOpts{}, director.SSHResult{Hosts: []director.Host{host}}, args); err == nil {
			t.Fatalf("accepted instance identifier %q", id)
		}
	}
	for _, id := range []string{"0", "instance-id", "b6a9a870-82da-4c2b-8ffd-60b0f55fa879"} {
		if err := args.validateHost(director.Host{IndexOrID: id}); err != nil {
			t.Fatal(err)
		}
	}
}
