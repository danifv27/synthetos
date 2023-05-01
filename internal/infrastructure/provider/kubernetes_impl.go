package provider

import (
	"context"
	"errors"

	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/provider"
	iClient "fry.org/cmo/cli/internal/infrastructure/client"
	"github.com/speijnik/go-errortree"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
)

type kubernetesClient struct {
	l             logger.Logger
	Client        *iClient.KubeDynamicClient
	Namespace     string
	LabelSelector string
}

// NewKubernetesProvider creates a new CucumberExporter
func NewKubernetesProvider(opts ...ProviderOption) (provider.ResourceProvider, error) {
	var rcerror error

	c := kubernetesClient{}
	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(&c); err != nil {
			return nil, errortree.Add(rcerror, "NewKubernetesProvider", err)
		}
	}

	return &c, nil
}

func WithKubernetesDynamicClient(path string, context *string) ProviderOption {

	return ProviderOptionFn(func(i interface{}) error {
		var rcerror, err error
		var c *kubernetesClient
		var ok bool
		var kubeconfig *rest.Config

		if c, ok = i.(*kubernetesClient); ok {
			if kubeconfig, _, err = iClient.NewKubeClusterConfig(path, context); err != nil {
				errortree.Add(rcerror, "provider.WithKubernetesDynamicClient", err)
			}
			if c.Client, err = iClient.NewKubeDynamicClient(kubeconfig); err != nil {
				errortree.Add(rcerror, "provider.WithKubernetesDynamicClient", err)
			}
			return nil
		}

		return errortree.Add(rcerror, "provider.WithKubernetesDynamicClient", errors.New("type mismatch, kustomizeClient expected"))
	})
}

func WithKubernetesNamespace(path string) ProviderOption {

	return ProviderOptionFn(func(i interface{}) error {
		var rcerror error
		var c *kubernetesClient
		var ok bool

		if c, ok = i.(*kubernetesClient); ok {
			c.Namespace = path
			return nil
		}

		return errortree.Add(rcerror, "provider.WithKubernetesNamespace", errors.New("type mismatch, kustomizeClient expected"))
	})
}

func WithKubernetesLabelSelector(selector string) ProviderOption {

	return ProviderOptionFn(func(i interface{}) error {
		var rcerror error
		var c *kubernetesClient
		var ok bool

		if c, ok = i.(*kubernetesClient); ok {
			c.LabelSelector = selector
			return nil
		}

		return errortree.Add(rcerror, "provider.WithK8sLabelSelector", errors.New("type mismatch, kustomizeClient expected"))
	})
}

func (c *kubernetesClient) GetResources(ctx context.Context, location string, selector string) ([]*unstructured.Unstructured, error) {
	var rcerror error

	return nil, errortree.Add(rcerror, "kubernetes.GetResources", errors.New("getresources method not implemented"))
}
