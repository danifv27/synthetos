package secretum

import (
	"errors"
	"fmt"
	"time"

	"fry.org/cmo/cli/internal/application"
	"fry.org/cmo/cli/internal/application/actions"
	"fry.org/cmo/cli/internal/application/printer"
	"fry.org/cmo/cli/internal/cli/common"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
	"github.com/workanator/go-floc/v3/guard"
	"github.com/workanator/go-floc/v3/run"
)

type KmsListFortanixGroupsCmd struct {
	Flags KmsListFortanixGroupsFlags `embed:""`
}

type KmsListFortanixGroupsFlags struct {
}

func initializeKmsListFortanixGroupsCmd(ctx floc.Context, ctrl floc.Control) error {
	var err, rcerror error
	var c *common.Cmdctx
	// var cli CLI

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsListFortanixGroupsCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsListFortanixGroupsCmd", e)
		}
		return err
	}
	// if cli, err = SecretumFlags(ctx); err != nil {
	// 	if e := SecretumSetRCErrorTree(ctx, "initializeKmsListFortanixGroupsCmd", err); e != nil {
	// 		return errortree.Add(rcerror, "initializeKmsListFortanixGroupsCmd", e)
	// 	}
	// 	return err
	// }

	if err = application.WithOptions(&c.Apps,
		application.WithListGroupsQuery(c.Apps.Logger, c.Adapters.KeyManager, c.Adapters.Printer),
	); err != nil {
		return errortree.Add(rcerror, "initializeKmsListFortanixGroupsCmd", err)
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

func kmsListFortanixGroupsJob(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	var cmd KmsCmd
	var err error

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		SecretumSetRCErrorTree(ctx, "secretum.startProbesServer", err)
		return err
	}
	if cmd, err = SecretumKmsCmd(ctx); err != nil {
		SecretumSetRCErrorTree(ctx, "secretum.startProbesServer", err)
		return err
	}
	req := actions.ListGroupsRequest{
		Mode: printer.PrinterModeNone,
	}
	m := cmd.List.Flags.Output
	switch {
	case m == "json":
		req.Mode = printer.PrinterModeJSON
	case m == "text":
		req.Mode = printer.PrinterModeText
	case m == "table":
		req.Mode = printer.PrinterModeTable
	}
	if _, err = c.Apps.Queries.ListGroups.Handle(req); err != nil {
		SecretumSetRCErrorTree(ctx, "kmsListFortanixGroupsJob", err)
		return err
	}

	return nil
}

func (cmd *KmsListFortanixGroupsCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	p := cli.Kms.Flags.Probes
	//We need to append at the beginning to traverse the initseq in the right order
	c.InitSeq = append([]floc.Job{initializeKmsListFortanixGroupsCmd}, c.InitSeq...)
	c.RunSeq = guard.OnTimeout(
		guard.ConstTimeout(5*time.Minute),
		nil, // No need for timeout data
		run.Sequence(
			run.If(p.AreProbesEnabled, run.Background(startProbesServer)),
			kmsListFortanixGroupsJob,
			func(ctx floc.Context, ctrl floc.Control) error {
				if rcerror, err := SecretumRCErrorTree(ctx); err != nil {
					ctrl.Fail(fmt.Sprintf("Command '%s' internal error", c.Cmd), err)
					return err
				} else if *rcerror != nil {
					ctrl.Fail(fmt.Sprintf("Command '%s' failed", c.Cmd), *rcerror)
					return *rcerror
				}
				ctrl.Complete(fmt.Sprintf("Command '%s' completed", c.Cmd))

				return nil
			},
		),
		func(ctx floc.Context, ctrl floc.Control, id interface{}) {
			// Fail the flow on timeout
			msg := fmt.Sprintf("Command '%s' timeout expired", c.Cmd)
			SecretumSetRCErrorTree(ctx, "timeout", errors.New(msg))
			if rcerror, err := SecretumRCErrorTree(ctx); err != nil {
				ctrl.Fail(fmt.Sprintf("Command '%s' internal error", c.Cmd), err)
			} else {
				ctrl.Fail(msg, *rcerror)
			}
		},
	)

	return nil
}
