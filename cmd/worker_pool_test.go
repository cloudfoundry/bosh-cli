package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

var _ = Describe("WorkerPool", func() {
	It("runs the given tasks", func() {
		pool := WorkerPool{
			WorkerCount: 2,
		}

		results, err := pool.ParallelDo(
			func() (interface{}, error) {
				return 1, nil
			},
			func() (interface{}, error) {
				return 2, nil
			},
			func() (interface{}, error) {
				return 3, nil
			},
		)
		Expect(err).ToNot(HaveOccurred())

		Expect(results).To(ConsistOf([]int{1, 2, 3}))
	})

	It("bubbles up any errors", func() {
		pool := WorkerPool{
			WorkerCount: 2,
		}

		_, err := pool.ParallelDo(
			func() (interface{}, error) {
				return 1, nil
			},
			func() (interface{}, error) {
				return -1, bosherr.ComplexError{
					Err:   errors.New("fake-error"),
					Cause: errors.New("fake-cause"),
				}
			},
			func() (interface{}, error) {
				return 3, nil
			},
		)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-error"))
		Expect(err.Error()).To(ContainSubstring("fake-cause"))
	})

	It("stops working after the first error", func() {
		pool := WorkerPool{
			WorkerCount: 1, // Force serial run
		}

		_, err := pool.ParallelDo(
			func() (interface{}, error) {
				return 1, nil
			},
			func() (interface{}, error) {
				return -1, bosherr.ComplexError{
					Err:   errors.New("fake-error"),
					Cause: errors.New("fake-cause"),
				}
			},
			func() (interface{}, error) {
				Fail("Expected third test to not run")
				return 3, nil
			},
		)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-error"))
		Expect(err.Error()).To(ContainSubstring("fake-cause"))
	})
})
