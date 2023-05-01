package kuberium

import (
	"errors"
	"fmt"
	"time"

	"fry.org/cmo/cli/internal/application"
	"fry.org/cmo/cli/internal/application/actions"
	"fry.org/cmo/cli/internal/application/printer"
	"fry.org/cmo/cli/internal/cli/common"
	"fry.org/cmo/cli/internal/infrastructure"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
	"github.com/workanator/go-floc/v3/guard"
	"github.com/workanator/go-floc/v3/run"
)

type KmzSummaryCmd struct {
	Flags KubeSummaryFlags `embed:""`
}

type KmzSummaryFlags struct {
	Output string `prefix:"k8s.list." help:"Format the output (table|json|text)." enum:"table,json,text" default:"table" env:"SC_KMZ_SUMMARY_OUTPUT"`
}

func initializeKmzSummaryCmd(ctx floc.Context, ctrl floc.Control) error {
	var err, rcerror error
	var c *common.Cmdctx
	// var cmd KmzCmd

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeKmzSummaryCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmzSummaryCmd", e)
		}
		return err
	}
	// if cmd, err = KuberiumKmzCmd(ctx); err != nil {
	// 	if e := KuberiumSetRCErrorTree(ctx, "initializeKmzSummaryCmd", err); e != nil {
	// 		return errortree.Add(rcerror, "initializeKmzSummaryCmd", e)
	// 	}
	// 	return err
	// }
	infraOptions := []infrastructure.AdapterOption{
		infrastructure.WithTablePrinter(),
	}
	if err = infrastructure.AdapterWithOptions(&c.Adapters, infraOptions...); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeKmzSummaryCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmzSummaryCmd", e)
		}
		return err
	}
	if err = application.WithOptions(&c.Apps,
		application.WithShowSummaryQuery(c.Apps.Logger, c.Adapters.Printer),
	); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeKmzSummaryCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmzSummaryCmd", e)
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

func kmzSummaryJob(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	var cmd KmzCmd
	var err error

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		KuberiumSetRCErrorTree(ctx, "kubeSummaryJob", err)
		return err
	}
	if cmd, err = KuberiumKmzCmd(ctx); err != nil {
		KuberiumSetRCErrorTree(ctx, "kmzSummaryJob", err)
		return err
	}
	req := actions.ShowSummaryRequest{
		Mode: printer.PrinterModeNone,
	}
	m := cmd.Summary.Flags.Output
	switch {
	case m == "json":
		req.Mode = printer.PrinterModeJSON
	case m == "text":
		req.Mode = printer.PrinterModeText
	case m == "table":
		req.Mode = printer.PrinterModeTable
	}
	if _, err = c.Apps.Queries.ShowSummary.Handle(req); err != nil {
		KuberiumSetRCErrorTree(ctx, "kmzSummaryJob", err)
		return err
	}

	return nil
}

func (cmd *KmzSummaryCmd) Run(c *common.Cmdctx, rcerror *error) error {

	//We need to append at the beginning to traverse the initseq in the right order
	c.InitSeq = append([]floc.Job{initializeKmzSummaryCmd}, c.InitSeq...)
	c.RunSeq = guard.OnTimeout(
		guard.ConstTimeout(5*time.Minute),
		nil, // No need for timeout data
		run.Sequence(
			run.Background(startProbesServer),
			kmzSummaryJob,
			func(ctx floc.Context, ctrl floc.Control) error {
				if rcerror, err := KuberiumRCErrorTree(ctx); err != nil {
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
			KuberiumSetRCErrorTree(ctx, "timeout", errors.New(msg))
			if rcerror, err := KuberiumRCErrorTree(ctx); err != nil {
				ctrl.Fail(fmt.Sprintf("Command '%s' internal error", c.Cmd), err)
			} else {
				ctrl.Fail(msg, *rcerror)
			}
		},
	)

	return nil
}
