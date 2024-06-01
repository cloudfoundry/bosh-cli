package cmd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
)

var _ = Describe("CancelTasksCmd", func() {
	var (
		director *fakedir.FakeDirector
		command  cmd.CancelTasksCmd
	)

	BeforeEach(func() {
		director = &fakedir.FakeDirector{}
		command = cmd.NewCancelTasksCmd(director)
	})

	Describe("Run", func() {
		var cancelTasksOpts opts.CancelTasksOpts

		act := func() error { return command.Run(cancelTasksOpts) }

		It("cancels all tasks given types", func() {
			cancelTasksOpts = opts.CancelTasksOpts{
				Types: []string{"fake-type1", "fake-type2"},
			}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.CancelTasksCallCount()).To(Equal(1))
			Expect(director.CancelTasksArgsForCall(0)).To(Equal(boshdir.TasksFilter{
				Types: []string{"fake-type1", "fake-type2"},
			}))
		})

		It("cancels all tasks given states", func() {
			cancelTasksOpts = opts.CancelTasksOpts{
				States: []string{"fake-state", "fake-state2"},
			}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.CancelTasksCallCount()).To(Equal(1))
			Expect(director.CancelTasksArgsForCall(0)).To(Equal(boshdir.TasksFilter{
				States: []string{"fake-state", "fake-state2"},
			}))
		})

		It("cancels all tasks given deployments", func() {
			cancelTasksOpts = opts.CancelTasksOpts{
				Deployment: "fake-deployment",
			}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.CancelTasksCallCount()).To(Equal(1))
			Expect(director.CancelTasksArgsForCall(0)).To(Equal(boshdir.TasksFilter{
				Deployment: "fake-deployment",
			}))
		})
	})
})
