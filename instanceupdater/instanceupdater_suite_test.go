package instanceupdater_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestInstanceupdater(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Instanceupdater Suite")
}
