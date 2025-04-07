package completion

import (
	"fmt"
	"strings"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	boshcmd "github.com/cloudfoundry/bosh-cli/v7/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
)

// TODO: type QueryFunc func(prefix string) ([]string, error)

type DirectorQueryInterface interface {
	listDirectorApiEndpoints(prefix string) ([]string, error)
	listDeploymentNames(prefix string) ([]string, error)
	listInstanceSlugs(prefix string, includeGroups bool) ([]string, error)
	listVmCIDs(prefix string) ([]string, error)
	listDiskCIDs(prefix string) ([]string, error)
	listOrphanedDiskCIDs(prefix string) ([]string, error)
	listActiveTasksCIDs(prefix string) ([]string, error)
	listErrands(prefix string) ([]string, error)
	listReleaseSlugs(prefix string) ([]string, error)
	listStemcells(prefix string) ([]string, error)
	listEnvAliases(prefix string) ([]string, error)
	listEventIds(prefix string) ([]string, error)
	listSnapshotCIDs(prefix string) ([]string, error)
	listNetworkNames(prefix string) ([]string, error)
	listConfigIDs(prefix string) ([]string, error)
}

func FilterQueryValues(values []string, prefix string) []string {
	var fv []string
	for _, v := range values {
		if strings.HasPrefix(v, prefix) {
			fv = append(fv, v)
		}
	}
	return fv
}

type DirectorQuery struct {
	cmdContext *CmdContext
	session    boshcmd.Session
	logger     boshlog.Logger
	logTag     string
}

func NewDirectorQuery(logger boshlog.Logger, cmdContext *CmdContext, session boshcmd.Session) *DirectorQuery {
	return &DirectorQuery{
		cmdContext: cmdContext,
		session:    session,
		logger:     logger,
		logTag:     "completion.DirectorQuery",
	}
}

func listDirectorApiEndpointsStatic(prefix string) []string {
	paths := map[string][]string{
		"/info":         {},
		"/configs":      {},
		"/configs/diff": {},
		"/tasks/":       {"{id}", "/output", "/cancel"},
		"/stemcells":    {},
		"/releases":     {},
		"/deployments/": {
			"{name}",
			"?exclude_configs=true&exclude_releases=true&exclude_stemcells=true",
			"/instances",
			"/instances?format=full",
			"/vms",
			"/vms?format=full"},
		"/events": {},
	}
	var matchedStrings []string
	for k, arr := range paths {
		if prefix == k || strings.HasPrefix(k, prefix) {
			// completion: "/ta" -> "/tasks"
			matchedStrings = append(matchedStrings, k)
			if strings.HasSuffix(k, "/") {
				matchedStrings = append(matchedStrings, k+"{}")
			}
		} else if strings.HasPrefix(prefix, k) {
			// completion: "/tasks/{id}/ou"
			s := strings.TrimPrefix(prefix, k)
			// s: "{id}/ou"
			completeId := ""
			completePathPrefix := ""
			sa := strings.Split(s, "/")
			if len(sa) > 0 {
				completeId = sa[0]
				if len(sa) > 1 {
					sa := strings.Split(s, "/")
					completePathPrefix = sa[1]
				}
			}
			for _, v := range arr {
				if strings.HasPrefix(v, "/"+completePathPrefix) {
					matchedStrings = append(matchedStrings, k+completeId+v)
				}
			}
		}
	}
	return matchedStrings
}

func (c *DirectorQuery) listDirectorApiEndpoints(prefix string) ([]string, error) {
	return listDirectorApiEndpointsStatic(prefix), nil
}

func (c *DirectorQuery) readFromCache(group string) (cache *CompleteCache, values []string, valid bool) {
	cache = NewCompleteCache(c.logger, c.cmdContext, group)
	if values, valid, _ = cache.GetValues(); valid { //nolint:errcheck
		c.logger.Debug(c.logTag, "'%' read from cache %s", group, cache)
		return cache, values, valid
	} else {
		return cache, []string{}, false
	}
}
func (c *DirectorQuery) writeToCache(cache *CompleteCache, values []string) {
	err := cache.PutValues(values)
	if err == nil {
		c.logger.Debug(c.logTag, "successful write to cache '%s'", cache)
	} else {
		c.logger.Debug(c.logTag, "write to cache '%s' error: %v", cache, err)
	}
}

func (c *DirectorQuery) director() (director boshdir.Director, err error) {
	if c.session == nil {
		err = fmt.Errorf("session is nil")
	} else {
		director, err = c.session.Director()
	}
	if err != nil {
		c.logger.Debug(logTag, "getting director error: %v", err)
	}
	return director, err
}

