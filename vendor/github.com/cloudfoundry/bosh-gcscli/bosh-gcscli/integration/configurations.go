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

package integration

import (
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/cloudfoundry/bosh-gcscli/config"

	. "github.com/onsi/ginkgo/extensions/table"
)

const regionalBucketEnv = "REGIONAL_BUCKET_NAME"
const multiRegionalBucketEnv = "MULTIREGIONAL_BUCKET_NAME"
const publicBucketEnv = "PUBLIC_BUCKET_NAME"

// noBucketMsg is the template used when a BucketEnv's environment variable
// has not been populated.
const noBucketMsg = "environment variable %s expected to contain a valid Google Cloud Storage bucket but was empty"

const getConfigErrMsg = "creating %s configs: %v"

func readBucketEnv(env string) (string, error) {
	bucket := os.Getenv(env)
	if len(bucket) == 0 {
		return "", fmt.Errorf(noBucketMsg, env)
	}
	return bucket, nil
}

func getBaseConfigs() ([]TableEntry, error) {
	var regional, multiRegional string
	var err error
	if regional, err = readBucketEnv(regionalBucketEnv); err != nil {
		return nil, fmt.Errorf(getConfigErrMsg, "base", err)
	}
	if multiRegional, err = readBucketEnv(multiRegionalBucketEnv); err != nil {
		return nil, fmt.Errorf(getConfigErrMsg, "base", err)
	}

	return []TableEntry{
		Entry("MultiRegional bucket, default StorageClass",
			&config.GCSCli{
				BucketName: multiRegional,
			}),
		Entry("Regional bucket, default StorageClass",
			&config.GCSCli{
				BucketName: regional,
			}),
		Entry("MultiRegional bucket, explicit StorageClass",
			&config.GCSCli{
				BucketName:   multiRegional,
				StorageClass: "MULTI_REGIONAL",
			}),
		Entry("Regional bucket, explicit StorageClass",
			&config.GCSCli{
				BucketName:   regional,
				StorageClass: "REGIONAL",
			}),
	}, nil
}

// encryptionKeyBytes are used as the key in tests requiring encryption.
var encryptionKeyBytes = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}

// encryptionKeyBytesHash is the has of the encryptionKeyBytes
//
// Typical usage is ensuring the encryption key is actually used by GCS.
var encryptionKeyBytesHash = sha256.Sum256(encryptionKeyBytes)

func getEncryptedConfigs() ([]TableEntry, error) {
	var regional, multiRegional string
	var err error
	if regional, err = readBucketEnv(regionalBucketEnv); err != nil {
		return nil, fmt.Errorf(getConfigErrMsg, "encrypted", err)
	}
	if multiRegional, err = readBucketEnv(multiRegionalBucketEnv); err != nil {
		return nil, fmt.Errorf(getConfigErrMsg, "encrypted", err)
	}

	return []TableEntry{
		Entry("MultiRegional bucket, default StorageClass, encrypted",
			&config.GCSCli{
				BucketName:    multiRegional,
				EncryptionKey: encryptionKeyBytes,
			}),
		Entry("Regional bucket, default StorageClass, encrypted",
			&config.GCSCli{
				BucketName:    regional,
				EncryptionKey: encryptionKeyBytes,
			}),
	}, nil
}

func getPublicConfigs() ([]TableEntry, error) {
	public, err := readBucketEnv(publicBucketEnv)
	if err != nil {
		return nil, fmt.Errorf(getConfigErrMsg, "public", err)
	}

	return []TableEntry{
		Entry("Public bucket",
			&config.GCSCli{
				BucketName: public,
			}),
	}, nil
}

func getStorageCompatConfigs() ([]TableEntry, error) {
	var regional, multiRegional string
	var err error
	if regional, err = readBucketEnv(regionalBucketEnv); err != nil {
		return nil, fmt.Errorf(getConfigErrMsg, "storage compat", err)
	}
	if multiRegional, err = readBucketEnv(multiRegionalBucketEnv); err != nil {
		return nil, fmt.Errorf(getConfigErrMsg, "storage compat", err)
	}

	return []TableEntry{
		Entry("MultiRegional bucket, regional StorageClass", &config.GCSCli{
			BucketName:   multiRegional,
			StorageClass: "REGIONAL",
		}),
		Entry("Regional bucket, multiregional StorageClass", &config.GCSCli{
			BucketName:   regional,
			StorageClass: "MULTI_REGIONAL",
		}),
	}, nil
}
