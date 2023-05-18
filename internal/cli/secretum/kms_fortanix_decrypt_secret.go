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

type KmsFortanixDecryptSecretCmd struct {
	ID    string                        `arg:"" help:"Secret ID or name of the secret to be decrypted"`
	Flags KmsFortanixDecryptSecretFlags `embed:"" prefix:"kms.fortanix.decrypt.secret."`
}

type KmsFortanixDecryptSecretFlags struct {
	Encode bool `help:"decode secret value?." env:"SC_KMS_FORTANIX_DECRYPT_SECRET_ENCODE" default:true negatable:""`
}

func initializeKmsFortanixDecryptSecretCmd(ctx floc.Context, ctrl floc.Control) error {
	var err, rcerror error
	var c *common.Cmdctx

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsFortanixDecryptSecretCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsFortanixDecryptSecretCmd", e)
		}
		return err
	}

	if err = application.WithOptions(&c.Apps,
		application.WithDecryptSecretsQuery(c.Apps.Logger, c.Adapters.KeyManager),
		application.WithPrintSecretCommand(c.Apps.Logger, c.Adapters.Printer),
	); err != nil {
		return errortree.Add(rcerror, "initializeKmsFortanixDecryptSecretCmd", err)
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

func kmsFortanixDecryptSecretJob(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	var cmd KmsCmd
	var err error

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		SecretumSetRCErrorTree(ctx, "secretum.kmsFortanixDecryptSecretJob", err)
		return err
	}
	if cmd, err = SecretumKmsCmd(ctx); err != nil {
		SecretumSetRCErrorTree(ctx, "secretum.kmsFortanixDecryptSecretJob", err)
		return err
	}
	secretCh := make(chan kms.Secret, 3)
	quit := make(chan struct{})
	// Let's start the printer consumer
	m := cmd.Flags.Output
	reqPrint := actions.PrintSecretRequest{
		Mode:      printer.PrinterModeNone,
		ReceiveCh: secretCh,
		Encode:    cmd.Fortanix.Decrypt.Secret.Flags.Encode,
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
			SecretumSetRCErrorTree(ctx, "kmsFortanixDecryptSecretJob", err)
		}
		close(quit)
	}(reqPrint)
	//Start the producer
	reqDecryptSecret := actions.DecryptSecretRequest{
		SendCh: secretCh,
		Name:   &cmd.Fortanix.Decrypt.Secret.ID,
	}
	go func(req actions.DecryptSecretRequest) {
		if err = c.Apps.Queries.DecryptSecret.Handle(req); err != nil {
			SecretumSetRCErrorTree(ctx, "kmsFortanixDecryptSecretJob", err)
		}
	}(reqDecryptSecret)
	//Wait until printer finish it work
	<-quit

	return nil
}

func (cmd *KmsFortanixDecryptSecretCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	p := cli.Kms.Flags.Probes
	//We need to append at the beginning to traverse the initseq in the right order
	c.InitSeq = append([]floc.Job{initializeKmsFortanixDecryptSecretCmd}, c.InitSeq...)
	c.RunSeq = guard.OnTimeout(
		guard.ConstTimeout(5*time.Minute),
		nil, // No need for timeout data
		run.Sequence(
			run.If(p.AreProbesEnabled, run.Background(startSecretumProbesServer)),
			kmsFortanixDecryptSecretJob,
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
