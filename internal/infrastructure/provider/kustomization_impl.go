package provider

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/provider"
	"github.com/speijnik/go-errortree"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
)

type kustomizeClient struct {
	l                 logger.Logger
	kustomizationPath string
	fSys              filesys.FileSystem
	kst               *krusty.Kustomizer
}

// honorKustomizeFlags feeds command line data to the krusty options.
func honorKustomizeFlags(kOpts *krusty.Options) *krusty.Options {

	kOpts.LoadRestrictions = types.LoadRestrictionsNone
	kOpts.PluginConfig.HelmConfig.Enabled = false
	kOpts.PluginConfig.HelmConfig.Command = ""
	// When true, a label
	//     app.kubernetes.io/managed-by: kustomize-<version>
	// is added to all the resources in the build out.
	kOpts.AddManagedbyLabel = false

	return kOpts
}

// NewKustomizationProvider creates a new CucumberExporter
func NewKustomizationProvider(opts ...ProviderOption) (provider.ResourceProvider, error) {
	var rcerror error

	c := kustomizeClient{}
	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(&c); err != nil {
			return nil, errortree.Add(rcerror, "NewKustomizationProvider", err)
		}
	}
	c.fSys = filesys.MakeFsOnDisk()
	c.kst = krusty.MakeKustomizer(
		honorKustomizeFlags(krusty.MakeDefaultOptions()),
	)
	return &c, nil
}

func exists(name string) (bool, error) {
	var rcerror, err error

	if _, err = os.Stat(name); err != nil {
		if os.IsNotExist((err)) {
			return false, nil
		}
		return false, errortree.Add(rcerror, "provider.exists", err)
	}

	return !os.IsNotExist(err), nil
}

func KustomizationExists(path string) (bool, error) {
	var rcerror, err error
	var b1, b2 bool

	if b1, err = exists(filepath.Join(path, "kustomization.yaml")); err != nil {
		return false, errortree.Add(rcerror, "provider.KustomizationExists", err)
	}
	if b2, err = exists(filepath.Join(path, "kustomization.yml")); err != nil {
		return false, errortree.Add(rcerror, "provider.KustomizationExists", err)
	}

	return b1 || b2, nil
}

func WithKustomizationPath(path string) ProviderOption {

	return ProviderOptionFn(func(i interface{}) error {
		var rcerror error
		var c *kustomizeClient
		var ok bool

		if c, ok = i.(*kustomizeClient); ok {
			b, err := KustomizationExists(path)
			if err != nil {
				return errortree.Add(rcerror, "provider.WithKustomizationPath", err)
			}
			if b {
				c.kustomizationPath = path
				return nil
			}
		}

		return errortree.Add(rcerror, "provider.WithKustomizationPath", errors.New("type mismatch, kustomizeClient expected"))
	})
}

func (c *kustomizeClient) GetResources(ctx context.Context, location string, selector string) ([]*unstructured.Unstructured, error) {
	var rcerror error

	return nil, errortree.Add(rcerror, "kustomization.GetResources", errors.New("getresources method not implemented"))
}
