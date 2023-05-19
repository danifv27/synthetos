package kuberium

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
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
	Flags     KubeFlags        `embed:"" prefix:"kube."`
	Images    KubeImagesCmd    `cmd:"" help:"List images used in deployed kubernetes pods."`
	Resources KubeResourcesCmd `cmd:"" help:"List resources associated with deployed kubernetes objects."`
	Summary   KubeSummaryCmd   `cmd:"" help:"Show a summary of the objects deployed in a namespace or present in a kubernetes manifests."`
}

type KubeFlags struct {
	Probes    common.Probes `embed:"" prefix:"probes."`
	Namespace string        `help:"namespace" env:"SC_KUBE_CONFIG_NAMESPACE" required:""`
	Path      string        `help:"path to the kubeconfig file to use for requests or host url" env:"SC_KUBE_CONFIG_PATH" required:""`
	Context   string        `help:"the name of the kubeconfig context to use" env:"SC_KUBE_CONTEXT" required:""`
	Selector  *string       `help:"selector (label query) to filter on," env:"SC_KUBE_SELECTOR" short:"l"`
	Output    string        `help:"Format the output (table|json|text)." enum:"table,json,text" default:"table" env:"SC_KUBE_OUTPUT"`
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
	// provider:k8s?path=<kubeconfig_path>&context=<kubernetes_context>&namespace=<kubernetes_namespace>&selector=<kubernetes_object_selector>
	uri := fmt.Sprintf("provider:k8s?path=%s&context=%s&namespace=%s",
		url.QueryEscape(cmd.Flags.Path),
		url.QueryEscape(cmd.Flags.Context),
		url.QueryEscape(cmd.Flags.Namespace))
	if cmd.Flags.Selector != nil {
		uri = fmt.Sprintf("%s&selector=%s", uri, url.QueryEscape(*cmd.Flags.Selector))
	}
	infraOptions := []infrastructure.AdapterOption{
		infrastructure.WithHealthchecker(cmd.Flags.Probes.RootPrefix),
		infrastructure.WithTablePrinter(),
		infrastructure.WithResourceProvider(uri, c.Apps.Logger),
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

func startKubeProbesServer(ctx floc.Context, ctrl floc.Control) error {
	var err, rcerror error
	var c *common.Cmdctx
	var cmd KubeCmd

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "startKubeProbesServer", err); e != nil {
			return errortree.Add(rcerror, "startKubeProbesServer", e)
		}
		return err
	}
	if cmd, err = KuberiumKubeCmd(ctx); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "startKubeProbesServer", err); e != nil {
			return errortree.Add(rcerror, "startKubeProbesServer", e)
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
			KuberiumSetRCErrorTree(ctx, "kuberium.startKubeProbesServer", err)
		}
	}()
	// Wait for the context to be canceled
	<-ctx.Done()
	// shut down the server gracefully
	// create a context with a timeout
	ct, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ct); err != nil {
		KuberiumSetRCErrorTree(ctx, "kuberium.startKubeProbesServer", err)
	}

	return nil
}

func (cmd *KubeCmd) Run(c *common.Cmdctx, rcerror *error) error {

	c.InitSeq = append([]floc.Job{initializeKubeCmd}, c.InitSeq...)

	return nil
}
