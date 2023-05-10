package secretum

import (
	"fry.org/cmo/cli/internal/cli/common"
	"fry.org/cmo/cli/internal/infrastructure"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
)

type KmsFortanixCmd struct {
	Flags KmsFortanixFlags   `embed:""`
	List  KmsFortanixListCmd `cmd:"" help:"List fortanix objects."`
}

type KmsFortanixFlags struct {
	Output string `prefix:"kms.fortanix." help:"Format the output (table|json|text)." enum:"table,json,text" default:"table" env:"SC_KMS_FORTANIX_OUTPUT"`
}

func initializeKmsFortanixCmd(ctx floc.Context, ctrl floc.Control) error {
	var err, rcerror error
	var c *common.Cmdctx

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsFortanixCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsFortanixCmd", e)
		}
		return err
	}
	infraOptions := []infrastructure.AdapterOption{
		infrastructure.WithTablePrinter(),
	}
	if err = infrastructure.AdapterWithOptions(&c.Adapters, infraOptions...); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsFortanixCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsFortanixCmd", e)
		}
		return err
	}
	*c = common.Cmdctx{
		Cmd:      c.Cmd,
		InitSeq:  c.InitSeq,
		Apps:     c.Apps,
		Adapters: c.Adapters,
		Ports:    c.Ports,
	}

	return nil
}

func (cmd *KmsFortanixCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	c.InitSeq = append([]floc.Job{initializeKmsFortanixCmd}, c.InitSeq...)

	return nil
}
