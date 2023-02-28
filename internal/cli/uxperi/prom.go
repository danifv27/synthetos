package uxperi

import (
	"context"
	"net/http"
	"time"

	"fry.org/cmo/cli/internal/application"
	"fry.org/cmo/cli/internal/application/healthchecker"
	"fry.org/cmo/cli/internal/cli/common"
	"fry.org/cmo/cli/internal/infrastructure"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
	"github.com/workanator/go-floc/v3/run"
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
	c.Adapters.Healthchecker.AddReadinessCheck(
		"google-http",
		healthchecker.HTTPGetCheck("https://www.google.es", 50*time.Second),
	)

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

func promRunHealthServer(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	// var cli CLI
	var err error

	if c, err = CmdCtx(ctx); err != nil {
		SetRCErrorTree(ctx, "promRunHealthServer", err)
		return err
	}
	// if cli, err = Flags(ctx); err != nil {
	// 	SetRCErrorTree(ctx, "promRunHealthServer", err)
	// 	return err
	// }

	//FIXME: past healthcheck address through flags
	// Start the server in a separate goroutine
	srv := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: c.Adapters.Healthchecker,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			SetRCErrorTree(ctx, "promRunHealthServer", err)
		}
	}()
	// Wait for the context to be canceled
	<-ctx.Done()
	// shut down the server gracefully
	// create a context with a timeout
	ct, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ct); err != nil {
		SetRCErrorTree(ctx, "promRunHealthServer", err)
	}

	return nil
}

func (cmd *PromCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	c.InitSeq = append(c.InitSeq, initializePromCmd)
	c.RunSeq = run.Parallel(
		promRunHealthServer,
		// run.Sequence(
		// 	func(ctx floc.Context, ctrl floc.Control) error {

		// 		if rcerror, err := RCErrorTree(ctx); err != nil {
		// 			ctrl.Fail(fmt.Sprintf("Command '%s' internal error", c.Cmd), err)
		// 			return err
		// 		} else if *rcerror != nil {
		// 			ctrl.Fail(fmt.Sprintf("Command '%s' failed", c.Cmd), *rcerror)
		// 			return *rcerror
		// 		}
		// 		ctrl.Complete(fmt.Sprintf("Command '%s' completed", c.Cmd))

		// 		return nil
		// 	},
		// ),
	)

	return nil
}
