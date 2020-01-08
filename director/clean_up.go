package director

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	gourl "net/url"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type CleanUpResponse struct {
	Releases         []CleanableRelease         `json:"releases"`
	Stemcells        []StemcellResp             `json:"stemcells"`
	CompiledPackages []CleanableCompiledPackage `json:"compiled_packages"`
	OrphanedDisks    []OrphanDiskResp           `json:"orphaned_disks"`
	OrphanedVMs      []OrphanedVMResponse       `json:"orphaned_vms"`
	ExportedReleases []string                   `json:"exported_releases"`
	DNSBlobs         []string                   `json:"dns_blobs"`
}

func (d DirectorImpl) CleanUp(all bool, dryRun bool, keepOrphanedDisks bool) (CleanUp, error) {
	return d.client.CleanUp(all, dryRun, keepOrphanedDisks)
}

func (c Client) CleanUp(all bool, dryRun bool, keepOrphanedDisks bool) (CleanUp, error) {
	if dryRun {
		return c.dryCleanUp(all, keepOrphanedDisks)
	} else {
		return CleanUp{}, c.cleanUp(all, keepOrphanedDisks)
	}
}

func (c Client) dryCleanUp(all bool, keepOrphanedDisks bool) (CleanUp, error) {
	query := gourl.Values{}
	query.Add("remove_all", strconv.FormatBool(all))
	query.Add("keep_orphaned_disks", strconv.FormatBool(keepOrphanedDisks))

	path := fmt.Sprintf("/cleanup/dryrun?%s", query.Encode())

	var resp CleanUpResponse

	err := c.clientRequest.Get(path, &resp)

	if err != nil {
		return CleanUp{}, bosherr.WrapErrorf(err, "Cleaning up resources")
	}

	orphanedVms, err := transformOrphanedVMs(resp.OrphanedVMs)
	if err != nil {
		return CleanUp{}, bosherr.WrapErrorf(err, "Cleaning up resources")
	}

	stemcells, err := transformStemcells(resp.Stemcells, c)
	if err != nil {
		return CleanUp{}, bosherr.WrapErrorf(err, "Cleaning up resources")
	}

	cleanUp := CleanUp{
		Releases:         resp.Releases,
		Stemcells:        stemcells,
		CompiledPackages: resp.CompiledPackages,
		OrphanedDisks:    resp.OrphanedDisks,
		OrphanedVMs:      orphanedVms,
		ExportedReleases: resp.ExportedReleases,
		DNSBlobs:         resp.DNSBlobs,
	}

	return cleanUp, nil
}

func (c Client) cleanUp(all bool, keepOrphanedDisks bool) error {
	body := map[string]interface{}{
		"config": map[string]bool{"remove_all": all, "keep_orphaned_disks": keepOrphanedDisks},
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return bosherr.WrapErrorf(err, "Marshaling request body")
	}

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}

	path := "/cleanup"
	_, err = c.taskClientRequest.PostResult(path, reqBody, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(err, "Cleaning up resources")
	}

	return nil
}