func (c *DirectorQuery) getCacheOrDeployment(group string, prefix string) (names []string, validCache bool, cache *CompleteCache, deployment boshdir.Deployment, err error) {
	cache, values, valid := c.readFromCache(group)
	if valid {
		return FilterQueryValues(values, prefix), valid, cache, nil, nil
	}

	deploymentName := c.cmdContext.DeploymentName
	director, err := c.director()
	if err != nil {
		return []string{}, false, cache, nil, err
	}
	if deploymentName == "" {
		return []string{}, false, cache, nil, fmt.Errorf("deployment name is empty")
	}
	deployment, err = director.FindDeployment(deploymentName)
	if err != nil {
		c.logger.Debug(c.logTag, "getting deployment '%s' error: %v", deploymentName, err)
		return []string{}, false, cache, nil, err
	}
	return []string{}, false, cache, deployment, err
}

func (c *DirectorQuery) listDeploymentNames(prefix string) (names []string, err error) {
	cache, values, valid := c.readFromCache("deployments")
	if valid {
		return FilterQueryValues(values, prefix), nil
	}

	director, err := c.director()
	if err != nil {
		return []string{}, err
	}
	ds, err := director.Deployments()
	if err != nil {
		c.logger.Debug(c.logTag, "listing deployments error: %v", err)
		return []string{}, err
	}

	c.logger.Debug(c.logTag, "Deployments get from director")
	var newValues []string
	for _, d := range ds {
		newValues = append(newValues, d.Name())
	}

	c.writeToCache(cache, newValues)
	return FilterQueryValues(newValues, prefix), nil
}

func (c *DirectorQuery) listInstanceSlugs(prefix string, includeGroups bool) (names []string, err error) {
	group := "instances"
	if includeGroups {
		group = group + "-with-groups"
	}
	names, validCache, cache, deployment, err := c.getCacheOrDeployment(group, prefix)
	if err != nil {
		return names, err
	}
	if validCache {
		return names, nil
	}

	instances, err := deployment.Instances()
	if err != nil {
		c.logger.Debug(c.logTag, "listing instances error: %v", err)
		return []string{}, err
	}

	c.logger.Debug(c.logTag, "Instance slugs get from director")
	var newValues []string
	groups := make(map[string]bool)
	for _, i := range instances {
		if includeGroups {
			if _, exists := groups[i.Group]; !exists {
				newValues = append(newValues, i.Group)
				groups[i.Group] = true
			}
		}
		slug := i.Group + "/" + i.ID
		newValues = append(newValues, slug)
	}
	c.writeToCache(cache, newValues)
	return FilterQueryValues(newValues, prefix), nil
}

func (c *DirectorQuery) listVmCIDs(prefix string) (values []string, err error) {
	cachedValues, validCache, cache, deployment, err := c.getCacheOrDeployment("vms", prefix)
	if err != nil {
		return cachedValues, err
	}
	if validCache {
		return cachedValues, nil
	}
	entities, err := deployment.VMInfos()
	if err != nil {
		c.logger.Debug(c.logTag, "listing vms error: %v", err)
		return []string{}, err
	}
	for _, ent := range entities {
		values = append(values, ent.ID)
	}
	c.writeToCache(cache, values)
	return FilterQueryValues(values, prefix), nil
}

func (c *DirectorQuery) listErrands(prefix string) (values []string, err error) {
	cachedValues, validCache, cache, deployment, err := c.getCacheOrDeployment("errands", prefix)
	if err != nil {
		return cachedValues, err
	}
	if validCache {
		return cachedValues, nil
	}
	entities, err := deployment.Errands()
	if err != nil {
		c.logger.Debug(c.logTag, "listing errands error: %v", err)
		return []string{}, err
	}
	for _, ent := range entities {
		values = append(values, ent.Name)
	}
	c.writeToCache(cache, values)
	return FilterQueryValues(values, prefix), nil
}

func (c *DirectorQuery) listDiskCIDs(prefix string) (values []string, err error) {
	//TODO implement me
	return values, nil
}

func (c *DirectorQuery) listOrphanedDiskCIDs(prefix string) (values []string, err error) {
	//TODO implement me
	return values, nil
}

func (c *DirectorQuery) listActiveTasksCIDs(prefix string) (values []string, err error) {
	//TODO implement me
	return values, nil
}

func (c *DirectorQuery) listReleaseSlugs(prefix string) (values []string, err error) {
	//TODO implement me
	return values, nil
}

func (c *DirectorQuery) listStemcells(prefix string) (values []string, err error) {
	//TODO implement me
	return values, nil
}

func (c *DirectorQuery) listSnapshotCIDs(prefix string) (values []string, err error) {
	//TODO implement me
	return values, nil
}
func (c *DirectorQuery) listEnvAliases(prefix string) (values []string, err error) {
	//TODO implement me
	return values, nil
}

func (c *DirectorQuery) listEventIds(prefix string) (values []string, err error) {
	//TODO implement me
	return values, nil
}

func (c *DirectorQuery) listNetworkNames(prefix string) (values []string, err error) {
	//TODO implement me
	return values, nil
}

func (c *DirectorQuery) listConfigIDs(prefix string) (values []string, err error) {
	//TODO implement me
	return values, nil
}
