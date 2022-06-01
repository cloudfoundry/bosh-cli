package cmd_test

import (
	"errors"

	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
)

var _ = Describe("ManifestCmd", func() {
	var (
		ui         *fakeui.FakeUI
		deployment *fakedir.FakeDeployment
		command    ManifestCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		deployment = &fakedir.FakeDeployment{}
	})

	JustBeforeEach(func() {
		command = NewManifestCmd(ui, deployment)
	})

	Describe("Run", func() {
		act := func() error { return command.Run() }

		It("shows deployment manifest", func() {
			deployment.ManifestReturns("some-manifest", nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(ui.Blocks).To(Equal([]string{"some-manifest"}))
		})

		It("returns error if manifest cannot be retrieved", func() {
			deployment.ManifestReturns("", errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
