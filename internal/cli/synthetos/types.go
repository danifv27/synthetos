package synthetos

import (
	"fry.org/cmo/cli/internal/cli/common"
	"fry.org/cmo/cli/internal/cli/kuberium"
	"fry.org/cmo/cli/internal/cli/secretum"
	"fry.org/cmo/cli/internal/cli/uxperi"
	"fry.org/cmo/cli/internal/cli/versio"
)

type CLI struct {
	Logging common.Log        `embed:"" prefix:"logging."`
	Version versio.VersionCmd `cmd:"" help:"Show version information"`
	Kms     secretum.KmsCmd   `cmd:"" help:"Manage KMS"`
	Kmz     kuberium.KmzCmd   `cmd:"" help:"Works with kustomize manifests"`
	Kube    kuberium.KubeCmd  `cmd:"" help:"Provides visibility to the resources running in a Kubernetes cluster"`
	Test    uxperi.TestCmd    `cmd:"" help:"Enter Prometheus mode"`
}
