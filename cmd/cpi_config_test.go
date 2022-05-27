package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v6/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/v6/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v6/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v6/ui/fakes"
)

var _ = Describe("CPIConfigCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  CPIConfigCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = NewCPIConfigCmd(ui, director)
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

			Expect(ui.Blocks).To(Equal([]string{"some-properties"}))
		})

		It("returns error if cpi config cannot be retrieved", func() {
			director.LatestCPIConfigReturns(boshdir.CPIConfig{}, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
