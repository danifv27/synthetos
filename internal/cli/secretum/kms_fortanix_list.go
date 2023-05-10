package secretum

import (
	"fry.org/cmo/cli/internal/cli/common"
	"github.com/workanator/go-floc/v3"
)

type KmsFortanixListCmd struct {
	Flags   KmsFortanixListFlags      `embed:""`
	Groups  KmsFortanixListGroupsCmd  `cmd:"" help:"List Fortanix groups."`
	Secrets KmsFortanixListSecretsCmd `cmd:"" help:"List Fortanix secrets."`
}

type KmsFortanixListFlags struct{}

func initializeKmsFortanixListCmd(ctx floc.Context, ctrl floc.Control) error {

	return nil
}

func (cmd *KmsFortanixListCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	c.InitSeq = append([]floc.Job{initializeKmsFortanixListCmd}, c.InitSeq...)

	return nil
}
