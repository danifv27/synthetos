package uxperi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"fry.org/cmo/cli/internal/application"
	// "fry.org/cmo/cli/internal/application/healthchecker"
	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/cli/common"
	"fry.org/cmo/cli/internal/infrastructure"
	iexporters "fry.org/cmo/cli/internal/infrastructure/exporters"
	ifeatures "fry.org/cmo/cli/internal/infrastructure/exporters/features"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
	"github.com/workanator/go-floc/v3/run"
)

type TestCmd struct {
	Flags TestFlags `embed:""`
}

type TestFlags struct {
	FeaturesFolder  string        `help:"path to gherkin features folder" prefix:"test." hidden:"" default:"./features" env:"SC_TEST_FEATURES_FOLDER"`
	SnapshotsFolder string        `help:"path to chromedp snapshots folder" prefix:"test." hidden:"" default:"./snapshots" env:"SC_TEST_SNAPSHOTS_FOLDER"`
	Timeout         time.Duration `help:"maximum amount of time that we should wait for a step or scenario to complete before timing out and marking the test as failed" prefix:"test." default:"1m" env:"SC_TEST_TIMEOUT"`
	// TargetURL      string        `help:"URL to check against" prefix:"test." env:"SC_TEST_TARGET_URL"`
	Auth struct {
		Id       string `help:"name used for authentication" prefix:"test." env:"SC_TEST_AZURE_USERNAME" hidden:""`
		Password string `help:"password used for authentication" prefix:"test." env:"SC_TEST_AZURE_PASSWORD" hidden:""`
	} `embed:"" group:"auth"`
	Probes struct {
		Enable     bool   `help:"enable actuator?." default:"true" prefix:"probes." env:"SC_TEST_PROBES_ENABLE" negatable:""`
		Address    string `help:"actuator adress with port" prefix:"probes." default:":8081" env:"SC_TEST_PROBES_ADDRESS" optional:""`
		RootPrefix string `help:"Prefix for the internal routes of web endpoints." prefix:"probes." env:"SC_TEST_PROBES_ROOT_PREFIX" default:"/actuator" optional:""`
		// Root           string  `help:"endpoint root" default:"/health" env:"SC_TEST_PROBES_ROOT" optional:"" group:"probes"`
	} `embed:"" group:"probes"`
	Metrics struct {
		Address    string `help:"actuator adress with port" prefix:"metrics." default:":8082" env:"SC_TEST_METRICS_ADDRESS" optional:"" `
		RootPrefix string `help:"Prefix for the internal routes of web endpoints." prefix:"metrics." env:"SC_TEST_METRICS_ROUTE_PREFIX" default:"/" optional:""`
	} `embed:"" group:"metrics"`
}

func initializeTestCmd(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	var err, rcerror error
	var cli CLI
	var login iexporters.CucumberPlugin

	if c, err = UxperiCmdCtx(ctx); err != nil {
		if e := UxperiSetRCErrorTree(ctx, "initializeExporterCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeExporterCmd", e)
		}
		return err
	}
	if cli, err = UxperiFlags(ctx); err != nil {
		if e := UxperiSetRCErrorTree(ctx, "initializeExporterCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeExporterCmd", e)
		}
		return err
	}

	if login, err = ifeatures.NewLoginPageFeature(cli.Test.Flags.FeaturesFolder,
		ifeatures.WithLoginPageAuth(cli.Test.Flags.Auth.Id, cli.Test.Flags.Auth.Password),
		ifeatures.WithLoginPageLogger(c.Apps.Logger),
		ifeatures.WithLoginPageSnapshotFolder(cli.Test.Flags.SnapshotsFolder),
	); err != nil {
		if e := UxperiSetRCErrorTree(ctx, "initializeExporterCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeTestCmd", e)
		}
		return err
	}
	infraOptions := []infrastructure.AdapterOption{
		infrastructure.WithHealthchecker(cli.Test.Flags.Probes.RootPrefix),
		infrastructure.WithCucumberExporter(
			iexporters.WithCucumberRootPrefix(cli.Test.Flags.Metrics.RootPrefix),
			iexporters.WithCucumberHistoryEndpoint(cli.Test.Flags.Metrics.RootPrefix),
			iexporters.WithCucumberTimeout(cli.Test.Flags.Timeout),
			iexporters.WithCucumberPlugin("loginPage", login),
		),
	}
	if err = infrastructure.AdapterWithOptions(&c.Adapters, infraOptions...); err != nil {
		if e := UxperiSetRCErrorTree(ctx, "initializeExporterCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeTestCmd", e)
		}
		return err
	}

	//TODO: Add proper k8s readiness and liveness
	// c.Adapters.Healthchecker.AddReadinessCheck(
	// 	"google-http",
	// 	healthchecker.HTTPGetCheck("https://www.google.es", 50*time.Second),
	// )

	if err = application.WithOptions(&c.Apps,
		application.WithHealthchecker(c.Adapters.Healthchecker),
	); err != nil {
		if e := UxperiSetRCErrorTree(ctx, "initializeExporterCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeTestCmd", e)
		}
		return err
	}
	if err = UxperiSetCmdCtx(ctx, common.Cmdctx{
		Cmd:      c.Cmd,
		InitSeq:  c.InitSeq,
		Apps:     c.Apps,
		Adapters: c.Adapters,
		Ports:    c.Ports,
	}); err != nil {
		if e := UxperiSetRCErrorTree(ctx, "initializeTestCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeExporterCmd", e)
		}
		return err
	}

	return nil
}

