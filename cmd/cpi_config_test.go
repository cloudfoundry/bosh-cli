package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("CPIConfigCmd", func() {
	var (
		testUI   *testui.Ui
		director *fakedir.FakeDirector
		command  cmd.CPIConfigCmd
	)

	BeforeEach(func() {
		testUI = &testui.Ui{}
		director = &fakedir.FakeDirector{}
		command = cmd.NewCPIConfigCmd(testUI, director)
	})

	Describe("Run", func() {
		act := func() error { return command.Run() }

		It("shows cpi config", func() {
			cpiConfig := boshdir.CPIConfig{
				Properties: "some-properties",
			}

			director.LatestCPIConfigReturns(cpiConfig, nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(testUI.Blocks).To(Equal([]string{"some-properties"}))
		})

		It("returns error if cpi config cannot be retrieved", func() {
			director.LatestCPIConfigReturns(boshdir.CPIConfig{}, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
