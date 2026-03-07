package cmd

import (
	"encoding/json"
	"strings"
	"time"

	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
)

type taskEvent struct {
	Stage    string `json:"stage"`
	State    string `json:"state"`
	Task     string `json:"task"`
	Time     int64  `json:"time"`
	Index    int    `json:"index"`
	Total    int    `json:"total"`
	Progress int    `json:"progress"`
}

type ErrandEventWatcher struct {
	deployment   boshdir.Deployment
	taskID       int
	pollDelay    time.Duration
	taskReporter boshdir.TaskReporter
}

func NewErrandEventWatcher(deployment boshdir.Deployment, taskID int, pollDelay time.Duration) *ErrandEventWatcher {
	return &ErrandEventWatcher{
		deployment: deployment,
		taskID:     taskID,
		pollDelay:  pollDelay,
	}
}

func (w *ErrandEventWatcher) WithTaskReporter(reporter boshdir.TaskReporter) *ErrandEventWatcher {
	w.taskReporter = reporter
	return w
}

// Watch polls the task event stream and sends discovered instance slugs
// (e.g. "smoke-tests/abc-123") on the returned channel. The channel is closed
// when the task is no longer running. If a TaskReporter is set, event chunks
// are also fed to it for real-time formatted output.
func (w *ErrandEventWatcher) Watch(stopCh <-chan struct{}) <-chan string {
	slugCh := make(chan string, 16)

	if w.taskReporter != nil {
		w.taskReporter.TaskStarted(w.taskID)
	}

	go func() {
		defer close(slugCh)

		var offset int
		seen := map[string]bool{}
		var lastState string

		for {
			select {
			case <-stopCh:
				return
			default:
			}

			chunk, newOffset, err := w.deployment.FetchTaskOutputChunk(w.taskID, offset, "event")
			if err == nil && len(chunk) > 0 {
				offset = newOffset
				if w.taskReporter != nil {
					w.taskReporter.TaskOutputChunk(w.taskID, chunk)
				}
				for _, slug := range parseEventChunk(chunk) {
					if !seen[slug] {
						seen[slug] = true
						select {
						case slugCh <- slug:
						case <-stopCh:
							return
						}
					}
				}
			}

			state, err := w.deployment.TaskState(w.taskID)
			if err != nil || !isTaskRunning(state) {
				if err == nil {
					lastState = state
				}
				// One final fetch to catch any remaining events
				chunk, _, err = w.deployment.FetchTaskOutputChunk(w.taskID, offset, "event")
				if err == nil && len(chunk) > 0 {
					if w.taskReporter != nil {
						w.taskReporter.TaskOutputChunk(w.taskID, chunk)
					}
					for _, slug := range parseEventChunk(chunk) {
						if !seen[slug] {
							seen[slug] = true
							select {
							case slugCh <- slug:
							case <-stopCh:
								return
							}
						}
					}
				}
				if w.taskReporter != nil {
					w.taskReporter.TaskFinished(w.taskID, lastState)
				}
				return
			}

			select {
			case <-time.After(w.pollDelay):
			case <-stopCh:
				return
			}
		}
	}()

	return slugCh
}

func parseEventChunk(chunk []byte) []string {
	var slugs []string

	for _, line := range strings.Split(string(chunk), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		slug := parseErrandEventLine(line)
		if slug != "" {
			slugs = append(slugs, slug)
		}
	}

	return slugs
}

func parseErrandEventLine(line string) string {
	var ev taskEvent
	if err := json.Unmarshal([]byte(line), &ev); err != nil {
		return ""
	}

	if ev.Stage != "Running errand" || ev.State != "started" {
		return ""
	}

	return ParseInstanceSlug(ev.Task)
}

// ParseInstanceSlug extracts "group/uuid" from a task field like "group/uuid (idx)".
func ParseInstanceSlug(task string) string {
	task = strings.TrimSpace(task)
	if task == "" {
		return ""
	}

	// Strip trailing " (N)" index suffix if present
	if idx := strings.LastIndex(task, " ("); idx >= 0 {
		task = task[:idx]
	}

	if !strings.Contains(task, "/") {
		return ""
	}

	return task
}

func isTaskRunning(state string) bool {
	return state == "queued" || state == "processing" || state == "cancelling"
}
