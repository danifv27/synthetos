package exporters

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"sync"
	"time"

	"fry.org/cmo/cli/internal/application/exporters"
	"github.com/chromedp/chromedp"
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

type CucumberStatsSet map[string]CucumberStatsItem

type CucumberStatsItem struct {
	Output string
	Stats  []CucumberStats
}

type CucumberStats struct {
	Id       string
	Start    time.Time
	Duration time.Duration
	Result   CucumberResult
}

type CucumberPlugin interface {
	// Do execute a godog test suite and returns the stats
	Do(ctx context.Context) (CucumberStatsSet, error)
	GetScenarioName() (string, error)
}

// cucumberHandler is a basic Healthchekcker implementation.
type cucumberHandler struct {
	http.ServeMux
	pluginMutex sync.RWMutex
	PluginSet   map[string]CucumberPlugin
	timeout     time.Duration
	templates   map[string]*template.Template
	history     historyBuffer
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

		return errortree.Add(rcerror, "WithCucumberTimeout", errors.New("type mismatch, cucumberHandler expected"))
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

	ctx, cancel := context.WithTimeout(r.Context(), c.timeout)
	defer cancel()
	reqWithTimeout := r.WithContext(ctx)
	c.handle(w, reqWithTimeout, c.PluginSet)
}

type PluginResponse struct {
	set CucumberStatsSet
	err error
}

func helper(ctx context.Context, plugin CucumberPlugin) <-chan PluginResponse {

	respChan := make(chan PluginResponse, 1)
	go func() {
		set, err := plugin.Do(ctx)
		respChan <- PluginResponse{
			set: set,
			err: err,
		}
	}()

	return respChan
}

func (c *cucumberHandler) handle(w http.ResponseWriter, r *http.Request, plugins map[string]CucumberPlugin) {
	var plugin CucumberPlugin
	var ok bool

	params := r.URL.Query()
	target := params.Get("target")
	if target == "" {
		http.Error(w, "missing target param", http.StatusBadRequest)
		return
	}
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

	scenarioSuccessGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "scenario_success",
		Help: "Displays whether or not the scenario test was succesful",
	}, []string{"feature_name", "scenario_name"})

	stepSuccessGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "step_success",
		Help: "Displays whether or not the step was a success",
	}, []string{"feature_name", "scenario_name", "step_name"})

	stepDurationGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "step_duration_seconds",
		Help: "Duration of test steps in seconds",
	}, []string{"feature_name", "scenario_name", "step_name", "step_status"})

	registry := prometheus.NewRegistry()
	registry.MustRegister(scenarioSuccessGaugeVec)
	registry.MustRegister(stepSuccessGaugeVec)
	registry.MustRegister(stepDurationGaugeVec)
	ctx, cancelFn := context.WithCancel(r.Context())
	ct := context.WithValue(ctx, ContextKeyTargetUrl, target)
	defer cancelFn()
	//Initialize chromedp context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.111 Safari/537.36"),
	)
	actx, _ := chromedp.NewExecAllocator(ct, opts...)
	plugingCtx, _ := chromedp.NewContext(actx)

	//FIXME: Add some values to history ring buffer when context fails
	select {
	case <-plugingCtx.Done():
		// Extract the reason for cancellation
		err := plugingCtx.Err()
		switch err {
		case context.Canceled:
			// Handle cancellation scenario
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("context cancel detected"))
		case context.DeadlineExceeded:
			// Handle max timeout
			if name, e := plugin.GetScenarioName(); e != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Scenario name not found for metrics %v", e)))
			} else {
				scenarioSuccessGaugeVec.WithLabelValues(strcase.ToCamel(featureName), name).Set(float64(CucumberFailure))
			}
		default:
			// Handle other errors

		}
	case pluginChan := <-helper(plugingCtx, plugin):
		c.addHistory(pluginChan.set)
		if pluginChan.err != nil {
			for k, v := range pluginChan.set {
				scenarioSuccessGaugeVec.WithLabelValues(strcase.ToCamel(featureName), k).Set(float64(CucumberFailure))
				for _, stats := range v.Stats {
					stepSuccessGaugeVec.WithLabelValues(strcase.ToCamel(featureName), k, stats.Id).Set(float64(stats.Result))
				}
			}
		} else {
			for k, v := range pluginChan.set {
				isSucceeded := true
				for _, stats := range v.Stats {
					stepDurationGaugeVec.WithLabelValues(strcase.ToCamel(featureName), k, stats.Id, stats.Result.String()).Set(stats.Duration.Seconds())
					stepSuccessGaugeVec.WithLabelValues(strcase.ToCamel(featureName), k, stats.Id).Set(float64(stats.Result))
					if stats.Result != CucumberSuccess {
						isSucceeded = false
					}
				}
				//0 failure
				if isSucceeded {
					scenarioSuccessGaugeVec.WithLabelValues(strcase.ToCamel(featureName), k).Set(float64(CucumberSuccess))
				} else {
					scenarioSuccessGaugeVec.WithLabelValues(strcase.ToCamel(featureName), k).Set(float64(CucumberFailure))
				}
			}
		}
	}
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}
