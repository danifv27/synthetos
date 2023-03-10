package exporters

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"sync"
	"time"

	"fry.org/cmo/cli/internal/application/exporters"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/speijnik/go-errortree"
)

type CucumberResult int

const (
	CucumberFailure     CucumberResult = iota //0
	CucumberSuccess                           //1
	CucumberNotExecuted                       //2
)

func (rc CucumberResult) String() string {

	return [...]string{"Failure", "Success", "Not executed"}[rc]
}

// map[scenarioID]
// type CucumberScenariosStats map[string]CucumberStats

// map[scenarioID][stepID]
// type CucumberStepsStats map[string]map[string]CucumberStats
type CucumberStatsSet map[string][]CucumberStats
type CucumberStats struct {
	Id       string
	Start    time.Time
	Duration time.Duration
	Result   CucumberResult
}

// type CucumberStatsSet struct {
// Feature   CucumberStats
// Scenarios CucumberScenariosStats
// Steps     CucumberStepsStats
// }

type CucumberPlugin interface {
	// Do execute a godog test suite and returns the stats
	Do(ctx context.Context) (CucumberStatsSet, error)
}

// cucumberHandler is a basic Healthchekcker implementation.
type cucumberHandler struct {
	http.ServeMux
	pluginMutex sync.RWMutex
	PluginSet   map[string]CucumberPlugin
}

// NewCucumberExporter creates a new CucumberExporter
func NewCucumberExporter(opts ...ExporterOption) (exporters.CucumberExporter, error) {
	var rcerror error

	h := cucumberHandler{
		PluginSet: make(map[string]CucumberPlugin),
	}
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
			c.Handle(path.Join(prefix, "/metrics"), promhttp.Handler())
			return nil
		}

		return errortree.Add(rcerror, "WithCucumberRootPrefix", errors.New("type mismatch, cucumberHandler expected"))
	})
}

func WithCucumberPlugin(name string, plugin CucumberPlugin) ExporterOption {

	return ExportOptionFn(func(i interface{}) error {
		var err, rcerror error
		var c *cucumberHandler
		var ok bool

		if c, ok = i.(*cucumberHandler); ok {
			if err = c.registerCucumberPlugin(name, plugin); err != nil {
				return errortree.Add(rcerror, "WithCucumberPlugin", err)
			}
			return nil
		}

		return errortree.Add(rcerror, "WithCucumberPlugin", errors.New("type mismatch, cucumberHandler expected"))
	})
}

func (c *cucumberHandler) registerCucumberPlugin(k string, v CucumberPlugin) error {
	var rcerror error

	c.pluginMutex.Lock()
	defer c.pluginMutex.Unlock()

	if _, dup := c.PluginSet[k]; dup {
		return errortree.Add(rcerror, "RegisterCucumberPlugin", fmt.Errorf("register called twice for driver %s", k))
	}

	c.PluginSet[k] = v

	return nil
}

func (c *cucumberHandler) ProbesEndpoint(w http.ResponseWriter, r *http.Request) {

	c.handle(w, r, c.PluginSet)
}

func (c *cucumberHandler) handle(w http.ResponseWriter, r *http.Request, plugins map[string]CucumberPlugin) {
	var plugin CucumberPlugin
	var stats CucumberStatsSet
	var ok bool
	var err error

	params := r.URL.Query()
	featureName := params.Get("feature")
	if featureName == "" {
		http.Error(w, "missing feature param", http.StatusBadRequest)
		return
	}

	//TODO: find timeout from headers or configuration
	ctx, cancel := context.WithTimeout(r.Context(), 11*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	plugin, ok = c.PluginSet[featureName]
	if !ok {
		http.Error(w, fmt.Sprintf("unknown feature %q", featureName), http.StatusBadRequest)
		return
	}

	featureSuccessGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "synthetos_feature_success",
		Help: "Displays whether or not the feature test was a success",
	})
	featureDurationGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "synthetos_feature_duration_seconds",
		Help: "Returns how long the probe took to complete in seconds",
	})

	registry := prometheus.NewRegistry()
	registry.MustRegister(featureSuccessGauge)
	registry.MustRegister(featureDurationGauge)
	stats, err = plugin.Do(ctx)
	fmt.Printf("[DBG]stats %v\n", stats)
	if err == nil {
		// Feature test succeeded
		featureSuccessGauge.Set(1)
	} else {
		// Feature test failed
		featureSuccessGauge.Set(0)
	}
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}
