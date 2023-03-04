package uxperi

import (
	"context"
	"fmt"
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

type TestCmd struct {
	Flags TestFlags `embed:""`
}

type TestFlags struct {
	Enable  bool   `help:"enable actuator?." default:"true" prefix:"actuator." env:"SC_TEST_ACTUATOR_ENABLE" group:"actuator" negatable:""`
	Address string `help:"actuator adress with port" prefix:"actuator." default:":8081" env:"SC_TEST_ACTUATOR_ADDRESS" optional:"" group:"actuator"`
	// Root           string  `help:"actuator root" default:"/probe" env:"SC_TEST_ACTUATOR_ROOT" optional:"" group:"actuator"`
}

func initializeTestCmd(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	var err, rcerror error

	if c, err = CmdCtx(ctx); err != nil {
		if e := SetRCErrorTree(ctx, "initializeTestCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeTestCmd", e)
		}
		return err
	}

	infraOptions := []infrastructure.AdapterOption{
		infrastructure.WithHealthchecker(),
	}
	if err = infrastructure.AdapterWithOptions(&c.Adapters, infraOptions...); err != nil {
		if e := SetRCErrorTree(ctx, "initializeTestCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeTestCmd", e)
		}
		return err
	}
	//FIXME: Add proper k8s readiness
	c.Adapters.Healthchecker.AddReadinessCheck(
		"google-http",
		healthchecker.HTTPGetCheck("https://www.google.es", 50*time.Second),
	)

	if err = application.WithOptions(&c.Apps,
		application.WithHealthchecker(c.Adapters.Healthchecker),
	); err != nil {
		if e := SetRCErrorTree(ctx, "initializeTestCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeTestCmd", e)
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
		if e := SetRCErrorTree(ctx, "initializeTestCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeTestCmd", e)
		}
		return err
	}

	return nil
}

func testRunHealthServer(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	var cli CLI
	var err error

	if c, err = CmdCtx(ctx); err != nil {
		SetRCErrorTree(ctx, "testRunHealthServer", err)
		return err
	}
	if cli, err = Flags(ctx); err != nil {
		SetRCErrorTree(ctx, "testRunHealthServer", err)
		return err
	}

	// Start the server in a separate goroutine
	srv := &http.Server{
		Addr:    cli.Test.Flags.Address,
		Handler: c.Adapters.Healthchecker,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			SetRCErrorTree(ctx, "testRunHealthServer", err)
		}
	}()
	// Wait for the context to be canceled
	<-ctx.Done()
	// shut down the server gracefully
	// create a context with a timeout
	ct, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ct); err != nil {
		SetRCErrorTree(ctx, "testRunHealthServer", err)
	}

	return nil
}

func (cmd *TestCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	isActuatorEnabled := func(ctx floc.Context) bool {

		return cli.Test.Flags.Enable
	}

	waitForCancel := func(ctx floc.Context, ctrl floc.Control) error {

		// Wait for the context to be canceled
		<-ctx.Done()

		return nil
	}

	c.InitSeq = append(c.InitSeq, initializeTestCmd)

	c.RunSeq = run.Sequence(
		run.If(isActuatorEnabled, run.Background(testRunHealthServer)),
		run.If(isActuatorEnabled, waitForCancel),
		func(ctx floc.Context, ctrl floc.Control) error {
			if rcerror, err := RCErrorTree(ctx); err != nil {
				ctrl.Fail(fmt.Sprintf("Command '%s' internal error", c.Cmd), err)
				return err
			} else if *rcerror != nil {
				ctrl.Fail(fmt.Sprintf("Command '%s' failed", c.Cmd), *rcerror)
				return *rcerror
			}
			ctrl.Complete(fmt.Sprintf("Command '%s' completed", c.Cmd))

			return nil
		},
	)

	return nil
}
