package admin

import (
	"fmt"
	"net/http"
)

// ServiceInfo represents information about a service that operators may find
// useful.
type ServiceInfo struct {
	Name        string
	Description string
	Team        string
}

// WithInfo configures the information endpoint to present information about
// the service.
func WithInfo(info ServiceInfo) OptionFunc {
	return func(s *server) {
		s.handler.HandleFunc("/info", func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(w, "Name: %s\n", info.Name)
			fmt.Fprintf(w, "Description: %s\n", info.Description)
			fmt.Fprintf(w, "Team: %s\n", info.Team)
		})
	}
}
