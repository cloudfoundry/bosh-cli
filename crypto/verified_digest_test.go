package crypto_test

import (
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"

	crypto "github.com/cloudfoundry/bosh-cli/v7/crypto"
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
)

func TestVerifiedDigestUsesOneStream(t *testing.T) {
	original := "the verified archive"
	calculator := crypto.NewDigestCalculator(nil, []boshcrypto.Algorithm{boshcrypto.DigestAlgorithmSHA1, boshcrypto.DigestAlgorithmSHA256})
	expected := boshcrypto.NewDigest(boshcrypto.DigestAlgorithmSHA1, fmt.Sprintf("%x", sha1.Sum([]byte(original))))
	actual, err := calculator.CalculateVerified(strings.NewReader(original), expected)
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("%x;sha256:%x", sha1.Sum([]byte(original)), sha256.Sum256([]byte(original)))
	if actual != want {
		t.Fatalf("digest=%s want=%s", actual, want)
	}
	if _, err := calculator.CalculateVerified(strings.NewReader("changed archive"), expected); err == nil {
		t.Fatal("changed bytes accepted")
	}
}
