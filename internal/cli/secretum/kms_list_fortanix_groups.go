package secretum

import (
	"fry.org/cmo/cli/internal/cli/common"
	"github.com/workanator/go-floc/v3"
)

type KmsListFortanixGroupsCmd struct {
	Flags KmsListFortanixGroupsFlags `embed:""`
}

type KmsListFortanixGroupsFlags struct {
}

func initializeKmsListFortanixGroupsCmd(ctx floc.Context, ctrl floc.Control) error {

	return nil
}

func (cmd *KmsListFortanixGroupsCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	c.InitSeq = append(c.InitSeq, initializeKmsListFortanixGroupsCmd)

	return nil
}
