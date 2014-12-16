package agentclient_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
)

var _ = Describe("AgentResponse", func() {
	Describe("TaskResponse", func() {
		var agentTaskResponse TaskResponse

		Describe("GetException", func() {
			BeforeEach(func() {
				agentResponseJSON := `{"exception":{"message":"fake-exception-message"}}`
				err := json.Unmarshal([]byte(agentResponseJSON), &agentTaskResponse)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns task id", func() {
				exceptionResponse := agentTaskResponse.GetException()
				Expect(exceptionResponse.Message).To(Equal("fake-exception-message"))
			})
		})

		Describe("TaskID", func() {
			BeforeEach(func() {
				agentResponseJSON := `{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`
				err := json.Unmarshal([]byte(agentResponseJSON), &agentTaskResponse)
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
					err := json.Unmarshal([]byte(agentResponseJSON), &agentTaskResponse)
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
					err := json.Unmarshal([]byte(agentResponseJSON), &agentTaskResponse)
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
					err := json.Unmarshal([]byte(agentResponseJSON), &agentTaskResponse)
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
