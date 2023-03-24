package exporters

import (
	"context"
	"embed"
	"fmt"
	"path"
	"path/filepath"

	"github.com/cucumber/godog"
	"github.com/speijnik/go-errortree"
)

//go:embed features/*.feature
var FeaturesFS embed.FS

// An ExporterOption applies optional changes to the Kong application.
type ExporterOption interface {
	Apply(t interface{}) error
}

// ExportOptionFn is function that adheres to the ExporterOption interface.
type ExportOptionFn func(t interface{}) error

func (o ExportOptionFn) Apply(t interface{}) error {

	return o(t)
}

var (
	ContextKeyTargetUrl    = ContextKey("targetUrl")
	ContextKeyScenarioName = ContextKey("scenarioName")
)

type ContextKey string

func (c ContextKey) String() string {
	return "exporters." + string(c)
}

func StringFromContext(ctx context.Context, key ContextKey) (string, error) {
	var value string
	var ok bool
	var rcerror error

	if value, ok = ctx.Value(key).(string); !ok {
		return "", errortree.Add(rcerror, "StringFromContext", fmt.Errorf("type mismatch with key %s", key))
	}

	return value, nil
}

func GetFeatures(fs embed.FS, dir string) ([]godog.Feature, error) {
	var rcerror error
	var features []godog.Feature

	if len(dir) == 0 {
		dir = "."
	}

	entries, err := fs.ReadDir(dir)
	if err != nil {
		return features, errortree.Add(rcerror, "GetFeatures", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if b, err := fs.ReadFile(path.Join(dir, entry.Name())); err != nil {
			return features, errortree.Add(rcerror, "GetFeatures", err)
		} else {
			f := godog.Feature{
				Name:     entry.Name(),
				Contents: b,
			}
			features = append(features, f)
		}
	}

	return features, nil
}

func GetFeature(fs embed.FS, path string) ([]godog.Feature, error) {
	var rcerror error
	var features []godog.Feature

	if b, err := fs.ReadFile(path); err != nil {
		return features, errortree.Add(rcerror, "GetFeature", err)
	} else {
		f := godog.Feature{
			Name:     filepath.Base(path),
			Contents: b,
		}
		return append(features, f), nil
	}
}
