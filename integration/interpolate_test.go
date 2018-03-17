package integration_test

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"path/filepath"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"gopkg.in/yaml.v2"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
)

var _ = Describe("interpolate command", func() {
	var (
		ui         *fakeui.FakeUI
		fs         *fakesys.FakeFileSystem
		cmdFactory Factory
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		logger := boshlog.NewLogger(boshlog.LevelNone)
		confUI := boshui.NewWrappingConfUI(ui, logger)

		fs = fakesys.NewFakeFileSystem()
		cmdFactory = NewFactory(NewBasicDepsWithFS(confUI, fs, logger))
	})

	It("interpolates manifest with variables", func() {
		err := fs.WriteFileString(filepath.Join("/", "file"), "file: ((key))")
		Expect(err).ToNot(HaveOccurred())

		cmd, err := cmdFactory.New([]string{"interpolate", filepath.Join("/", "file"), "-v", "key=val"})
		Expect(err).ToNot(HaveOccurred())

		err = cmd.Execute()
		Expect(err).ToNot(HaveOccurred())
		Expect(ui.Blocks).To(Equal([]string{"file: val\n"}))
	})

	It("interpolates manifest with variables provided piece by piece via dot notation", func() {
		err := fs.WriteFileString(filepath.Join("/", "template"), "file: ((key))\nfile2: ((key.subkey2))\n")
		Expect(err).ToNot(HaveOccurred())

		err = fs.WriteFileString(filepath.Join("/", "file-val"), "file-val-content")
		Expect(err).ToNot(HaveOccurred())

		cmd, err := cmdFactory.New([]string{
			"interpolate", filepath.Join("/", "template"),
			"-v", "key.subkey=val",
			"-v", "key.subkey2=val2",
			"--var-file", "key.subkey3=" + filepath.Join("/", "file-val"),
		})
		Expect(err).ToNot(HaveOccurred())

		err = cmd.Execute()
		Expect(err).ToNot(HaveOccurred())
		Expect(ui.Blocks).To(Equal([]string{"file:\n  subkey: val\n  subkey2: val2\n  subkey3: file-val-content\nfile2: val2\n"}))
	})

	It("returns portion of the template when --path flag is provided", func() {
		err := fs.WriteFileString(filepath.Join("/", "file"), "file: ((key))")
		Expect(err).ToNot(HaveOccurred())

		cmd, err := cmdFactory.New([]string{"interpolate", filepath.Join("/", "file"), "-v", `key={"nested": true}`, "--path", "/file/nested"})
		Expect(err).ToNot(HaveOccurred())

		err = cmd.Execute()
		Expect(err).ToNot(HaveOccurred())
		Expect(ui.Blocks).To(Equal([]string{"true\n"}))
	})

	It("generates and stores missing password variable when --vars-store is provided", func() {
		err := fs.WriteFileString(filepath.Join("/", "file"), `password: ((key))
variables:
- name: key
  type: password
`)
		Expect(err).ToNot(HaveOccurred())

		var genedPass string

		{ // running command first time
			cmd, err := cmdFactory.New([]string{"interpolate", filepath.Join("/", "file"), "--vars-store", filepath.Join("/", "vars"), "--path", "/password"})
			Expect(err).ToNot(HaveOccurred())

			err = cmd.Execute()
			Expect(err).ToNot(HaveOccurred())
			Expect(ui.Blocks).To(HaveLen(1))

			genedPass = ui.Blocks[0]
			Expect(len(genedPass)).To(BeNumerically(">", 10))

			contents, err := fs.ReadFileString(filepath.Join("/", "vars"))
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(Equal("key: " + genedPass))
		}

		ui.Blocks = []string{}

		{ // running command second time
			cmd, err := cmdFactory.New([]string{"interpolate", filepath.Join("/", "file"), "--vars-store", filepath.Join("/", "vars"), "--path", "/password"})
			Expect(err).ToNot(HaveOccurred())

			err = cmd.Execute()
			Expect(err).ToNot(HaveOccurred())
			Expect(ui.Blocks[0]).To(Equal(genedPass))
		}
	})

	It("generates and stores missing certificate variable when --vars-store is provided", func() {
		err := fs.WriteFileString("/file", `
ca:
  certificate: ((ca.certificate))
server:
  certificate: ((server.certificate))

variables:
- name: ca
  type: certificate
  options:
    is_ca: true
    common_name: ca
- name: server
  type: certificate
  options:
    ca: ca
    common_name: ((common_name))
`)
		Expect(err).ToNot(HaveOccurred())

		cmd, err := cmdFactory.New([]string{"interpolate", filepath.Join("/", "file"), "--vars-store", filepath.Join("/", "vars"), "-v", "common_name=test.com"})
		Expect(err).ToNot(HaveOccurred())

		err = cmd.Execute()
		Expect(err).ToNot(HaveOccurred())
		Expect(ui.Blocks).To(HaveLen(1))

		type expectedCert struct {
			Certificate string
		}

		type expectedStore struct {
			CA     expectedCert
			Server expectedCert
		}

		var store, output expectedStore

		{
			contents, err := fs.ReadFileString(filepath.Join("/", "vars"))
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).ToNot(BeEmpty())

			err = yaml.Unmarshal([]byte(contents), &store)
			Expect(err).ToNot(HaveOccurred())

			err = yaml.Unmarshal([]byte(ui.Blocks[0]), &output)
			Expect(err).ToNot(HaveOccurred())

			Expect(output.CA.Certificate).To(Equal(store.CA.Certificate))
			Expect(output.Server.Certificate).To(Equal(store.Server.Certificate))
		}

		{
			roots := x509.NewCertPool()

			ok := roots.AppendCertsFromPEM([]byte(store.CA.Certificate))
			Expect(ok).To(BeTrue())

			block, _ := pem.Decode([]byte(store.Server.Certificate))
			Expect(block).ToNot(BeNil())

			cert, err := x509.ParseCertificate(block.Bytes)
			Expect(err).ToNot(HaveOccurred())

			_, err = cert.Verify(x509.VerifyOptions{DNSName: "test.com", Roots: roots})
			Expect(err).ToNot(HaveOccurred())

			_, err = cert.Verify(x509.VerifyOptions{DNSName: "not-test.com", Roots: roots})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("certificate is valid"))
		}
	})

	It("returns errors if there are missing variables and --var-errs is provided", func() {
		err := fs.WriteFileString("/file", `
ca: ((ca2.certificate))
used_key: ((missing_key))

variables:
- name: ca
  type: certificate
  options:
    is_ca: true
    common_name: ca
- name: server
  type: certificate
  options:
    ca: ca
    common_name: ((common_name))
`)
		Expect(err).ToNot(HaveOccurred())

		err = fs.WriteFileString(filepath.Join("/", "ro-vars"), "used_key: true\nunused_file: true")
		Expect(err).ToNot(HaveOccurred())

		cmd, err := cmdFactory.New([]string{
			"interpolate", filepath.Join("/", "file"),
			"-v", "used_key=val",
			"--vars-store", filepath.Join("/", "vars"),
			"--var-errs",
		})
		Expect(err).ToNot(HaveOccurred())

		err = cmd.Execute()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Expected to find variables: ca2\ncommon_name\nmissing_key"))
	})

	It("returns errors if there are unused variables and --var-errs-unused is provided", func() {
		err := fs.WriteFileString(filepath.Join("/", "file"), `
ca: ((ca.certificate))
used_key: ((used_key))

variables:
- name: ca
  type: certificate
  options:
    is_ca: true
    common_name: ca
- name: server
  type: certificate
  options:
    ca: ca
    common_name: ((common_name))
`)
		Expect(err).ToNot(HaveOccurred())

		err = fs.WriteFileString(filepath.Join("/", "ro-vars"), "used_key: true\nunused_file: true")
		Expect(err).ToNot(HaveOccurred())

		cmd, err := cmdFactory.New([]string{
			"interpolate", filepath.Join("/", "file"),
			"-v", "common_name=name",
			"-v", "used_key=val",
			"-v", "unused_flag=val",
			"-l", filepath.Join("/", "ro-vars"),
			"--vars-store", filepath.Join("/", "vars"),
			"--var-errs-unused",
		})
		Expect(err).ToNot(HaveOccurred())

		err = cmd.Execute()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Expected to use variables: unused_file\nunused_flag"))
	})

	It("allows to use config server as vars store when --vars-store is configured with config-server:// schema", func() {
		err := fs.WriteFileString(filepath.Join("/", "file"), `
# ensure that vars are interpolated inside an array to guarantee particular order
interpolation_order_guarantee:
- ca: ((ca.certificate))
- missing_key: ((missing_key))

variables:
- name: ca
  type: certificate
  options:
    is_ca: true
    common_name: ca
- name: server
  type: certificate
  options:
    ca: ca
    common_name: ((common_name))
`)
		Expect(err).ToNot(HaveOccurred())

		caCert, configServer := BuildHTTPSServer()
		defer configServer.Close()

		configServer.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v1/data", "name=config-server-ns/ca"),
				ghttp.RespondWith(http.StatusNotFound, ``),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/api/v1/data"),
				VerifyBodyContains(`"name":"config-server-ns/ca"`),
				ghttp.RespondWith(http.StatusOK, `{"value":{"ca": "ca", "certificate": "ca-cert", "private_key": "ca-key"}}`),
			),

			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v1/data", "name=config-server-ns/server"),
				ghttp.RespondWith(http.StatusNotFound, ``),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/api/v1/data"),
				VerifyBodyContains(`"name":"config-server-ns/server"`),
				ghttp.RespondWith(http.StatusOK, `{"value":{"ca": "server-ca", "certificate": "server-cert", "private_key": "server-key"}}`),
			),

			// After generating all variables, follow manifest order

			// read for ((ca.certificate))
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v1/data", "name=config-server-ns/ca"),
				ghttp.RespondWith(http.StatusOK, `{"data": [{"value":{"ca": "ca", "certificate": "ca-cert", "private_key": "ca-key"}}]}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v1/data", "name=config-server-ns/ca"),
				ghttp.RespondWith(http.StatusOK, `{"data": [{"value":{"ca": "ca", "certificate": "ca-cert", "private_key": "ca-key"}}]}`),
			),

			// read for ((missing_key))
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v1/data", "name=config-server-ns/missing_key"),
				ghttp.RespondWith(http.StatusNotFound, ``),
			),

			// Re-read all variables

			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v1/data", "name=config-server-ns/ca"),
				ghttp.RespondWith(http.StatusOK, `{"data": [{"value":{"ca": "ca", "certificate": "ca-cert", "private_key": "ca-key"}}]}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v1/data", "name=config-server-ns/ca"),
				ghttp.RespondWith(http.StatusOK, `{"data": [{"value":{"ca": "ca", "certificate": "ca-cert", "private_key": "ca-key"}}]}`),
			),

			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v1/data", "name=config-server-ns/server"),
				ghttp.RespondWith(http.StatusOK, `{"data": [{"value":{"ca": "server-ca", "certificate": "server-ca-cert", "private_key": "server-ca-key"}}]}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v1/data", "name=config-server-ns/server"),
				ghttp.RespondWith(http.StatusOK, `{"data": [{"value":{"ca": "server-ca", "certificate": "server-ca-cert", "private_key": "server-ca-key"}}]}`),
			),
		)

		cmd, err := cmdFactory.New([]string{
			"interpolate", filepath.Join("/", "file"),
			"-v", "common_name=name",
			"--vars-store", "config-server://",
			"--config-server-url", configServer.URL(),
			"--config-server-tls-ca", caCert,
			// not correct key pair since test server does not verify
			"--config-server-tls-certificate", "-----BEGIN CERTIFICATE-----\nMIIDSjCCAjKgAwIBAgIQWyNlE989qt4Jq5e3uPnEvzANBgkqhkiG9w0BAQsFADAz\nMQwwCgYDVQQGEwNVU0ExFjAUBgNVBAoTDUNsb3VkIEZvdW5kcnkxCzAJBgNVBAMT\nAmNhMB4XDTE4MDEzMTAyMDQwNloXDTE5MDEzMTAyMDQwNlowPDEMMAoGA1UEBhMD\nVVNBMRYwFAYDVQQKEw1DbG91ZCBGb3VuZHJ5MRQwEgYDVQQDEwtzZXJ2ZXItY2Vy\ndDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALZCZGt0DlFNSiRCZU83\nefftOa+EaXmkrMMr1chlb4rdSE5ft3G6Jw9cmdMDB4ZrA3pbQo6ENcFQn1AI6h/2\nC8q2XASUzzO7vDo/oYfsK4YGXXCgVEKB+aQagKDBhPCFW2m6SaoaHjxfZtS4TVBL\nepr7SwZfFBxFKLlInTOdy+i/TARFNf+xszrffVlhRTbTozCwI0wrURQeXPU0V5kh\nM4vOyDrN0/K9GmgO9dNC5X7T1JS2BQ7vtceAf7eSCyiuP3Zi8Gja9/51l3JKAXA7\nceY0+r0U+c4yd2XhaY0gHPYzJegzFilF+bC311heMJpQYm5KdAmeHMOtWW1m2zeU\nxLECAwEAAaNRME8wDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMB\nMAwGA1UdEwEB/wQCMAAwGgYDVR0RBBMwEYIJbG9jYWxob3N0hwR/AAABMA0GCSqG\nSIb3DQEBCwUAA4IBAQANsC2yzCvdQFW0Lr7iVNr44O4J3HLVdMuMGoNoYBGGk7+k\njiExDkY1BNvXtYxGTQo8x0a9i/DzrT1qTsxYQbmSUa35vh08bMhht6aRr4G5LkEq\nneHgJF3oo0G7sEu06dokKZLWkpHtHU9s2csXrLWYCIcejyWhkJGlKTN6WYgum3dS\nWjMSq20Hn+SW/PcUTBpB+gJDKnz2XCX9Hu0elVJOIv+DsRBNlrU0LUyF79Dbnl6j\n/nW3vZF9r7+LfYgLdLB661hvyv5F2e97ic7mmekG7nRop5j074vNPCj8yMdADqVn\nue3UUC2R6+gWoOxY0HVhVlGX9vVJap+/1O7rHBLr\n-----END CERTIFICATE-----",
			"--config-server-tls-private-key", "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAtkJka3QOUU1KJEJlTzd59+05r4RpeaSswyvVyGVvit1ITl+3\ncbonD1yZ0wMHhmsDeltCjoQ1wVCfUAjqH/YLyrZcBJTPM7u8Oj+hh+wrhgZdcKBU\nQoH5pBqAoMGE8IVbabpJqhoePF9m1LhNUEt6mvtLBl8UHEUouUidM53L6L9MBEU1\n/7GzOt99WWFFNtOjMLAjTCtRFB5c9TRXmSEzi87IOs3T8r0aaA7100LlftPUlLYF\nDu+1x4B/t5ILKK4/dmLwaNr3/nWXckoBcDtx5jT6vRT5zjJ3ZeFpjSAc9jMl6DMW\nKUX5sLfXWF4wmlBibkp0CZ4cw61ZbWbbN5TEsQIDAQABAoIBADL52tBbA24l6ei+\nUUuYvppjVVEL/dwx/MgRyJdmF46FWaXiC5LZd/dJ9RQZss8buztLrw/hVo+dFxHx\njFooHSAzZQU7AcD8bybziSBVI882lIfdr/NyGvqVFwjfV2lWQz0NB3F2IKLOJBq2\n+ZjNo5sZUeCUUzGc/kjkUGORbOjJr4OBkZM2hRqtdDbdaPK7OSiSFEpCwTuTLbhD\nCnP6ay+WWGADcgUSeYKrREohnjAhL8VuAOiQcveU4gmPfdMDZcq1cSdyEEO+NHs2\nFXLBKfdU42C61PRDfa5NNd1suTLBYAZWcr8AEnEvUEp/BeGhEBAdAlgVRHkjKs3G\nuD4GDGECgYEA03EQzRUbog6S2u8cIplUc5bJ9Wj+b+zmix22ve3FIxLTGfPb3cLL\nQe3JIkxLICYP7ZO1MaYy34+30lbPUbogkEzLP7fpt/qUk938CwLFL53ikOiWVA7k\np4hAC2G4Q20qP2biH6G4/h4eRhuXiJALu54GLtjMk2v9bQM2biYoWQUCgYEA3Kr8\n68vjIMU/xemllpafgVWICQMPOPCh6qdL9Fn7Atdt0BkjfDo3h1iaedEhKYQDZNvF\nxaiScSiByk9IFPyvIs0JEcywrKy5NQ4/0dMSQEOzfxBg9bPKI7/etHi9/ALW+XKB\naRZ4GNsr6YjC9szcE8LfQrCAsUe3BOX2zSUcnL0CgYAbiH+dlQASLD+nTrelMb4z\nhxEpadCoFns25lmjhdDD7nGa0Yxx5im9ng8w7ipiN1KfpzpTCsdZIUfYlgFNLSWM\nZNOaqoI+uNycHK3zaRrwRmj4YbEhpQbVYgKk+Mab0R1NQEJ1yANk49shWfpzh/5f\nIgbAFu8cy1Um2uI9ma5rWQKBgQDBdWankuhdIpD2ghCaJRNR4BqTTAtccBqEDoeY\nggp+Q0AS4PcrQh7MmfFUOvRH4WTYV5Tb5R399vVS2I7pV15ztC3vXPTHbeYxjXyG\nB/ZIQRJso39d6XGeReiJcBGfjx3JM4ohB4HiyMOGyk+i75dB++agIP2ybp0VvkbR\nM2gSQQKBgEm8KxlNQoI8FsV5I+9maL9ArFwUQX7Ck+5Z0/ldCviZdSAwDThY7Zng\nbY7my03T1A+gtVsgc/u8hsWuQh6wFPsB20wZoI0zDqcHnvmjgUUyQv18qhsAxiuO\nq3/JxAGV043+WleFaD9jGgBgPeqR3Aoch0U1AqDHDat5jVg2KLLT\n-----END RSA PRIVATE KEY-----",
			"--config-server-namespace", "config-server-ns",
		})
		Expect(err).ToNot(HaveOccurred())

		err = cmd.Execute()
		Expect(err).ToNot(HaveOccurred())
		Expect(ui.Blocks).To(Equal([]string{`interpolation_order_guarantee:
- ca: ca-cert
- missing_key: ((missing_key))
variables: []
`}))
	})

	It("generates and does not store missing password variable when --vars-store is configured with memory:// schema", func() {
		err := fs.WriteFileString(filepath.Join("/", "file"), `password: ((key))
variables:
- name: key
  type: password
`)
		Expect(err).ToNot(HaveOccurred())

		var genedPass string

		{ // running command first time
			cmd, err := cmdFactory.New([]string{
				"interpolate", filepath.Join("/", "file"),
				"--vars-store", "memory://" + filepath.Join("/", "vars1"),
				"--path", "/password",
			})
			Expect(err).ToNot(HaveOccurred())

			err = cmd.Execute()
			Expect(err).ToNot(HaveOccurred())
			Expect(ui.Blocks).To(HaveLen(1))

			genedPass = ui.Blocks[0]
			Expect(len(genedPass)).To(BeNumerically(">", 10))

			Expect(fs.FileExists(filepath.Join("/", "vars1"))).ToNot(BeTrue())
		}

		ui.Blocks = []string{}

		{ // running command second time
			cmd, err := cmdFactory.New([]string{
				"interpolate", filepath.Join("/", "file"),
				// use different memory store since this test uses same cmdFactory
				"--vars-store", "memory://",
				"--path", "/password",
			})
			Expect(err).ToNot(HaveOccurred())

			err = cmd.Execute()
			Expect(err).ToNot(HaveOccurred())
			Expect(ui.Blocks[0]).ToNot(Equal(genedPass))
		}
	})
})
