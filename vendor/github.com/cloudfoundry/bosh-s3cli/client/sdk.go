package client

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cloudfoundry/bosh-s3cli/config"
)

func NewSDK(c config.S3Cli) (*s3.S3, error) {
	// Wiring in the httpclient.ClientFactory
	// proved difficult, so we bailed and just did
	// this instead for now... maybe there's a better way?
	//
	// Notably, the gcscli just uses the http.DefaultClient.
	// We're guessing we just have logic to skip verify because
	// there may be s3-compatible blobstores without real certs,
	// but even for gsc, we might want to take a look at adding
	// sane timeouts, etc.
	httpClient := http.DefaultClient
	httpClient.Transport = &http.Transport{
		TLSHandshakeTimeout: 30 * time.Second,
		DisableKeepAlives:   true,
	}

	if !c.SSLVerifyPeer {
		httpClient.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	s3Config := aws.NewConfig().
		WithLogLevel(aws.LogOff).
		WithS3ForcePathStyle(true).
		WithDisableSSL(!c.UseSSL).
		WithHTTPClient(httpClient)

	if c.UseRegion() {
		s3Config = s3Config.WithRegion(c.Region).WithEndpoint(c.S3Endpoint())
	} else {
		s3Config = s3Config.WithRegion(config.EmptyRegion).WithEndpoint(c.S3Endpoint())
	}

	if c.CredentialsSource == config.StaticCredentialsSource {
		s3Config = s3Config.WithCredentials(credentials.NewStaticCredentials(c.AccessKeyID, c.SecretAccessKey, ""))
	}

	if c.CredentialsSource == config.NoneCredentialsSource {
		s3Config = s3Config.WithCredentials(credentials.AnonymousCredentials)
	}

	s3Session := session.New(s3Config)
	s3Client := s3.New(s3Session)

	if c.UseV2SigningMethod {
		setv2Handlers(s3Client)
	}

	return s3Client, nil
}
