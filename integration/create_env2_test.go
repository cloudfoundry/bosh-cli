package integration_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
)

var _ = Describe("create-env command (2)", func() {
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

	It("allows to use config server for state file storage when --state is configured with config-server:// schema", func() {
		err := fs.WriteFileString(filepath.Join("/", "file"), ``)
		Expect(err).ToNot(HaveOccurred())

		caCert, configServer := BuildHTTPSServer()
		defer configServer.Close()

		stateFileValue := &configServerValue{
			ExpectedBodyStr: `"name":"config-server-ns/state-json"`,
		}

		configServer.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v1/data", "name=config-server-ns/state-json"),
				ghttp.RespondWith(http.StatusNotFound, ``),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v1/data", "name=config-server-ns/state-json"),
				ghttp.RespondWith(http.StatusNotFound, ``),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/api/v1/data"),
				stateFileValue.Write,
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v1/data", "name=config-server-ns/state-json"),
				stateFileValue.Read,
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v1/data", "name=config-server-ns/state-json"),
				stateFileValue.Read,
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/api/v1/data"),
				stateFileValue.Write,
			),
		)

		cmd, err := cmdFactory.New([]string{
			"create-env", filepath.Join("/", "file"),
			"--state", "config-server://state.json",
			"--config-server-url", configServer.URL(),
			"--config-server-tls-ca", caCert,
			// not correct key pair since test server does not verify
			"--config-server-tls-certificate", "-----BEGIN CERTIFICATE-----\nMIIDSjCCAjKgAwIBAgIQWyNlE989qt4Jq5e3uPnEvzANBgkqhkiG9w0BAQsFADAz\nMQwwCgYDVQQGEwNVU0ExFjAUBgNVBAoTDUNsb3VkIEZvdW5kcnkxCzAJBgNVBAMT\nAmNhMB4XDTE4MDEzMTAyMDQwNloXDTE5MDEzMTAyMDQwNlowPDEMMAoGA1UEBhMD\nVVNBMRYwFAYDVQQKEw1DbG91ZCBGb3VuZHJ5MRQwEgYDVQQDEwtzZXJ2ZXItY2Vy\ndDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALZCZGt0DlFNSiRCZU83\nefftOa+EaXmkrMMr1chlb4rdSE5ft3G6Jw9cmdMDB4ZrA3pbQo6ENcFQn1AI6h/2\nC8q2XASUzzO7vDo/oYfsK4YGXXCgVEKB+aQagKDBhPCFW2m6SaoaHjxfZtS4TVBL\nepr7SwZfFBxFKLlInTOdy+i/TARFNf+xszrffVlhRTbTozCwI0wrURQeXPU0V5kh\nM4vOyDrN0/K9GmgO9dNC5X7T1JS2BQ7vtceAf7eSCyiuP3Zi8Gja9/51l3JKAXA7\nceY0+r0U+c4yd2XhaY0gHPYzJegzFilF+bC311heMJpQYm5KdAmeHMOtWW1m2zeU\nxLECAwEAAaNRME8wDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMB\nMAwGA1UdEwEB/wQCMAAwGgYDVR0RBBMwEYIJbG9jYWxob3N0hwR/AAABMA0GCSqG\nSIb3DQEBCwUAA4IBAQANsC2yzCvdQFW0Lr7iVNr44O4J3HLVdMuMGoNoYBGGk7+k\njiExDkY1BNvXtYxGTQo8x0a9i/DzrT1qTsxYQbmSUa35vh08bMhht6aRr4G5LkEq\nneHgJF3oo0G7sEu06dokKZLWkpHtHU9s2csXrLWYCIcejyWhkJGlKTN6WYgum3dS\nWjMSq20Hn+SW/PcUTBpB+gJDKnz2XCX9Hu0elVJOIv+DsRBNlrU0LUyF79Dbnl6j\n/nW3vZF9r7+LfYgLdLB661hvyv5F2e97ic7mmekG7nRop5j074vNPCj8yMdADqVn\nue3UUC2R6+gWoOxY0HVhVlGX9vVJap+/1O7rHBLr\n-----END CERTIFICATE-----",
			"--config-server-tls-private-key", "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAtkJka3QOUU1KJEJlTzd59+05r4RpeaSswyvVyGVvit1ITl+3\ncbonD1yZ0wMHhmsDeltCjoQ1wVCfUAjqH/YLyrZcBJTPM7u8Oj+hh+wrhgZdcKBU\nQoH5pBqAoMGE8IVbabpJqhoePF9m1LhNUEt6mvtLBl8UHEUouUidM53L6L9MBEU1\n/7GzOt99WWFFNtOjMLAjTCtRFB5c9TRXmSEzi87IOs3T8r0aaA7100LlftPUlLYF\nDu+1x4B/t5ILKK4/dmLwaNr3/nWXckoBcDtx5jT6vRT5zjJ3ZeFpjSAc9jMl6DMW\nKUX5sLfXWF4wmlBibkp0CZ4cw61ZbWbbN5TEsQIDAQABAoIBADL52tBbA24l6ei+\nUUuYvppjVVEL/dwx/MgRyJdmF46FWaXiC5LZd/dJ9RQZss8buztLrw/hVo+dFxHx\njFooHSAzZQU7AcD8bybziSBVI882lIfdr/NyGvqVFwjfV2lWQz0NB3F2IKLOJBq2\n+ZjNo5sZUeCUUzGc/kjkUGORbOjJr4OBkZM2hRqtdDbdaPK7OSiSFEpCwTuTLbhD\nCnP6ay+WWGADcgUSeYKrREohnjAhL8VuAOiQcveU4gmPfdMDZcq1cSdyEEO+NHs2\nFXLBKfdU42C61PRDfa5NNd1suTLBYAZWcr8AEnEvUEp/BeGhEBAdAlgVRHkjKs3G\nuD4GDGECgYEA03EQzRUbog6S2u8cIplUc5bJ9Wj+b+zmix22ve3FIxLTGfPb3cLL\nQe3JIkxLICYP7ZO1MaYy34+30lbPUbogkEzLP7fpt/qUk938CwLFL53ikOiWVA7k\np4hAC2G4Q20qP2biH6G4/h4eRhuXiJALu54GLtjMk2v9bQM2biYoWQUCgYEA3Kr8\n68vjIMU/xemllpafgVWICQMPOPCh6qdL9Fn7Atdt0BkjfDo3h1iaedEhKYQDZNvF\nxaiScSiByk9IFPyvIs0JEcywrKy5NQ4/0dMSQEOzfxBg9bPKI7/etHi9/ALW+XKB\naRZ4GNsr6YjC9szcE8LfQrCAsUe3BOX2zSUcnL0CgYAbiH+dlQASLD+nTrelMb4z\nhxEpadCoFns25lmjhdDD7nGa0Yxx5im9ng8w7ipiN1KfpzpTCsdZIUfYlgFNLSWM\nZNOaqoI+uNycHK3zaRrwRmj4YbEhpQbVYgKk+Mab0R1NQEJ1yANk49shWfpzh/5f\nIgbAFu8cy1Um2uI9ma5rWQKBgQDBdWankuhdIpD2ghCaJRNR4BqTTAtccBqEDoeY\nggp+Q0AS4PcrQh7MmfFUOvRH4WTYV5Tb5R399vVS2I7pV15ztC3vXPTHbeYxjXyG\nB/ZIQRJso39d6XGeReiJcBGfjx3JM4ohB4HiyMOGyk+i75dB++agIP2ybp0VvkbR\nM2gSQQKBgEm8KxlNQoI8FsV5I+9maL9ArFwUQX7Ck+5Z0/ldCviZdSAwDThY7Zng\nbY7my03T1A+gtVsgc/u8hsWuQh6wFPsB20wZoI0zDqcHnvmjgUUyQv18qhsAxiuO\nq3/JxAGV043+WleFaD9jGgBgPeqR3Aoch0U1AqDHDat5jVg2KLLT\n-----END RSA PRIVATE KEY-----",
			"--config-server-namespace", "config-server-ns",
		})
		Expect(err).ToNot(HaveOccurred())

		// Expect cmd to error as we do not want to continue with create-env
		err = cmd.Execute()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("releases must contain at least 1 release"))

		{
			type stateFileType struct {
				DirectorID     string `json:"director_id"`
				InstallationID string `json:"installation_id"`
			}
			var stateFile stateFileType

			err = json.Unmarshal([]byte(stateFileValue.Value()), &stateFile)
			Expect(err).ToNot(HaveOccurred())

			Expect(stateFile.DirectorID).ToNot(BeEmpty())
			Expect(stateFile.InstallationID).ToNot(BeEmpty())
		}
	})

	It("allows to use memory for state file storage when --state is configured with memory:// schema", func() {
		err := fs.WriteFileString(filepath.Join("/", "file"), ``)
		Expect(err).ToNot(HaveOccurred())

		cmd, err := cmdFactory.New([]string{
			"create-env", filepath.Join("/", "file"),
			"--state", "memory://" + filepath.Join("/", "state.json"),
		})
		Expect(err).ToNot(HaveOccurred())

		// Expect cmd to error as we do not want to continue with create-env
		err = cmd.Execute()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("releases must contain at least 1 release"))

		Expect(fs.FileExists(filepath.Join("/", "state.json"))).To(BeFalse())
	})
})

type configServerValue struct {
	val             []byte
	ExpectedBodyStr string
}

func (v *configServerValue) Write(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	Expect(err).ToNot(HaveOccurred())
	Expect(body).To(ContainSubstring(v.ExpectedBodyStr))

	v.val = body

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(v.val))
}

func (v *configServerValue) Read(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"data": [%s]}`, v.val)))
}

func (v *configServerValue) Value() string {
	type envelopeType struct {
		Value string
	}
	var envelope envelopeType

	err := json.Unmarshal(v.val, &envelope)
	Expect(err).ToNot(HaveOccurred())

	return envelope.Value
}
