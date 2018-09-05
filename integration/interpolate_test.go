package integration_test

import (
	"crypto/x509"
	"encoding/pem"
	"path/filepath"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

	It("generates and stores missing password variable with custom length when --vars-store is provided", func() {
		err := fs.WriteFileString(filepath.Join("/", "file"), `password: ((key))
variables:
- name: key
  type: password
  options:
    length: 42
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
			Expect(len(genedPass)).To(Equal(42 + 1))

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

			caBlock, _ := pem.Decode([]byte(store.CA.Certificate))
			ca, err := x509.ParseCertificate(caBlock.Bytes)
			Expect(err).ToNot(HaveOccurred())

			Expect(cert.SubjectKeyId).ToNot(BeNil())
			Expect(cert.AuthorityKeyId).To(Equal(ca.SubjectKeyId))

			_, err = cert.Verify(x509.VerifyOptions{DNSName: "test.com", Roots: roots})
			Expect(err).ToNot(HaveOccurred())

			_, err = cert.Verify(x509.VerifyOptions{DNSName: "not-test.com", Roots: roots})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("certificate is valid"))
		}
	})

	It("generates and stores missing certificate variable with organization when --vars-store is provided", func() {
		err := fs.WriteFileString("/file", `
ca:
  certificate: ((ca.certificate))

variables:
- name: ca
  type: certificate
  options:
    is_ca: true
    common_name: ca
    organization: "org-AB"
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
			CA expectedCert
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
		}

		{
			roots := x509.NewCertPool()

			ok := roots.AppendCertsFromPEM([]byte(store.CA.Certificate))
			Expect(ok).To(BeTrue())

			caBlock, _ := pem.Decode([]byte(store.CA.Certificate))
			ca, err := x509.ParseCertificate(caBlock.Bytes)
			Expect(err).ToNot(HaveOccurred())

			Expect(ca.Subject).To(ContainSubstring("org-AB"))
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
})
