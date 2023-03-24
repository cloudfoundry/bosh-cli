package oci

import (
	"context"
	"fmt"
	"os"

	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/image"
)

// TarballImageProvider is an image.Provider for an OCI image (V1) for an existing tar on disk (from a buildah push <img> oci-archive:<name>.tar command).
type TarballImageProvider struct {
	path      string
	tmpDirGen *file.TempDirGenerator
}

// NewProviderFromTarball creates a new provider instance for the specific image tarball already at the given path.
func NewProviderFromTarball(path string, tmpDirGen *file.TempDirGenerator) *TarballImageProvider {
	return &TarballImageProvider{
		path:      path,
		tmpDirGen: tmpDirGen,
	}
}

// Provide an image object that represents the OCI image from a tarball.
func (p *TarballImageProvider) Provide(ctx context.Context, metadata ...image.AdditionalMetadata) (*image.Image, error) {
	// note: we are untaring the image and using the existing directory provider, we could probably enhance the google
	// container registry lib to do this without needing to untar to a temp dir (https://github.com/google/go-containerregistry/issues/726)
	f, err := os.Open(p.path)
	if err != nil {
		return nil, fmt.Errorf("unable to open OCI tarball: %w", err)
	}

	tempDir, err := p.tmpDirGen.NewDirectory("oci-tarball-image")
	if err != nil {
		return nil, err
	}

	if err = file.UntarToDirectory(f, tempDir); err != nil {
		return nil, err
	}

	return NewProviderFromPath(tempDir, p.tmpDirGen).Provide(ctx, metadata...)
}
