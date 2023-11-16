package completion

import (
	"reflect"

	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/spf13/cobra"
)

type CompleteFunctions struct {
	directorQuery DirectorQueryInterface
	logger        boshlog.Logger
	logTag        string
}
type CompleteFunctionsMap map[string]func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

func NewCompleteFunctionsMap(logger boshlog.Logger, directorQuery DirectorQueryInterface) *CompleteFunctionsMap {
	c := &CompleteFunctions{
		directorQuery: directorQuery,
		logger:        logger,
		logTag:        "completion.CompleteFunctions",
	}
	cfm := &CompleteFunctionsMap{
		"--deployment": c.listDeploymentNames,
		"--dir":        c.listDirectories,
		"--ops-file":   c.listFiles,
		"--var-file":   c.listFiles,
		"--vars-file":  c.listFiles,
		"--vars-store": c.listFiles,
		reflect.TypeOf(opts.AddBlobArgs{}).Name():                          c.noFile,
		reflect.TypeOf(opts.AliasEnvArgs{}).Name():                         c.listEnvAliases,
		reflect.TypeOf(opts.AllOrInstanceGroupOrInstanceSlugArgs{}).Name(): c.listInstanceGroupsOrSlugs,
		reflect.TypeOf(opts.AttachDiskArgs{}).Name():                       c.listDiskCIDs,
		reflect.TypeOf(opts.ConfigArgs{}).Name():                           c.listConfigIDs,
		reflect.TypeOf(opts.CreateEnvArgs{}).Name():                        c.listFiles,
		reflect.TypeOf(opts.CreateRecoveryPlanArgs{}).Name():               c.listFiles,
		reflect.TypeOf(opts.CreateReleaseArgs{}).Name():                    c.listFiles,
		reflect.TypeOf(opts.CurlArgs{}).Name():                             c.listDirectorApiEndpoints,
		reflect.TypeOf(opts.DeleteConfigArgs{}).Name():                     c.listConfigIDs,
		reflect.TypeOf(opts.DeleteDiskArgs{}).Name():                       c.listOrphanedDiskCIDs,
		reflect.TypeOf(opts.DeleteEnvArgs{}).Name():                        c.listFiles,
		reflect.TypeOf(opts.DeleteNetworkArgs{}).Name():                    c.listNetworkNames,
		reflect.TypeOf(opts.DeleteReleaseArgs{}).Name():                    c.listReleaseSlugs,
		reflect.TypeOf(opts.DeleteSnapshotArgs{}).Name():                   c.listSnapshotCIDs,
		reflect.TypeOf(opts.DeleteStemcellArgs{}).Name():                   c.listStemcells,
		reflect.TypeOf(opts.DeleteVMArgs{}).Name():                         c.listVmCIDs,
		reflect.TypeOf(opts.DeployArgs{}).Name():                           c.listFiles,
		reflect.TypeOf(opts.EventArgs{}).Name():                            c.listEventIds,
		reflect.TypeOf(opts.ExportReleaseArgs{}).Name():                    c.noFile,
		reflect.TypeOf(opts.FinalizeReleaseArgs{}).Name():                  c.noFile,
		reflect.TypeOf(opts.GenerateJobArgs{}).Name():                      c.noFile,
		reflect.TypeOf(opts.GeneratePackageArgs{}).Name():                  c.noFile,
		reflect.TypeOf(opts.InspectLocalReleaseArgs{}).Name():              c.noFile,
		reflect.TypeOf(opts.InspectReleaseArgs{}).Name():                   c.listReleaseSlugs,
		reflect.TypeOf(opts.InspectStemcellTarballArgs{}).Name():           c.listFiles,
		reflect.TypeOf(opts.InstanceSlugArgs{}).Name():                     c.listInstanceSlugs,
		reflect.TypeOf(opts.InterpolateArgs{}).Name():                      c.listFiles,
		reflect.TypeOf(opts.OrphanDiskArgs{}).Name():                       c.listDiskCIDs,
		reflect.TypeOf(opts.RecoverArgs{}).Name():                          c.listFiles,
		reflect.TypeOf(opts.RedigestReleaseArgs{}).Name():                  c.listFiles,
		reflect.TypeOf(opts.RemoveBlobArgs{}).Name():                       c.listFiles,
		reflect.TypeOf(opts.RepackStemcellArgs{}).Name():                   c.listFiles,
		reflect.TypeOf(opts.RunErrandArgs{}).Name():                        c.listErrands,
		reflect.TypeOf(opts.SCPArgs{}).Name():                              c.listFiles,
		reflect.TypeOf(opts.SshSlugArgs{}).Name():                          c.listInstanceGroupsOrSlugs,
		reflect.TypeOf(opts.StartStopEnvArgs{}).Name():                     c.listFiles,
		reflect.TypeOf(opts.TaskArgs{}).Name():                             c.listActiveTaskIDs,
		reflect.TypeOf(opts.UnaliasEnvArgs{}).Name():                       c.listEnvAliases,
		reflect.TypeOf(opts.UpdateCloudConfigArgs{}).Name():                c.listFiles,
		reflect.TypeOf(opts.UpdateConfigArgs{}).Name():                     c.listFiles,
		reflect.TypeOf(opts.UpdateCPIConfigArgs{}).Name():                  c.listFiles,
		reflect.TypeOf(opts.UpdateResurrectionArgs{}).Name():               c.noFile,
		reflect.TypeOf(opts.UpdateRuntimeConfigArgs{}).Name():              c.listFiles,
		reflect.TypeOf(opts.UploadReleaseArgs{}).Name():                    c.listFiles,
		reflect.TypeOf(opts.UploadStemcellArgs{}).Name():                   c.listFiles,
		reflect.TypeOf(opts.VendorPackageArgs{}).Name():                    c.noFile,
	}
	return cfm
}

