package exporters

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/antifuchs/o"
	"github.com/robert-nix/ansihtml"
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

	if pt, err = template.New("layout.gohtml").Funcs(funcs).ParseFS(htmlFS, "html/layout.gohtml", "html/css/layout_*.gocss"); err != nil {
		return errortree.Add(rcerror, "loadTemplates", err)
	}
	c.templates[pt.Name()] = pt

	if pt, err = template.New("terminal.gohtml").Funcs(funcs).ParseFS(htmlFS, "html/terminal.gohtml", "html/css/terminal_*.gocss"); err != nil {
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
			c.newHistoryBuffer(25)

			return nil
		}

		return errortree.Add(rcerror, "WithCucumberHistory", errors.New("type mismatch, cucumberHandler expected"))
	})
}

func (c *cucumberHandler) HistoryEndpoint(w http.ResponseWriter, r *http.Request) {
	var t *template.Template
	var ok bool

	params := r.URL.Query()
	id := params.Get("id")
	if id == "" {
		if t, ok = c.templates["layout.gohtml"]; !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Layout template not found"))
			return
		}
		if err := t.Execute(w, c.history.data); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Template %s Error: '%s'", t.Name(), err.Error())))
			return
		}
	} else {
		if i, err := strconv.Atoi(id); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Converting %s Error: '%s'", id, err.Error())))
			return
		} else {
			scenario := params.Get("scenario")
			if scenario == "" {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Missing scenario param"))
				return
			}
			if t, ok = c.templates["terminal.gohtml"]; !ok {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Terminal template not found"))
				return
			}
			//Translate ansi to html
			html := string(ansihtml.ConvertToHTMLWithClasses([]byte(c.history.data[i][scenario].Output), "term-", false))
			if err := t.Execute(w, template.HTML(html)); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Template %s Error: '%s'", t.Name(), err.Error())))
				return
			}
		}
	}
}
