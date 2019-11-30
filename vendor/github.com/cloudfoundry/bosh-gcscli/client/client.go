/*
 * Copyright 2017 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"golang.org/x/oauth2/google"

	"log"

	"cloud.google.com/go/storage"
	"github.com/cloudfoundry/bosh-gcscli/config"
)

// ErrInvalidROWriteOperation is returned when credentials associated with the
// client disallow an attempted write operation.
var ErrInvalidROWriteOperation = errors.New("the client operates in read only mode. Change 'credentials_source' parameter value ")

// GCSBlobstore encapsulates interaction with the GCS blobstore
type GCSBlobstore struct {
	authenticatedGCS *storage.Client
	publicGCS        *storage.Client
	config           *config.GCSCli
}

// validateRemoteConfig determines if the configuration of the client matches
// against the remote configuration and the StorageClass is valid for the location.
//
// If operating in read-only mode, no mutations can be performed
// so the remote bucket location is always compatible.
func (client *GCSBlobstore) validateRemoteConfig() error {
	if client.readOnly() {
		return nil
	}

	bucket := client.authenticatedGCS.Bucket(client.config.BucketName)
	attrs, err := bucket.Attrs(context.Background())
	if err != nil {
		return err
	}
	return client.config.FitCompatibleLocation(attrs.Location)
}

// getObjectHandle returns a handle to an object named src
func (client *GCSBlobstore) getObjectHandle(gcs *storage.Client, src string) *storage.ObjectHandle {
	handle := gcs.Bucket(client.config.BucketName).Object(src)
	if client.config.EncryptionKey != nil {
		handle = handle.Key(client.config.EncryptionKey)
	}
	return handle
}

// New returns a GCSBlobstore configured to operate using the given config
//
// non-nil error is returned on invalid Client or config. If the configuration
// is incompatible with the GCS bucket, a non-nil error is also returned.
func New(ctx context.Context, cfg *config.GCSCli) (*GCSBlobstore, error) {
	if cfg == nil {
		return nil, errors.New("expected non-nill config object")
	}

	authenticatedGCS, publicGCS, err := newStorageClients(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("creating storage client: %v", err)
	}

	return &GCSBlobstore{authenticatedGCS: authenticatedGCS, publicGCS: publicGCS, config: cfg}, nil
}

// Get fetches a blob from the GCS blobstore.
// Destination will be overwritten if it already exists.
func (client *GCSBlobstore) Get(src string, dest io.Writer) error {
	reader, err := client.getReader(client.publicGCS, src)

	// If the public client fails, try using it as an authenticated actor
	if err != nil && client.authenticatedGCS != nil {
		reader, err = client.getReader(client.authenticatedGCS, src)
	}

	if err != nil {
		return err
	}

	_, err = io.Copy(dest, reader)
	return err
}

func (client *GCSBlobstore) getReader(gcs *storage.Client, src string) (*storage.Reader, error) {
	return client.getObjectHandle(gcs, src).NewReader(context.Background())
}

// Put uploads a blob to the GCS blobstore.
// Destination will be overwritten if it already exists.
//
// Put retries retryAttempts times
const retryAttempts = 3

func (client *GCSBlobstore) Put(src io.ReadSeeker, dest string) error {
	if client.readOnly() {
		return ErrInvalidROWriteOperation
	}

	if err := client.validateRemoteConfig(); err != nil {
		return err
	}

	pos, err := src.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("finding buffer position: %v", err)
	}

	var errs []error
	for i := 0; i < retryAttempts; i++ {
		err := client.putOnce(src, dest)
		if err == nil {
			return nil
		}

		errs = append(errs, err)
		log.Printf("upload failed for %s, attempt %d/%d: %v\n", dest, i+1, retryAttempts, err)

		if _, err := src.Seek(pos, io.SeekStart); err != nil {
			return fmt.Errorf("restting buffer position after failed upload: %v", err)
		}
	}

	return fmt.Errorf("upload failed for %s after %d attempts: %v", dest, retryAttempts, errs)
}

func (client *GCSBlobstore) putOnce(src io.ReadSeeker, dest string) error {
	remoteWriter := client.getObjectHandle(client.authenticatedGCS, dest).NewWriter(context.Background())
	remoteWriter.ObjectAttrs.StorageClass = client.config.StorageClass

	if _, err := io.Copy(remoteWriter, src); err != nil {
		remoteWriter.CloseWithError(err)
		return err
	}

	return remoteWriter.Close()
}

// Delete removes a blob from from the GCS blobstore.
//
// If the object does not exist, Delete returns a nil error.
func (client *GCSBlobstore) Delete(dest string) error {
	if client.readOnly() {
		return ErrInvalidROWriteOperation
	}

	err := client.getObjectHandle(client.authenticatedGCS, dest).Delete(context.Background())
	if err == storage.ErrObjectNotExist {
		return nil
	}
	return err
}

// Exists checks if a blob exists in the GCS blobstore.
func (client *GCSBlobstore) Exists(dest string) (exists bool, err error) {
	if exists, err = client.exists(client.publicGCS, dest); err == nil {
		return exists, nil
	}

	// If the public client fails, try using it as an authenticated actor
	if client.authenticatedGCS != nil {
		return client.exists(client.authenticatedGCS, dest)
	}

	return
}

func (client *GCSBlobstore) exists(gcs *storage.Client, dest string) (bool, error) {
	_, err := client.getObjectHandle(gcs, dest).Attrs(context.Background())
	if err == nil {
		log.Printf("File '%s' exists in bucket '%s'\n", dest, client.config.BucketName)
		return true, nil
	} else if err == storage.ErrObjectNotExist {
		log.Printf("File '%s' does not exist in bucket '%s'\n", dest, client.config.BucketName)
		return false, nil
	}
	return false, err
}

func (client *GCSBlobstore) readOnly() bool {
	return client.authenticatedGCS == nil
}

func (client *GCSBlobstore) Sign(id string, action string, expiry time.Duration) (string, error) {
	token, err := google.JWTConfigFromJSON([]byte(client.config.ServiceAccountFile), storage.ScopeFullControl)
	if err != nil {
		return "", err
	}
	options := storage.SignedURLOptions{
		Method:         action,
		Expires:        time.Now().Add(expiry),
		PrivateKey:     token.PrivateKey,
		GoogleAccessID: token.Email,
		Scheme:         storage.SigningSchemeV4,
	}

	// GET/PUT to the resultant signed url must include, in addition to the below:
	// 'x-goog-encryption-key' and 'x-goog-encryption-key-hash'
	willEncrypt := len(client.config.EncryptionKey) > 0
	if willEncrypt {
		options.Headers = []string{
			"x-goog-encryption-algorithm: AES256",
		}
	}
	return storage.SignedURL(client.config.BucketName, id, &options)
}
