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

	stepSuccessGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "step_success",
		Help: "Displays whether or not the step test was a success",
	}, []string{"feature_name", "scenario_name"})

	// stepDurationHistogramVec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
	// 	Name:    "step_bucket_duration_seconds",
	// 	Help:    "Duration of test steps in seconds.",
	// 	Buckets: prometheus.ExponentialBuckets(1, 1.5, 5),
	// }, []string{"feature_name", "scenario_name", "step_name", "step_status"})

	stepDurationGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "step_duration_seconds",
		Help: "Duration of http request by phase, summed over all redirects",
	}, []string{"feature_name", "scenario_name", "step_name", "step_status"})

	registry := prometheus.NewRegistry()
	registry.MustRegister(stepSuccessGaugeVec)
	registry.MustRegister(stepDurationGaugeVec)
	// registry.MustRegister(stepDurationHistogramVec)
	if stats, err = plugin.Do(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("%d - Something bad happened!\n\n%s", http.StatusInternalServerError, err.Error())))
		return
	}
	for k, v := range stats {
		success := 0
		for _, stats := range v {
			// fmt.Printf("[DBG]key[%s] value[%s]\n", k, v)
			// stepDurationHistogramVec.WithLabelValues("loginPage", k, stats.Id, stats.Result.String()).Observe(stats.Duration.Seconds())
			stepDurationGaugeVec.WithLabelValues("loginPage", k, stats.Id, stats.Result.String()).Set(stats.Duration.Seconds())
			success += int(stats.Result)
		}
		//0 failure
		if success > 0 {
			stepSuccessGaugeVec.WithLabelValues("loginPage", k).Set(1)
		} else {
			stepSuccessGaugeVec.WithLabelValues("loginPage", k).Set(0)
		}
	}
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}
