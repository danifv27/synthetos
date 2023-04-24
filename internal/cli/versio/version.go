package versio

import (
	"errors"
	"fmt"
	"time"

	"fry.org/cmo/cli/internal/application"
	"fry.org/cmo/cli/internal/application/actions"
	"fry.org/cmo/cli/internal/cli/common"
	"fry.org/cmo/cli/internal/infrastructure"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
	"github.com/workanator/go-floc/v3/guard"
	"github.com/workanator/go-floc/v3/run"
)

type VersionCmd struct {
	Flags VersionFlags `embed:""`
}

type VersionFlags struct {
	Output string `env:"SC_VERSION_OUTPUT" prefix:"version." help:"Format the output (pretty|json)." enum:"pretty,json" default:"pretty"`
}

func initializeVersionCmd(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	var err, rcerror error

	if c, err = VersioCmdCtx(ctx); err != nil {
		if e := VersioSetRCErrorTree(ctx, "initializeVersionCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeVersionCmd", e)
		}
		return err
	}

	infraOptions := []infrastructure.AdapterOption{
		infrastructure.WithVersion("version:embed"),
		infrastructure.WithTablePrinter(),
	}
	if err = infrastructure.AdapterWithOptions(&c.Adapters, infraOptions...); err != nil {
		if e := VersioSetRCErrorTree(ctx, "initializeVersionCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeVersionCmd", e)
		}
		return err
	}
	if err = application.WithOptions(&c.Apps,
		application.WithPrintVersionCommand(c.Adapters.Version, c.Adapters.Printer),
	); err != nil {
		if e := VersioSetRCErrorTree(ctx, "initializeVersionCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeVersionCmd", e)
		}
		return err
	}
	if err = VersioSetCmdCtx(ctx, common.Cmdctx{
		Cmd:      c.Cmd,
		InitSeq:  c.InitSeq,
		Apps:     c.Apps,
		Adapters: c.Adapters,
		Ports:    c.Ports,
	}); err != nil {
		if e := VersioSetRCErrorTree(ctx, "initializeVersionCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeVersionCmd", e)
		}
		return err
	}

	return nil
}

func versionPrintJob(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	var cmd VersionCmd
	var err error

	if c, err = VersioCmdCtx(ctx); err != nil {
		VersioSetRCErrorTree(ctx, "versionPrintJob", err)
		return err
	}
	if cmd, err = VersioVersionCmd(ctx); err != nil {
		VersioSetRCErrorTree(ctx, "versionPrintJob", err)
		return err
	}
	req := actions.PrintVersionRequest{
		Format: cmd.Flags.Output,
	}
	if err = c.Apps.Commands.PrintVersion.Handle(req); err != nil {
		VersioSetRCErrorTree(ctx, "versionPrintJob", err)
		return err
	}

	return nil
}

func (cmd *VersionCmd) Run(c *common.Cmdctx, rcerror *error) error {

	c.InitSeq = append(c.InitSeq, initializeVersionCmd)

	c.RunSeq = guard.OnTimeout(
		guard.ConstTimeout(5*time.Minute),
		nil, // No need for timeout data
		run.Sequence(
			versionPrintJob,
			func(ctx floc.Context, ctrl floc.Control) error {

				if rcerror, err := VersioRCErrorTree(ctx); err != nil {
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
			VersioSetRCErrorTree(ctx, "timeout", errors.New(msg))
			if rcerror, err := VersioRCErrorTree(ctx); err != nil {
				ctrl.Fail(fmt.Sprintf("Command '%s' internal error", c.Cmd), err)
			} else {
				ctrl.Fail(msg, *rcerror)
			}
		},
	)

	return nil
}
