package director_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	. "github.com/cloudfoundry/bosh-init/director"
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

	FDescribe("Events", func() {
		blankOpts := make(map[string]interface{})
		It("returns events", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/events"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[
	{
		"id": 1,
		"timestamp": 1440318199,
		"user": "admin",
		"action": "cleanup ssh",
		"objectType": "instance",
		"objectName": "33d",
		"task": "303",
		"deployment": "test-bosh",
		"instance": "reporter/e",
		"context": {"user":"bosh_z$"}
	},
	{
		"id": 2,
		"timestamp": 1440318200,
		"user": "admin2",
		"action": "delete",
		"objectType": "vm",
		"objectName": "33f",
		"task": "302",
		"deployment": "test-bosh-2",
		"instance": "compilation-6",
		"context": {}
	}
]`),
				),
			)

			events, err := director.Events(blankOpts)

			Expect(err).ToNot(HaveOccurred())
			Expect(events).To(HaveLen(2))

			expectEvent1(events)
			expectEvent2(events)
		})

		It("filters events based on 'before-id' option", func() {
			opts := make(map[string]interface{})
			opts["before-id"] = 1
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/events", "before-id=1"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[
					{
		"id": 1,
		"timestamp": 1440318199,
		"user": "admin",
		"action": "cleanup ssh",
		"objectType": "instance",
		"objectName": "33d",
		"task": "303",
		"deployment": "test-bosh",
		"instance": "reporter/e",
		"context": {"user":"bosh_z$"}
	}
					]`),
				),
			)

			events, err := director.Events(opts)

			Expect(events).To(HaveLen(1))

			expectEvent1(events)

			Expect(err).ToNot(HaveOccurred())
		})

		It("filters events based on 'before' option", func() {
			opts := make(map[string]interface{})
			opts["before"] = time.Date(2015, time.August, 23, 8, 23, 19, 0, time.UTC)
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/events", "before=1440318199"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[
					{
		"id": 1,
		"timestamp": 1440318199,
		"user": "admin",
		"action": "cleanup ssh",
		"objectType": "instance",
		"objectName": "33d",
		"task": "303",
		"deployment": "test-bosh",
		"instance": "reporter/e",
		"context": {"user":"bosh_z$"}
	}
					]`),
				),
			)

			events, err := director.Events(opts)

			Expect(events).To(HaveLen(1))

			expectEvent1(events)

			Expect(err).ToNot(HaveOccurred())
		})

		It("filters events based on 'after' option", func() {
			opts := make(map[string]interface{})
			opts["after"] = time.Date(2015, time.August, 23, 8, 23, 20, 0, time.UTC)
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/events", "after=1440318200"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[

	{
		"id": 2,
		"timestamp": 1440318200,
		"user": "admin2",
		"action": "delete",
		"objectType": "vm",
		"objectName": "33f",
		"task": "302",
		"deployment": "test-bosh-2",
		"instance": "compilation-6",
		"context": {}
	}
					]`),
				),
			)

			events, err := director.Events(opts)

			Expect(events).To(HaveLen(1))

			expectEvent2(events)

			Expect(err).ToNot(HaveOccurred())
		})

		It("filters events based on 'deployment' option", func() {
			opts := make(map[string]interface{})
			opts["deployment"] = "test-bosh-2"
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/events", "deployment=test-bosh-2"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[
	{
		"id": 2,
		"timestamp": 1440318200,
		"user": "admin2",
		"action": "delete",
		"objectType": "vm",
		"objectName": "33f",
		"task": "302",
		"deployment": "test-bosh-2",
		"instance": "compilation-6",
		"context": {}
	}
					]`),
				),
			)

			events, err := director.Events(opts)

			Expect(events).To(HaveLen(1))

			expectEvent2(events)

			Expect(err).ToNot(HaveOccurred())
		})

		It("filters events based on 'task' option", func() {
			opts := make(map[string]interface{})
			opts["task"] = "303"
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/events", "task=303"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[
					{
		"id": 1,
		"timestamp": 1440318199,
		"user": "admin",
		"action": "cleanup ssh",
		"objectType": "instance",
		"objectName": "33d",
		"task": "303",
		"deployment": "test-bosh",
		"instance": "reporter/e",
		"context": {"user":"bosh_z$"}
	}
					]`),
				),
			)

			events, err := director.Events(opts)

			Expect(events).To(HaveLen(1))

			expectEvent1(events)

			Expect(err).ToNot(HaveOccurred())
		})

		It("filters events based on 'instance' option", func() {
			opts := make(map[string]interface{})
			opts["instance"] = "compilation-6"
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/events", "instance=compilation-6"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[
	{
		"id": 2,
		"timestamp": 1440318200,
		"user": "admin2",
		"action": "delete",
		"objectType": "vm",
		"objectName": "33f",
		"task": "302",
		"deployment": "test-bosh-2",
		"instance": "compilation-6",
		"context": {}
	}
					]`),
				),
			)

			events, err := director.Events(opts)

			Expect(events).To(HaveLen(1))

			expectEvent2(events)

			Expect(err).ToNot(HaveOccurred())
		})

		It("returns a single event based on multiple options", func() {
			opts := make(map[string]interface{})
			opts["instance"] = "compilation-6"
			opts["deployment"] = "test-bosh-2"
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/events", "instance=compilation-6&deployment=test-bosh-2"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[
	{
		"id": 2,
		"timestamp": 1440318200,
		"user": "admin2",
		"action": "delete",
		"objectType": "vm",
		"objectName": "33f",
		"task": "302",
		"deployment": "test-bosh-2",
		"instance": "compilation-6",
		"context": {}
	}
					]`),
				),
			)

			events, err := director.Events(opts)

			Expect(events).To(HaveLen(1))

			expectEvent2(events)

			Expect(err).ToNot(HaveOccurred())
		})

		It("returns no events based on multiple options", func() {
			opts := make(map[string]interface{})
			opts["instance"] = "compilation-6"
			opts["deployment"] = "test-bosh"
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/events"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, "[]"),
				),
			)

			events, err := director.Events(opts)

			Expect(events).To(HaveLen(0))

			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if response is non-200", func() {
			AppendBadRequest(ghttp.VerifyRequest("GET", "/events"), server)

			_, err := director.Events(blankOpts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Finding events: Director responded with non-successful status code"))
		})

		It("returns error if response cannot be unmarshalled", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/events"),
					ghttp.RespondWith(http.StatusOK, ``),
				),
			)

			_, err := director.Events(blankOpts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Finding events: Unmarshaling Director response"))
		})
	})

})

