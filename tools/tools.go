//go:build tools
// +build tools

package tools

import (
	_ "github.com/golang/mock/mockgen/model"     // comment to make golint happy
	_ "github.com/maxbrunsfeld/counterfeiter/v6" // comment to make golint happy
	_ "github.com/onsi/ginkgo/ginkgo"            // comment to make golint happy
)

// This file imports packages that are used when running go generate, or used
// during the development process but not otherwise depended on by built code.
