package main

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"

	"fry.org/cmo/cli/internal/application"
	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/cli/common"
	"fry.org/cmo/cli/internal/cli/uxperi"
	"fry.org/cmo/cli/internal/infrastructure"

	"github.com/alecthomas/kong"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
	"github.com/workanator/go-floc/v3/run"
)

func initializeCmd(cli *uxperi.CLI, cmd string) (common.Cmdctx, error) {
	var err, rcerror error
	var output string

	ctx := common.Cmdctx{
		Cmd: cmd,
	}

	if cli.Logging.Json {
		output = "json"
	} else {
		output = "plain"
	}
	if ctx.Adapters, err = infrastructure.NewAdapters(
		infrastructure.WithLogger(fmt.Sprintf("logger:logrus?level=%s&output=%s", url.QueryEscape(cli.Logging.Level), url.QueryEscape(output))),
	); err != nil {
		rcerror = errortree.Add(rcerror, "context", err)
		rcerror = errortree.Add(rcerror, "cmd", fmt.Errorf("%s", cmd))
		rcerror = errortree.Add(rcerror, "msg", fmt.Errorf("can not initialize %s", cmd))
		return ctx, rcerror
	}
	if ctx.Apps, err = application.NewApplications(
		application.WithLogger(ctx.Adapters.Logger),
	); err != nil {
		return common.Cmdctx{}, fmt.Errorf("initializeCmd: %w", err)
	}
	if ctx.Ports, err = infrastructure.NewPorts(); err != nil {
		return common.Cmdctx{}, fmt.Errorf("initializeCmd: %w", err)
	}

	return ctx, nil
}

func main() {
	var err, rcerror error
	var pCtxcmd *common.Cmdctx
	var result floc.Result
	var data interface{}

	cli := uxperi.CLI{
		Logging: common.Log{},
	}

	bin := filepath.Base(os.Args[0])
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	exBin := filepath.Base(ex)
	//fmt.Printf("[DBG]path: %s, bin: %s\n", exPath, exBin)
	pCtxcmd = new(common.Cmdctx)
	//config file has precedence over envars
	ctx := kong.Parse(&cli,
		kong.Bind(pCtxcmd),
		kong.Bind(&rcerror),
		kong.Name(bin),
		kong.Description("Cucumber based Prometheus exporter"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Tree: true,
		}),
		// kong.TypeMapper(reflect.TypeOf([]common.K8sResource{}), common.K8sResource{}),
		kong.Configuration(kong.JSON, fmt.Sprintf("/etc/%s.json", bin), fmt.Sprintf("~/.%s.json", bin), fmt.Sprintf("%s/.%s.json", exPath, exBin)),
	)
	if *pCtxcmd, err = initializeCmd(&cli, ctx.Command()); err != nil {
		ctx.FatalIfErrorf(err)
	}
	pCtxcmd.Apps.Logger.WithFields(logger.Fields{
		"folder":     exPath,
		"executable": exBin,
		"cmd":        pCtxcmd.Cmd,
	}).Debug("Run CLI command")
	//Run should create the job flow that will be executed as a sequence
	if err = ctx.Run(&cli); err != nil {
		rcerror = errortree.Add(rcerror, "context", err)
		rcerror = errortree.Add(rcerror, "cmd", fmt.Errorf("%s", ctx.Command()))
		rcerror = errortree.Add(rcerror, "msg", fmt.Errorf("can not execute '%s' command", ctx.Command()))
		ctx.FatalIfErrorf(rcerror)
	}

	flocCtx := floc.NewContext()
	uxperi.UxperiSetCmdCtx(flocCtx, *pCtxcmd)
	uxperi.UxperiSetFlags(flocCtx, cli)
	ctrl := floc.NewControl(flocCtx)

	// Wait for SIGINT OS signal and cancel the flow
	waitInterrupt := func(ctx floc.Context, ctrl floc.Control) error {
		c := make(chan os.Signal, 1)
		defer close(c)

		signal.Notify(c, os.Interrupt)

		// Wait for OS signal or flow finished signal
		select {
		case s := <-c:
			// OS signal was caught
			ctrl.Cancel(s)

		case <-ctx.Done():
			// The flow is finished
		}

		return nil
	}

	flow := run.Parallel(
		waitInterrupt,
		run.Sequence(append(pCtxcmd.InitSeq, pCtxcmd.RunSeq)...),
	)

	//TODO: validate RunWith when the job finish with errors
	if result, data, err = floc.RunWith(flocCtx, ctrl, flow); err != nil {
		if rcerr, e := uxperi.UxperiRCErrorTree(flocCtx); e != nil {
			rcerror = errortree.Add(rcerror, "context", e)
			rcerror = errortree.Add(rcerror, "cmd", fmt.Errorf("%s", ctx.Command()))
			rcerror = errortree.Add(rcerror, "msg", fmt.Errorf("error retrieving context values"))
		} else {
			rcerror = errortree.Add(*rcerr, "context", err)
			rcerror = errortree.Add(*rcerr, "cmd", fmt.Errorf("%s", ctx.Command()))
			rcerror = errortree.Add(*rcerr, "msg", fmt.Errorf("error running job sequence"))
		}
		ctx.FatalIfErrorf(rcerror)
	}
	// At this point the job has finished properly.
	// FIXME: Validate the way the result of the job is processed
	switch {
	case result.IsCanceled():
		pCtxcmd.Apps.Logger.WithFields(logger.Fields{
			"reason": data,
		}).Debug("Flow canceled by user")
	case result.IsCompleted():
		pCtxcmd.Apps.Logger.Debug("Flow succcesfully completed")
	case result.IsFailed():
		pCtxcmd.Apps.Logger.Debug("Flow failure")
	default:
		pCtxcmd.Apps.Logger.Debug("Flow finished with improper state")
		if rcerror, err := uxperi.UxperiRCErrorTree(flocCtx); err != nil {
			ctx.FatalIfErrorf(err)
		} else {
			ctx.FatalIfErrorf(*rcerror)
		}
	}
}
