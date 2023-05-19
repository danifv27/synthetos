package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/provider"
	iClient "fry.org/cmo/cli/internal/infrastructure/client"
	"github.com/avast/retry-go/v4"
	"github.com/speijnik/go-errortree"
	"github.com/tidwall/pretty"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

type kubernetesClient struct {
	l             logger.Logger
	Client        *iClient.KubeDynamicClient
	Namespace     string
	LabelSelector string
}

type kubernetesClientItem struct {
	obj     unstructured.Unstructured
	rcerror error
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
				return errortree.Add(rcerror, "provider.WithKubernetesDynamicClient", err)
			}
			if c.Client, err = iClient.NewKubeDynamicClient(kubeconfig); err != nil {
				return errortree.Add(rcerror, "provider.WithKubernetesDynamicClient", err)
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

func containerToImage(img string, imgName string, statuses []v1.ContainerStatus) (*provider.Image, error) {
	if img == "" {
		return nil, fmt.Errorf("container %s has no image", img)
	}

	res := &provider.Image{
		Name: img,
	}

	if strings.Contains(img, "@") {
		res.Digest = strings.Split(img, "@")[1]
		return res, nil
	}
	// search in statuses for ImageID to get digest
	for i := range statuses {
		if imgName == statuses[i].Name {
			if statuses[i].State.Running == nil && statuses[i].State.Terminated == nil {
				break // We can get valid digest only from running or terminated containers
			}
			if strings.Contains(statuses[i].ImageID, "@") {
				res.Digest = strings.Split(statuses[i].ImageID, "@")[1]
			}
			break
		}
	}

	return res, nil
}

func (c *kubernetesClient) AllImages(ctx context.Context, sendCh chan<- provider.Image, selector string) error {
	var rcerror error

	//We do not have enough rights to list namespaces.
	// namespaces, e := c.Client.KubeInterface.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	// if err != nil {
	// 	return images, errortree.Add(rcerror, "AllImages", err)
	// }
	defer close(sendCh)
	listOptions := metav1.ListOptions{
		LabelSelector: selector,
	}
	// for i := range namespaces.Items {
	// namespace := namespaces.Items[i].Name
	pods, err := c.Client.KubeInterface.CoreV1().Pods(c.Namespace).List(ctx, listOptions)
	if err != nil {
		return errortree.Add(rcerror, "AllImages", err)
	}
	images := make(map[string]bool)
	for j := range pods.Items {
		pod := pods.Items[j]
		for k := range pod.Spec.InitContainers {
			img, err := containerToImage(pod.Spec.InitContainers[k].Image, pod.Spec.InitContainers[k].Name, pod.Status.InitContainerStatuses)
			if err != nil {
				return errortree.Add(rcerror, "AllImages", err)
			}
			if _, ok := images[img.Digest]; !ok {
				images[img.Digest] = true
				sendCh <- *img
			}
		}
		for k := range pod.Spec.Containers {
			img, err := containerToImage(pod.Spec.Containers[k].Image, pod.Spec.Containers[k].Name, pod.Status.ContainerStatuses)
			if err != nil {
				return errortree.Add(rcerror, "AllImages", err)
			}
			if _, ok := images[img.Digest]; !ok {
				images[img.Digest] = true
				sendCh <- *img
			}
		}
		for k := range pod.Spec.EphemeralContainers {
			img, err := containerToImage(pod.Spec.EphemeralContainers[k].Image,
				pod.Spec.EphemeralContainers[k].Name, pod.Status.EphemeralContainerStatuses)
			if err != nil {
				return errortree.Add(rcerror, "AllImages", err)
			}
			if _, ok := images[img.Digest]; !ok {
				images[img.Digest] = true
				sendCh <- *img
			}
		}
	}
	// }

	// toReturn := make([]provider.Image, 0)
	// for _, v := range images {
	// 	toReturn = append(toReturn, v)
	// }

	return nil
}

func (c *kubernetesClient) AllResources(ctx context.Context, ch chan<- provider.Resource, ns string, selector string, concise bool) error {
	var rcerror, err error
	var apiResourceList []*metav1.APIResourceList

	if apiResourceList, err = c.Client.KubeInterface.Discovery().ServerPreferredResources(); err != nil {
		return errortree.Add(rcerror, "provider.AllResources", err)
	}

	resourceMap := make(map[string]provider.ResourceList)
	for _, apiResource := range apiResourceList {

		gv, err := schema.ParseGroupVersion(apiResource.GroupVersion)
		if err != nil {
			return errortree.Add(rcerror, "provider.AllResources", err)
		}

		for i := range apiResource.APIResources {
			res := apiResource.APIResources[i]
			gvr := schema.GroupVersionResource{
				Group:    gv.Group,
				Version:  gv.Version,
				Resource: res.Name,
			}

			resourceList, err := c.Client.DynamicInterface.Resource(gvr).Namespace(ns).List(ctx, metav1.ListOptions{})
			if err != nil {
				c.l.WithFields(logger.Fields{
					"err": err,
					"gvr": gvr,
				}).Debug("failed to list resources")
				continue
			}
			c.l.WithFields(logger.Fields{
				"gvr":   gvr,
				"count": len(resourceList.Items),
			}).Debug("found resources")

			if len(resourceList.Items) > 0 {
				resourceMap[gvr.String()] = provider.ResourceList{
					Kind:           resourceList.Items[0].GetKind(),
					APIVersion:     gvr.GroupVersion().String(),
					Namespaced:     res.Namespaced,
					ResourcesCount: len(resourceList.Items),
					Resources:      make([]provider.Resource, 0),
				}
			}
			if !concise {
				for _, item := range resourceList.Items {
					res := provider.Resource{
						Name:      item.GetName(),
						Namespace: item.GetNamespace(),
					}

					val := resourceMap[gvr.String()]
					val.Resources = append(val.Resources, res)
					resourceMap[gvr.String()] = val
				}
			}
		}
	}

	if j, err := json.Marshal(resourceMap); err != nil {
		return errortree.Add(rcerror, "provider.AllResources", err)
	} else {
		fmt.Printf("%s\n", pretty.Pretty(j))
	}

	return nil
}

func (c *kubernetesClient) GetResources(ctx context.Context, location string, selector string) ([]*unstructured.Unstructured, error) {
	var rcerror, err error
	var list []*metav1.APIResourceList
	var uItems []*unstructured.Unstructured

	//The v1.APIResourceList object is used to represent a list of available Kubernetes API resources for a particular group version (GV).
	list, err = c.Client.DiscoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, errortree.Add(rcerror, "GetResources", err)
	}
	listOptions := metav1.ListOptions{
		LabelSelector: selector,
	}
	ch := make(chan kubernetesClientItem, 10)
	wg := sync.WaitGroup{}
	//Start the consumer first to keep the number of producer goroutines low
	quit := make(chan struct{})
	c.l.Debug("Starting kubernetes consumer...")
	go func(ch chan kubernetesClientItem, q chan struct{}) {
	consumerLoop:
		for {
			select {
			case <-ctx.Done(): // closes when the caller cancels the ctx
				c.l.WithFields(logger.Fields{
					"err": ctx.Err(),
				}).Warn("Kubernetes consumer stopped")
				break consumerLoop
			case u, ok := <-ch:
				if !ok {
					break consumerLoop
				}
				if u.rcerror == nil {
					uItems = append(uItems, &u.obj)
					// c.l.WithFields(logger.Fields{
					// 	"obj": u.obj.GetName(),
					// }).Debug("Kubernetes consumed...")
				}
				// } else {
				// 	c.l.WithFields(logger.Fields{
				// 		"err": u.rcerror,
				// 		"obj": u.obj.GetName(),
				// 	}).Warn("Error consuming Kubernetes cluster object")
				// }
			} //select
		} //for
		close(q)
		c.l.Debug("Finished kubernetes consumer. Channel closed")
	}(ch, quit)
	// Start the producer
	for idx := range list {
		wg.Add(1)
		go func(ch chan kubernetesClientItem, meta *metav1.APIResourceList) {
			var u kubernetesClientItem
			var ulist *unstructured.UnstructuredList
			var rcerror error

			defer wg.Done()
			select {
			case <-ctx.Done(): // closes when the caller cancels the ctx
				c.l.WithFields(logger.Fields{
					"err": ctx.Err(),
				}).Warn("Kubernetes producer stopped")
				break
			default:
				if gv, err := schema.ParseGroupVersion(meta.GroupVersion); err != nil {
					c.l.WithFields(logger.Fields{
						"err": err,
						"gv":  meta.GroupVersion,
					}).Warn("Error parsing group version")
					u.rcerror = errortree.Add(rcerror, "GetResources", err)
				} else {
					for _, res := range meta.APIResources {
						gvr := gv.WithResource(res.Name)
						err := retry.Do(
							func() error {
								var err error
								ulist, err = c.Client.DynamicInterface.Resource(gvr).Namespace(location).List(context.TODO(), listOptions)
								// c.l.WithFields(logger.Fields{
								// 	"err":       err,
								// 	"gvr":       gvr,
								// 	"namespace": location,
								// }).Warn("Error listing kubernetes objects")
								return err
							},
							retry.Attempts(3),
						)
						if err != nil {
							u.rcerror = errortree.Add(rcerror, "GetResources", fmt.Errorf("listing resource %v (%v): %w", gvr, location, err))

						} else if ulist != nil && len(ulist.Items) > 0 {
							for _, item := range ulist.Items {
								u.obj = item
								// c.l.WithFields(logger.Fields{
								// 	"gvr":     gvr,
								// 	"gv":      meta.GroupVersion,
								// 	"rcerror": u.rcerror,
								// }).Debug("Send to kubernetes consumer")
								ch <- u
							}
						}
					}
				}
			} //select
		}(ch, list[idx])
	} //for range
	// c.l.Debug("Waiting for kubernetes producers to stop...")
	wg.Wait()
	// c.l.Debug("Kubernetes producers closed. Closing channel...")
	close(ch)
	// c.l.Debug("Channel closed. Waiting for kubernetes consumer to close")
	<-quit
	// c.l.Debug("Kubernetes consumer closed")

	return uItems, nil
}
