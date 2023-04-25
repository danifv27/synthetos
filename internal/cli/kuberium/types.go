package kuberium

import (
	"fry.org/cmo/cli/internal/cli/common"
	"fry.org/cmo/cli/internal/cli/versio"
)

type CLI struct {
	Logging common.Log        `embed:"" prefix:"logging."`
	Version versio.VersionCmd `cmd:"" help:"Show version information"`
	K8s     K8sCmd            `cmd:"" help:"Provides visibility into the resources running in a Kubernetes cluster"`
	Kmz     KmzCmd            `cmd:"" help:"Read a Kubernetes manifest and list its contents"`
}
