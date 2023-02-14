package template_test

import (
	"crypto/tls"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/director/template"
	"github.com/cloudfoundry/bosh-cli/v7/testutils"
)

var (
	cert        tls.Certificate
	cacertBytes []byte
	validCACert string
)
var _ = BeforeSuite(func() {
	var err error
	cert, cacertBytes, err = testutils.CertSetup()
	validCACert = string(cacertBytes)
	Expect(err).ToNot(HaveOccurred())
})

func TestReg(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "director/template")
}

type FakeVariables struct {
	GetFunc      func(VariableDefinition) (interface{}, bool, error)
	GetVarDef    VariableDefinition
	GetErr       error
	GetCallCount int
}

func (v *FakeVariables) Get(varDef VariableDefinition) (interface{}, bool, error) {
	v.GetCallCount++
	v.GetVarDef = varDef
	if v.GetFunc != nil {
		return v.GetFunc(varDef)
	}
	return nil, false, v.GetErr
}

func (v *FakeVariables) List() ([]VariableDefinition, error) {
	return nil, nil
}
