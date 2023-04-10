package secretum

import (
	"fmt"

	"fry.org/cmo/cli/internal/cli/common"
	"github.com/workanator/go-floc/v3"
	"github.com/workanator/go-floc/v3/run"
)

type KmsCmd struct {
	Flags KmsFlags   `embed:""`
	List  KmsListCmd `cmd:"" help:"KMS list."`
}

type KmsFlags struct {
}

func initializeKmsCmd(ctx floc.Context, ctrl floc.Control) error {

	return nil
}

func (cmd *KmsCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	waitForCancel := func(ctx floc.Context, ctrl floc.Control) error {

		// Wait for the context to be canceled
		<-ctx.Done()

		return nil
	}

	c.InitSeq = append(c.InitSeq, initializeKmsCmd)
	c.RunSeq = run.Sequence(
		waitForCancel,
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
	)

	return nil
}
