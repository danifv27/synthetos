package synthetos

import (
	"fry.org/cmo/cli/internal/cli/common"
	"fry.org/cmo/cli/internal/cli/kuberium"
	"fry.org/cmo/cli/internal/cli/secretum"
	"fry.org/cmo/cli/internal/cli/uxperi"
)

type CLI struct {
	Logging common.Log      `embed:"" prefix:"logging."`
	Version VersionCmd      `cmd:"" help:"Show version information"`
	Kms     secretum.KmsCmd `cmd:"" help:"Manage KMS"`
	Test    uxperi.TestCmd  `cmd:"" help:"Enter Prometheus mode"`
	Kube    kuberium.K8sCmd `cmd:"" help:"Provides visibility into the resources running in a Kubernetes cluster"`
}
