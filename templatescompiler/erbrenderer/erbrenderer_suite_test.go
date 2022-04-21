package erbrenderer_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestErbrenderer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Erbrenderer Suite")
}
