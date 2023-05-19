package secretum

import (
	"fry.org/cmo/cli/internal/cli/common"
	"github.com/workanator/go-floc/v3"
)

type KmsFortanixDecryptCmd struct {
	Flags     KmsFortanixDecryptFlags        `embed:""`
	Secret    KmsFortanixDecryptSecretCmd    `cmd:"" help:"Decrypt Fortanix secrets."`
	Manifests KmsFortanixDecryptManifestsCmd `cmd:"" help:"Decrypt manifests. Only manifests with proper Kubernetes annotations will be processed. The rest will be output unmodified."`
}

type KmsFortanixDecryptFlags struct{}

func initializeKmsFortanixDecryptCmd(ctx floc.Context, ctrl floc.Control) error {

	return nil
}

func (cmd *KmsFortanixDecryptCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	c.InitSeq = append([]floc.Job{initializeKmsFortanixDecryptCmd}, c.InitSeq...)

	return nil
}