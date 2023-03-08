package stereoscope

import (
	"github.com/anchore/stereoscope/pkg/image"
)

type config struct {
	Registry           image.RegistryOptions
	AdditionalMetadata []image.AdditionalMetadata
	Platform           *image.Platform
}