func exporterRunHealthServer(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	var cli CLI
	var err error

	if c, err = UxperiCmdCtx(ctx); err != nil {
		UxperiSetRCErrorTree(ctx, "exporterRunHealthServer", err)
		return err
	}
	if cli, err = UxperiFlags(ctx); err != nil {
		UxperiSetRCErrorTree(ctx, "exporterRunHealthServer", err)
		return err
	}

	// Start the server in a separate goroutine
	srv := &http.Server{
		Addr:    cli.Test.Flags.Probes.Address,
		Handler: c.Adapters.Healthchecker,
	}
	go func() {
		c.Apps.Logger.WithFields(logger.Fields{
			"address": cli.Test.Flags.Probes.Address,
		}).Debug("Starting health server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			UxperiSetRCErrorTree(ctx, "exporterRunHealthServer", err)
		}
	}()
	// Wait for the context to be canceled
	<-ctx.Done()
	// shut down the server gracefully
	// create a context with a timeout
	ct, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ct); err != nil {
		UxperiSetRCErrorTree(ctx, "exporterRunHealthServer", err)
	}

	return nil
}

func exporterRunMetricsServer(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	var cli CLI
	var err error

	if c, err = UxperiCmdCtx(ctx); err != nil {
		UxperiSetRCErrorTree(ctx, "exporterRunMetricsServer", err)
		return err
	}
	if cli, err = UxperiFlags(ctx); err != nil {
		UxperiSetRCErrorTree(ctx, "exporterRunMetricsServer", err)
		return err
	}

	// Start the server in a separate goroutine
	srv := &http.Server{
		Addr:    cli.Test.Flags.Metrics.Address,
		Handler: c.Adapters.CucumberExporter,
	}
	go func() {
		c.Apps.Logger.WithFields(logger.Fields{
			"address": cli.Test.Flags.Metrics.Address,
		}).Debug("Starting metrics server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			UxperiSetRCErrorTree(ctx, "exporterRunMetricsServer", err)
		}
	}()
	// Wait for the context to be canceled
	<-ctx.Done()
	// shut down the server gracefully
	// create a context with a timeout
	ct, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ct); err != nil {
		UxperiSetRCErrorTree(ctx, "exporterRunMetricsServer<", err)
	}

	return nil
}

func (cmd *TestCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	isActuatorEnabled := func(ctx floc.Context) bool {

		return cli.Test.Flags.Probes.Enable
	}

	waitForCancel := func(ctx floc.Context, ctrl floc.Control) error {

		// Wait for the context to be canceled
		<-ctx.Done()

		return nil
	}

	c.InitSeq = append(c.InitSeq, initializeTestCmd)

	c.RunSeq = run.Sequence(
		run.Background(exporterRunMetricsServer),
		run.If(isActuatorEnabled, run.Background(exporterRunHealthServer)),
		waitForCancel,
		func(ctx floc.Context, ctrl floc.Control) error {
			if rcerror, err := UxperiRCErrorTree(ctx); err != nil {
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
