package client

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	"github.com/cloudfoundry/bosh-s3cli/config"
)

var errorInvalidCredentialsSourceValue = errors.New("the client operates in read only mode. Change 'credentials_source' parameter value ")
var oneTB = int64(1000 * 1024 * 1024 * 1024)

// awsS3Client encapsulates AWS S3 blobstore interactions
type awsS3Client struct {
	s3Client    *s3.Client
	s3cliConfig *config.S3Cli
}

// Get fetches a blob, destination will be overwritten if exists
func (b *awsS3Client) Get(src string, dest io.WriterAt) error {
	downloader := manager.NewDownloader(b.s3Client)

	_, err := downloader.Download(context.TODO(), dest, &s3.GetObjectInput{
		Bucket: aws.String(b.s3cliConfig.BucketName),
		Key:    b.key(src),
	})

	if err != nil {
		return err
	}

	return nil
}

// Put uploads a blob
func (b *awsS3Client) Put(src io.ReadSeeker, dest string) error {
	cfg := b.s3cliConfig
	if cfg.CredentialsSource == config.NoneCredentialsSource {
		return errorInvalidCredentialsSourceValue
	}

	uploader := manager.NewUploader(b.s3Client, func(u *manager.Uploader) {
		u.LeavePartsOnError = false

		if !cfg.MultipartUpload {
			// disable multipart uploads by way of large PartSize configuration
			u.PartSize = oneTB
		}

		if cfg.ShouldDisableUploaderRequestChecksumCalculation() {
			// Disable checksum calculation for Alicloud OSS (Object Storage Service)
			// Alicloud doesn't support AWS chunked encoding with checksum calculation
			u.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		}
	})
	uploadInput := &s3.PutObjectInput{
		Body:   src,
		Bucket: aws.String(cfg.BucketName),
		Key:    b.key(dest),
	}
	if cfg.ServerSideEncryption != "" {
		uploadInput.ServerSideEncryption = types.ServerSideEncryption(cfg.ServerSideEncryption)
	}
	if cfg.SSEKMSKeyID != "" {
		uploadInput.SSEKMSKeyId = aws.String(cfg.SSEKMSKeyID)
	}

	retry := 0
	maxRetries := 3
	for {
		putResult, err := uploader.Upload(context.TODO(), uploadInput)
		if err != nil {
			if _, ok := err.(manager.MultiUploadFailure); ok {
				if retry == maxRetries {
					log.Println("Upload retry limit exceeded:", err.Error())
					return fmt.Errorf("upload retry limit exceeded: %s", err.Error())
				}
				retry++
				time.Sleep(time.Second * time.Duration(retry))
				continue
			}
			log.Println("Upload failed:", err.Error())
			return fmt.Errorf("upload failure: %s", err.Error())
		}

		log.Println("Successfully uploaded file to", putResult.Location)
		return nil
	}
}

// Delete removes a blob - no error is returned if the object does not exist
func (b *awsS3Client) Delete(dest string) error {
	if b.s3cliConfig.CredentialsSource == config.NoneCredentialsSource {
		return errorInvalidCredentialsSourceValue
	}

	deleteParams := &s3.DeleteObjectInput{
		Bucket: aws.String(b.s3cliConfig.BucketName),
		Key:    b.key(dest),
	}

	_, err := b.s3Client.DeleteObject(context.TODO(), deleteParams)

	if err == nil {
		return nil
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && (apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "NoSuchKey") {
		return nil
	}
	return err
}

// Exists checks if blob exists
func (b *awsS3Client) Exists(dest string) (bool, error) {
	existsParams := &s3.HeadObjectInput{
		Bucket: aws.String(b.s3cliConfig.BucketName),
		Key:    b.key(dest),
	}

	_, err := b.s3Client.HeadObject(context.TODO(), existsParams)

	if err == nil {
		log.Printf("File '%s' exists in bucket '%s'\n", dest, b.s3cliConfig.BucketName)
		return true, nil
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NotFound" {
		log.Printf("File '%s' does not exist in bucket '%s'\n", dest, b.s3cliConfig.BucketName)
		return false, nil
	}
	return false, err
}

// Sign creates a presigned URL
func (b *awsS3Client) Sign(objectID string, action string, expiration time.Duration) (string, error) {
	action = strings.ToUpper(action)
	switch action {
	case "GET":
		return b.getSigned(objectID, expiration)
	case "PUT":
		return b.putSigned(objectID, expiration)
	default:
		return "", fmt.Errorf("action not implemented: %s", action)
	}
}

func (b *awsS3Client) key(srcOrDest string) *string {
	formattedKey := aws.String(srcOrDest)
	if len(b.s3cliConfig.FolderName) != 0 {
		formattedKey = aws.String(fmt.Sprintf("%s/%s", b.s3cliConfig.FolderName, srcOrDest))
	}

	return formattedKey
}

func (b *awsS3Client) getSigned(objectID string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(b.s3Client)
	signParams := &s3.GetObjectInput{
		Bucket: aws.String(b.s3cliConfig.BucketName),
		Key:    b.key(objectID),
	}

	req, err := presignClient.PresignGetObject(context.TODO(), signParams, s3.WithPresignExpires(expiration))
	if err != nil {
		return "", err
	}

	return req.URL, nil
}

func (b *awsS3Client) putSigned(objectID string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(b.s3Client)
	signParams := &s3.PutObjectInput{
		Bucket: aws.String(b.s3cliConfig.BucketName),
		Key:    b.key(objectID),
	}

	req, err := presignClient.PresignPutObject(context.TODO(), signParams, s3.WithPresignExpires(expiration))
	if err != nil {
		return "", err
	}

	return req.URL, nil
}
