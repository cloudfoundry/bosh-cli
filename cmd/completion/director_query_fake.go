package completion

type DirectorQueryFake struct {
	cmdContext *CmdContext
}

func NewDirectorQueryFake(cmdContext *CmdContext) *DirectorQueryFake {
	return &DirectorQueryFake{cmdContext: cmdContext}
}

func (c *DirectorQueryFake) listDirectorApiEndpoints(prefix string) ([]string, error) {
	return listDirectorApiEndpointsStatic(prefix), nil
}
func (c *DirectorQueryFake) listDeploymentNames(prefix string) ([]string, error) {
	return FilterQueryValues([]string{"fake-d1", "fake-d2"}, prefix), nil
}

func (c *DirectorQueryFake) listInstanceSlugs(prefix string, includeGroups bool) (values []string, err error) {
	deploymentName := c.cmdContext.DeploymentName
	if includeGroups {
		values = []string{
			"fake-" + deploymentName + "-g1",
			"fake-" + deploymentName + "-g1/fake-i1",
			"fake-" + deploymentName + "-g1/fake-i2",
			"fake-" + deploymentName + "-g2",
			"fake-" + deploymentName + "-g2/fake-i1",
			"fake-" + deploymentName + "-g2/fake-i2",
		}
	} else {
		values = []string{
			"fake-" + deploymentName + "-g1/fake-i1",
			"fake-" + deploymentName + "-g1/fake-i2",
			"fake-" + deploymentName + "-g2/fake-i1",
			"fake-" + deploymentName + "-g2/fake-i2",
		}
	}
	return FilterQueryValues(values, prefix), nil
}

func (c *DirectorQueryFake) listVmCIDs(prefix string) (values []string, err error) {
	return FilterQueryValues([]string{"fake-vm"}, prefix), nil
}

func (c *DirectorQueryFake) listDiskCIDs(prefix string) (values []string, err error) {
	return FilterQueryValues([]string{"fake-disk"}, prefix), nil
}

func (c *DirectorQueryFake) listOrphanedDiskCIDs(prefix string) (values []string, err error) {
	return FilterQueryValues([]string{"fake-orphaned-disk"}, prefix), nil
}

func (c *DirectorQueryFake) listActiveTasksCIDs(prefix string) (values []string, err error) {
	return FilterQueryValues([]string{"fake-task"}, prefix), nil
}

func (c *DirectorQueryFake) listErrands(prefix string) (values []string, err error) {
	return FilterQueryValues([]string{"fake-errand"}, prefix), nil
}

func (c *DirectorQueryFake) listReleaseSlugs(prefix string) (values []string, err error) {
	return FilterQueryValues([]string{"fake-rel"}, prefix), nil
}

func (c *DirectorQueryFake) listStemcells(prefix string) (values []string, err error) {
	return FilterQueryValues([]string{"fake-stemcell"}, prefix), nil
}
func (c *DirectorQueryFake) listEnvAliases(prefix string) (values []string, err error) {
	return FilterQueryValues([]string{"fake-env-alias"}, prefix), nil
}

func (c *DirectorQueryFake) listEventIds(prefix string) (values []string, err error) {
	return FilterQueryValues([]string{"fake-event"}, prefix), nil
}

func (c *DirectorQueryFake) listSnapshotCIDs(prefix string) (values []string, err error) {
	return FilterQueryValues([]string{"fake-snap"}, prefix), nil
}

func (c *DirectorQueryFake) listNetworkNames(prefix string) (values []string, err error) {
	return FilterQueryValues([]string{"fake-net"}, prefix), nil
}

func (c *DirectorQueryFake) listConfigIDs(prefix string) (values []string, err error) {
	return FilterQueryValues([]string{"fake-cfg"}, prefix), nil
}
