package healthchecker

import (
	"context"
	"net/http"
)

// Check is a health/readiness check.
type Check func(ctx context.Context) error

// Handler is an http.Handler with additional methods that register health and
// readiness checks. It handles handle "/live" and "/ready" HTTP
// endpoints.
type Healthchecker interface {
	// The Handler is an http.Handler, so it can be exposed directly and handle
	// /live and /ready endpoints.
	http.Handler
	// AddLivenessCheck adds a check that indicates that this instance of the
	// application should be destroyed or restarted.
	AddLivenessCheck(name string, check Check)
	// AddReadinessCheck adds a check that indicates that this instance of the
	// application is currently unable to serve requests because of an upstream
	// or some transient failure.
	AddReadinessCheck(name string, check Check)
	// LiveEndpoint is the HTTP handler for just the /live endpoint.
	LiveEndpoint(http.ResponseWriter, *http.Request)
	// ReadyEndpoint is the HTTP handler for just the /ready endpoint.
	ReadyEndpoint(http.ResponseWriter, *http.Request)
}
