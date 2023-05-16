package provider

import (
	"errors"
	"fmt"
	"net/url"

	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/provider"
	"github.com/speijnik/go-errortree"
) // A ProviderOption applies optional changes to the provider implementation

type ProviderOption interface {
	Apply(t interface{}) error
}

// ProviderOptionFn is function that adheres to the ProviderOption interface.
type ProviderOptionFn func(t interface{}) error

func (o ProviderOptionFn) Apply(t interface{}) error {

	return o(t)
}

// ParseResourceProvider the uri string and returns the proper provider.GetResources implementation
// Available uris:
// provider:k8s?path=<kubeconfig_path>&context=<kubernetes_context>&namespace=<kubernetes_namespace>&selector=<kubernetes_object_selector>
// provider:kustomize?kustomization=<path_to_kustomize_yaml>
func ParseResourceProvider(URI string, l logger.Logger) (provider.ResourceProvider, error) {
	var k provider.ResourceProvider
	var err, rcerror error
	var u *url.URL
	var context *string

	u, err = url.Parse(URI)
	if err != nil {
		rcerror = errortree.Add(rcerror, "provider.ParseResourceProvider", err)
		return nil, rcerror
	}
	if u.Scheme != "provider" {
		rcerror = errortree.Add(rcerror, "provider.ParseResourceProvider", fmt.Errorf("invalid scheme %s", URI))
		return nil, rcerror
	}
	options := []ProviderOption{
		WithLogger(l),
	}
	switch u.Opaque {
	case "k8s":
		path := u.Query().Get("path")
		ctx := u.Query().Get("context")
		if ctx == "" {
			context = nil
		} else {
			context = &ctx
		}
		options = append(options,
			WithKubernetesDynamicClient(path, context),
		)
		ns := u.Query().Get("namespace")
		if ns != "" {
			options = append(options,
				WithKubernetesNamespace(ns),
			)
		}
		selector := u.Query().Get("selector")
		if selector != "" {
			options = append(options,
				WithKubernetesLabelSelector(selector),
			)
		}
		if k, err = NewKubernetesProvider(options...); err != nil {
			rcerror = errortree.Add(rcerror, "provider.ParseResourceProvider", err)
			return nil, rcerror
		}
	case "kustomize":
		path := u.Query().Get("kustomization")
		if path == "" {
			rcerror = errortree.Add(rcerror, "provider.ParseResourceProvider", errors.New("missing kustomization query argument"))
			return nil, rcerror
		}
		options = append(options,
			WithKustomizationPath(path),
		)

		if k, err = NewKustomizationProvider(options...); err != nil {
			rcerror = errortree.Add(rcerror, "provider.ParseResourceProvider", err)
			return nil, rcerror
		}
	default:
		rcerror = errortree.Add(rcerror, "provider.ParseResourceProvider", fmt.Errorf("unsupported provider implementation %q", u.Opaque))
		return nil, rcerror
	}

	return k, nil
}

// ParseResourceProvider the uri string and returns the proper provider.GetResources implementation
// Available uris:
// provider:reader?path=<input_path>
func ParseManifestProvider(URI string, l logger.Logger) (provider.ManifestProvider, error) {
	var k provider.ManifestProvider
	var err, rcerror error
	var u *url.URL

	u, err = url.Parse(URI)
	if err != nil {
		rcerror = errortree.Add(rcerror, "provider.ParseManifestProvider", err)
		return nil, rcerror
	}
	if u.Scheme != "provider" {
		rcerror = errortree.Add(rcerror, "provider.ParseManifestProvider", fmt.Errorf("invalid scheme %s", URI))
		return nil, rcerror
	}
	options := []ProviderOption{
		WithLogger(l),
	}
	switch u.Opaque {
	case "reader":
		path := u.Query().Get("path")
		if path == "" {
			rcerror = errortree.Add(rcerror, "provider.ParseManifestProvider", errors.New("missing path query argument"))
			return nil, rcerror
		}
		options = append(options,
			WithReaderProviderInputType(path),
		)

		if k, err = NewReaderProvider(options...); err != nil {
			rcerror = errortree.Add(rcerror, "provider.ParseManifestProvider", err)
			return nil, rcerror
		}
	default:
		rcerror = errortree.Add(rcerror, "provider.ParseManifestProvider", fmt.Errorf("unsupported manifest provider implementation %q", u.Opaque))
		return nil, rcerror
	}

	return k, nil
}

func WithLogger(l logger.Logger) ProviderOption {

	return ProviderOptionFn(func(i interface{}) error {
		var rcerror error
		var f *kustomizeClient
		var k *kubernetesClient
		var m *yamlReader
		var ok bool

		if f, ok = i.(*kustomizeClient); ok {
			f.l = l
			return nil
		} else if k, ok = i.(*kubernetesClient); ok {
			k.l = l
			return nil
		} else if m, ok = i.(*yamlReader); ok {
			m.l = l
			return nil
		}

		return errortree.Add(rcerror, "provider.WithLogger", errors.New("type mismatch, kustomizeClient or k8sClient expected"))
	})
}
