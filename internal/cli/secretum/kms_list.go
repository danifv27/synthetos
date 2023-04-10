package secretum

import (
	"fry.org/cmo/cli/internal/cli/common"
	"github.com/workanator/go-floc/v3"
)

type KmsListCmd struct {
	Flags    KmsListFlags       `embed:""`
	Fortanix KmsListFortanixCmd `cmd:"" help:"List fortanix objects."`
}

type KmsListFlags struct {
}

func initializeKmsListCmd(ctx floc.Context, ctrl floc.Control) error {

	return nil
}

func (cmd *KmsListCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	c.InitSeq = append(c.InitSeq, initializeKmsListCmd)

	return nil
}
