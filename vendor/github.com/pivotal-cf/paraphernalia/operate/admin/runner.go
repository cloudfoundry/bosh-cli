// Package admin simplifies the setup of various administrative endpoints such
// as healthchecks, debugging, and service information.
package admin

import (
	"net"
	"net/http"
	"net/http/pprof"
	"os"

	"golang.org/x/net/trace"

	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/http_server"
)

// OptionFunc is used to configure the runner.
type OptionFunc func(*server)

// Runner builds and returns a runner that, when run, will start an HTTP server
// listening on the specified port that provides the Go pprof endpoints from
// the standard library along with the additional trace endpoints.
func Runner(port string, options ...OptionFunc) ifrit.Runner {
	server := &server{
		port:    port,
		handler: defaultDebugEndpoints(),
	}

	for _, option := range options {
		option(server)
	}

	return server
}

type server struct {
	port    string
	handler *http.ServeMux
}

// Run starts the server.
func (s *server) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	address := net.JoinHostPort("localhost", s.port)

	return http_server.New(address, s.handler).Run(signals, ready)
}

func defaultDebugEndpoints() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	mux.HandleFunc("/debug/requests", func(w http.ResponseWriter, req *http.Request) {
		any, sensitive := trace.AuthRequest(req)
		if !any {
			http.Error(w, "not allowed", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		trace.Render(w, req, sensitive)
	})

	mux.HandleFunc("/debug/events", func(w http.ResponseWriter, req *http.Request) {
		any, sensitive := trace.AuthRequest(req)
		if !any {
			http.Error(w, "not allowed", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		trace.RenderEvents(w, req, sensitive)
	})

	return mux
}
