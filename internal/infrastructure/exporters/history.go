package exporters

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"

	"github.com/speijnik/go-errortree"
)

func (c *cucumberHandler) loadTemplates() error {
	var rcerror, err error
	var tmplFiles []fs.DirEntry
	var pt *template.Template

	if tmplFiles, err = fs.ReadDir(c.files, "templates"); err != nil {
		return errortree.Add(rcerror, "loadTemplates", err)
	}

	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			continue
		}

		if pt, err = template.ParseFS(c.files, "templates/"+tmpl.Name(), "templates/layouts/*.gohtml"); err != nil {
			return errortree.Add(rcerror, "loadTemplates", err)
		}

		c.templates[tmpl.Name()] = pt
	}

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

	if _, ok := c.templates["layout.gohtml"]; !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Template not found")))
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
