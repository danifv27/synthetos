package exporters

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strings"

	"github.com/speijnik/go-errortree"
)

//go:embed all:html
var htmlFS embed.FS

func (c *cucumberHandler) loadTemplates() error {
	var rcerror, err error
	var pt *template.Template

	funcs := template.FuncMap{
		"uppercase": func(v string) string {
			return strings.ToUpper(v)
		},
	}

	if pt, err = template.New("layout.gohtml").Funcs(funcs).ParseFS(htmlFS, "html/layout.gohtml", "html/log.gohtml"); err != nil {
		return errortree.Add(rcerror, "loadTemplates", err)
	}

	c.templates[pt.Name()] = pt

	return nil
}

func WithCucumberHistoryEndpoint(prefix string) ExporterOption {

	return ExportOptionFn(func(i interface{}) error {
		var rcerror error
		var c *cucumberHandler
		var ok bool

		if c, ok = i.(*cucumberHandler); ok {
			c.templates = make(map[string]*template.Template)
			c.Handle(path.Join(prefix, "/history"), http.HandlerFunc(c.HistoryEndpoint))
			if err := c.loadTemplates(); err != nil {
				return errortree.Add(rcerror, "WithCucumberHistory", err)
			}
			return nil
		}

		return errortree.Add(rcerror, "WithCucumberHistory", errors.New("type mismatch, cucumberHandler expected"))
	})
}

func (c *cucumberHandler) HistoryEndpoint(w http.ResponseWriter, r *http.Request) {
	var t *template.Template
	var ok bool

	if t, ok = c.templates["layout.gohtml"]; !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Template not found"))
		return
	}
	if err := t.Execute(w, struct {
		Title   string
		Message string
	}{
		Title:   "Page Title",
		Message: "This is the message",
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Template %s Error: '%s'", t.Name(), err.Error())))
		return
	}

	// w.Header().Set("Content-Type", "text/html")
	// w.Write([]byte(`<html>
	// 	<head><title>Uxperi - A Cucumber Based Exporter</title></head>
	// 	<body>
	// 	<h1>User Experience Exporter</h1>
	// 	<p><a href="metrics">Metrics</a></p>
	// 	<p><a href="config">Configuration</a></p>
	// 	<h2>Recent Probes</h2>
	// 	<table border='1'><tr><th>Module</th><th>Target</th><th>Result</th><th>Debug</th>`))
	// // results := rh.List()
	// // for i := len(results) - 1; i >= 0; i-- {
	// // 	r := results[i]
	// // 	success := "Success"
	// // 	if !r.Success {
	// // 		success = "<strong>Failure</strong>"
	// // 	}
	// // 	fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td><td><a href='logs?id=%d'>Logs</a></td></td>",
	// // 		html.EscapeString(r.ModuleName), html.EscapeString(r.Target), success, r.Id)
	// // }
	// w.Write([]byte(`</table></body></html>`))
}
