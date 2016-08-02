package director_test

import (
	//"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	. "github.com/cloudfoundry/bosh-init/director"
	//fakedir "github.com/cloudfoundry/bosh-init/director/fakes"
)

var _ = Describe("Director", func() {
	var (
		director Director
		server   *ghttp.Server
	)

	BeforeEach(func() {
		director, server = BuildServer()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Events", func() {
		It("returns events", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/events", "state=processing,cancelling,queued&verbose=1"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[
	{
		"id": 165,
		"timestamp": 1440318199,
		"user": "admin",
		"action": "cleanup ssh",
		"object_type": "instance",
		"object_name": "reporter/e3fed6d4-a245-442d-9380-875aeb184951",
		"task: "303",
		"deployment": "test-bosh",
		"instance": "reporter/e3fed6d4-a245-442d-9380-875aeb184951",
		"context": {"user":"^bosh_zquoxju62$"}
	},
	{
		"id": 166,
		"timestamp": 1440318200,
		"user": "admin2",
		"action": "delete",
		"object_type": "vm",
		"object_name": "33f13ac6-94f5-4890-6fa7-1ce7750dc956",
		"task: "302",
		"deployment": "test-bosh-2",
		"instance": "compilation-617c5bbe-eef5-4215-a9d0-8e8688f560e8/f5667668-a16f-44d6-baf1-500dada0c406",
		"context": {}
	}
]`),
				),
			)

			events, err := director.Events(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(events).To(HaveLen(2))

			Expect(events[0].Id()).To(Equal(165))
			Expect(events[0].Timestamp()).To(Equal(time.Date(2015, time.August, 23, 8, 23, 19, 0, time.UTC)))
			Expect(events[0].User()).To(Equal("admin"))
			Expect(events[0].Action()).To(Equal("cleanup ssh"))
			Expect(events[0].ObjectType()).To(Equal("instance"))
			Expect(events[0].ObjectName()).To(Equal("reporter/e3fed6d4-a245-442d-9380-875aeb184951"))
			Expect(events[0].Task()).To(Equal("303"))
			Expect(events[0].Deployment()).To(Equal("test-bosh"))
			Expect(events[0].Instance()).To(Equal("reporter/e3fed6d4-a245-442d-9380-875aeb184951"))
			Expect(events[0].Context()).To(Equal("{'user':'^bosh_zquoxju62$'}"))

			Expect(events[1].Id()).To(Equal(166))
			Expect(events[1].Timestamp()).To(Equal(time.Date(2015, time.August, 23, 8, 23, 20, 0, time.UTC)))
			Expect(events[1].User()).To(Equal("admin2"))
			Expect(events[1].Action()).To(Equal("delete"))
			Expect(events[1].ObjectType()).To(Equal("vm"))
			Expect(events[1].ObjectName()).To(Equal("33f13ac6-94f5-4890-6fa7-1ce7750dc956"))
			Expect(events[1].Task()).To(Equal("302"))
			Expect(events[1].Deployment()).To(Equal("test-bosh-2"))
			Expect(events[1].Instance()).To(Equal("compilation-617c5bbe-eef5-4215-a9d0-8e8688f560e8/f5667668-a16f-44d6-baf1-500dada0c406"))
			Expect(events[1].Context()).To(Equal(""))
		})

		It("includes all tasks when requested", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/tasks", ""),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, "[]"),
				),
			)

			_, err := director.CurrentTasks(true)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if response is non-200", func() {
			AppendBadRequest(ghttp.VerifyRequest("GET", "/tasks"), server)

			_, err := director.CurrentTasks(false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Finding current tasks: Director responded with non-successful status code"))
		})

		It("returns error if response cannot be unmarshalled", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/tasks"),
					ghttp.RespondWith(http.StatusOK, ``),
				),
			)

			_, err := director.CurrentTasks(false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Finding current tasks: Unmarshaling Director response"))
		})
	})

})

var _ = Describe("Event", func() {
	var (
		director Director
		//event     Event
		server   *ghttp.Server
	)

	BeforeEach(func() {
		director, server = BuildServer()

		//var err error

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/events/123"),
				ghttp.RespondWith(http.StatusOK, `{"id":123}`),
			),
		)

		////event, err = director.FindEvent(123)
		//Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		server.Close()
	})

	//Describe("TaskOutput", func() {
	//	var (
	//		reporter *fakedir.FakeTaskReporter
	//	)
	//
	//	BeforeEach(func() {
	//		reporter = &fakedir.FakeTaskReporter{}
	//	})
	//
	//	types := map[string]func(Task) error{
	//		"event":  func(t Task) error { return task.EventOutput(reporter) },
	//		"cpi":    func(t Task) error { return task.CPIOutput(reporter) },
	//		"debug":  func(t Task) error { return task.DebugOutput(reporter) },
	//		"result": func(t Task) error { return task.ResultOutput(reporter) },
	//		"raw":    func(t Task) error { return task.RawOutput(reporter) },
	//	}
	//
	//	for type_, typeFunc := range types {
	//		type_ := type_
	//		typeFunc := typeFunc
	//
	//		It(fmt.Sprintf("reports task '%s' output", type_), func() {
	//			server.AppendHandlers(
	//				ghttp.CombineHandlers(
	//					ghttp.VerifyRequest("GET", "/tasks/123"),
	//					ghttp.VerifyBasicAuth("username", "password"),
	//					ghttp.RespondWith(http.StatusOK, `{"id":123, "state":"done"}`),
	//				),
	//				ghttp.CombineHandlers(
	//					ghttp.VerifyRequest("GET", "/tasks/123/output", fmt.Sprintf("type=%s", type_)),
	//					ghttp.VerifyBasicAuth("username", "password"),
	//					ghttp.RespondWith(http.StatusOK, "chunk"),
	//				),
	//			)
	//
	//			Expect(typeFunc(task)).ToNot(HaveOccurred())
	//
	//			taskID := reporter.TaskStartedArgsForCall(0)
	//			Expect(taskID).To(Equal(123))
	//
	//			taskID, chunk := reporter.TaskOutputChunkArgsForCall(0)
	//			Expect(taskID).To(Equal(123))
	//			Expect(chunk).To(Equal([]byte("chunk")))
	//
	//			taskID, state := reporter.TaskFinishedArgsForCall(0)
	//			Expect(taskID).To(Equal(123))
	//			Expect(state).To(Equal("done"))
	//		})
	//
	//		It(fmt.Sprintf("returns error if task '%s' response is non-200", type_), func() {
	//			AppendBadRequest(ghttp.VerifyRequest("GET", "/tasks/123"), server)
	//
	//			err := typeFunc(task)
	//			Expect(err).To(HaveOccurred())
	//			Expect(err.Error()).To(ContainSubstring("Capturing task '123' output"))
	//		})
	//	}
	//})
	//
	//Describe("Cancel", func() {
	//	It("cancels task", func() {
	//		server.AppendHandlers(
	//			ghttp.CombineHandlers(
	//				ghttp.VerifyRequest("DELETE", "/task/123"),
	//				ghttp.RespondWith(http.StatusOK, ``),
	//			),
	//		)
	//
	//		Expect(task.Cancel()).ToNot(HaveOccurred())
	//	})
	//
	//	It("returns error if response is non-200", func() {
	//		AppendBadRequest(ghttp.VerifyRequest("DELETE", "/task/123"), server)
	//
	//		err := task.Cancel()
	//		Expect(err).To(HaveOccurred())
	//		Expect(err.Error()).To(ContainSubstring(
	//			"Cancelling task '123': Director responded with non-successful status code"))
	//	})
	//})
})
