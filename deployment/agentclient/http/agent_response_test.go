package http_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/http"
)

var _ = Describe("AgentResponse", func() {
	Describe("TaskResponse", func() {
		var agentTaskResponse TaskResponse

		Describe("ServerError", func() {
			BeforeEach(func() {
				agentResponseJSON := `{"exception":{"message":"fake-exception-message"}}`
				err := agentTaskResponse.Unmarshal([]byte(agentResponseJSON))
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns task id", func() {
				err := agentTaskResponse.ServerError()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Agent responded with error: fake-exception-message"))
			})
		})

		Describe("TaskID", func() {
			BeforeEach(func() {
				agentResponseJSON := `{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`
				err := agentTaskResponse.Unmarshal([]byte(agentResponseJSON))
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns task id", func() {
				taskID, err := agentTaskResponse.TaskID()
				Expect(err).ToNot(HaveOccurred())
				Expect(taskID).To(Equal("fake-agent-task-id"))
			})
		})

		Describe("TaskState", func() {
			Context("when task value is a map and has agent_task_id", func() {
				BeforeEach(func() {
					agentResponseJSON := `{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`
					err := agentTaskResponse.Unmarshal([]byte(agentResponseJSON))
					Expect(err).ToNot(HaveOccurred())
				})

				It("returns task state", func() {
					taskState, err := agentTaskResponse.TaskState()
					Expect(err).ToNot(HaveOccurred())
					Expect(taskState).To(Equal("running"))
				})
			})

			Context("when task value is a map and does not have agent_task_id", func() {
				BeforeEach(func() {
					agentResponseJSON := `{"value":{}}`
					err := agentTaskResponse.Unmarshal([]byte(agentResponseJSON))
					Expect(err).ToNot(HaveOccurred())
				})

				It("returns task state", func() {
					taskState, err := agentTaskResponse.TaskState()
					Expect(err).ToNot(HaveOccurred())
					Expect(taskState).To(Equal("finished"))
				})
			})

			Context("when task value is a string", func() {
				BeforeEach(func() {
					agentResponseJSON := `{"value":"stopped"}`
					err := agentTaskResponse.Unmarshal([]byte(agentResponseJSON))
					Expect(err).ToNot(HaveOccurred())
				})

				It("returns task state", func() {
					taskState, err := agentTaskResponse.TaskState()
					Expect(err).ToNot(HaveOccurred())
					Expect(taskState).To(Equal("finished"))
				})
			})
		})
	})
})
