package cmd_test

import (
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
)

var _ = Describe("ErrandEventWatcher", func() {
	Describe("ParseInstanceSlug", func() {
		It("extracts group/uuid from 'group/uuid (idx)'", func() {
			Expect(cmd.ParseInstanceSlug("smoke-tests/abc-123 (0)")).To(Equal("smoke-tests/abc-123"))
		})

		It("extracts group/uuid when no index suffix", func() {
			Expect(cmd.ParseInstanceSlug("smoke-tests/abc-123")).To(Equal("smoke-tests/abc-123"))
		})

		It("returns empty for group-only (no slash)", func() {
			Expect(cmd.ParseInstanceSlug("smoke-tests")).To(Equal(""))
		})

		It("returns empty for empty string", func() {
			Expect(cmd.ParseInstanceSlug("")).To(Equal(""))
		})

		It("handles whitespace", func() {
			Expect(cmd.ParseInstanceSlug("  smoke-tests/abc-123 (2)  ")).To(Equal("smoke-tests/abc-123"))
		})
	})

	Describe("Watch", func() {
		var (
			deployment *fakedir.FakeDeployment
		)

		BeforeEach(func() {
			deployment = &fakedir.FakeDeployment{}
		})

		makeEvent := func(stage, state, task string) string {
			ev := map[string]any{
				"stage": stage,
				"state": state,
				"task":  task,
				"time":  1772657703,
			}
			b, err := json.Marshal(ev)
			Expect(err).NotTo(HaveOccurred())
			return string(b)
		}

		It("parses 'Running errand' started events and emits instance slugs", func() {
			events := strings.Join([]string{
				makeEvent("Preparing deployment", "started", "Preparing deployment"),
				makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)"),
			}, "\n")

			callCount := 0
			deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
				callCount++
				if callCount == 1 {
					return []byte(events), len(events), nil
				}
				return nil, offset, nil
			}
			deployment.TaskStateReturns("done", nil)

			watcher := cmd.NewErrandEventWatcher(deployment, 42, 0)
			stopCh := make(chan struct{})
			slugCh := watcher.Watch(stopCh)

			var slugs []string
			for s := range slugCh {
				slugs = append(slugs, s)
			}

			Expect(slugs).To(Equal([]string{"smoke-tests/abc-123"}))
		})

		It("handles multiple instances", func() {
			events := strings.Join([]string{
				makeEvent("Running errand", "started", "mysql/aaa-111 (0)"),
				makeEvent("Running errand", "started", "mysql/bbb-222 (1)"),
				makeEvent("Running errand", "started", "mysql/ccc-333 (2)"),
			}, "\n")

			callCount := 0
			deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
				callCount++
				if callCount == 1 {
					return []byte(events), len(events), nil
				}
				return nil, offset, nil
			}
			deployment.TaskStateReturns("done", nil)

			watcher := cmd.NewErrandEventWatcher(deployment, 42, 0)
			stopCh := make(chan struct{})
			slugCh := watcher.Watch(stopCh)

			var slugs []string
			for s := range slugCh {
				slugs = append(slugs, s)
			}

			Expect(slugs).To(ConsistOf("mysql/aaa-111", "mysql/bbb-222", "mysql/ccc-333"))
		})

		It("ignores non-errand events", func() {
			events := strings.Join([]string{
				makeEvent("Preparing deployment", "started", "Preparing deployment"),
				makeEvent("Creating missing vms", "started", "smoke-tests/abc-123 (0)"),
				makeEvent("Fetching logs", "started", "smoke-tests/abc-123 (0)"),
			}, "\n")

			deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
				return []byte(events), len(events), nil
			}
			deployment.TaskStateReturns("done", nil)

			watcher := cmd.NewErrandEventWatcher(deployment, 42, 0)
			stopCh := make(chan struct{})
			slugCh := watcher.Watch(stopCh)

			var slugs []string
			for s := range slugCh {
				slugs = append(slugs, s)
			}

			Expect(slugs).To(BeEmpty())
		})

		It("ignores finished/failed states for Running errand", func() {
			events := strings.Join([]string{
				makeEvent("Running errand", "finished", "smoke-tests/abc-123 (0)"),
				makeEvent("Running errand", "failed", "smoke-tests/abc-123 (0)"),
			}, "\n")

			deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
				return []byte(events), len(events), nil
			}
			deployment.TaskStateReturns("done", nil)

			watcher := cmd.NewErrandEventWatcher(deployment, 42, 0)
			stopCh := make(chan struct{})
			slugCh := watcher.Watch(stopCh)

			var slugs []string
			for s := range slugCh {
				slugs = append(slugs, s)
			}

			Expect(slugs).To(BeEmpty())
		})

		It("handles malformed JSON gracefully", func() {
			events := strings.Join([]string{
				"this is not json",
				"",
				makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)"),
				"{broken json",
			}, "\n")

			deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
				return []byte(events), len(events), nil
			}
			deployment.TaskStateReturns("done", nil)

			watcher := cmd.NewErrandEventWatcher(deployment, 42, 0)
			stopCh := make(chan struct{})
			slugCh := watcher.Watch(stopCh)

			var slugs []string
			for s := range slugCh {
				slugs = append(slugs, s)
			}

			Expect(slugs).To(Equal([]string{"smoke-tests/abc-123"}))
		})

		It("deduplicates slugs", func() {
			events := strings.Join([]string{
				makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)"),
				makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)"),
			}, "\n")

			deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
				return []byte(events), len(events), nil
			}
			deployment.TaskStateReturns("done", nil)

			watcher := cmd.NewErrandEventWatcher(deployment, 42, 0)
			stopCh := make(chan struct{})
			slugCh := watcher.Watch(stopCh)

			var slugs []string
			for s := range slugCh {
				slugs = append(slugs, s)
			}

			Expect(slugs).To(Equal([]string{"smoke-tests/abc-123"}))
		})

		It("polls incrementally using offset", func() {
			event1 := makeEvent("Running errand", "started", "mysql/aaa-111 (0)")
			event2 := makeEvent("Running errand", "started", "mysql/bbb-222 (1)")

			callCount := 0
			deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
				callCount++
				switch callCount {
				case 1:
					return []byte(event1 + "\n"), len(event1) + 1, nil
				case 2:
					return []byte(event2 + "\n"), len(event1) + 1 + len(event2) + 1, nil
				default:
					return nil, offset, nil
				}
			}

			taskStateCallCount := 0
			deployment.TaskStateStub = func(id int) (string, error) {
				taskStateCallCount++
				if taskStateCallCount <= 2 {
					return "processing", nil
				}
				return "done", nil
			}

			watcher := cmd.NewErrandEventWatcher(deployment, 42, 0)
			stopCh := make(chan struct{})
			slugCh := watcher.Watch(stopCh)

			var slugs []string
			for s := range slugCh {
				slugs = append(slugs, s)
			}

			Expect(slugs).To(ConsistOf("mysql/aaa-111", "mysql/bbb-222"))
			Expect(deployment.FetchTaskOutputChunkCallCount()).To(BeNumerically(">=", 2))

			_, secondOffset, _ := deployment.FetchTaskOutputChunkArgsForCall(1)
			Expect(secondOffset).To(Equal(len(event1) + 1))
		})

		It("feeds event chunks to TaskReporter when set", func() {
			events := strings.Join([]string{
				makeEvent("Preparing deployment", "started", "Preparing deployment"),
				makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)"),
			}, "\n")

			callCount := 0
			deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
				callCount++
				if callCount == 1 {
					return []byte(events), len(events), nil
				}
				return nil, offset, nil
			}
			deployment.TaskStateReturns("done", nil)

			reporter := &fakedir.FakeTaskReporter{}
			watcher := cmd.NewErrandEventWatcher(deployment, 42, 0)
			watcher.WithTaskReporter(reporter)
			stopCh := make(chan struct{})
			slugCh := watcher.Watch(stopCh)

			for range slugCh {
			}

			Expect(reporter.TaskStartedCallCount()).To(Equal(1))
			startedID := reporter.TaskStartedArgsForCall(0)
			Expect(startedID).To(Equal(42))

			Expect(reporter.TaskOutputChunkCallCount()).To(BeNumerically(">=", 1))
			chunkID, chunkData := reporter.TaskOutputChunkArgsForCall(0)
			Expect(chunkID).To(Equal(42))
			Expect(string(chunkData)).To(ContainSubstring("Running errand"))

			Expect(reporter.TaskFinishedCallCount()).To(Equal(1))
			finishedID, finishedState := reporter.TaskFinishedArgsForCall(0)
			Expect(finishedID).To(Equal(42))
			Expect(finishedState).To(Equal("done"))
		})

		It("discovers slugs from the final fetch after task completes and reports them", func() {
			lateEvent := makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)")

			// The goroutine calls FetchTaskOutputChunk twice per iteration
			// (once in the normal poll, once in the final fetch when the task
			// is done). We need the data to appear only in the final fetch
			// (call 3), not the normal poll (calls 1 and 2).
			//
			// Iteration 1: fetch(1)=empty, TaskState="done" -> final fetch(2)=event
			fetchCallCount := 0
			deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
				fetchCallCount++
				if fetchCallCount == 2 {
					return []byte(lateEvent), len(lateEvent), nil
				}
				return nil, offset, nil
			}

			deployment.TaskStateReturns("done", nil)

			reporter := &fakedir.FakeTaskReporter{}
			watcher := cmd.NewErrandEventWatcher(deployment, 42, 0)
			watcher.WithTaskReporter(reporter)
			stopCh := make(chan struct{})
			slugCh := watcher.Watch(stopCh)

			var slugs []string
			for s := range slugCh {
				slugs = append(slugs, s)
			}

			Expect(slugs).To(Equal([]string{"smoke-tests/abc-123"}))

			Expect(reporter.TaskOutputChunkCallCount()).To(Equal(1))
			chunkID, chunkData := reporter.TaskOutputChunkArgsForCall(0)
			Expect(chunkID).To(Equal(42))
			Expect(string(chunkData)).To(ContainSubstring("Running errand"))
		})

		It("stops when stopCh is closed", func() {
			deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
				return nil, offset, nil
			}
			deployment.TaskStateReturns("processing", nil)

			watcher := cmd.NewErrandEventWatcher(deployment, 42, 0)
			stopCh := make(chan struct{})
			slugCh := watcher.Watch(stopCh)

			close(stopCh)

			var slugs []string
			for s := range slugCh {
				slugs = append(slugs, s)
			}

			Expect(slugs).To(BeEmpty())
		})
	})
})
