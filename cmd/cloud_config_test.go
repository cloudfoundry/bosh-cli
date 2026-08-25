package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("CloudConfigCmd", func() {
	var (
		testUI   *testui.Ui
		director *fakedir.FakeDirector
		command  cmd.CloudConfigCmd
	)

	BeforeEach(func() {
		testUI = &testui.Ui{}
		director = &fakedir.FakeDirector{}
		command = cmd.NewCloudConfigCmd(testUI, director)
	})

	Describe("Run", func() {
		var (
			cloudConfigOpts opts.CloudConfigOpts
		)

		act := func() error { return command.Run(cloudConfigOpts) }

		It("shows cloud config", func() {
			cloudConfig := boshdir.CloudConfig{
				Properties: "some-properties",
			}

			director.LatestCloudConfigReturns(cloudConfig, nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(testUI.Blocks).To(Equal([]string{"some-properties"}))
		})

		It("returns error if cloud config cannot be retrieved", func() {
			director.LatestCloudConfigReturns(boshdir.CloudConfig{}, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
