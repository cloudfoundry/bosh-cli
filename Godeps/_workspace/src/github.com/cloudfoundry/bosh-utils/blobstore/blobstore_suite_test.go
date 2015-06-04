package blobstore_test

import (
	. "github.com/cloudfoundry/bosh-utils/internal/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/bosh-utils/internal/github.com/onsi/gomega"

	"testing"
)

func TestBlobstore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Blobstore Suite")
}
