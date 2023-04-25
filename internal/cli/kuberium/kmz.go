package kuberium

import (
	"time"

	"fry.org/cmo/cli/internal/application"
	"fry.org/cmo/cli/internal/application/healthchecker"
	"fry.org/cmo/cli/internal/cli/common"
	"fry.org/cmo/cli/internal/infrastructure"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
)

type KmzCmd struct {
	Flags KmzFlags `embed:""`
	// Summary KmzSummaryCmd `cmd:"" help:"Show a summary of the objects present in a kubernetes manifests."`
}

type KmzFlags struct {
	Probes common.Probes `embed:"" group:"probes"`
}

func initializeKmzCmd(ctx floc.Context, ctrl floc.Control) error {
	var err, rcerror error
	var c *common.Cmdctx
	var cmd KmzCmd

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeKmsCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsCmd", e)
		}
		return err
	}
	if cmd, err = KuberiumKmzCmd(ctx); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeKmsCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsCmd", e)
		}
		return err
	}
	infraOptions := []infrastructure.AdapterOption{
		infrastructure.WithHealthchecker(cmd.Flags.Probes.RootPrefix),
	}
	if err = infrastructure.AdapterWithOptions(&c.Adapters, infraOptions...); err != nil {
		if e := KuberiumSetRCErrorTree(ctx, "initializeKmzCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmzCmd", e)
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
		if e := KuberiumSetRCErrorTree(ctx, "initializeKmzCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmzCmd", e)
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

func (cmd *KmzCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	c.InitSeq = append([]floc.Job{initializeKmzCmd}, c.InitSeq...)

	return nil
}
