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
	"github.com/iancoleman/strcase"
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
	Do(ctx context.Context, cancel context.CancelFunc) (CucumberStatsSet, error)
}

// cucumberHandler is a basic Healthchekcker implementation.
type cucumberHandler struct {
	http.ServeMux
	pluginMutex sync.RWMutex
	PluginSet   map[string]CucumberPlugin
	timeout     time.Duration
}

// NewCucumberExporter creates a new CucumberExporter
func NewCucumberExporter(opts ...ExporterOption) (exporters.CucumberExporter, error) {
	var rcerror error

	h := cucumberHandler{
		PluginSet: make(map[string]CucumberPlugin),
		timeout:   2 * time.Second,
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

func WithCucumberTimeout(t time.Duration) ExporterOption {

	return ExportOptionFn(func(i interface{}) error {
		var rcerror error
		var c *cucumberHandler
		var ok bool

		if c, ok = i.(*cucumberHandler); ok {
			c.timeout = t
			return nil
		}

		return errortree.Add(rcerror, "WithCucumberRootPrefix", errors.New("type mismatch, cucumberHandler expected"))
	})
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

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Millisecond)
	defer cancel()
	reqWithTimeout := r.WithContext(ctx)
	c.handle(w, reqWithTimeout, c.PluginSet)
}

type PluginResponse struct {
	stats CucumberStatsSet
	err   error
}

func helper(ctx context.Context, cancelFn context.CancelFunc, plugin CucumberPlugin) <-chan PluginResponse {

	respChan := make(chan PluginResponse, 1)
	go func() {
		stats, err := plugin.Do(ctx, cancelFn)
		respChan <- PluginResponse{
			stats: stats,
			err:   err,
		}
	}()

	return respChan
}

func (c *cucumberHandler) handle(w http.ResponseWriter, r *http.Request, plugins map[string]CucumberPlugin) {
	var plugin CucumberPlugin
	var ok bool

	params := r.URL.Query()
	featureName := params.Get("feature")
	if featureName == "" {
		http.Error(w, "missing feature param", http.StatusBadRequest)
		return
	}

	plugin, ok = c.PluginSet[featureName]
	if !ok {
		http.Error(w, fmt.Sprintf("unknown feature %q", featureName), http.StatusBadRequest)
		return
	}

	stepSuccessGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "step_success",
		Help: "Displays whether or not the test was a success",
	}, []string{"feature_name", "scenario_name"})

	stepDurationGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "step_duration_seconds",
		Help: "Duration of test steps in seconds",
	}, []string{"feature_name", "scenario_name", "step_name", "step_status"})

	registry := prometheus.NewRegistry()
	registry.MustRegister(stepSuccessGaugeVec)
	registry.MustRegister(stepDurationGaugeVec)

	ctx, cancelFn := context.WithCancel(r.Context())
	defer cancelFn()
	select {
	case <-ctx.Done():
		cancelFn()
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Max deadline %v exceeded", c.timeout)))
		return
	case pluginChan := <-helper(ctx, cancelFn, plugin):
		if pluginChan.err != nil {
			cancelFn()
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%d - Something bad happened!\n\n%s", http.StatusInternalServerError, pluginChan.err.Error())))
			return
		}
		for k, v := range pluginChan.stats {
			success := 0
			for _, stats := range v {
				// fmt.Printf("[DBG]key[%s] value[%s]\n", k, v)
				// stepDurationHistogramVec.WithLabelValues(strcase.ToCamel(featureName), k, stats.Id, stats.Result.String()).Observe(stats.Duration.Seconds())
				stepDurationGaugeVec.WithLabelValues(strcase.ToCamel(featureName), k, stats.Id, stats.Result.String()).Set(stats.Duration.Seconds())
				success += int(stats.Result)
			}
			//0 failure
			if success > 0 {
				stepSuccessGaugeVec.WithLabelValues(strcase.ToCamel(featureName), k).Set(1)
			} else {
				stepSuccessGaugeVec.WithLabelValues(strcase.ToCamel(featureName), k).Set(0)
			}
		}
		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}
