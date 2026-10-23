package crypto

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"strings"

	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . DigestCalculator

type DigestCalculator interface {
	Calculate(string) (string, error)
	CalculateVerified(io.Reader, boshcrypto.Digest) (string, error)
	CalculateString(string) string
}

type digestCalculator struct {
	fs         boshsys.FileSystem
	algorithms []boshcrypto.Algorithm
}

func NewDigestCalculator(fs boshsys.FileSystem, algos []boshcrypto.Algorithm) DigestCalculator {
	return digestCalculator{
		fs:         fs,
		algorithms: algos,
	}
}

func (c digestCalculator) Calculate(filePath string) (string, error) {
	digest, err := boshcrypto.NewMultipleDigestFromPath(filePath, c.fs, c.algorithms)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Calculating digest for %s", filePath)
	}
	return digest.String(), nil
}

func (c digestCalculator) CalculateString(data string) string {
	digest, err := boshcrypto.NewMultipleDigest(strings.NewReader(data), c.algorithms)
	if err != nil {
		panic("According to the docs sha1.Write will never return an error")
	}
	return digest.String()
}

// CalculateVerified computes replacement digests over exactly the bytes checked
// against the prior digest, without reopening the archive path.
func (c digestCalculator) CalculateVerified(reader io.Reader, expected boshcrypto.Digest) (string, error) {
	if len(c.algorithms) == 0 {
		return "", bosherr.Error("must provide at least one algorithm")
	}
	hashes := make([]hash.Hash, 0, len(c.algorithms))
	writers := make([]io.Writer, 0, len(c.algorithms))
	for _, algorithm := range c.algorithms {
		var h hash.Hash
		switch algorithm.Name() {
		case "sha1":
			h = sha1.New()
		case "sha256":
			h = sha256.New()
		case "sha512":
			h = sha512.New()
		default:
			return "", bosherr.Errorf("unsupported digest algorithm %s", algorithm.Name())
		}
		hashes = append(hashes, h)
		writers = append(writers, h)
	}
	if err := expected.Verify(io.TeeReader(reader, io.MultiWriter(writers...))); err != nil {
		return "", err
	}
	digests := make([]boshcrypto.Digest, 0, len(hashes))
	for i, h := range hashes {
		digests = append(digests, boshcrypto.NewDigest(c.algorithms[i], fmt.Sprintf("%x", h.Sum(nil))))
	}
	return boshcrypto.MustNewMultipleDigest(digests...).String(), nil
}
