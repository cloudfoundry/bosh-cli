package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
)

var _ = Describe("PauseTasksCmd", func() {
	var (
		director *fakedir.FakeDirector
		command  PauseTasksCmd
	)

	BeforeEach(func() {
		director = &fakedir.FakeDirector{}
		command = NewPauseTasksCmd(director)
	})

	Describe("Run", func() {

		act := func() error { return command.Run() }

		It("pauses tasks", func() {

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.PauseTasksCallCount()).To(Equal(1))
		})

		It("returns error if changing resurrection fails", func() {
			director.PauseTasksReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
