package cmd_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	cmd "github.com/cloudfoundry/bosh-micro-cli/cmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeploymentCmd", func() {
	var command cmd.Cmd
	var manifestPath string
	var args []string

	BeforeEach(func() {
		command = cmd.NewDeploymentCmd()
		Expect(command).ToNot(BeNil())
	})

	Context("#Run", func() {
		Context("ran with valid args", func() {
			BeforeEach(func() {
				file, err := ioutil.TempFile("", "bosh-micro-cli-manifest")
				Expect(err).ToNot(HaveOccurred())

				manifestPath = file.Name()
				args = []string{manifestPath}
			})

			AfterEach(func() {
				err := os.RemoveAll(manifestPath)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("storing the file", func() {
				var expectedFilePath string
				var expectedJsonContent string

				BeforeEach(func() {
					usr, err := user.Current()
					Expect(err).ToNot(HaveOccurred())

					expectedFilePath = path.Join(usr.HomeDir, ".bosh_micro.json")
					expectedJsonContent = fmt.Sprintf(`
					{
						"deployment" : "%s"
					}
					`, manifestPath)
				})

				It("stores the manifest file path to a hidden file at the home dir", func() {
					err := command.Run(args)
					Expect(err).ToNot(HaveOccurred())

					actualFileContent, err := ioutil.ReadFile(expectedFilePath)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(actualFileContent)).To(ContainSubstring(manifestPath))
				})

				It("store the file in json format", func() {
					err := command.Run(args)
					Expect(err).ToNot(HaveOccurred())

					actualFileContent, err := ioutil.ReadFile(expectedFilePath)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(actualFileContent)).To(MatchJSON(expectedJsonContent))
				})

			})
		})

		Context("ran with invalid args", func() {
			It("fails when manifest file path is nil", func() {
				err := command.Run(nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Deployment command argument cannot be nil"))
			})

			It("fails when manifest file path is empty", func() {
				err := command.Run([]string{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Deployment command arguments must have at least one arg"))
			})

			It("fails when manifest file path does not exist", func() {
				err := command.Run([]string{"fake/manifest/path"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Deployment command manifest path fake/manifest/path does not exist"))
			})
		})
	})
})
