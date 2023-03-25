package exporters

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strings"

	"github.com/antifuchs/o"
	"github.com/speijnik/go-errortree"
)

//go:embed all:html
var htmlFS embed.FS

type historyBuffer struct {
	ring o.Ring
	data []CucumberStatsSet
}

func (c *cucumberHandler) newHistoryBuffer(size uint) error {

	c.history = historyBuffer{
		ring: o.NewRing(size),
		data: make([]CucumberStatsSet, size),
	}

	return nil
}

func (c *cucumberHandler) addHistory(s CucumberStatsSet) error {

	c.history.data[c.history.ring.ForcePush()] = s

	return nil
}

// const cucumberHistorySize 50

func (c *cucumberHandler) loadTemplates() error {
	var rcerror, err error
	var pt *template.Template

	funcs := template.FuncMap{
		"uppercase": func(v string) string {
			return strings.ToUpper(v)
		},
	}

	if pt, err = template.New("layout.gohtml").Funcs(funcs).ParseFS(htmlFS, "html/layout.gohtml", "html/css/*.gocss"); err != nil {
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
			c.newHistoryBuffer(2)

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
	if err := t.Execute(w, c.history.data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Template %s Error: '%s'", t.Name(), err.Error())))
		return
	}
}
