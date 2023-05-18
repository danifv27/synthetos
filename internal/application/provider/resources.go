package provider

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ResourceProvider interface {
	GetResources(ctx context.Context, location string, selector string) ([]*unstructured.Unstructured, error)
	AllImages(ctx context.Context, ch chan<- Image, selector string) error
	AllResources(ctx context.Context, ch chan<- Resource, namespace string, selector string, full bool) error
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
}

type Resources struct {
	Images    []Image                 `json:"images,omitempty"`
	Resources map[string]ResourceList `json:"resources"`
}

type ResourceList struct {
	Kind           string     `json:"kind"`
	APIVersion     string     `json:"api_version"`
	Namespaced     bool       `json:"namespaced"`
	ResourcesCount int        `json:"count"`
	Resources      []Resource `json:"resources,omitempty"`
}

type Resource struct {
	Kind       string `json:"kind,omitempty"`
	APIVersion string `json:"api_version,omitempty"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace,omitempty"`
}
type Image struct {
	Name   string `json:"name"`
	Digest string `json:"digest"`
}
