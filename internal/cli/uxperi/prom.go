package uxperi

import (
	"fry.org/cmo/cli/internal/application"
	"fry.org/cmo/cli/internal/cli/common"
	"fry.org/cmo/cli/internal/infrastructure"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
)

type PromCmd struct {
	Flags VersionFlags `embed:""`
}

type PromFlags struct{}

func initializePromCmd(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	var err, rcerror error

	if c, err = CmdCtx(ctx); err != nil {
		if e := SetRCErrorTree(ctx, "initializePromCmd", err); e != nil {
			return errortree.Add(rcerror, "initializePromCmd", e)
		}
		return err
	}

	infraOptions := []infrastructure.AdapterOption{
		infrastructure.WithHealthchecker(),
	}
	if err = infrastructure.AdapterWithOptions(&c.Adapters, infraOptions...); err != nil {
		if e := SetRCErrorTree(ctx, "initializePromCmd", err); e != nil {
			return errortree.Add(rcerror, "initializePromCmd", e)
		}
		return err
	}
	if err = application.WithOptions(&c.Apps,
		application.WithHealthchecker(c.Adapters.Healthchecker),
	); err != nil {
		if e := SetRCErrorTree(ctx, "initializePromCmd", err); e != nil {
			return errortree.Add(rcerror, "initializePromCmd", e)
		}
		return err
	}
	if err = SetCmdCtx(ctx, common.Cmdctx{
		Cmd:      c.Cmd,
		InitSeq:  c.InitSeq,
		Apps:     c.Apps,
		Adapters: c.Adapters,
		Ports:    c.Ports,
	}); err != nil {
		if e := SetRCErrorTree(ctx, "initializePromCmd", err); e != nil {
			return errortree.Add(rcerror, "initializePromCmd", e)
		}
		return err
	}

	return nil
}

func (cmd *PromCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	c.InitSeq = append(c.InitSeq, initializePromCmd)

	return nil
}
