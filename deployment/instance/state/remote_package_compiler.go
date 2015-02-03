package state

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
)

type remotePackageCompiler struct {
	blobstore   bmblobstore.Blobstore
	agentClient bmagentclient.AgentClient
}

func NewRemotePackageCompiler(blobstore bmblobstore.Blobstore, agentClient bmagentclient.AgentClient) PackageCompiler {
	return &remotePackageCompiler{
		blobstore:   blobstore,
		agentClient: agentClient,
	}
}

func (c *remotePackageCompiler) Compile(releasePackage *bmrelpkg.Package, compiledPackageRefs map[string]PackageRef) (PackageRef, error) {

	blobID, err := c.blobstore.Add(releasePackage.ArchivePath)
	if err != nil {
		return PackageRef{}, bosherr.WrapErrorf(err, "Adding release package archive '%s' to blobstore", releasePackage.ArchivePath)
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
		packageRef, found := compiledPackageRefs[dependency.Name]
		if !found {
			return PackageRef{}, bosherr.Errorf("Remote compilation failure: Package '%s' requires package '%s', but it has not been compiled", releasePackage.Name, dependency.Name)
		}
		packageDependencies[i] = bmagentclient.BlobRef{
			Name:        packageRef.Name,
			Version:     packageRef.Version,
			SHA1:        packageRef.Archive.SHA1,
			BlobstoreID: packageRef.Archive.BlobstoreID,
		}
	}

	compiledPackageRef, err := c.agentClient.CompilePackage(packageSource, packageDependencies)
	if err != nil {
		return PackageRef{}, bosherr.WrapErrorf(err, "Remotely compiling package '%s' with the agent", releasePackage.Name)
	}

	return PackageRef{
		Name:    compiledPackageRef.Name,
		Version: compiledPackageRef.Version,
		Archive: BlobRef{
			BlobstoreID: compiledPackageRef.BlobstoreID,
			SHA1:        compiledPackageRef.SHA1,
		},
	}, nil
}
