package exporters

import (
	"errors"
	"net/http"
	"path"

	"fry.org/cmo/cli/internal/application/exporters"
	"github.com/speijnik/go-errortree"
)

// cucumberHandler is a basic Healthchekcker implementation.
type cucumberHandler struct {
	http.ServeMux
}

// NewCucumberExporter creates a new CucumberExporter
func NewCucumberExporter(opts ...ExporterOption) (exporters.CucumberExporter, error) {
	var rcerror error

	h := cucumberHandler{}
	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(&h); err != nil {
			return nil, errortree.Add(rcerror, "NewCucumberExporter", err)
		}
	}

	return &h, nil
}

// WithOptions
func WithCucumberOptions(c *exporters.CucumberExporter, opts ...ExporterOption) error {
	var rcerror error

	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(c); err != nil {
			return errortree.Add(rcerror, "WithOptions", err)
		}
	}

	return nil
}

func WithCucumberRootPrefix(prefix string) ExporterOption {

	return ExportOptionFn(func(i interface{}) error {
		var rcerror error
		var c *cucumberHandler
		var ok bool

		if c, ok = i.(*cucumberHandler); ok {
			c.Handle(path.Join(prefix, "/probes"), http.HandlerFunc(c.ProbesEndpoint))
			return nil
		}

		return errortree.Add(rcerror, "WithCucumberRootPrefix", errors.New("type mismatch, cucumberHandler expected"))
	})
}

func (c *cucumberHandler) ProbesEndpoint(w http.ResponseWriter, r *http.Request) {

	c.handle(w, r)
}

func (c *cucumberHandler) handle(w http.ResponseWriter, r *http.Request) {

	// start := time.Now()
	// registry := prometheus.NewRegistry()
	// registry.MustRegister(probeSuccessGauge)
	// registry.MustRegister(probeDurationGauge)
	// success := prober(ctx, target, module, registry, sl)
	// duration := time.Since(start).Seconds()
	// probeDurationGauge.Set(duration)
	// if success {
	// 	probeSuccessGauge.Set(1)
	// 	level.Info(sl).Log("msg", "Probe succeeded", "duration_seconds", duration)
	// } else {
	// 	level.Error(sl).Log("msg", "Probe failed", "duration_seconds", duration)
	// }
}
