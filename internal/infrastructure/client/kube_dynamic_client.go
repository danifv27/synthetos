package client

import (
	"github.com/speijnik/go-errortree"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type KubeDynamicClient struct {
	DiscoveryClient  *discovery.DiscoveryClient
	DiscoveryMapper  *restmapper.DeferredDiscoveryRESTMapper
	DynamicInterface dynamic.Interface
}

func NewKubeDynamicClient(config *rest.Config) (*KubeDynamicClient, error) {
	var rcerror error

	// 1. Prepare a RESTMapper to find GVR
	// DiscoveryClient queries API server about the resources
	disC, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, errortree.Add(rcerror, "NewKubeDynamicClient", err)
	}
	cacheC := memory.NewMemCacheClient(disC)
	cacheC.Invalidate()

	dm := restmapper.NewDeferredDiscoveryRESTMapper(cacheC)

	// client.DynamicClient
	// 2. Prepare the dynamic client
	dynC, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, errortree.Add(rcerror, "NewKubeDynamicClient", err)
	}

	return &KubeDynamicClient{
		DiscoveryClient:  disC,
		DiscoveryMapper:  dm,
		DynamicInterface: dynC,
	}, nil
}
