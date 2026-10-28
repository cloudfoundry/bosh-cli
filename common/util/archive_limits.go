package util

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
)

var archiveBudgetMu sync.Mutex
var archiveBudgets = map[string]*boshcmd.ArchiveBudget{}

func ArchiveExtractionOptions(destination string) (boshcmd.CompressorOptions, error) {
	limit := int64(128 * 1024 * 1024 * 1024)
	members := 10000
	manifestLimit := int64(32 * 1024 * 1024)
	if free, err := archiveAvailableSpace(destination); err == nil {
		if free < uint64(limit)*5/4 {
			limit = int64(free * 4 / 5)
		}
		// Leave room for filesystem blocks and directory metadata as well as
		// file contents when extracting onto an almost-full filesystem.
		if allowance := free / 5 / 8192; allowance < uint64(members) {
			members = int(allowance)
		}
	}

	if raw := os.Getenv("BOSH_RELEASE_MAX_EXPANDED_BYTES"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value <= 0 {
			return boshcmd.CompressorOptions{}, fmt.Errorf("BOSH_RELEASE_MAX_EXPANDED_BYTES must be a positive byte count")
		}
		limit = value
	}
	if raw := os.Getenv("BOSH_RELEASE_MAX_ARCHIVE_MEMBERS"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return boshcmd.CompressorOptions{}, fmt.Errorf("BOSH_RELEASE_MAX_ARCHIVE_MEMBERS must be positive")
		}
		members = value
	}
	if raw := os.Getenv("BOSH_RELEASE_MAX_MANIFEST_BYTES"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value <= 0 {
			return boshcmd.CompressorOptions{}, fmt.Errorf("BOSH_RELEASE_MAX_MANIFEST_BYTES must be a positive byte count")
		}
		manifestLimit = value
	}
	if limit <= 0 || members <= 0 {
		return boshcmd.CompressorOptions{}, fmt.Errorf("no extraction space available")
	}
	key := os.Getenv("BOSH_RELEASE_MAX_EXPANDED_BYTES") + "/" + os.Getenv("BOSH_RELEASE_MAX_ARCHIVE_MEMBERS") + "/" + os.Getenv("BOSH_RELEASE_MAX_MANIFEST_BYTES")
	archiveBudgetMu.Lock()
	defer archiveBudgetMu.Unlock()
	budget := archiveBudgets[key]
	if budget == nil {
		var err error
		budget, err = boshcmd.NewArchiveBudget(limit, members, manifestLimit)
		if err != nil {
			return boshcmd.CompressorOptions{}, err
		}
		archiveBudgets[key] = budget
	}
	return boshcmd.CompressorOptions{MaxExpandedBytes: limit, MaxArchiveMembers: members, MaxManifestBytes: manifestLimit, ArchiveBudget: budget}, nil
}
