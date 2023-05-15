package secretum

import (
	"errors"
	"fmt"
	"time"

	"fry.org/cmo/cli/internal/application"
	"fry.org/cmo/cli/internal/application/actions"
	"fry.org/cmo/cli/internal/application/kms"
	"fry.org/cmo/cli/internal/application/printer"
	"fry.org/cmo/cli/internal/cli/common"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
	"github.com/workanator/go-floc/v3/guard"
	"github.com/workanator/go-floc/v3/run"
)

type KmsFortanixListSecretsCmd struct {
	GroupID *string                     `arg:"" help:"Group ID to be scanned" optional:""`
	Flags   KmsFortanixListSecretsFlags `embed:"" prefix:"kms.fortanix.list.secrets."`
}

type KmsFortanixListSecretsFlags struct {
	Decode bool `help:"decode secret value?." env:"SC_KMS_FORTANIX_LIST_SECRETS_DECODE" default:"false"`
}

func initializeKmsFortanixListSecretsCmd(ctx floc.Context, ctrl floc.Control) error {
	var err, rcerror error
	var c *common.Cmdctx
	// var cli CLI

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsFortanixListSecretsCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsFortanixListSecretsCmd", e)
		}
		return err
	}
	if err = application.WithOptions(&c.Apps,
		application.WithListSecretsQuery(c.Apps.Logger, c.Adapters.KeyManager),
		application.WithPrintSecretCommand(c.Apps.Logger, c.Adapters.Printer),
	); err != nil {
		return errortree.Add(rcerror, "initializeKmsFortanixListSecretsCmd", err)
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

func KmsFortanixListSecretsJob(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	var cmd KmsCmd
	var err error

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		SecretumSetRCErrorTree(ctx, "secretum.KmsFortanixListSecretsJob", err)
		return err
	}
	if cmd, err = SecretumKmsCmd(ctx); err != nil {
		SecretumSetRCErrorTree(ctx, "secretum.KmsFortanixListSecretsJob", err)
		return err
	}
	secretCh := make(chan kms.Secret, 3)
	quit := make(chan struct{})
	// Let's start the printer consumer
	m := cmd.Flags.Output
	reqPrint := actions.PrintSecretRequest{
		Mode:      printer.PrinterModeNone,
		ReceiveCh: secretCh,
		Decode:    cmd.Fortanix.List.Secrets.Flags.Decode,
	}
	switch {
	case m == "json":
		reqPrint.Mode = printer.PrinterModeJSON
	case m == "text":
		reqPrint.Mode = printer.PrinterModeText
	case m == "table":
		reqPrint.Mode = printer.PrinterModeTable
	}
	go func(req actions.PrintSecretRequest) {
		if err = c.Apps.Commands.PrintSecret.Handle(reqPrint); err != nil {
			SecretumSetRCErrorTree(ctx, "KmsFortanixListSecretsJob", err)
		}
		close(quit)
	}(reqPrint)
	//Start the producer
	reqListSecrets := actions.ListSecretsRequest{
		SendCh:  secretCh,
		GroupID: cmd.Fortanix.List.Secrets.GroupID,
	}
	go func(req actions.ListSecretsRequest) {
		if err = c.Apps.Queries.ListSecrets.Handle(req); err != nil {
			SecretumSetRCErrorTree(ctx, "KmsFortanixListSecretsJob", err)
		}
	}(reqListSecrets)
	//Wait until printer finish it work
	<-quit

	return nil
}

func (cmd *KmsFortanixListSecretsCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	p := cli.Kms.Flags.Probes
	//We need to append at the beginning to traverse the initseq in the right order
	c.InitSeq = append([]floc.Job{initializeKmsFortanixListSecretsCmd}, c.InitSeq...)
	c.RunSeq = guard.OnTimeout(
		guard.ConstTimeout(5*time.Minute),
		nil, // No need for timeout data
		run.Sequence(
			run.If(p.AreProbesEnabled, run.Background(startSecretumProbesServer)),
			KmsFortanixListSecretsJob,
			func(ctx floc.Context, ctrl floc.Control) error {
				if rcerror, err := SecretumRCErrorTree(ctx); err != nil {
					ctrl.Fail(fmt.Sprintf("Command '%s' internal error", c.Cmd), err)
					return err
				} else if *rcerror != nil {
					ctrl.Fail(fmt.Sprintf("Command '%s' failed", c.Cmd), *rcerror)
					return *rcerror
				}
				ctrl.Complete(fmt.Sprintf("Command '%s' completed", c.Cmd))

				return nil
			},
		),
		func(ctx floc.Context, ctrl floc.Control, id interface{}) {
			// Fail the flow on timeout
			msg := fmt.Sprintf("Command '%s' timeout expired", c.Cmd)
			SecretumSetRCErrorTree(ctx, "timeout", errors.New(msg))
			if rcerror, err := SecretumRCErrorTree(ctx); err != nil {
				ctrl.Fail(fmt.Sprintf("Command '%s' internal error", c.Cmd), err)
			} else {
				ctrl.Fail(msg, *rcerror)
			}
		},
	)

	return nil
}
