package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
)

var _ = Describe("TaskConfigCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  TaskConfigCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = NewTaskConfigCmd(ui, director)
	})

	Describe("Run", func() {
		act := func() error { return command.Run() }

		It("shows task config", func() {
			taskConfig := boshdir.TaskConfig{
				Properties: "some-properties",
			}

			director.LatestTaskConfigReturns(taskConfig, nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(ui.Blocks).To(Equal([]string{"some-properties"}))
		})

		It("returns error if task config cannot be retrieved", func() {
			director.LatestTaskConfigReturns(boshdir.TaskConfig{}, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
