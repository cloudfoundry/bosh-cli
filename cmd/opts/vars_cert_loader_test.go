package opts_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
)

var _ = Describe("VarsCertLoader", func() {
	var (
		vars   boshtpl.StaticVariables
		loader VarsCertLoader
	)

	BeforeEach(func() {
		vars = boshtpl.StaticVariables{}
		loader = NewVarsCertLoader(vars)
	})

	Describe("LoadCerts", func() {
		It("returns error if getting variable failed", func() {
			loader = NewVarsCertLoader(&FakeVariables{GetErr: errors.New("fake-err")})

			_, _, err := loader.LoadCerts("unknown")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if variable by that name is not found", func() {
			_, _, err := loader.LoadCerts("unknown")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected to find variable 'unknown' with a certificate"))
		})

		It("returns error if variable cannot be parsed", func() {
			vars["cert"] = 123

			_, _, err := loader.LoadCerts("cert")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected variable 'cert' to be deserializable"))
		})

		const cert = "-----BEGIN CERTIFICATE-----\nMIIDtzCCAp+gAwIBAgIJAMZ/qRdRamluMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV\nBAYTAkFVMRMwEQYDVQQIEwpTb21lLVN0YXRlMSEwHwYDVQQKExhJbnRlcm5ldCBX\naWRnaXRzIFB0eSBMdGQwIBcNMTYwODI2MjIzMzE5WhgPMjI5MDA2MTAyMjMzMTla\nMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIEwpTb21lLVN0YXRlMSEwHwYDVQQKExhJ\nbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw\nggEKAoIBAQDN/bv70wDn6APMqiJZV7ESZhUyGu8OzuaeEfb+64SNvQIIME0s9+i7\nD9gKAZjtoC2Tr9bJBqsKdVhREd/X6ePTaopxL8shC9GxXmTqJ1+vKT6UxN4kHr3U\n+Y+LK2SGYUAvE44nv7sBbiLxDl580P00ouYTf6RJgW6gOuKpIGcvsTGA4+u0UTc+\ny4pj6sT0+e3xj//Y4wbLdeJ6cfcNTU63jiHpKc9Rgo4Tcy97WeEryXWz93rtRh8d\npvQKHVDU/26EkNsPSsn9AHNgaa+iOA2glZ2EzZ8xoaMPrHgQhcxoi8maFzfM2dX2\nXB1BOswa/46yqfzc4xAwaW0MLZLg3NffAgMBAAGjgacwgaQwHQYDVR0OBBYEFNRJ\nPYFebixALIR2Ee+yFoSqurxqMHUGA1UdIwRuMGyAFNRJPYFebixALIR2Ee+yFoSq\nurxqoUmkRzBFMQswCQYDVQQGEwJBVTETMBEGA1UECBMKU29tZS1TdGF0ZTEhMB8G\nA1UEChMYSW50ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkggkAxn+pF1FqaW4wDAYDVR0T\nBAUwAwEB/zANBgkqhkiG9w0BAQUFAAOCAQEAoPTwU2rm0ca5b8xMni3vpjYmB9NW\noSpGcWENbvu/p7NpiPAe143c5EPCuEHue/AbHWWxBzNAZvhVZBeFirYNB3HYnCla\njP4WI3o2Q0MpGy3kMYigEYG76WeZAM5ovl0qDP6fKuikZofeiygb8lPs7Hv4/88x\npSsZYBm7UPTS3Pl044oZfRJdqTpyHVPDqwiYD5KQcI0yHUE9v5KC0CnqOrU/83PE\nb0lpHA8bE9gQTQjmIa8MIpaP3UNTxvmKfEQnk5UAZ5xY2at5mmyj3t8woGdzoL98\nyDd2GtrGsguQXM2op+4LqEdHef57g7vwolZejJqN776Xu/lZtCTp01+HTA==\n-----END CERTIFICATE-----\n"

		It("returns error if pem encoded certificate cannot be found", func() {
			vars["cert"] = map[interface{}]interface{}{
				"certificate": "not-cert",
			}

			_, _, err := loader.LoadCerts("cert")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Certificate did not contain PEM formatted block"))
		})

		It("returns error if certificate cannot be parsed", func() {
			vars["cert"] = map[interface{}]interface{}{
				"certificate": "-----BEGIN CERTIFICATE-----\nMIIDtzCCAp+gAwIBAgIJAMZ/qRdR\n-----END CERTIFICATE-----\n",
			}

			_, _, err := loader.LoadCerts("cert")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Parsing certificate: x509: malformed certificate"))
		})

		It("returns error if pem encoded private key cannot be found", func() {
			vars["cert"] = map[interface{}]interface{}{
				"certificate": cert,
				"private_key": "not-priv-key",
			}

			_, _, err := loader.LoadCerts("cert")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Private key did not contain PEM formatted block"))
		})

		It("returns error if private key cannot be parsed", func() {
			vars["cert"] = map[interface{}]interface{}{
				"certificate": cert,
				"private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEAzf27+9MA5+gDzKoiWVex\n-----END RSA PRIVATE KEY-----\n",
			}

			_, _, err := loader.LoadCerts("cert")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Parsing private key: asn1: syntax error: data truncated"))
		})
	})
})

type FakeVariables struct {
	GetErr error
}

func (v *FakeVariables) Get(_ boshtpl.VariableDefinition) (interface{}, bool, error) {
	return nil, false, v.GetErr
}

func (v *FakeVariables) List() ([]boshtpl.VariableDefinition, error) {
	return nil, nil
}
