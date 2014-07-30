package cmd_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	cmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	uifakes "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeploymentCmd", func() {
	var command cmd.Cmd
	var manifestPath string
	var args []string
	var boshMicroPath string
	var boshMicroFile string
	var fakeUI *uifakes.FakeUI

	BeforeEach(func() {
		var err error
		boshMicroPath, err = ioutil.TempDir("", "bosh-micro-cli")
		Expect(err).NotTo(HaveOccurred())

		fakeUI = &uifakes.FakeUI{}
		command = cmd.NewDeploymentCmd(fakeUI, boshMicroPath)
		Expect(command).ToNot(BeNil())
	})

	AfterEach(func() {
		err := os.RemoveAll(boshMicroPath)
		Expect(err).NotTo(HaveOccurred())
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

			Context("a bosh micro JSON is present", func() {
				Context("and valid", func() {
					var expectedFilePath string

					BeforeEach(func() {
						expectedFilePath = path.Join(boshMicroPath, ".bosh_micro.json")
						err := ioutil.WriteFile(expectedFilePath, []byte("{}"), os.ModePerm)

						err = command.Run(args)
						Expect(err).ToNot(HaveOccurred())
					})

					It("stores the manifest file path to a hidden file at the home dir", func() {
						actualFileContent, err := ioutil.ReadFile(expectedFilePath)
						Expect(err).NotTo(HaveOccurred())
						Expect(string(actualFileContent)).To(ContainSubstring(manifestPath))
					})

					It("stores the manifiest file in JSON format", func() {
						var expectedJsonContent string
						expectedJsonContent = fmt.Sprintf(`
						{
							"deployment" : "%s"
						}
						`, manifestPath)

						actualFileContent, err := ioutil.ReadFile(expectedFilePath)
						Expect(err).NotTo(HaveOccurred())
						Expect(string(actualFileContent)).To(MatchJSON(expectedJsonContent))
					})

					It("says 'deployment set..' to the UI", func() {
						Expect(fakeUI.Said).To(ContainElement(ContainSubstring(fmt.Sprintf("Deployment set to '%s'", manifestPath))))
					})
				})

				Context("and invalid", func() {
					BeforeEach(func() {
						file, err := ioutil.TempFile("", "bosh-micro-cli-manifest")
						Expect(err).ToNot(HaveOccurred())
						manifestPath = file.Name()

						boshMicroFile = path.Join(boshMicroPath, ".bosh_micro.json")
						err = ioutil.WriteFile(boshMicroFile, []byte("---invalid JSON---"), os.ModePerm)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						err := os.RemoveAll(manifestPath)
						Expect(err).ToNot(HaveOccurred())

						err = os.Remove(boshMicroFile)
						Expect(err).ToNot(HaveOccurred())
					})

					It("errors with cannot parse JSON error", func() {
						err := command.Run([]string{manifestPath})
						Expect(err).To(HaveOccurred())
						errorMessage := fmt.Sprintf("Could not unmarshal JSON content '%s'", boshMicroFile)
						Expect(err.Error()).To(ContainSubstring(errorMessage))
					})
				})
			})
		})

		Context("ran without args", func() {
			Context("when the bosh file exists", func() {
				BeforeEach(func() {
					file, err := ioutil.TempFile("", "bosh-micro-cli-manifest")
					Expect(err).ToNot(HaveOccurred())
					manifestPath = file.Name()

					expectedJsonContent := fmt.Sprintf(`
						{
							"deployment" : "%s"
						}
						`, manifestPath)
					boshMicroFile = path.Join(boshMicroPath, ".bosh_micro.json")

					err = ioutil.WriteFile(boshMicroFile, []byte(expectedJsonContent), os.ModePerm)
					Expect(err).ToNot(HaveOccurred())
				})

				AfterEach(func() {
					err := os.RemoveAll(manifestPath)
					Expect(err).ToNot(HaveOccurred())

					err = os.Remove(boshMicroFile)
					Expect(err).ToNot(HaveOccurred())
				})

				It("says `Deployment set to '<manifest_path>'`", func() {
					err := command.Run(nil)
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeUI.Said).To(ContainElement(fmt.Sprintf("Current deployment is '%s'", manifestPath)))
				})
			})

			Context("when the bosh file does not exist", func() {
				It("says `Deployment not set' to UI stderr when called with nil", func() {
					err := command.Run(nil)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Deployment not set"))
					Expect(fakeUI.Errors).To(ContainElement("Deployment not set"))
				})

				It("says `Deployment not set' to UI stderr when called with empty string slice", func() {
					err := command.Run([]string{})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Deployment not set"))
					Expect(fakeUI.Errors).To(ContainElement("Deployment not set"))
				})

				It("fails when manifest file path does not exist", func() {
					err := command.Run([]string{"fake/manifest/path"})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Deployment command manifest path fake/manifest/path does not exist"))
				})
			})

		})
	})
})