func (c *CompleteFunctions) listDirectorApiEndpoints(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	names, err := c.directorQuery.listDirectorApiEndpoints(toComplete)
	if err == nil {
		return names, cobra.ShellCompDirectiveNoFileComp
	} else {
		return []string{"{error: " + err.Error() + "}"}, cobra.ShellCompDirectiveError
	}
}

func (c *CompleteFunctions) noFile(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}
func (c *CompleteFunctions) listFiles(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveFilterFileExt
}
func (c *CompleteFunctions) listDirectories(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveFilterDirs
}
func (c *CompleteFunctions) listDeploymentNames(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	names, err := c.directorQuery.listDeploymentNames(toComplete)
	if err == nil {
		return names, cobra.ShellCompDirectiveNoFileComp
	} else {
		c.logger.Debug(c.logTag, "listing error: %v", err)
		return []string{}, cobra.ShellCompDirectiveError
	}
}
func (c *CompleteFunctions) listInstanceSlugs(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	names, err := c.directorQuery.listInstanceSlugs(toComplete, false)
	if err == nil {
		return names, cobra.ShellCompDirectiveNoFileComp
	} else {
		c.logger.Debug(c.logTag, "listing error: %v", err)
		return names, cobra.ShellCompDirectiveError
	}
}
func (c *CompleteFunctions) listInstanceGroupsOrSlugs(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	names, err := c.directorQuery.listInstanceSlugs(toComplete, true)
	if err == nil {
		return names, cobra.ShellCompDirectiveNoFileComp
	} else {
		c.logger.Debug(c.logTag, "listing error: %v", err)
		return names, cobra.ShellCompDirectiveError
	}
}
func (c *CompleteFunctions) listVmCIDs(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ids, err := c.directorQuery.listVmCIDs(toComplete)
	if err == nil {
		return ids, cobra.ShellCompDirectiveNoFileComp
	} else {
		c.logger.Debug(c.logTag, "listing error: %v", err)
		return ids, cobra.ShellCompDirectiveError
	}
}
func (c *CompleteFunctions) listOrphanedDiskCIDs(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ids, err := c.directorQuery.listOrphanedDiskCIDs(toComplete)
	if err == nil {
		return ids, cobra.ShellCompDirectiveNoFileComp
	} else {
		c.logger.Debug(c.logTag, "listing error: %v", err)
		return ids, cobra.ShellCompDirectiveError
	}
}
func (c *CompleteFunctions) listDiskCIDs(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ids, err := c.directorQuery.listDiskCIDs(toComplete)
	if err == nil {
		return ids, cobra.ShellCompDirectiveNoFileComp
	} else {
		c.logger.Debug(c.logTag, "listing error: %v", err)
		return ids, cobra.ShellCompDirectiveError
	}
}
func (c *CompleteFunctions) listActiveTaskIDs(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ids, err := c.directorQuery.listActiveTasksCIDs(toComplete)
	if err == nil {
		return ids, cobra.ShellCompDirectiveNoFileComp
	} else {
		c.logger.Debug(c.logTag, "listing error: %v", err)
		return ids, cobra.ShellCompDirectiveError
	}
}
func (c *CompleteFunctions) listErrands(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	values, err := c.directorQuery.listErrands(toComplete)
	if err == nil {
		return values, cobra.ShellCompDirectiveNoFileComp
	} else {
		c.logger.Debug(c.logTag, "listing error: %v", err)
		return values, cobra.ShellCompDirectiveError
	}
}
func (c *CompleteFunctions) listReleaseSlugs(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	values, err := c.directorQuery.listReleaseSlugs(toComplete)
	if err == nil {
		return values, cobra.ShellCompDirectiveNoFileComp
	} else {
		c.logger.Debug(c.logTag, "listing error: %v", err)
		return values, cobra.ShellCompDirectiveError
	}
}
func (c *CompleteFunctions) listStemcells(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	values, err := c.directorQuery.listStemcells(toComplete)
	if err == nil {
		return values, cobra.ShellCompDirectiveNoFileComp
	} else {
		c.logger.Debug(c.logTag, "listing error: %v", err)
		return values, cobra.ShellCompDirectiveError
	}
}
func (c *CompleteFunctions) listEnvAliases(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	values, err := c.directorQuery.listEnvAliases(toComplete)
	if err == nil {
		return values, cobra.ShellCompDirectiveNoFileComp
	} else {
		c.logger.Debug(c.logTag, "listing error: %v", err)
		return values, cobra.ShellCompDirectiveError
	}
}
func (c *CompleteFunctions) listEventIds(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	values, err := c.directorQuery.listEventIds(toComplete)
	if err == nil {
		return values, cobra.ShellCompDirectiveNoFileComp
	} else {
		c.logger.Debug(c.logTag, "listing error: %v", err)
		return values, cobra.ShellCompDirectiveError
	}
}
func (c *CompleteFunctions) listSnapshotCIDs(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	values, err := c.directorQuery.listSnapshotCIDs(toComplete)
	if err == nil {
		return values, cobra.ShellCompDirectiveNoFileComp
	} else {
		c.logger.Debug(c.logTag, "listing error: %v", err)
		return values, cobra.ShellCompDirectiveError
	}
}
func (c *CompleteFunctions) listNetworkNames(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	values, err := c.directorQuery.listNetworkNames(toComplete)
	if err == nil {
		return values, cobra.ShellCompDirectiveNoFileComp
	} else {
		c.logger.Debug(c.logTag, "listing error: %v", err)
		return values, cobra.ShellCompDirectiveError
	}
}
func (c *CompleteFunctions) listConfigIDs(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	values, err := c.directorQuery.listConfigIDs(toComplete)
	if err == nil {
		return values, cobra.ShellCompDirectiveNoFileComp
	} else {
		c.logger.Debug(c.logTag, "listing error: %v", err)
		return values, cobra.ShellCompDirectiveError
	}
}
