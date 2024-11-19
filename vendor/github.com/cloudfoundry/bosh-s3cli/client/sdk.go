package client

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"

	"github.com/cloudfoundry/bosh-s3cli/config"
)

func NewAwsS3Client(c *config.S3Cli) (*s3.S3, error) {
	var httpClient *http.Client

	if c.SSLVerifyPeer {
		httpClient = boshhttp.CreateDefaultClient(nil)
	} else {
		httpClient = boshhttp.CreateDefaultClientInsecureSkipVerify()
	}

	awsConfig := aws.NewConfig().
		WithLogLevel(aws.LogOff).
		WithS3ForcePathStyle(!c.HostStyle).
		WithDisableSSL(!c.UseSSL).
		WithHTTPClient(httpClient)

	if c.UseRegion() {
		awsConfig = awsConfig.WithRegion(c.Region).WithEndpoint(c.S3Endpoint())
	} else {
		awsConfig = awsConfig.WithRegion(config.EmptyRegion).WithEndpoint(c.S3Endpoint())
	}

	if c.CredentialsSource == config.StaticCredentialsSource {
		awsConfig = awsConfig.WithCredentials(credentials.NewStaticCredentials(c.AccessKeyID, c.SecretAccessKey, ""))
	}

	if c.CredentialsSource == config.NoneCredentialsSource {
		awsConfig = awsConfig.WithCredentials(credentials.AnonymousCredentials)
	}

	s3Session, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}

	if c.AssumeRoleArn != "" {
		awsConfig = awsConfig.WithCredentials(stscreds.NewCredentials(s3Session, c.AssumeRoleArn))
	}

	s3Client := s3.New(s3Session, awsConfig)

	if c.UseV2SigningMethod {
		setv2Handlers(s3Client)
	}

	return s3Client, nil
}
