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
	Probes  common.Probes `embed:"" group:"probes"`
	Metrics struct {
		Address    string `help:"actuator adress with port" prefix:"metrics." default:":8082" env:"SC_TEST_METRICS_ADDRESS" optional:"" `
		RootPrefix string `help:"Prefix for the internal routes of web endpoints." prefix:"metrics." env:"SC_TEST_METRICS_ROUTE_PREFIX" default:"/" optional:""`
	} `embed:"" group:"metrics"`
}

func initializeTestCmd(ctx floc.Context, ctrl floc.Control) error {
	var err, rcerror error
	var c *common.Cmdctx
	var cmd TestCmd
	var login, products iexporters.CucumberPlugin

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		if e := UxperiSetRCErrorTree(ctx, "initializeTestCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeTestCmd", e)
		}
		return err
	}
	if cmd, err = UxperiTestCmd(ctx); err != nil {
		if e := UxperiSetRCErrorTree(ctx, "initializeTestCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeTestCmd", e)
		}
		return err
	}

	if login, err = ifeatures.NewLoginPageFeature(cmd.Flags.FeaturesFolder,
		ifeatures.WithLoginPageAuth(cmd.Flags.Auth.Id, cmd.Flags.Auth.Password),
		ifeatures.WithLoginPageLogger(c.Apps.Logger),
		ifeatures.WithLoginPageSnapshotFolder(cmd.Flags.SnapshotsFolder),
	); err != nil {
		if e := UxperiSetRCErrorTree(ctx, "initializeTestCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeTestCmd", e)
		}
		return err
	}

	if products, err = ifeatures.NewProductsTabFeature(cmd.Flags.FeaturesFolder,
		ifeatures.WithProductsTabAuth(cmd.Flags.Auth.Id, cmd.Flags.Auth.Password),
		ifeatures.WithProductsTabLogger(c.Apps.Logger),
		ifeatures.WithProductsTabSnapshotFolder(cmd.Flags.SnapshotsFolder),
	); err != nil {
		if e := UxperiSetRCErrorTree(ctx, "initializeExporterCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeExporterCmd", e)
		}
		return err
	}

	infraOptions := []infrastructure.AdapterOption{
		infrastructure.WithHealthchecker(cmd.Flags.Probes.RootPrefix),
		infrastructure.WithCucumberExporter(
			iexporters.WithCucumberRootPrefix(cmd.Flags.Metrics.RootPrefix),
			iexporters.WithCucumberHistoryEndpoint(cmd.Flags.Metrics.RootPrefix),
			iexporters.WithCucumberTimeout(cmd.Flags.Timeout),
			iexporters.WithCucumberPlugin("loginPage", login),
			iexporters.WithCucumberPlugin("productsTab", products),
		),
	}
	if err = infrastructure.AdapterWithOptions(&c.Adapters, infraOptions...); err != nil {
		if e := UxperiSetRCErrorTree(ctx, "initializeTestCmd", err); e != nil {
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
		if e := UxperiSetRCErrorTree(ctx, "initializeTestCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeTestCmd", e)
		}
		return err
	}
	if err = common.CommonSetCmdCtx(ctx, common.Cmdctx{
		Cmd:      c.Cmd,
		InitSeq:  c.InitSeq,
		Apps:     c.Apps,
		Adapters: c.Adapters,
		Ports:    c.Ports,
	}); err != nil {
		if e := UxperiSetRCErrorTree(ctx, "initializeTestCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeTestCmd", e)
		}
		return err
	}

	return nil
}

func startTestProbesServer(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	var cmd TestCmd
	var err error

	if cmd, err = UxperiTestCmd(ctx); err != nil {
		UxperiSetRCErrorTree(ctx, "uxperi.startTestProbesServer", err)
		return err
	}
	if !cmd.Flags.Probes.AreProbesEnabled(ctx) {
		return nil
	}
	if c, err = common.CommonCmdCtx(ctx); err != nil {
		UxperiSetRCErrorTree(ctx, "uxperi.startTestProbesServer", err)
		return err
	}
	// Start the server in a separate goroutine
	srv := &http.Server{
		Addr:    cmd.Flags.Probes.Address,
		Handler: c.Adapters.Healthchecker,
	}
	go func() {
		c.Apps.Logger.WithFields(logger.Fields{
			"rootPrefix": cmd.Flags.Probes.RootPrefix,
			"address":    cmd.Flags.Probes.Address,
		}).Info("Starting health probes endpoints")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			UxperiSetRCErrorTree(ctx, "uxperi.startTestProbesServer", err)
		}
	}()
	// Wait for the context to be canceled
	<-ctx.Done()
	// shut down the server gracefully
	// create a context with a timeout
	ct, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ct); err != nil {
		UxperiSetRCErrorTree(ctx, "uxperi.startTestProbesServer", err)
	}

	return nil
}

func exporterRunMetricsServer(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	var cmd TestCmd
	var err error

	if cmd, err = UxperiTestCmd(ctx); err != nil {
		UxperiSetRCErrorTree(ctx, "exporterRunMetricsServer", err)
		return err
	}
	if c, err = common.CommonCmdCtx(ctx); err != nil {
		UxperiSetRCErrorTree(ctx, "uxperi.exporterRunMetricsServer", err)
		return err
	}
	// Start the server in a separate goroutine
	srv := &http.Server{
		Addr:    cmd.Flags.Metrics.Address,
		Handler: c.Adapters.CucumberExporter,
	}
	go func() {
		c.Apps.Logger.WithFields(logger.Fields{
			"rootPrefix": cmd.Flags.Metrics.RootPrefix,
			"address":    cmd.Flags.Metrics.Address,
		}).Info("Starting Prometheus metrics endpoint")
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

func (cmd *TestCmd) Run(c *common.Cmdctx, rcerror *error) error {

	waitForCancel := func(ctx floc.Context, ctrl floc.Control) error {

		// Wait for the context to be canceled
		<-ctx.Done()

		return nil
	}

	c.InitSeq = append(c.InitSeq, initializeTestCmd)

	c.RunSeq = run.Sequence(
		run.Background(exporterRunMetricsServer),
		run.Background(startTestProbesServer),
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