func expectEvent1(events []Event) {
	Expect(events[0].Id()).To(Equal(1))
	Expect(events[0].Timestamp()).To(Equal(time.Date(2015, time.August, 23, 8, 23, 19, 0, time.UTC)))
	Expect(events[0].User()).To(Equal("admin"))
	Expect(events[0].Action()).To(Equal("cleanup ssh"))
	Expect(events[0].ObjectType()).To(Equal("instance"))
	Expect(events[0].ObjectName()).To(Equal("33d"))
	Expect(events[0].Task()).To(Equal("303"))
	Expect(events[0].Deployment()).To(Equal("test-bosh"))
	Expect(events[0].Instance()).To(Equal("reporter/e"))

	context := make(map[string]interface{})
	context["user"] = "bosh_z$"

	Expect(events[0].Context()).To(Equal(context))
}

func expectEvent2(events []Event) {
	i := len(events) - 1
	Expect(events[i].Id()).To(Equal(2))
	Expect(events[i].Timestamp()).To(Equal(time.Date(2015, time.August, 23, 8, 23, 20, 0, time.UTC)))
	Expect(events[i].User()).To(Equal("admin2"))
	Expect(events[i].Action()).To(Equal("delete"))
	Expect(events[i].ObjectType()).To(Equal("vm"))
	Expect(events[i].ObjectName()).To(Equal("33f"))
	Expect(events[i].Task()).To(Equal("302"))
	Expect(events[i].Deployment()).To(Equal("test-bosh-2"))
	Expect(events[i].Instance()).To(Equal("compilation-6"))

	blankContext := make(map[string]interface{})

	Expect(events[i].Context()).To(Equal(blankContext))
}
