package net_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestProperty(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Common Net Suite")
}
