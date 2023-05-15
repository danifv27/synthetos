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
	"fry.org/cmo/cli/internal/cli/secretum"
	"fry.org/cmo/cli/internal/cli/versio"
	"fry.org/cmo/cli/internal/infrastructure"

	"github.com/alecthomas/kong"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
	"github.com/workanator/go-floc/v3/run"
)

func initializeCmd(cli *secretum.CLI, cmd string) (common.Cmdctx, error) {
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

func getConfigPaths() ([]string, error) {
	var rcerror, err error
	var ex string
	var paths []string

	bin := filepath.Base(os.Args[0])
	if ex, err = os.Executable(); err != nil {
		return paths, errortree.Add(rcerror, "getConfigPaths", err)
	}
	exPath := filepath.Dir(ex)
	exBin := filepath.Base(ex)
	paths = append(paths, fmt.Sprintf("/etc/%s.json", bin))
	paths = append(paths, fmt.Sprintf("~/.%s.json", bin))
	paths = append(paths, fmt.Sprintf("%s/.%s.json", exPath, exBin))

	c := secretum.CLI{}
	kong.Parse(&c)
	if c.Config.Path != "" {
		paths = append(paths, c.Config.Path)
	}

	return paths, nil
}

func main() {
	var err, rcerror error
	var pCtxcmd *common.Cmdctx
	var result floc.Result
	var data interface{}
	var configs []string

	cli := secretum.CLI{
		Logging: common.Log{},
	}
	pCtxcmd = new(common.Cmdctx)
	if configs, err = getConfigPaths(); err != nil {
		panic(err)
	}
	bin := filepath.Base(os.Args[0])
	//config file has precedence over envars
	ctx := kong.Parse(&cli,
		kong.Bind(pCtxcmd),
		kong.Bind(&rcerror),
		kong.Name(bin),
		kong.Description("KMS manager"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Tree: true,
		}),
		// kong.TypeMapper(reflect.TypeOf([]common.K8sResource{}), common.K8sResource{}),
		kong.Configuration(kong.JSON, configs...),
	)
	if *pCtxcmd, err = initializeCmd(&cli, ctx.Command()); err != nil {
		ctx.FatalIfErrorf(err)
	}
	pCtxcmd.Apps.Logger.WithFields(logger.Fields{
		"cmd": pCtxcmd.Cmd,
	}).Debug("Run CLI command")
	//Run should create the job flow that will be executed as a sequence
	if err = ctx.Run(&cli); err != nil {
		rcerror = errortree.Add(rcerror, "context", err)
		rcerror = errortree.Add(rcerror, "cmd", fmt.Errorf("%s", ctx.Command()))
		rcerror = errortree.Add(rcerror, "msg", fmt.Errorf("can not execute '%s' command", ctx.Command()))
		ctx.FatalIfErrorf(rcerror)
	}

	flocCtx := floc.NewContext()
	common.CommonSetCmdCtx(flocCtx, *pCtxcmd)
	secretum.SecretumSetKmsCmd(flocCtx, cli.Kms)
	versio.VersioSetVersionCmd(flocCtx, cli.Version)
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

	seq := append(pCtxcmd.InitSeq, pCtxcmd.RunSeq)
	jobs := make([]floc.Job, 0)
	for _, item := range seq {
		if item != nil {
			jobs = append(jobs, item)
		}
	}
	//Last command
	jobs = append(jobs,
		func(ctx floc.Context, ctrl floc.Control) error {
			if rcerror, err := secretum.SecretumRCErrorTree(ctx); err != nil {
				ctrl.Fail(fmt.Sprintf("Command '%s' internal error", pCtxcmd.Cmd), err)
				return err
			} else if *rcerror != nil {
				ctrl.Fail(fmt.Sprintf("Command '%s' failed", pCtxcmd.Cmd), *rcerror)
				return *rcerror
			}
			ctrl.Complete(fmt.Sprintf("Command '%s' completed", pCtxcmd.Cmd))

			return nil
		},
	)
	//Run command are traversed starting from kms/list/fortanix/groups to kms
	flow := run.Parallel(
		waitInterrupt,
		run.Sequence(jobs...),
	)

	//TODO: validate RunWith when the job finish with errors
	if result, data, err = floc.RunWith(flocCtx, ctrl, flow); err != nil {
		if rcerr, e := secretum.SecretumRCErrorTree(flocCtx); e != nil {
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
		if rcerror, err := secretum.SecretumRCErrorTree(flocCtx); err != nil {
			ctx.FatalIfErrorf(err)
		} else {
			ctx.FatalIfErrorf(*rcerror)
		}
	}
}
