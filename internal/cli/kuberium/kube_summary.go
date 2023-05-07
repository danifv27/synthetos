package kuberium

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"fry.org/cmo/cli/internal/application"
	"fry.org/cmo/cli/internal/application/actions"
	"fry.org/cmo/cli/internal/application/printer"
	"fry.org/cmo/cli/internal/application/provider"
	"fry.org/cmo/cli/internal/cli/common"
	"fry.org/cmo/cli/internal/infrastructure"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
	"github.com/workanator/go-floc/v3/guard"
	"github.com/workanator/go-floc/v3/run"
)

type KubeSummaryCmd struct {
	Flags KubeSummaryFlags `embed:""`
}

type KubeSummaryFlags struct {
	Output string `prefix:"kube.summary." help:"Format the output (table|json|text)." enum:"table,json,text" default:"table" env:"SC_KUBE_SUMMARY_OUTPUT"`
}

func initializeKubeSummaryCmd(ctx floc.Context, ctrl floc.Control) error {
	var err, rcerror error
	var c *common.Cmdctx
	var cmd KubeCmd

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeKubeSummaryCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKubeSummaryCmd", e)
		}
		return err
	}
	if cmd, err = KuberiumKubeCmd(ctx); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeKubeSummaryCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKubeSummaryCmd", e)
		}
		return err
	}
	// provider:k8s?path=<kubeconfig_path>&context=<kubernetes_context>&namespace=<kubernetes_namespace>&selector=<kubernetes_object_selector>
	uri := fmt.Sprintf("provider:k8s?path=%s&context=%s&namespace=%s",
		url.QueryEscape(cmd.Flags.Path),
		url.QueryEscape(cmd.Flags.Context),
		url.QueryEscape(cmd.Flags.Namespace))
	if cmd.Flags.Selector != nil {
		uri = fmt.Sprintf("%s&selector=%s", uri, url.QueryEscape(*cmd.Flags.Selector))
	}
	infraOptions := []infrastructure.AdapterOption{
		infrastructure.WithTablePrinter(),
		infrastructure.WithResourceProvider(uri, c.Apps.Logger),
	}
	if err = infrastructure.AdapterWithOptions(&c.Adapters, infraOptions...); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeKubeSummaryCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKubeSummaryCmd", e)
		}
		return err
	}
	if err = application.WithOptions(&c.Apps,
		application.WithShowSummaryQuery(c.Apps.Logger, c.Adapters.ResourceProvider),
		application.WithPrintResourceSummaryCommand(c.Apps.Logger, c.Adapters.Printer),
	); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeKubeSummaryCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKubeSummaryCmd", e)
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

func kubeSummaryJob(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	var cmd KubeCmd
	var err error

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		KuberiumSetRCErrorTree(ctx, "kubeSummaryJob", err)
		return err
	}
	if cmd, err = KuberiumKubeCmd(ctx); err != nil {
		KuberiumSetRCErrorTree(ctx, "kubeSummaryJob", err)
		return err
	}
	summaryCh := make(chan provider.Summary, 3)
	quit := make(chan struct{})

	// Let's start the printer consumer
	m := cmd.Summary.Flags.Output
	reqPrint := actions.PrintResourceSummaryRequest{
		Mode: printer.PrinterModeNone,
		Ch:   summaryCh,
	}
	switch {
	case m == "json":
		reqPrint.Mode = printer.PrinterModeJSON
	case m == "text":
		reqPrint.Mode = printer.PrinterModeText
	case m == "table":
		reqPrint.Mode = printer.PrinterModeTable
	}
	go func(req actions.PrintResourceSummaryRequest) {
		if err = c.Apps.Commands.PrintResourceSummary.Handle(req); err != nil {
			KuberiumSetRCErrorTree(ctx, "kubeSummaryJob", err)
		}
		close(quit)
	}(reqPrint)
	//Start the producer
	reqShow := actions.ShowSummaryRequest{
		Location: cmd.Flags.Namespace,
		Ch:       summaryCh,
	}
	if cmd.Flags.Selector != nil {
		reqShow.Selector = *cmd.Flags.Selector
	}
	go func(req actions.ShowSummaryRequest) {
		if err = c.Apps.Queries.ShowSummary.Handle(req); err != nil {
			KuberiumSetRCErrorTree(ctx, "kubeSummaryJob", err)
		}
	}(reqShow)
	//Wait until printer finish it work
	<-quit

	return nil
}

func (cmd *KubeSummaryCmd) Run(c *common.Cmdctx, rcerror *error) error {

	//We need to append at the beginning to traverse the initseq in the right order
	c.InitSeq = append([]floc.Job{initializeKubeSummaryCmd}, c.InitSeq...)
	c.RunSeq = guard.OnTimeout(
		guard.ConstTimeout(5*time.Minute),
		nil, // No need for timeout data
		run.Sequence(
			run.Background(startProbesServer),
			kubeSummaryJob,
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
