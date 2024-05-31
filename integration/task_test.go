package integration_test

import (
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("task command", func() {
	It("streams task output", func() {
		directorCACert, director := buildHTTPSServer()
		defer director.Close()

		processing := ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/tasks/123"),
			ghttp.RespondWith(http.StatusOK, `{"id":123, "state":"processing"}`),
		)

		director.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/info"),
				ghttp.RespondWith(http.StatusOK, `{"user_authentication":{"type":"basic","options":{}}}`),
			),
			processing,
			processing,
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/tasks/123/output", "type=event"),
				ghttp.RespondWith(http.StatusRequestedRangeNotSatisfiable, "Byte range unsatisfiable\n"),
			),
			processing,
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/tasks/123/output", "type=event"),
				ghttp.RespondWith(http.StatusOK, `{}`+"\n"),
			),
			processing,
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/tasks/123/output", "type=event"),
				ghttp.RespondWith(http.StatusOK, `{"time":1503082451,"stage":"event-one`),
			),
			processing,
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/tasks/123/output", "type=event"),
				ghttp.RespondWith(http.StatusOK, ""),
			),
			processing,
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/tasks/123/output", "type=event"),
				ghttp.RespondWith(http.StatusOK, `","tags":[],"total":1,"task":"event-one-task","state":"started","progress":0}`+"\n{}\n"),
			),
			processing,
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/tasks/123/output", "type=event"),
				ghttp.RespondWith(http.StatusOK, ""),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/tasks/123"),
				ghttp.RespondWith(http.StatusOK, `{"id":123, "state":"done"}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/tasks/123/output", "type=event"),
				ghttp.RespondWith(http.StatusOK, `{"time":1503082451,"stage":"event-two","tags":[],"total":1,"task":"event-two-task","index":1,"state":"started","progress":0}`+"\n"),
			),
		)

		createAndExecCommand(cmdFactory, []string{"task", "123", "-e", director.URL(), "--ca-cert", directorCACert})

		output := strings.Join(ui.Blocks, "\n")
		Expect(output).To(ContainSubstring("event-one"))
		Expect(output).To(ContainSubstring("event-two"))
	})
})
