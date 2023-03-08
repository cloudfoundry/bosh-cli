package oci

import (
	"context"
	"fmt"

	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/google/go-containerregistry/pkg/v1/layout"
)

// DirectoryImageProvider is an image.Provider for an OCI image (V1) for an existing tar on disk (from a buildah push <img> oci:<img> command).
type DirectoryImageProvider struct {
	path      string
	tmpDirGen *file.TempDirGenerator
}

// NewProviderFromPath creates a new provider instance for the specific image already at the given path.
func NewProviderFromPath(path string, tmpDirGen *file.TempDirGenerator) *DirectoryImageProvider {
	return &DirectoryImageProvider{
		path:      path,
		tmpDirGen: tmpDirGen,
	}
}

// Provide an image object that represents the OCI image as a directory.
func (p *DirectoryImageProvider) Provide(_ context.Context, userMetadata ...image.AdditionalMetadata) (*image.Image, error) {
	pathObj, err := layout.FromPath(p.path)
	if err != nil {
		return nil, fmt.Errorf("unable to read image from OCI directory path %q: %w", p.path, err)
	}

	index, err := layout.ImageIndexFromPath(p.path)
	if err != nil {
		return nil, fmt.Errorf("unable to parse OCI directory index: %w", err)
	}

	indexManifest, err := index.IndexManifest()
	if err != nil {
		return nil, fmt.Errorf("unable to parse OCI directory indexManifest: %w", err)
	}

	// for now, lets only support one image indexManifest (it is not clear how to handle multiple manifests)
	if len(indexManifest.Manifests) != 1 {
		return nil, fmt.Errorf("unexpected number of OCI directory manifests (found %d)", len(indexManifest.Manifests))
	}

	manifest := indexManifest.Manifests[0]
	img, err := pathObj.Image(manifest.Digest)
	if err != nil {
		return nil, fmt.Errorf("unable to parse OCI directory as an image: %w", err)
	}

	var metadata = []image.AdditionalMetadata{
		image.WithManifestDigest(manifest.Digest.String()),
	}

	// make a best-effort attempt at getting the raw indexManifest
	rawManifest, err := img.RawManifest()
	if err == nil {
		metadata = append(metadata, image.WithManifest(rawManifest))
	}

	// apply user-supplied metadata last to override any default behavior
	metadata = append(metadata, userMetadata...)

	contentTempDir, err := p.tmpDirGen.NewDirectory("oci-dir-image")
	if err != nil {
		return nil, err
	}

	return image.New(img, contentTempDir, metadata...), nil
}
