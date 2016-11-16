package integration_test

import (
	"crypto/x509"
	"encoding/pem"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
)

var _ = Describe("build-manifest command", func() {
	var (
		ui         *fakeui.FakeUI
		fs         *fakesys.FakeFileSystem
		cmdFactory Factory
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()

		ui = &fakeui.FakeUI{}
		logger := boshlog.NewLogger(boshlog.LevelNone)
		confUI := boshui.NewWrappingConfUI(ui, logger)

		deps := NewBasicDeps(confUI, logger)
		deps.FS = fs

		cmdFactory = NewFactory(deps)
	})

	It("interpolates manifest with variables", func() {
		err := fs.WriteFileString("/file", "file: ((key))")
		Expect(err).ToNot(HaveOccurred())

		cmd, err := cmdFactory.New([]string{"build-manifest", "/file", "-v", "key=val"})
		Expect(err).ToNot(HaveOccurred())

		err = cmd.Execute()
		Expect(err).ToNot(HaveOccurred())
		Expect(ui.Blocks).To(Equal([]string{"file: val\n"}))
	})

	It("returns portion of the template when --path flag is provided", func() {
		err := fs.WriteFileString("/file", "file: ((key))")
		Expect(err).ToNot(HaveOccurred())

		cmd, err := cmdFactory.New([]string{"build-manifest", "/file", "-v", `key={"nested": true}`, "--path", "/file/nested"})
		Expect(err).ToNot(HaveOccurred())

		err = cmd.Execute()
		Expect(err).ToNot(HaveOccurred())
		Expect(ui.Blocks).To(Equal([]string{"true\n"}))
	})

	It("generates and stores missing password variable when --vars-store is provided", func() {
		err := fs.WriteFileString("/file", `password: ((key))
variables:
- name: key
  type: password
`)
		Expect(err).ToNot(HaveOccurred())

		var genedPass string

		{ // running command first time
			cmd, err := cmdFactory.New([]string{"build-manifest", "/file", "--vars-store", "/vars", "--path", "/password"})
			Expect(err).ToNot(HaveOccurred())

			err = cmd.Execute()
			Expect(err).ToNot(HaveOccurred())
			Expect(ui.Blocks).To(HaveLen(1))

			genedPass = ui.Blocks[0]
			Expect(len(genedPass)).To(BeNumerically(">", 10))

			contents, err := fs.ReadFileString("/vars")
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(Equal("key: " + genedPass))
		}

		ui.Blocks = []string{}

		{ // running command second time
			cmd, err := cmdFactory.New([]string{"build-manifest", "/file", "--vars-store", "/vars", "--path", "/password"})
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
    common_name: ca
- name: server
  type: certificate
  options:
    ca: ca
    common_name: ((common_name))
`)
		Expect(err).ToNot(HaveOccurred())

		cmd, err := cmdFactory.New([]string{"build-manifest", "/file", "--vars-store", "/vars", "-v", "common_name=test.com"})
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
			contents, err := fs.ReadFileString("/vars")
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
})
