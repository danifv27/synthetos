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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
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

func (c *kustomizeClient) AllImages(ctx context.Context, sendCh chan<- provider.Image, selector string) error {
	var rcerror error

	return errortree.Add(rcerror, "provider.AllImages", errors.New("AllImages method not implemented"))
}

func (c *kustomizeClient) AllResources(ctx context.Context, ch chan<- provider.Resource, ns string, selector string, full bool) error {
	var rcerror error

	return errortree.Add(rcerror, "provider.AllImages", errors.New("AllImages method not implemented"))
}

func (c *kustomizeClient) GetResources(ctx context.Context, location string, selector string) ([]*unstructured.Unstructured, error) {
	var err, rcerror error
	var resources []*unstructured.Unstructured
	var m resmap.ResMap

	m, err = c.kst.Run(c.fSys, c.kustomizationPath)
	if err != nil {
		return nil, errortree.Add(rcerror, "kustomize.GetResources", err)
	}

	// Convert the resources to unstructured objects.
	for _, res := range m.Resources() {

		// Get the Resource's ResId
		resId := res.OrgId()

		// Create an Unstructured object with the Resource's contents
		u := &unstructured.Unstructured{}
		u.SetName(resId.Name)
		u.SetNamespace(resId.Namespace)
		u.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   resId.Group,
			Kind:    resId.Kind,
			Version: resId.Version,
		})
		if u.Object, err = res.Map(); err != nil {
			return nil, errortree.Add(rcerror, "kustomize.GetResources", err)
		}

		resources = append(resources, u)
	}

	return resources, nil
}
