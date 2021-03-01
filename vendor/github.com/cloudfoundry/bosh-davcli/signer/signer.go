package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type Signer interface {
	GenerateSignedURL(endpoint, prefixedBlobID, verb string, timeStamp time.Time, expiresAfter time.Duration) (string, error)
}

type signer struct {
	secret string
}

func NewSigner(secret string) Signer {
	return &signer{
		secret: secret,
	}
}

func (s *signer) generateSignature(prefixedBlobID, verb string, timeStamp time.Time, expires int) string {
	verb = strings.ToUpper(verb)
	signature := fmt.Sprintf("%s%s%d%d", verb, prefixedBlobID, timeStamp.Unix(), expires)
	hmac := hmac.New(sha256.New, []byte(s.secret))
	hmac.Write([]byte(signature))
	sigBytes := hmac.Sum(nil)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(sigBytes)
}

func (s *signer) GenerateSignedURL(endpoint, prefixedBlobID, verb string, timeStamp time.Time, expiresAfter time.Duration) (string, error) {
	verb = strings.ToUpper(verb)
	if verb != "GET" && verb != "PUT" {
		return "", fmt.Errorf("action not implemented: %s. Available actions are 'GET' and 'PUT'", verb)
	}

	endpoint = strings.TrimSuffix(endpoint, "/")
	expiresAfterSeconds := int(expiresAfter.Seconds())
	signature := s.generateSignature(prefixedBlobID, verb, timeStamp, expiresAfterSeconds)

	blobURL, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	blobURL.Path = path.Join(blobURL.Path, "signed", prefixedBlobID)
	req, err := http.NewRequest(verb, blobURL.String(), nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	q.Add("st", signature)
	q.Add("ts", fmt.Sprintf("%d", timeStamp.Unix()))
	q.Add("e", fmt.Sprintf("%d", expiresAfterSeconds))
	req.URL.RawQuery = q.Encode()
	return req.URL.String(), nil
}
