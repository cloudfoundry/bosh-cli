package client

import (
	"time"

	"github.com/cloudfoundry/bosh-s3cli/config"
)

type SignURLProvider interface {
	Sign(action string, objectID string, expiration time.Duration) (string, error)
}

func NewSignURLProvider(s3BlobstoreClient S3Blobstore, s3cliConfig *config.S3Cli) (SignURLProvider, error) {
	if s3cliConfig.SwiftAuthAccount != "" {
		client := NewSwiftClient(s3cliConfig)
		return &client, nil
	} else {
		return &s3BlobstoreClient, nil
	}
}
