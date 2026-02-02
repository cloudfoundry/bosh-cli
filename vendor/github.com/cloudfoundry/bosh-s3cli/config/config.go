package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// The S3Cli represents configuration for the s3cli
type S3Cli struct {
	AccessKeyID                               string `json:"access_key_id"`
	SecretAccessKey                           string `json:"secret_access_key"`
	BucketName                                string `json:"bucket_name"`
	FolderName                                string `json:"folder_name"`
	CredentialsSource                         string `json:"credentials_source"`
	Host                                      string `json:"host"`
	Port                                      int    `json:"port"` // 0 means no custom port
	Region                                    string `json:"region"`
	SSLVerifyPeer                             bool   `json:"ssl_verify_peer"`
	UseSSL                                    bool   `json:"use_ssl"`
	ServerSideEncryption                      string `json:"server_side_encryption"`
	SSEKMSKeyID                               string `json:"sse_kms_key_id"`
	AssumeRoleArn                             string `json:"assume_role_arn"`
	MultipartUpload                           bool   `json:"multipart_upload"`
	HostStyle                                 bool   `json:"host_style"`
	SwiftAuthAccount                          string `json:"swift_auth_account"`
	SwiftTempURLKey                           string `json:"swift_temp_url_key"`
	requestChecksumCalculationEnabled         bool
	uploaderRequestChecksumCalculationEnabled bool
}

const defaultAWSRegion = "us-east-1" //nolint:unused

// StaticCredentialsSource specifies that credentials will be supplied using access_key_id and secret_access_key
const StaticCredentialsSource = "static"

// NoneCredentialsSource specifies that credentials will be empty. The blobstore client operates in read only mode.
const NoneCredentialsSource = "none"

const credentialsSourceEnvOrProfile = "env_or_profile"

// Nothing was provided in configuration
const noCredentialsSourceProvided = ""

var errorStaticCredentialsMissing = errors.New("access_key_id and secret_access_key must be provided")

type errorStaticCredentialsPresent struct {
	credentialsSource string
}

func (e errorStaticCredentialsPresent) Error() string {
	return fmt.Sprintf("can't use access_key_id and secret_access_key with %s credentials_source", e.credentialsSource)
}

func newStaticCredentialsPresentError(desiredSource string) error {
	return errorStaticCredentialsPresent{credentialsSource: desiredSource}
}

// NewFromReader returns a new s3cli configuration struct from the contents of reader.
// reader.Read() is expected to return valid JSON
func NewFromReader(reader io.Reader) (S3Cli, error) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return S3Cli{}, err
	}

	c := S3Cli{
		SSLVerifyPeer:                     true,
		UseSSL:                            true,
		MultipartUpload:                   true,
		requestChecksumCalculationEnabled: true,
		uploaderRequestChecksumCalculationEnabled: true,
	}

	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return S3Cli{}, err
	}

	if c.BucketName == "" {
		return S3Cli{}, errors.New("bucket_name must be set")
	}

	switch c.CredentialsSource {
	case StaticCredentialsSource:
		if c.AccessKeyID == "" || c.SecretAccessKey == "" {
			return S3Cli{}, errorStaticCredentialsMissing
		}
	case credentialsSourceEnvOrProfile:
		if c.AccessKeyID != "" || c.SecretAccessKey != "" {
			return S3Cli{}, newStaticCredentialsPresentError(credentialsSourceEnvOrProfile)
		}
	case NoneCredentialsSource:
		if c.AccessKeyID != "" || c.SecretAccessKey != "" {
			return S3Cli{}, newStaticCredentialsPresentError(NoneCredentialsSource)
		}

	case noCredentialsSourceProvided:
		if c.SecretAccessKey != "" && c.AccessKeyID != "" {
			c.CredentialsSource = StaticCredentialsSource
		} else if c.SecretAccessKey == "" && c.AccessKeyID == "" {
			c.CredentialsSource = NoneCredentialsSource
		} else {
			return S3Cli{}, errorStaticCredentialsMissing
		}
	default:
		return S3Cli{}, fmt.Errorf("invalid credentials_source: %s", c.CredentialsSource)
	}

	switch Provider(c.Host) {
	case "aws":
		c.configureAWS()
	case "alicloud":
		c.configureAlicloud()
	case "google":
		c.configureGoogle()
	default:
		c.configureDefault()
	}

	return c, nil
}

// Provider returns one of (aws, alicloud, google) based on a host name.
// Returns "" if a known provider cannot be detected.
func Provider(host string) string {
	for provider, regex := range providerRegex {
		if regex.MatchString(host) {
			return provider
		}
	}

	return ""
}

func (c *S3Cli) configureAWS() {
	c.MultipartUpload = true

	if c.Region == "" {
		if region := AWSHostToRegion(c.Host); region != "" {
			c.Region = region
		} else {
			c.Region = defaultAWSRegion
		}
	}
}

func (c *S3Cli) configureAlicloud() {
	c.MultipartUpload = true
	c.HostStyle = true

	c.Host = strings.Split(c.Host, ":")[0]
	if c.Region == "" {
		c.Region = AlicloudHostToRegion(c.Host)
	}
	c.requestChecksumCalculationEnabled = false
	c.uploaderRequestChecksumCalculationEnabled = false
}

func (c *S3Cli) configureGoogle() {
	c.MultipartUpload = false
	c.requestChecksumCalculationEnabled = false
}

func (c *S3Cli) configureDefault() {
	// No specific configuration needed for default/unknown providers
}

// S3Endpoint returns the S3 URI to use if custom host information has been provided
func (c *S3Cli) S3Endpoint() string {
	if c.Host == "" {
		return ""
	}
	if c.Port == 80 && !c.UseSSL {
		return c.Host
	}
	if c.Port == 443 && c.UseSSL {
		return c.Host
	}
	if c.Port != 0 {
		return fmt.Sprintf("%s:%d", c.Host, c.Port)
	}
	return c.Host
}

func (c *S3Cli) IsGoogle() bool {
	return Provider(c.Host) == "google"
}

func (c *S3Cli) ShouldDisableRequestChecksumCalculation() bool {
	return !c.requestChecksumCalculationEnabled
}

func (c *S3Cli) ShouldDisableUploaderRequestChecksumCalculation() bool {
	return !c.uploaderRequestChecksumCalculationEnabled
}
