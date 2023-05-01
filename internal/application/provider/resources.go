package provider

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ResourceProvider interface {
	GetResources(ctx context.Context, location string, selector string) ([]*unstructured.Unstructured, error)
}

type Summary struct {
	APIVersion string `json:"api_version"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}
