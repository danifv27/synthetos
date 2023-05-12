package secretum

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"fry.org/cmo/cli/internal/application"
	"fry.org/cmo/cli/internal/application/actions"
	"fry.org/cmo/cli/internal/application/provider"
	"fry.org/cmo/cli/internal/cli/common"
	"fry.org/cmo/cli/internal/infrastructure"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
	"github.com/workanator/go-floc/v3/guard"
	"github.com/workanator/go-floc/v3/run"
)

type KmsFortanixDecryptManifestsCmd struct {
	InputPath string                           `arg:"" help:"Input file or - for stdin" default:"-" type:"path"`
	Flags     KmsFortanixDecryptManifestsFlags `embed:""`
}

type KmsFortanixDecryptManifestsFlags struct{}

func initializeKmsFortanixDecryptManifestsCmd(ctx floc.Context, ctrl floc.Control) error {
	var err, rcerror error
	var c *common.Cmdctx
	var cmd KmsCmd

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsFortanixDecryptManifestsCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsFortanixDecryptManifestsCmd", e)
		}
		return errortree.Add(rcerror, "initializeKmsFortanixDecryptManifestsCmd", err)
	}
	if cmd, err = SecretumKmsCmd(ctx); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsFortanixDecryptManifestsCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsFortanixDecryptManifestsCmd", e)
		}
		return errortree.Add(rcerror, "initializeKmsFortanixDecryptManifestsCmd", err)
	}
	uri := fmt.Sprintf("provider:reader?path=%s", url.QueryEscape(cmd.Fortanix.Decrypt.Manifests.InputPath))
	infraOptions := []infrastructure.AdapterOption{
		infrastructure.WithManifestProvider(uri, c.Apps.Logger),
	}
	if err = infrastructure.AdapterWithOptions(&c.Adapters, infraOptions...); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsFortanixDecryptManifestsCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsFortanixDecryptManifestsCmd", e)
		}
		return errortree.Add(rcerror, "initializeKmsFortanixDecryptManifestsCmd", err)
	}
	if err = application.WithOptions(&c.Apps,
		application.WithListManifestsCommand(c.Apps.Logger, c.Adapters.ManifestProvider),
		application.WithPrintManifestsCommand(c.Apps.Logger, c.Adapters.Printer),
	); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsFortanixDecryptManifestsCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsFortanixDecryptManifestsCmd", e)
		}
		return errortree.Add(rcerror, "initializeKmsFortanixDecryptManifestsCmd", err)
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

func kmsFortanixDecryptManifestsJob(ctx floc.Context, ctrl floc.Control) error {
	var c *common.Cmdctx
	// var cmd KmsCmd
	var rcerror, err error

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "kmsFortanixDecryptManifestsJob", err); e != nil {
			return errortree.Add(rcerror, "kmsFortanixDecryptManifestsJob", e)
		}
		return errortree.Add(rcerror, "kmsFortanixDecryptManifestsJob", err)
	}
	// if cmd, err = SecretumKmsCmd(ctx); err != nil {
	// 	if e := SecretumSetRCErrorTree(ctx, "kmsFortanixDecryptManifestsJob", err); e != nil {
	// 		return errortree.Add(rcerror, "kmsFortanixDecryptManifestsJob", e)
	// 	}
	// 	return errortree.Add(rcerror, "kmsFortanixDecryptManifestsJob", err)
	// }

	manifestCh := make(chan provider.Manifest, 3)
	quit := make(chan struct{})
	// // Let's start the printer consumer
	reqPrint := actions.PrintManifestsRequest{
		ReceiveCh: manifestCh,
	}
	go func(req actions.PrintManifestsRequest) {
		if err = c.Apps.Commands.PrintManifests.Handle(reqPrint); err != nil {
			SecretumSetRCErrorTree(ctx, "kmsFortanixDecryptManifestsJob", err)
		}
		close(quit)
	}(reqPrint)
	//Start the producer
	reqListObjects := actions.ListManifestsObjectsRequest{
		SendCh:    manifestCh,
		InputPath: "-",
	}
	go func(req actions.ListManifestsObjectsRequest) {
		if err = c.Apps.Queries.ListManifests.Handle(req); err != nil {
			SecretumSetRCErrorTree(ctx, "kmsFortanixDecryptManifestsJob", err)
		}
	}(reqListObjects)
	//Wait until printer finish it work
	<-quit

	return nil
}

func (cmd *KmsFortanixDecryptManifestsCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	p := cli.Kms.Flags.Probes
	//We need to append at the beginning to traverse the initseq in the right order
	c.InitSeq = append([]floc.Job{initializeKmsFortanixDecryptManifestsCmd}, c.InitSeq...)
	c.RunSeq = guard.OnTimeout(
		guard.ConstTimeout(5*time.Minute),
		nil, // No need for timeout data
		run.Sequence(
			run.If(p.AreProbesEnabled, run.Background(startSecretumProbesServer)),
			kmsFortanixDecryptManifestsJob,
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
