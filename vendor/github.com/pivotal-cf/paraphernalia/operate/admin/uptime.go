package admin

import (
	"fmt"
	"net/http"
	"time"
)

var bootTime = time.Now()

// WithUptime adds a handler to the administrative endpoint that returns the
// uptime of the service. It is registered at /uptime.
func WithUptime() OptionFunc {
	return func(s *server) {
		s.handler.HandleFunc("/uptime", func(w http.ResponseWriter, req *http.Request) {
			uptime := time.Now().Sub(bootTime)
			fmt.Fprintf(w, "%s", uptime)
		})
	}
}
