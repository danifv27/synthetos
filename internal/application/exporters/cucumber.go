package exporters

import (
	"net/http"
)

// Handler is an http.Handler with additional methods that register Prometheus endpoints.
type CucumberExporter interface {
	// The Handler is an http.Handler, so it can be exposed directly and handle endpoints.
	http.Handler
}
