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

type KubeCmd struct {
	Flags   KubeFlags      `embed:""`
	Summary KubeSummaryCmd `cmd:"" help:"Show a summary of the objects deployed in a namespace or present in a kubernetes manifests."`
}

type KubeFlags struct {
	Probes    common.Probes `embed:"" group:"probes"`
	Namespace string        `help:"namespace" prefix:"kube." env:"SC_KUBE_CONFIG_NAMESPACE" required:""`
	Path      string        `help:"path to the kubeconfig file to use for requests or host url" prefix:"kube." env:"SC_KUBE_CONFIG_PATH" required:""`
	Context   string        `help:"the name of the kubeconfig context to use" prefix:"kube." env:"SC_KUBE_CONTEXT" required:""`
	Selector  *string       `help:"selector (label query) to filter on," prefix:"kube." env:"SC_KUBE_SELECTOR" short:"l"`
}

func initializeKubeCmd(ctx floc.Context, ctrl floc.Control) error {
	var err, rcerror error
	var c *common.Cmdctx
	var cmd KubeCmd

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeKubeCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKubeCmd", e)
		}
		return err
	}
	if cmd, err = KuberiumKubeCmd(ctx); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeKubeCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKubeCmd", e)
		}
		return err
	}
	infraOptions := []infrastructure.AdapterOption{
		infrastructure.WithHealthchecker(cmd.Flags.Probes.RootPrefix),
	}
	if err = infrastructure.AdapterWithOptions(&c.Adapters, infraOptions...); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeKubeCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKubeCmd", e)
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
		if e := KuberiumSetRCErrorTree(ctx, "initializeKubeCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKubeCmd", e)
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
	var err, rcerror error
	var c *common.Cmdctx
	var cmd KubeCmd

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "startProbesServer", err); e != nil {
			return errortree.Add(rcerror, "startProbesServer", e)
		}
		return err
	}
	if cmd, err = KuberiumKubeCmd(ctx); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "startProbesServer", err); e != nil {
			return errortree.Add(rcerror, "startProbesServer", e)
		}
		return err
	}
	p := cmd.Flags.Probes
	if !p.AreProbesEnabled(ctx) {
		c.Apps.Logger.Debug("Probes not enabled")
		return nil
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

func (cmd *KubeCmd) Run(c *common.Cmdctx, rcerror *error) error {

	c.InitSeq = append([]floc.Job{initializeKubeCmd}, c.InitSeq...)

	return nil
}
