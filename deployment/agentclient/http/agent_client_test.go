package http_test

import (
	"encoding/json"
	"net/http"

	. "github.com/cloudfoundry/bosh-init/deployment/agentclient/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmagentclient "github.com/cloudfoundry/bosh-init/deployment/agentclient"
	bmas "github.com/cloudfoundry/bosh-init/deployment/applyspec"
	fakebmhttpclient "github.com/cloudfoundry/bosh-init/deployment/httpclient/fakes"
)

var _ = Describe("AgentClient", func() {
	var (
		fakeHTTPClient *fakebmhttpclient.FakeHTTPClient
		agentClient    bmagentclient.AgentClient
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fakeHTTPClient = fakebmhttpclient.NewFakeHTTPClient()
		agentClient = NewAgentClient("http://localhost:6305", "fake-uuid", 0, fakeHTTPClient, logger)
	})

	Describe("Ping", func() {
		Context("when agent responds with a value", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior(`{"value":"pong"}`, 200, nil)
			})

			It("makes a POST request to the endpoint", func() {
				_, err := agentClient.Ping()
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeHTTPClient.PostInputs).To(HaveLen(1))
				Expect(fakeHTTPClient.PostInputs[0].Endpoint).To(Equal("http://localhost:6305/agent"))

				var request AgentRequestMessage
				err = json.Unmarshal(fakeHTTPClient.PostInputs[0].Payload, &request)
				Expect(err).ToNot(HaveOccurred())

				Expect(request).To(Equal(AgentRequestMessage{
					Method:    "ping",
					Arguments: []interface{}{},
					ReplyTo:   "fake-uuid",
				}))
			})

			It("returns the value", func() {
				responseValue, err := agentClient.Ping()
				Expect(err).ToNot(HaveOccurred())
				Expect(responseValue).To(Equal("pong"))
			})
		})

		Context("when agent does not respond with 200", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior("", http.StatusInternalServerError, nil)
			})

			It("returns an error", func() {
				_, err := agentClient.Ping()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("status code: 500"))
			})
		})

		Context("when agent responds with exception", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior(`{"exception":{"message":"bad request"}}`, 200, nil)
			})

			It("returns an error", func() {
				_, err := agentClient.Ping()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("bad request"))
			})
		})
	})

	Describe("Stop", func() {
		Context("when agent responds with a value", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`, 200, nil)
				fakeHTTPClient.SetPostBehavior(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`, 200, nil)
				fakeHTTPClient.SetPostBehavior(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`, 200, nil)
				fakeHTTPClient.SetPostBehavior(`{"value":"stopped"}`, 200, nil)
			})

			It("makes a POST request to the endpoint", func() {
				err := agentClient.Stop()
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeHTTPClient.PostInputs).To(HaveLen(4))
				Expect(fakeHTTPClient.PostInputs[0].Endpoint).To(Equal("http://localhost:6305/agent"))

				var request AgentRequestMessage
				err = json.Unmarshal(fakeHTTPClient.PostInputs[0].Payload, &request)
				Expect(err).ToNot(HaveOccurred())

				Expect(request).To(Equal(AgentRequestMessage{
					Method:    "stop",
					Arguments: []interface{}{},
					ReplyTo:   "fake-uuid",
				}))
			})

			It("waits for the task to be finished", func() {
				err := agentClient.Stop()
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeHTTPClient.PostInputs).To(HaveLen(4))
				Expect(fakeHTTPClient.PostInputs[1].Endpoint).To(Equal("http://localhost:6305/agent"))

				var request AgentRequestMessage
				err = json.Unmarshal(fakeHTTPClient.PostInputs[1].Payload, &request)
				Expect(err).ToNot(HaveOccurred())

				Expect(request).To(Equal(AgentRequestMessage{
					Method:    "get_task",
					Arguments: []interface{}{"fake-agent-task-id"},
					ReplyTo:   "fake-uuid",
				}))
			})
		})

		Context("when agent does not respond with 200", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior("", http.StatusInternalServerError, nil)
			})

			It("returns an error", func() {
				err := agentClient.Stop()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("status code: 500"))
			})
		})

		Context("when agent responds with exception", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior(`{"exception":{"message":"bad request"}}`, 200, nil)
			})

			It("returns an error", func() {
				err := agentClient.Stop()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("bad request"))
			})
		})
	})

	Describe("Apply", func() {
		var (
			specJSON []byte
			spec     bmas.ApplySpec
		)

		BeforeEach(func() {
			spec = bmas.ApplySpec{
				Deployment: "fake-deployment-name",
			}
			var err error
			specJSON, err = json.Marshal(spec)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when agent responds with a value", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`, 200, nil)
				fakeHTTPClient.SetPostBehavior(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`, 200, nil)
				fakeHTTPClient.SetPostBehavior(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`, 200, nil)
				fakeHTTPClient.SetPostBehavior(`{"value":"stopped"}`, 200, nil)
			})

			It("makes a POST request to the endpoint", func() {
				err := agentClient.Apply(spec)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeHTTPClient.PostInputs).To(HaveLen(4))
				Expect(fakeHTTPClient.PostInputs[0].Endpoint).To(Equal("http://localhost:6305/agent"))

				var request AgentRequestMessage
				err = json.Unmarshal(fakeHTTPClient.PostInputs[0].Payload, &request)
				Expect(err).ToNot(HaveOccurred())

				var specArgument interface{}
				err = json.Unmarshal(specJSON, &specArgument)
				Expect(err).ToNot(HaveOccurred())

				Expect(request).To(Equal(AgentRequestMessage{
					Method:    "apply",
					Arguments: []interface{}{specArgument},
					ReplyTo:   "fake-uuid",
				}))
			})

			It("waits for the task to be finished", func() {
				err := agentClient.Apply(spec)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeHTTPClient.PostInputs).To(HaveLen(4))
				Expect(fakeHTTPClient.PostInputs[1].Endpoint).To(Equal("http://localhost:6305/agent"))

				var request AgentRequestMessage
				err = json.Unmarshal(fakeHTTPClient.PostInputs[1].Payload, &request)
				Expect(err).ToNot(HaveOccurred())

				Expect(request).To(Equal(AgentRequestMessage{
					Method:    "get_task",
					Arguments: []interface{}{"fake-agent-task-id"},
					ReplyTo:   "fake-uuid",
				}))
			})
		})

		Context("when agent does not respond with 200", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior("", http.StatusInternalServerError, nil)
			})

			It("returns an error", func() {
				err := agentClient.Apply(spec)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("status code: 500"))
			})
		})

		Context("when agent responds with exception", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior(`{"exception":{"message":"bad request"}}`, 200, nil)
			})

			It("returns an error", func() {
				err := agentClient.Apply(spec)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("bad request"))
			})
		})
	})

	Describe("Start", func() {
		Context("when agent responds with a value", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior(`{"value":"started"}`, 200, nil)
			})

			It("makes a POST request to the endpoint", func() {
				err := agentClient.Start()
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeHTTPClient.PostInputs).To(HaveLen(1))
				Expect(fakeHTTPClient.PostInputs[0].Endpoint).To(Equal("http://localhost:6305/agent"))

				var request AgentRequestMessage
				err = json.Unmarshal(fakeHTTPClient.PostInputs[0].Payload, &request)
				Expect(err).ToNot(HaveOccurred())

				Expect(request).To(Equal(AgentRequestMessage{
					Method:    "start",
					Arguments: []interface{}{},
					ReplyTo:   "fake-uuid",
				}))
			})
		})

		Context("when agent does not respond with 200", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior("", http.StatusInternalServerError, nil)
			})

			It("returns an error", func() {
				err := agentClient.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("status code: 500"))
			})
		})

		Context("when agent responds with exception", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior(`{"exception":{"message":"bad request"}}`, 200, nil)
			})

			It("returns an error", func() {
				err := agentClient.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("bad request"))
			})
		})
	})

	Describe("GetState", func() {
		Context("when agent responds with a value", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior(`{"value":{"job_state":"running"}}`, 200, nil)
			})

			It("makes a POST request to the endpoint", func() {
				stateResponse, err := agentClient.GetState()
				Expect(err).ToNot(HaveOccurred())
				Expect(stateResponse).To(Equal(bmagentclient.AgentState{JobState: "running"}))

				Expect(fakeHTTPClient.PostInputs).To(HaveLen(1))
				Expect(fakeHTTPClient.PostInputs[0].Endpoint).To(Equal("http://localhost:6305/agent"))

				var request AgentRequestMessage
				err = json.Unmarshal(fakeHTTPClient.PostInputs[0].Payload, &request)
				Expect(err).ToNot(HaveOccurred())

				Expect(request).To(Equal(AgentRequestMessage{
					Method:    "get_state",
					Arguments: []interface{}{},
					ReplyTo:   "fake-uuid",
				}))
			})
		})

		Context("when agent does not respond with 200", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior("", http.StatusInternalServerError, nil)
			})

			It("returns an error", func() {
				stateResponse, err := agentClient.GetState()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("status code: 500"))
				Expect(stateResponse).To(Equal(bmagentclient.AgentState{}))
			})
		})

		Context("when agent responds with exception", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior(`{"exception":{"message":"bad request"}}`, 200, nil)
			})

			It("returns an error", func() {
				stateResponse, err := agentClient.GetState()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("bad request"))
				Expect(stateResponse).To(Equal(bmagentclient.AgentState{}))
			})
		})
	})

	Describe("MountDisk", func() {
		Context("when agent responds with a value", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`, 200, nil)
				fakeHTTPClient.SetPostBehavior(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`, 200, nil)
				fakeHTTPClient.SetPostBehavior(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`, 200, nil)
				fakeHTTPClient.SetPostBehavior(`{"value":{}}`, 200, nil)
			})

			It("makes a POST request to the endpoint", func() {
				err := agentClient.MountDisk("fake-disk-cid")
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeHTTPClient.PostInputs).To(HaveLen(4))
				Expect(fakeHTTPClient.PostInputs[0].Endpoint).To(Equal("http://localhost:6305/agent"))

				var request AgentRequestMessage
				err = json.Unmarshal(fakeHTTPClient.PostInputs[0].Payload, &request)
				Expect(err).ToNot(HaveOccurred())

				Expect(request).To(Equal(AgentRequestMessage{
					Method:    "mount_disk",
					Arguments: []interface{}{"fake-disk-cid"},
					ReplyTo:   "fake-uuid",
				}))
			})

			It("waits for the task to be finished", func() {
				err := agentClient.MountDisk("fake-disk-cid")
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeHTTPClient.PostInputs).To(HaveLen(4))
				Expect(fakeHTTPClient.PostInputs[1].Endpoint).To(Equal("http://localhost:6305/agent"))

				var request AgentRequestMessage
				err = json.Unmarshal(fakeHTTPClient.PostInputs[1].Payload, &request)
				Expect(err).ToNot(HaveOccurred())

				Expect(request).To(Equal(AgentRequestMessage{
					Method:    "get_task",
					Arguments: []interface{}{"fake-agent-task-id"},
					ReplyTo:   "fake-uuid",
				}))
			})
		})

		Describe("UnmountDisk", func() {
			Context("when agent responds with a value", func() {
				BeforeEach(func() {
					fakeHTTPClient.SetPostBehavior(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`, 200, nil)
					fakeHTTPClient.SetPostBehavior(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`, 200, nil)
					fakeHTTPClient.SetPostBehavior(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`, 200, nil)
					fakeHTTPClient.SetPostBehavior(`{"value":{}}`, 200, nil)
				})

				It("makes a POST request to the endpoint", func() {
					err := agentClient.UnmountDisk("fake-disk-cid")
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeHTTPClient.PostInputs).To(HaveLen(4))
					Expect(fakeHTTPClient.PostInputs[0].Endpoint).To(Equal("http://localhost:6305/agent"))

					var request AgentRequestMessage
					err = json.Unmarshal(fakeHTTPClient.PostInputs[0].Payload, &request)
					Expect(err).ToNot(HaveOccurred())

					Expect(request).To(Equal(AgentRequestMessage{
						Method:    "unmount_disk",
						Arguments: []interface{}{"fake-disk-cid"},
						ReplyTo:   "fake-uuid",
					}))
				})

				It("waits for the task to be finished", func() {
					err := agentClient.UnmountDisk("fake-disk-cid")
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeHTTPClient.PostInputs).To(HaveLen(4))
					Expect(fakeHTTPClient.PostInputs[1].Endpoint).To(Equal("http://localhost:6305/agent"))

					var request AgentRequestMessage
					err = json.Unmarshal(fakeHTTPClient.PostInputs[1].Payload, &request)
					Expect(err).ToNot(HaveOccurred())

					Expect(request).To(Equal(AgentRequestMessage{
						Method:    "get_task",
						Arguments: []interface{}{"fake-agent-task-id"},
						ReplyTo:   "fake-uuid",
					}))
				})
			})
		})

		Context("when agent does not respond with 200", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior("", http.StatusInternalServerError, nil)
			})

			It("returns an error", func() {
				err := agentClient.MountDisk("fake-disk-cid")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("status code: 500"))
			})
		})

		Context("when agent responds with exception", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior(`{"exception":{"message":"bad request"}}`, 200, nil)
			})

			It("returns an error", func() {
				err := agentClient.MountDisk("fake-disk-cid")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("bad request"))
			})
		})
	})

	Describe("ListDisk", func() {
		Context("when agent responds with a value", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior(`{"value":["fake-disk-1", "fake-disk-2"]}`, 200, nil)
			})

			It("makes a POST request to the endpoint", func() {
				_, err := agentClient.ListDisk()
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeHTTPClient.PostInputs).To(HaveLen(1))
				Expect(fakeHTTPClient.PostInputs[0].Endpoint).To(Equal("http://localhost:6305/agent"))

				var request AgentRequestMessage
				err = json.Unmarshal(fakeHTTPClient.PostInputs[0].Payload, &request)
				Expect(err).ToNot(HaveOccurred())

				Expect(request).To(Equal(AgentRequestMessage{
					Method:    "list_disk",
					Arguments: []interface{}{},
					ReplyTo:   "fake-uuid",
				}))
			})

			It("returns disks", func() {
				disks, err := agentClient.ListDisk()
				Expect(err).ToNot(HaveOccurred())
				Expect(disks).To(Equal([]string{"fake-disk-1", "fake-disk-2"}))
			})
		})

		Context("when agent does not respond with 200", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior("", http.StatusInternalServerError, nil)
			})

			It("returns an error", func() {
				_, err := agentClient.ListDisk()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("status code: 500"))
			})
		})

		Context("when agent responds with exception", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior(`{"exception":{"message":"bad request"}}`, 200, nil)
			})

			It("returns an error", func() {
				_, err := agentClient.ListDisk()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("bad request"))
			})
		})
	})

	Describe("MigrateDisk", func() {
		Context("when agent responds with a value", func() {
			BeforeEach(func() {
				fakeHTTPClient.SetPostBehavior(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`, 200, nil)
				fakeHTTPClient.SetPostBehavior(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`, 200, nil)
				fakeHTTPClient.SetPostBehavior(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`, 200, nil)
				fakeHTTPClient.SetPostBehavior(`{"value":{}}`, 200, nil)
			})

			It("makes a POST request to the endpoint", func() {
				err := agentClient.MigrateDisk()
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeHTTPClient.PostInputs).To(HaveLen(4))
				Expect(fakeHTTPClient.PostInputs[0].Endpoint).To(Equal("http://localhost:6305/agent"))

				var request AgentRequestMessage
				err = json.Unmarshal(fakeHTTPClient.PostInputs[0].Payload, &request)
				Expect(err).ToNot(HaveOccurred())

				Expect(request).To(Equal(AgentRequestMessage{
					Method:    "migrate_disk",
					Arguments: []interface{}{},
					ReplyTo:   "fake-uuid",
				}))
			})

			It("waits for the task to be finished", func() {
				err := agentClient.MigrateDisk()
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeHTTPClient.PostInputs).To(HaveLen(4))
				Expect(fakeHTTPClient.PostInputs[1].Endpoint).To(Equal("http://localhost:6305/agent"))

				var request AgentRequestMessage
				err = json.Unmarshal(fakeHTTPClient.PostInputs[1].Payload, &request)
				Expect(err).ToNot(HaveOccurred())

				Expect(request).To(Equal(AgentRequestMessage{
					Method:    "get_task",
					Arguments: []interface{}{"fake-agent-task-id"},
					ReplyTo:   "fake-uuid",
				}))
			})
		})
	})

	Describe("CompilePackage", func() {
		BeforeEach(func() {
			fakeHTTPClient.SetPostBehavior(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`, 200, nil)
			fakeHTTPClient.SetPostBehavior(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`, 200, nil)
			fakeHTTPClient.SetPostBehavior(`{"value":{"agent_task_id":"fake-agent-task-id","state":"running"}}`, 200, nil)
			fakeHTTPClient.SetPostBehavior(`{
	"value": {
		"result": {
			"sha1": "fake-compiled-package-sha1",
			"blobstore_id": "fake-compiled-package-blobstore-id"
		}
	}
}
`, 200, nil)
		})

		It("makes a compile_package request and waits for the task to be done", func() {
			packageSource := bmagentclient.BlobRef{
				Name:        "fake-package-name",
				Version:     "fake-package-version",
				SHA1:        "fake-package-sha1",
				BlobstoreID: "fake-package-blobstore-id",
			}
			dependencies := []bmagentclient.BlobRef{
				{
					Name:        "fake-compiled-package-dep-name",
					Version:     "fake-compiled-package-dep-version",
					SHA1:        "fake-compiled-package-dep-sha1",
					BlobstoreID: "fake-compiled-package-dep-blobstore-id",
				},
			}
			_, err := agentClient.CompilePackage(packageSource, dependencies)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeHTTPClient.PostInputs).To(HaveLen(4))
			Expect(fakeHTTPClient.PostInputs[0].Endpoint).To(Equal("http://localhost:6305/agent"))

			var request AgentRequestMessage
			err = json.Unmarshal(fakeHTTPClient.PostInputs[0].Payload, &request)
			Expect(err).ToNot(HaveOccurred())

			Expect(request).To(Equal(AgentRequestMessage{
				Method: "compile_package",
				Arguments: []interface{}{
					"fake-package-blobstore-id",
					"fake-package-sha1",
					"fake-package-name",
					"fake-package-version",
					map[string]interface{}{
						"fake-compiled-package-dep-name": map[string]interface{}{
							"name":         "fake-compiled-package-dep-name",
							"version":      "fake-compiled-package-dep-version",
							"sha1":         "fake-compiled-package-dep-sha1",
							"blobstore_id": "fake-compiled-package-dep-blobstore-id",
						},
					},
				},
				ReplyTo: "fake-uuid",
			}))
		})
	})
})
