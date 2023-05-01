package client

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/speijnik/go-errortree"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure" // auth for AKS clusters
	krest "k8s.io/client-go/rest"
	kcmd "k8s.io/client-go/tools/clientcmd"
)

const kube_namespace_envar = "OC_KUBE_K8S_NAMESPACE"

// kubeIsValidUrl tests a string to determine if it is a well-structured url or not.
func kubeIsValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

func kubeBuildConfigFromFlags(kubeconfigPath string, context string) (*krest.Config, string, error) {
	var config *krest.Config
	var err error
	var ns string

	c := kcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&kcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&kcmd.ConfigOverrides{
			CurrentContext: context,
		})
	if config, err = c.ClientConfig(); err != nil {
		return nil, "", fmt.Errorf("buildConfigFromFlags: %w", err)
	}
	ns, _, err = c.Namespace()

	return config, ns, err
}

// NewKubeClusterConfig creates the rest configuration. If neither masterUrl or kubeconfigPath are passed in path argument we fallback to inClusterConfig
func NewKubeClusterConfig(path string, context *string) (*krest.Config, string, error) {
	var config *krest.Config
	var rcerror, err error
	var ns string

	if kubeIsValidUrl(path) {
		// masterUrl detectected
		config, err = kcmd.BuildConfigFromFlags(path, "")
		ns = KubeNamespace(kube_namespace_envar)
	} else {
		if context != nil {
			// // kubeconfig plus context
			config, ns, err = kubeBuildConfigFromFlags(path, *context)
		} else {
			// inClusterConfig
			rules := kcmd.NewDefaultClientConfigLoadingRules()
			rules.DefaultClientConfig = &kcmd.DefaultClientConfig
			overrides := &kcmd.ConfigOverrides{ClusterDefaults: kcmd.ClusterDefaults}
			if context != nil {
				overrides.CurrentContext = *context
			}
			c := kcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
			config, err = c.ClientConfig()
			if err != nil {
				return nil, "", errortree.Add(rcerror, "NewClusterConfig", err)
			}
			ns, _, err = c.Namespace()
		}
	}
	if err != nil {
		return nil, "", errortree.Add(rcerror, "NewClusterConfig", err)
	}
	//Modify QPS and Burst to minimize Waited for 5.354167719s due to client-side throttling, not priority and fairness,
	config.QPS = 100
	config.Burst = 100

	return config, ns, nil
}

func KubeNamespace(envar string) string {
	// This way assumes you've set the envar environment variable using the downward API.
	// This check has to be done first for backwards compatibility with the way InClusterConfig was originally set up
	if ns, ok := os.LookupEnv(envar); ok {
		return ns
	}

	// Fall back to the namespace associated with the service account token, if available
	if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}

	return "default"
}
