package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type Signer interface {
	GenerateSignature(objectID, verb string, timeStamp, expires time.Time) string
	GenerateSignedURL(path, objectID, verb string, timeStamp time.Time, expiration time.Duration) (string, error)
}

type signer struct {
	secret string
}

func NewSigner(secret string) Signer {
	return &signer{
		secret: secret,
	}
}

func (s *signer) GenerateSignature(objectID, verb string, timeStamp, expires time.Time) string {
	verb = strings.ToUpper(verb)
	signature := fmt.Sprintf("%s\n%s\n%d\n%d", verb, objectID, timeStamp.Unix(), expires.Unix())
	hmac := hmac.New(sha256.New, []byte(s.secret))
	hmac.Write([]byte(signature))
	return hex.EncodeToString(hmac.Sum(nil))
}

func (s *signer) GenerateSignedURL(blobstorePath, objectID, verb string, timeStamp time.Time, expiration time.Duration) (string, error) {
	verb = strings.ToUpper(verb)
	if verb != "GET" && verb != "PUT" {
		return "", fmt.Errorf("action not implemented: %s. Available actions are 'GET' and 'PUT'", verb)
	}

	blobstorePath = strings.TrimSuffix(blobstorePath, "/")
	invalidAfter := timeStamp.Add(expiration)
	signature := s.GenerateSignature(objectID, verb, timeStamp, invalidAfter)

	return fmt.Sprintf("%s/%s?st=%s&ts=%d&e=%d", blobstorePath, objectID, signature, timeStamp.Unix(), invalidAfter.Unix()), nil
}
