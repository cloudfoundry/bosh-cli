package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
)

var _ = Describe("CancelTaskCmd", func() {
	var (
		director *fakedir.FakeDirector
		command  cmd.CancelTaskCmd
	)

	BeforeEach(func() {
		director = &fakedir.FakeDirector{}
		command = cmd.NewCancelTaskCmd(director)
	})

	Describe("Run", func() {
		var (
			cancelTaskOpts opts.CancelTaskOpts
			task           *fakedir.FakeTask
		)

		BeforeEach(func() {
			cancelTaskOpts = opts.CancelTaskOpts{Args: opts.TaskArgs{ID: 123}}
			task = &fakedir.FakeTask{}
			director.FindTaskReturns(task, nil)
		})

		act := func() error { return command.Run(cancelTaskOpts) }

		It("fetches and cancels given task", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.FindTaskCallCount()).To(Equal(1))
			Expect(director.FindTaskArgsForCall(0)).To(Equal(123))

			Expect(task.CancelCallCount()).To(Equal(1))
		})

		It("returns error if task cancellation fails", func() {
			task.CancelReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if task cannot be retrieved", func() {
			director.FindTaskReturns(nil, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
