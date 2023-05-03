package kuberium

import (
	"errors"
	"fmt"
	"net/url"
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
	Flags KmzSummaryFlags `embed:""`
}

type KmzSummaryFlags struct {
	Output string `prefix:"kmz.summary." help:"Format the output (table|json|text)." enum:"table,json,text" default:"table" env:"SC_KMZ_SUMMARY_OUTPUT"`
}

func initializeKmzSummaryCmd(ctx floc.Context, ctrl floc.Control) error {
	var err, rcerror error
	var c *common.Cmdctx
	var cmd KmzCmd

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeKmzSummaryCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmzSummaryCmd", e)
		}
		return err
	}
	if cmd, err = KuberiumKmzCmd(ctx); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeKmzSummaryCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmzSummaryCmd", e)
		}
		return err
	}
	uri := fmt.Sprintf("provider:kustomize?kustomization=%s", url.QueryEscape(cmd.Flags.KustomizationPath))
	infraOptions := []infrastructure.AdapterOption{
		infrastructure.WithTablePrinter(),
		infrastructure.WithResourceProvider(uri, c.Apps.Logger),
	}
	if err = infrastructure.AdapterWithOptions(&c.Adapters, infraOptions...); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeKmzSummaryCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmzSummaryCmd", e)
		}
		return err
	}
	if err = application.WithOptions(&c.Apps,
		application.WithShowSummaryQuery(c.Apps.Logger, c.Adapters.ResourceProvider),
		application.WithPrintResourceSummaryCommand(c.Apps.Logger, c.Adapters.Printer),
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
	var showSummaryRC actions.ShowSummaryResult

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		KuberiumSetRCErrorTree(ctx, "kubeSummaryJob", err)
		return err
	}
	if cmd, err = KuberiumKmzCmd(ctx); err != nil {
		KuberiumSetRCErrorTree(ctx, "kmzSummaryJob", err)
		return err
	}
	req := actions.ShowSummaryRequest{
		Location: cmd.Flags.KustomizationPath,
	}
	if showSummaryRC, err = c.Apps.Queries.ShowSummary.Handle(req); err != nil {
		KuberiumSetRCErrorTree(ctx, "kmzSummaryJob", err)
		return err
	}
	m := cmd.Summary.Flags.Output
	reqPrint := actions.PrintResourceSummaryRequest{
		Mode:  printer.PrinterModeNone,
		Items: showSummaryRC.Items,
	}
	switch {
	case m == "json":
		reqPrint.Mode = printer.PrinterModeJSON
	case m == "text":
		reqPrint.Mode = printer.PrinterModeText
	case m == "table":
		reqPrint.Mode = printer.PrinterModeTable
	}
	if _, err = c.Apps.Commands.PrintResourceSummary.Handle(reqPrint); err != nil {
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
