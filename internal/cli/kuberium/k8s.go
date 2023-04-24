package kuberium

import (
	"context"
	"net/http"
	"time"

	"fry.org/cmo/cli/internal/application"
	"fry.org/cmo/cli/internal/application/healthchecker"
	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/cli/common"
	"fry.org/cmo/cli/internal/infrastructure"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
)

type K8sCmd struct {
	Flags   K8sFlags      `embed:""`
	Summary K8sSummaryCmd `cmd:"" help:"Show a summary of the objects deployed in a namespace or present in a kubernetes manifests."`
}

type K8sFlags struct {
	Probes common.Probes `embed:"" group:"probes"`
}

func initializeK8sCmd(ctx floc.Context, ctrl floc.Control) error {
	var err, rcerror error
	var c *common.Cmdctx
	var cli CLI

	if c, err = KuberiumCmdCtx(ctx); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeK8sCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeK8sCmd", e)
		}
		return err
	}
	if cli, err = KuberiumFlags(ctx); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeK8sCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeK8sCmd", e)
		}
		return err
	}
	infraOptions := []infrastructure.AdapterOption{
		infrastructure.WithHealthchecker(cli.K8s.Flags.Probes.RootPrefix),
	}
	if err = infrastructure.AdapterWithOptions(&c.Adapters, infraOptions...); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeK8sCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeK8sCmd", e)
		}
		return err
	}
	//TODO: Add proper k8s readiness and liveness
	c.Adapters.Healthchecker.AddReadinessCheck(
		"google-http",
		healthchecker.HTTPGetCheck("https://www.google.es", 10*time.Second),
	)
	c.Adapters.Healthchecker.AddLivenessCheck(
		"google-dns",
		healthchecker.DNSResolveCheck("www.google.es", 25*time.Second),
	)
	if err = application.WithOptions(&c.Apps,
		application.WithHealthchecker(c.Adapters.Healthchecker),
	); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeK8sCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeK8sCmd", e)
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

func startProbesServer(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	var cli CLI
	var err error

	if c, err = KuberiumCmdCtx(ctx); err != nil {
		KuberiumSetRCErrorTree(ctx, "kuberium.startProbesServer", err)
		return err
	}
	if cli, err = KuberiumFlags(ctx); err != nil {
		KuberiumSetRCErrorTree(ctx, "kuberium.startProbesServer", err)
		return err
	}

	// Start the server in a separate goroutine
	srv := &http.Server{
		Addr:    cli.K8s.Flags.Probes.Address,
		Handler: c.Adapters.Healthchecker,
	}
	go func() {
		c.Apps.Logger.WithFields(logger.Fields{
			"rootPrefix": cli.K8s.Flags.Probes.RootPrefix,
			"address":    cli.K8s.Flags.Probes.Address,
		}).Info("Starting health endpoints")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			KuberiumSetRCErrorTree(ctx, "kuberium.startProbesServer", err)
		}
	}()
	// Wait for the context to be canceled
	<-ctx.Done()
	// shut down the server gracefully
	// create a context with a timeout
	ct, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ct); err != nil {
		KuberiumSetRCErrorTree(ctx, "kuberium.startProbesServer", err)
	}

	return nil
}

func (cmd *K8sCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	c.InitSeq = append([]floc.Job{initializeK8sCmd}, c.InitSeq...)

	return nil
}