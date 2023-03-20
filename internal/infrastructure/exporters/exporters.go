package exporters

import (
	"context"
	"fmt"

	"github.com/speijnik/go-errortree"
)

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
