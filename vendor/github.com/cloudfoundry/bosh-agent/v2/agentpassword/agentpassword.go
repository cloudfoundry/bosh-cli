package agentpassword

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
)

const (
	Feature = "http-password-hmac-sha256"

	prefix     = "bosh-hmac-sha256$"
	saltLength = 16
	keyLength  = 32
)

type Verifier struct {
	salt []byte
	key  []byte
}

func IsVerifier(password string) bool {
	return strings.HasPrefix(password, prefix)
}

func ParseVerifier(s string) (*Verifier, error) {
	if !strings.HasPrefix(s, prefix) {
		return nil, errors.New("malformed HTTP password verifier: missing prefix")
	}
	parts := strings.Split(s, "$")
	if len(parts) != 3 {
		return nil, errors.New("malformed HTTP password verifier: invalid part count")
	}
	salt, err := base64.RawURLEncoding.Strict().DecodeString(parts[1])
	if err != nil || len(salt) != saltLength {
		return nil, errors.New("malformed HTTP password verifier: invalid salt")
	}
	key, err := base64.RawURLEncoding.Strict().DecodeString(parts[2])
	if err != nil || len(key) != keyLength {
		return nil, errors.New("malformed HTTP password verifier: invalid key")
	}
	return &Verifier{salt: salt, key: key}, nil
}

func (v *Verifier) Matches(password string) bool {
	if v == nil {
		return false
	}
	mac := hmac.New(sha256.New, v.salt)
	_, _ = mac.Write([]byte(password))
	return subtle.ConstantTimeCompare(v.key, mac.Sum(nil)) == 1
}

func HashURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", errors.New("invalid URL")
	}
	if u.Scheme != "https" {
		return raw, nil
	}
	if u.User == nil {
		return "", errors.New("URL missing userinfo")
	}
	password, hasPassword := u.User.Password()
	if !hasPassword || password == "" {
		return "", errors.New("URL missing password")
	}
	if IsVerifier(password) {
		return "", errors.New("URL password is already a verifier")
	}

	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", errors.New("failed to generate random salt")
	}

	mac := hmac.New(sha256.New, salt)
	_, _ = mac.Write([]byte(password))
	key := mac.Sum(nil)

	encodedSalt := base64.RawURLEncoding.EncodeToString(salt)
	encodedKey := base64.RawURLEncoding.EncodeToString(key)
	verifier := prefix + encodedSalt + "$" + encodedKey

	u.User = url.UserPassword(u.User.Username(), verifier)
	return u.String(), nil
}
