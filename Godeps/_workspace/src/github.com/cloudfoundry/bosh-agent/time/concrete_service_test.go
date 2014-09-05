package time_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-agent/time"
)

var _ = Describe("concreteService", func() {
	Describe("Now", func() {
		It("returns current time", func() {
			service := NewConcreteService()
			t1 := service.Now()
			t2 := service.Now()
			Expect(float64(t2.Sub(t1))).To(BeNumerically(">", 0))
		})
	})

	Describe("Sleep", func() {
		It("sleeps for the given time", func() {
			service := NewConcreteService()

			t1 := time.Now()
			service.Sleep(1 * time.Millisecond)
			t2 := time.Now()
			Expect(float64(t2.Sub(t1))).To(BeNumerically(">", 1*time.Millisecond))
		})
	})
})
