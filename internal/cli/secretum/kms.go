package secretum

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

type KmsCmd struct {
	Flags KmsFlags   `embed:""`
	List  KmsListCmd `cmd:"" help:"KMS list."`
}

type KmsFlags struct {
	Probes common.Probes `embed:"" group:"probes"`
}

func initializeKmsCmd(ctx floc.Context, ctrl floc.Control) error {
	var err, rcerror error
	var c *common.Cmdctx
	var cli CLI

	if c, err = SecretumCmdCtx(ctx); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeTestCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeTestCmd", e)
		}
		return err
	}
	if cli, err = SecretumFlags(ctx); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeTestCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeTestCmd", e)
		}
		return err
	}
	infraOptions := []infrastructure.AdapterOption{
		infrastructure.WithHealthchecker(cli.Kms.Flags.Probes.RootPrefix),
	}
	if err = infrastructure.AdapterWithOptions(&c.Adapters, infraOptions...); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeTestCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeTestCmd", e)
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
		if e := SecretumSetRCErrorTree(ctx, "initializeTestCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeTestCmd", e)
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

	if c, err = SecretumCmdCtx(ctx); err != nil {
		SecretumSetRCErrorTree(ctx, "secretum.startProbesServer", err)
		return err
	}
	if cli, err = SecretumFlags(ctx); err != nil {
		SecretumSetRCErrorTree(ctx, "secretum.startProbesServer", err)
		return err
	}

	// Start the server in a separate goroutine
	srv := &http.Server{
		Addr:    cli.Kms.Flags.Probes.Address,
		Handler: c.Adapters.Healthchecker,
	}
	go func() {
		c.Apps.Logger.WithFields(logger.Fields{
			"rootPrefix": cli.Kms.Flags.Probes.RootPrefix,
			"address":    cli.Kms.Flags.Probes.Address,
		}).Info("Starting health endpoints")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			SecretumSetRCErrorTree(ctx, "secretum.startProbesServer", err)
		}
	}()
	// Wait for the context to be canceled
	<-ctx.Done()
	// shut down the server gracefully
	// create a context with a timeout
	ct, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ct); err != nil {
		SecretumSetRCErrorTree(ctx, "secretum.startProbesServer", err)
	}

	return nil
}

func (cmd *KmsCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	c.InitSeq = append([]floc.Job{initializeKmsCmd}, c.InitSeq...)

	return nil
}
