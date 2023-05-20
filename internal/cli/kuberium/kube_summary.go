package kuberium

import (
	"errors"
	"fmt"
	"time"

	"fry.org/cmo/cli/internal/application"
	"fry.org/cmo/cli/internal/application/actions"
	"fry.org/cmo/cli/internal/application/printer"
	"fry.org/cmo/cli/internal/application/provider"
	"fry.org/cmo/cli/internal/cli/common"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
	"github.com/workanator/go-floc/v3/guard"
	"github.com/workanator/go-floc/v3/run"
)

type KubeSummaryCmd struct {
	Flags KubeSummaryFlags `embed:""`
}

type KubeSummaryFlags struct {
	Exclude []common.K8sResource `help:"Regexp matching resources to be excluded from the comparision" prefix:"kube.summary." env:"SC_KUBE_SUMMARY_EXCLUDE_RESOURCE" xor:"Exclude,Include"`
	Include []common.K8sResource `help:"Regexp matching resources to be included to the comparision" prefix:"kube.summary." env:"SC_KUBE_SUMMARY_INCLUDE_RESOURCE" xor:"Exclude,Include"`
	Output  string               `help:"Format the output (table|json|text)." enum:"table,json,text" prefix:"kube.summary." default:"table" env:"SC_KUBE_SUMMARY_OUTPUT"`
}

func initializeKubeSummaryCmd(ctx floc.Context, ctrl floc.Control) error {
	var err, rcerror error
	var c *common.Cmdctx
	// var cmd KubeCmd

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeKubeSummaryCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKubeSummaryCmd", e)
		}
		return err
	}
	// if cmd, err = KuberiumKubeCmd(ctx); err != nil {
	// 	if e := KuberiumSetRCErrorTree(ctx, "initializeKubeSummaryCmd", err); e != nil {
	// 		return errortree.Add(rcerror, "initializeKubeSummaryCmd", e)
	// 	}
	// 	return err
	// }

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
	printerCh := make(chan provider.Summary, 3)
	summaryCh := make(chan provider.Summary, 3)
	quit := make(chan struct{})

	// Let's start the printer consumer
	m := cmd.Flags.Output
	reqPrint := actions.PrintResourceSummaryRequest{
		Mode:    printer.PrinterModeNone,
		InputCh: summaryCh,
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
	//Start the filtering middleware

	//Start the producer
	reqShow := actions.ShowSummaryRequest{
		Location: cmd.Flags.Namespace,
		OutputCh: summaryCh,
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
			run.Background(startKubeProbesServer),
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
