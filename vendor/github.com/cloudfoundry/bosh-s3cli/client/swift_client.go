package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/bosh-s3cli/config"
)

type SwiftBlobstore struct {
	s3cliConfig *config.S3Cli
}

func NewSwiftClient(s3cliConfig *config.S3Cli) SwiftBlobstore {
	return SwiftBlobstore{s3cliConfig: s3cliConfig}
}

func (client *SwiftBlobstore) Sign(objectID string, action string, expiration time.Duration) (string, error) {
	action = strings.ToUpper(action)
	switch action {
	case "GET", "PUT":
		return client.SignedURL(action, objectID, expiration)
	default:
		return "", fmt.Errorf("action not implemented: %s", action)
	}
}

func (client *SwiftBlobstore) SignedURL(action string, objectID string, expiration time.Duration) (string, error) {
	path := fmt.Sprintf("/v1/%s/%s/%s", client.s3cliConfig.SwiftAuthAccount, client.s3cliConfig.BucketName, objectID)

	expires := time.Now().Add(expiration).Unix()
	hmacBody := action + "\n" + strconv.FormatInt(expires, 10) + "\n" + path

	h := hmac.New(sha256.New, []byte(client.s3cliConfig.SwiftTempURLKey))
	h.Write([]byte(hmacBody))
	signature := hex.EncodeToString(h.Sum(nil))

	url := fmt.Sprintf("https://%s%s?temp_url_sig=%s&temp_url_expires=%d\n", client.s3cliConfig.Host, path, signature, expires)

	return url, nil
}
