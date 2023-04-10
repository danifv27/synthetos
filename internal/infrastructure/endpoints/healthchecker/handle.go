package healthchecker

import (
	"context"
	"encoding/json"
	"net/http"
	"path"
	"sync"

	"fry.org/cmo/cli/internal/application/healthchecker"
)

// healthcheckHandler is a basic Healthchekcker implementation.
type healthcheckHandler struct {
	http.ServeMux
	checksMutex     sync.RWMutex
	livenessChecks  map[string]healthchecker.Check
	readinessChecks map[string]healthchecker.Check
}

// NewHealthchecker creates a new Healthchecker
func NewHealthchecker(root string) healthchecker.Healthchecker {
	h := &healthcheckHandler{
		livenessChecks:  make(map[string]healthchecker.Check),
		readinessChecks: make(map[string]healthchecker.Check),
	}
	h.Handle(path.Join(root, "/liveness"), http.HandlerFunc(h.LiveEndpoint))
	h.Handle(path.Join(root, "/readiness"), http.HandlerFunc(h.ReadyEndpoint))

	return h
}

func (s *healthcheckHandler) LiveEndpoint(w http.ResponseWriter, r *http.Request) {

	s.handle(w, r, s.livenessChecks)
}

func (s *healthcheckHandler) ReadyEndpoint(w http.ResponseWriter, r *http.Request) {

	s.handle(w, r, s.readinessChecks, s.livenessChecks)
}

func (s *healthcheckHandler) AddLivenessCheck(name string, check healthchecker.Check) {

	s.checksMutex.Lock()
	defer s.checksMutex.Unlock()
	s.livenessChecks[name] = check
}

func (s *healthcheckHandler) AddReadinessCheck(name string, check healthchecker.Check) {

	s.checksMutex.Lock()
	defer s.checksMutex.Unlock()
	s.readinessChecks[name] = check
}

func (s *healthcheckHandler) collectChecks(checks map[string]healthchecker.Check, resultsOut map[string]string, statusOut *int) {

	s.checksMutex.RLock()
	defer s.checksMutex.RUnlock()
	for name, check := range checks {
		if err := check(context.TODO()); err != nil {
			*statusOut = http.StatusServiceUnavailable
			resultsOut[name] = err.Error()
		} else {
			resultsOut[name] = "OK"
		}
	}
}

func (s *healthcheckHandler) handle(w http.ResponseWriter, r *http.Request, checks ...map[string]healthchecker.Check) {

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	checkResults := make(map[string]string)
	status := http.StatusOK
	for _, c := range checks {
		s.collectChecks(c, checkResults, &status)
	}

	// write out the response code and content type header
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	// unless ?full=1, return an empty body. Kubernetes only cares about the
	// HTTP status code, so we won't waste bytes on the full body.
	if r.URL.Query().Get("full") != "1" {
		w.Write([]byte("{}\n"))
		return
	}

	// otherwise, write the JSON body ignoring any encoding errors (which
	// shouldn't really be possible since we're encoding a map[string]string).
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(checkResults)
}
