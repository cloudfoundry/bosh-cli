package grouper_test

import (
	"errors"
	"os"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/fake_runner"
	"github.com/tedsuo/ifrit/grouper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ordered Group", func() {
	var (
		started chan struct{}

		groupRunner  ifrit.Runner
		groupProcess ifrit.Process
		members      grouper.Members

		childRunner1 *fake_runner.TestRunner
		childRunner2 *fake_runner.TestRunner
		childRunner3 *fake_runner.TestRunner

		Δ time.Duration = 10 * time.Millisecond
	)

	Describe("Start", func() {
		BeforeEach(func() {
			childRunner1 = fake_runner.NewTestRunner()
			childRunner2 = fake_runner.NewTestRunner()
			childRunner3 = fake_runner.NewTestRunner()

			members = grouper.Members{
				{"child1", childRunner1},
				{"child2", childRunner2},
				{"child3", childRunner3},
			}

			groupRunner = grouper.NewOrdered(os.Interrupt, members)
		})

		AfterEach(func() {
			childRunner1.EnsureExit()
			childRunner2.EnsureExit()
			childRunner3.EnsureExit()

			Eventually(started).Should(BeClosed())
			groupProcess.Signal(os.Kill)
			Eventually(groupProcess.Wait()).Should(Receive())
		})

		BeforeEach(func() {
			started = make(chan struct{})
			go func() {
				groupProcess = ifrit.Invoke(groupRunner)
				close(started)
			}()
		})

		It("runs the first runner, then the second, then becomes ready", func() {
			Eventually(childRunner1.RunCallCount).Should(Equal(1))
			Consistently(childRunner2.RunCallCount, Δ).Should(BeZero())
			Consistently(started, Δ).ShouldNot(BeClosed())

			childRunner1.TriggerReady()

			Eventually(childRunner2.RunCallCount).Should(Equal(1))
			Consistently(childRunner3.RunCallCount, Δ).Should(BeZero())
			Consistently(started, Δ).ShouldNot(BeClosed())

			childRunner2.TriggerReady()

			Eventually(childRunner3.RunCallCount).Should(Equal(1))
			Consistently(started, Δ).ShouldNot(BeClosed())

			childRunner3.TriggerReady()

			Eventually(started).Should(BeClosed())
		})

		Describe("when all the runners are ready", func() {
			var (
				signal1 <-chan os.Signal
				signal2 <-chan os.Signal
				signal3 <-chan os.Signal
			)

			BeforeEach(func() {
				signal1 = childRunner1.WaitForCall()
				childRunner1.TriggerReady()
				signal2 = childRunner2.WaitForCall()
				childRunner2.TriggerReady()
				signal3 = childRunner3.WaitForCall()
				childRunner3.TriggerReady()

				Eventually(started).Should(BeClosed())
			})

			Describe("when it receives a signal", func() {
				BeforeEach(func() {
					groupProcess.Signal(syscall.SIGUSR2)
				})

				It("doesn't send any more signals to remaining child processes", func() {
					Eventually(signal3).Should(Receive(Equal(syscall.SIGUSR2)))
					childRunner2.TriggerExit(nil)
					Consistently(signal3).ShouldNot(Receive())
				})
			})

			Describe("when a process exits cleanly", func() {
				BeforeEach(func() {
					childRunner1.TriggerExit(nil)
				})

				It("sends an interrupt signal to the other processes", func() {
					Eventually(signal3).Should(Receive(Equal(os.Interrupt)))
					childRunner3.TriggerExit(nil)
					Eventually(signal2).Should(Receive(Equal(os.Interrupt)))
				})

				It("does not exit", func() {
					Consistently(groupProcess.Wait(), Δ).ShouldNot(Receive())
				})

				Describe("when another process exits", func() {
					BeforeEach(func() {
						childRunner2.TriggerExit(nil)
					})

					It("doesn't send any more signals to remaining child processes", func() {
						Eventually(signal3).Should(Receive(Equal(os.Interrupt)))
						Consistently(signal3).ShouldNot(Receive())
					})
				})

				Describe("when all of the processes have exited cleanly", func() {
					BeforeEach(func() {
						childRunner2.TriggerExit(nil)
						childRunner3.TriggerExit(nil)
					})

					It("exits cleanly", func() {
						Eventually(groupProcess.Wait()).Should(Receive(BeNil()))
					})
				})

				Describe("when one of the processes exits with an error", func() {
					BeforeEach(func() {
						childRunner2.TriggerExit(errors.New("Fail"))
						childRunner3.TriggerExit(nil)
					})

					It("returns an error indicating which child processes failed", func() {
						var err error
						Eventually(groupProcess.Wait()).Should(Receive(&err))
						errTrace := err.(grouper.ErrorTrace)
						Ω(errTrace).Should(HaveLen(3))

						Ω(errTrace).Should(ContainElement(grouper.ExitEvent{grouper.Member{"child1", childRunner1}, nil}))
						Ω(errTrace).Should(ContainElement(grouper.ExitEvent{grouper.Member{"child2", childRunner2}, errors.New("Fail")}))
					})
				})
			})
		})

		Describe("when the first member is started", func() {
			var signals <-chan os.Signal

			BeforeEach(func() {
				childRunner1.WaitForCall()
				childRunner1.TriggerReady()
			})

			Describe("and the first member exits while second member is setting up", func() {
				BeforeEach(func() {
					signals = childRunner2.WaitForCall()
					childRunner1.TriggerExit(nil)
				})

				It("should terminate", func() {
					var signal os.Signal
					Eventually(signals).Should(Receive(&signal))
					Expect(signal).To(Equal(syscall.SIGINT))
				})
			})

			Describe("and the second member exits before becoming ready", func() {
				BeforeEach(func() {
					signals = childRunner1.WaitForCall()
					childRunner2.TriggerExit(nil)
				})

				It("should terminate the first runner", func() {
					var signal os.Signal
					Eventually(signals).Should(Receive(&signal))
					Expect(signal).To(Equal(syscall.SIGINT))
				})

				It("should not return an error", func() {
					var err error
					Eventually(groupProcess.Wait()).Should(Receive(&err))
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Describe("Failed start", func() {
			BeforeEach(func() {
				signal1 := childRunner1.WaitForCall()
				childRunner1.TriggerReady()
				childRunner2.TriggerExit(errors.New("Fail"))
				Eventually(signal1).Should(Receive(Equal(os.Interrupt)))
				childRunner1.TriggerExit(nil)
				Eventually(started).Should(BeClosed())
			})

			It("exits without starting further processes", func() {
				var err error

				Eventually(groupProcess.Wait()).Should(Receive(&err))
				errTrace := err.(grouper.ErrorTrace)
				Ω(errTrace).Should(ContainElement(grouper.ExitEvent{grouper.Member{"child1", childRunner1}, nil}))
				Ω(errTrace).Should(ContainElement(grouper.ExitEvent{grouper.Member{"child2", childRunner2}, errors.New("Fail")}))
				Ω(exitIndex("child1", errTrace)).Should(BeNumerically(">", exitIndex("child2", errTrace)))
			})
		})
	})

	Describe("Stop", func() {

		var runnerIndex int64
		var startOrder chan int64
		var stopOrder chan int64
		var receivedSignals chan os.Signal

		makeRunner := func(waitTime time.Duration) (ifrit.Runner, chan struct{}) {
			quickExit := make(chan struct{})
			return ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
				index := atomic.AddInt64(&runnerIndex, 1)
				startOrder <- index
				close(ready)

				select {
				case <-quickExit:
				case <-signals:
				}
				time.Sleep(waitTime)
				stopOrder <- index

				return nil
			}), quickExit
		}

		makeSignalEchoRunner := func(waitTime time.Duration, name string) ifrit.Runner {
			return ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
				close(ready)
				done := make(chan bool)
				go func() {
					time.Sleep(waitTime)
					done <- true
				}()
			L:
				for {
					select {
					case s := <-signals:
						receivedSignals <- s
					case _ = <-done:
						break L
					}
				}
				return nil
			})
		}

		Context("when runner receives a single signal", func() {
			BeforeEach(func() {
				startOrder = make(chan int64, 3)
				stopOrder = make(chan int64, 3)

				r1, _ := makeRunner(0)
				r2, _ := makeRunner(30 * time.Millisecond)
				r3, _ := makeRunner(50 * time.Millisecond)
				members = grouper.Members{
					{"child1", r1},
					{"child2", r2},
					{"child3", r3},
				}
			})

			AfterEach(func() {
				groupProcess.Signal(os.Kill)
				Eventually(groupProcess.Wait()).Should(Receive())
			})

			JustBeforeEach(func() {
				groupRunner = grouper.NewOrdered(os.Interrupt, members)

				started = make(chan struct{})
				go func() {
					groupProcess = ifrit.Invoke(groupRunner)
					close(started)
				}()

				Eventually(started).Should(BeClosed())
			})

			It("stops in reverse order", func() {
				groupProcess.Signal(os.Kill)
				Eventually(groupProcess.Wait()).Should(Receive())
				close(startOrder)
				close(stopOrder)

				Ω(startOrder).To(HaveLen(len(stopOrder)))

				order := []int64{}
				for r := range startOrder {
					order = append(order, r)
				}

				for i := len(stopOrder) - 1; i >= 0; i-- {
					Ω(order[i]).To(Equal(<-stopOrder))
				}
			})

			Context("when a runner stops", func() {
				var quickExit chan struct{}

				BeforeEach(func() {
					var r1 ifrit.Runner
					r1, quickExit = makeRunner(0)
					members[0].Runner = r1
				})

				It("stops in reverse order", func() {
					close(quickExit)
					Eventually(groupProcess.Wait()).Should(Receive())
					close(startOrder)
					close(stopOrder)

					Ω(startOrder).To(HaveLen(len(stopOrder)))

					order := []int64{}
					for r := range startOrder {
						order = append(order, r)
					}

					firstDeath := <-stopOrder
					for i := len(order) - 1; i >= 0; i-- {
						if order[i] == firstDeath {
							continue
						}
						Ω(order[i]).To(Equal(<-stopOrder))
					}
				})
			})
		})

		Context("when a runner receives multiple signals", func() {
			BeforeEach(func() {
				startOrder = make(chan int64, 2)
				stopOrder = make(chan int64, 2)

				r1 := makeSignalEchoRunner(200*time.Millisecond, "child1")
				r2 := makeSignalEchoRunner(100*time.Millisecond, "child2")
				members = grouper.Members{
					{"child1", r1},
					{"child2", r2},
				}
			})

			AfterEach(func() {
				groupProcess.Signal(os.Kill)
				Eventually(groupProcess.Wait()).Should(Receive())
			})

			JustBeforeEach(func() {
				groupRunner = grouper.NewOrdered(os.Interrupt, members)

				started = make(chan struct{})
				go func() {
					groupProcess = ifrit.Invoke(groupRunner)
					close(started)
				}()

				Eventually(started).Should(BeClosed())
			})

			Context("of different types", func() {

				BeforeEach(func() {
					receivedSignals = make(chan os.Signal, 4)
				})

				It("allows the process to finish gracefully", func() {
					groupProcess.Signal(syscall.SIGINT)
					Consistently(groupProcess.Wait(), 20*time.Millisecond, 10*time.Millisecond).ShouldNot(Receive())
					groupProcess.Signal(syscall.SIGUSR1)
					Consistently(groupProcess.Wait(), 20*time.Millisecond, 10*time.Millisecond).ShouldNot(Receive())
					groupProcess.Signal(syscall.SIGUSR2)
					Consistently(groupProcess.Wait(), 20*time.Millisecond, 10*time.Millisecond).ShouldNot(Receive())

					Eventually(groupProcess.Wait()).Should(Receive())

					signals := []os.Signal{syscall.SIGINT, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGUSR2}
					for _, expectedSig := range signals {
						actualSig := <-receivedSignals
						Expect(actualSig).Should(Equal(expectedSig))
					}
				})
			})

			Context("of same type", func() {

				BeforeEach(func() {
					receivedSignals = make(chan os.Signal, 2)
				})

				It("allows the process to finish gracefully", func() {
					groupProcess.Signal(syscall.SIGUSR1)
					Consistently(groupProcess.Wait(), 20*time.Millisecond, 10*time.Millisecond).ShouldNot(Receive())
					groupProcess.Signal(syscall.SIGUSR1)
					Consistently(groupProcess.Wait(), 20*time.Millisecond, 10*time.Millisecond).ShouldNot(Receive())
					groupProcess.Signal(syscall.SIGUSR1)
					Consistently(groupProcess.Wait(), 20*time.Millisecond, 10*time.Millisecond).ShouldNot(Receive())

					Eventually(groupProcess.Wait()).Should(Receive())

					signals := []os.Signal{syscall.SIGUSR1, syscall.SIGUSR1}
					for _, expectedSig := range signals {
						actualSig := <-receivedSignals
						Expect(actualSig).Should(Equal(expectedSig))
					}
				})
			})
		})
	})
})

func exitIndex(name string, errTrace grouper.ErrorTrace) int {
	for i, exitTrace := range errTrace {
		if exitTrace.Member.Name == name {
			return i
		}
	}

	return -1
}
