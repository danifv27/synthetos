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
	Flags    KmsFlags       `embed:""`
	Fortanix KmsFortanixCmd `cmd:"" help:"Manage Fortanix KMS instance."`
}

type KmsFlags struct {
	Output string        `prefix:"kms." help:"Format the output (table|json|text)." enum:"table,json,text" default:"table" env:"SC_KMS_OUTPUT"`
	Probes common.Probes `embed:"" prefix:"probes."`
}

func initializeKmsCmd(ctx floc.Context, ctrl floc.Control) error {
	var err, rcerror error
	var c *common.Cmdctx
	var cmd KmsCmd

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsCmd", e)
		}
		return err
	}
	if cmd, err = SecretumKmsCmd(ctx); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsCmd", e)
		}
		return err
	}
	infraOptions := []infrastructure.AdapterOption{
		infrastructure.WithHealthchecker(cmd.Flags.Probes.RootPrefix),
	}
	if err = infrastructure.AdapterWithOptions(&c.Adapters, infraOptions...); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsCmd", e)
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
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsCmd", e)
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

func startSecretumProbesServer(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	var cmd KmsCmd
	var err error

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		SecretumSetRCErrorTree(ctx, "secretum.startSecretumProbesServer", err)
		return err
	}
	if cmd, err = SecretumKmsCmd(ctx); err != nil {
		SecretumSetRCErrorTree(ctx, "secretum.startSecretumProbesServer", err)
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
		}).Info("Starting health endpoints")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			SecretumSetRCErrorTree(ctx, "secretum.startSecretumProbesServer", err)
		}
	}()
	// Wait for the context to be canceled
	<-ctx.Done()
	// shut down the server gracefully
	// create a context with a timeout
	ct, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ct); err != nil {
		SecretumSetRCErrorTree(ctx, "secretum.startSecretumProbesServer", err)
	}

	return nil
}

func (cmd *KmsCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	c.InitSeq = append([]floc.Job{initializeKmsCmd}, c.InitSeq...)

	return nil
}
