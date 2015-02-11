package state

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
	bmstatepkg "github.com/cloudfoundry/bosh-micro-cli/state/pkg"
)

type remotePackageCompiler struct {
	blobstore   bmblobstore.Blobstore
	agentClient bmagentclient.AgentClient
	packageRepo bmstatepkg.CompiledPackageRepo
}

func NewRemotePackageCompiler(blobstore bmblobstore.Blobstore, agentClient bmagentclient.AgentClient, packageRepo bmstatepkg.CompiledPackageRepo) bmstatepkg.Compiler {
	return &remotePackageCompiler{
		blobstore:   blobstore,
		agentClient: agentClient,
		packageRepo: packageRepo,
	}
}

func (c *remotePackageCompiler) Compile(releasePackage *bmrelpkg.Package) (record bmstatepkg.CompiledPackageRecord, err error) {

	blobID, err := c.blobstore.Add(releasePackage.ArchivePath)
	if err != nil {
		return bmstatepkg.CompiledPackageRecord{}, bosherr.WrapErrorf(err, "Adding release package archive '%s' to blobstore", releasePackage.ArchivePath)
	}

	packageSource := bmagentclient.BlobRef{
		Name:        releasePackage.Name,
		Version:     releasePackage.Fingerprint,
		SHA1:        releasePackage.SHA1,
		BlobstoreID: blobID,
	}

	// Resolve dependencies from map of previously compiled packages.
	// Only install the package's immediate dependencies when compiling (not all transitive dependencies).
	packageDependencies := make([]bmagentclient.BlobRef, len(releasePackage.Dependencies), len(releasePackage.Dependencies))
	for i, dependency := range releasePackage.Dependencies {
		compiledPackageRecord, found, err := c.packageRepo.Find(*dependency)
		if err != nil {
			return record, bosherr.WrapErrorf(
				err,
				"Finding compiled package '%s/%s' as dependency for '%s/%s'",
				dependency.Name,
				dependency.Fingerprint,
				releasePackage.Name,
				releasePackage.Fingerprint,
			)
		}
		if !found {
			return record, bosherr.Errorf(
				"Remote compilation failure: Package '%s' requires package '%s', but it has not been compiled",
				releasePackage.Name,
				dependency.Name,
			)
		}
		packageDependencies[i] = bmagentclient.BlobRef{
			Name:        dependency.Name,
			Version:     dependency.Fingerprint,
			BlobstoreID: compiledPackageRecord.BlobID,
			SHA1:        compiledPackageRecord.BlobSHA1,
		}
	}

	compiledPackageRef, err := c.agentClient.CompilePackage(packageSource, packageDependencies)
	if err != nil {
		return record, bosherr.WrapErrorf(err, "Remotely compiling package '%s' with the agent", releasePackage.Name)
	}

	record = bmstatepkg.CompiledPackageRecord{
		BlobID:   compiledPackageRef.BlobstoreID,
		BlobSHA1: compiledPackageRef.SHA1,
	}

	err = c.packageRepo.Save(*releasePackage, record)
	if err != nil {
		return record, bosherr.WrapErrorf(err, "Saving compiled package record %#v of package %#v", record, releasePackage)
	}

	return record, nil
}
