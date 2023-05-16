package provider

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ResourceProvider interface {
	GetResources(ctx context.Context, location string, selector string) ([]*unstructured.Unstructured, error)
}

type ManifestProvider interface {
	GetManifests(ctx context.Context, sendCh chan<- Manifest) error
}

type Summary struct {
	APIVersion string `json:"api_version"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

type Manifest struct {
	Yaml string `json:"yaml,omitempty"`
	// Obj  runtime.Object `json:"object,omitempty"`
}
