package sif

import (
	"context"

	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/google/go-containerregistry/pkg/v1/partial"
)

// SingularityImageProvider is an image.Provider for a Singularity Image Format (SIF) image.
type SingularityImageProvider struct {
	path      string
	tmpDirGen *file.TempDirGenerator
}

// NewProviderFromPath creates a new provider instance for the Singularity Image Format (SIF) image
// at path.
func NewProviderFromPath(path string, tmpDirGen *file.TempDirGenerator) *SingularityImageProvider {
	return &SingularityImageProvider{
		path:      path,
		tmpDirGen: tmpDirGen,
	}
}

// Provide returns an Image that represents a Singularity Image Format (SIF) image.
func (p *SingularityImageProvider) Provide(ctx context.Context, userMetadata ...image.AdditionalMetadata) (*image.Image, error) {
	// We need to map the SIF to a GGCR v1.Image. Start with an implementation of the GGCR
	// partial.UncompressedImageCore interface.
	si, err := newSIFImage(p.path)
	if err != nil {
		return nil, err
	}

	// Promote our partial.UncompressedImageCore implementation to an v1.Image.
	ui, err := partial.UncompressedToImage(si)
	if err != nil {
		return nil, err
	}

	// The returned image must reference a content cache dir.
	contentCacheDir, err := p.tmpDirGen.NewDirectory()
	if err != nil {
		return nil, err
	}

	// Apply user-supplied metadata last to override any default behavior.
	metadata := []image.AdditionalMetadata{
		image.WithOS("linux"),
		image.WithArchitecture(si.arch, ""),
	}
	metadata = append(metadata, userMetadata...)

	return image.New(ui, contentCacheDir, metadata...), nil
}
