package kuberium

import (
	"fry.org/cmo/cli/internal/cli/common"
	"fry.org/cmo/cli/internal/cli/versio"
)

type CLI struct {
	Logging common.Log        `embed:"" prefix:"logging."`
	Version versio.VersionCmd `cmd:"" help:"Show version information"`
	Kube    KubeCmd           `cmd:"" help:"Provides visibility to the resources running in a Kubernetes cluster"`
	Kmz     KmzCmd            `cmd:"" help:"Works with kustomize manifests"`
}
