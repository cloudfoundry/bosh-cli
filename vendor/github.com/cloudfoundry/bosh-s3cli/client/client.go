package client

import (
	"io"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/cloudfoundry/bosh-s3cli/config"
)

type S3CompatibleClient interface {
	Get(src string, dest io.WriterAt) error
	Put(src io.ReadSeeker, dest string) error
	Delete(dest string) error
	Exists(dest string) (bool, error)
	Sign(objectID string, action string, expiration time.Duration) (string, error)
}

// New returns an S3CompatibleClient
func New(s3Client *s3.S3, s3cliConfig *config.S3Cli) S3CompatibleClient {
	return &s3CompatibleClient{
		s3cliConfig: s3cliConfig,
		openstackSwiftBlobstore: &openstackSwiftS3Client{
			s3cliConfig: s3cliConfig,
		},
		awsS3BlobstoreClient: &awsS3Client{
			s3Client:    s3Client,
			s3cliConfig: s3cliConfig,
		},
	}
}

type s3CompatibleClient struct {
	s3cliConfig             *config.S3Cli
	awsS3BlobstoreClient    *awsS3Client
	openstackSwiftBlobstore *openstackSwiftS3Client
}

func (c *s3CompatibleClient) Get(src string, dest io.WriterAt) error {
	return c.awsS3BlobstoreClient.Get(src, dest)
}

func (c *s3CompatibleClient) Put(src io.ReadSeeker, dest string) error {
	return c.awsS3BlobstoreClient.Put(src, dest)
}

func (c *s3CompatibleClient) Delete(dest string) error {
	return c.awsS3BlobstoreClient.Delete(dest)
}

func (c *s3CompatibleClient) Exists(dest string) (bool, error) {
	return c.awsS3BlobstoreClient.Exists(dest)
}

func (c *s3CompatibleClient) Sign(objectID string, action string, expiration time.Duration) (string, error) {
	if c.s3cliConfig.SwiftAuthAccount != "" {
		return c.openstackSwiftBlobstore.Sign(objectID, action, expiration)
	}

	return c.awsS3BlobstoreClient.Sign(objectID, action, expiration)
}
