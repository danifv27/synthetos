package secretum

import (
	"fmt"

	"fry.org/cmo/cli/internal/cli/common"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
)

var (
	secretumContextKeyCLI     = secretumContextKey("cli")
	secretumContextKeyRCError = secretumContextKey("rcerror")
	secretumContextKeyCmdCtx  = secretumContextKey("cmdctx")
)

type secretumContextKey string

func (c secretumContextKey) String() string {
	return "secretum." + string(c)
}

// SecretumFlags gets a pointer to CLI structure
func SecretumFlags(ctx floc.Context) (CLI, error) {
	var cli CLI
	var ok bool
	var rcerror error

	if cli, ok = ctx.Value(secretumContextKeyCLI).(CLI); !ok {
		return CLI{}, errortree.Add(rcerror, "Flags", fmt.Errorf("type mismatch with key %s", secretumContextKeyCLI))
	}

	return cli, nil
}

func SecretumSetFlags(ctx floc.Context, c CLI) error {

	ctx.AddValue(secretumContextKeyCLI, c)

	return nil
}

// SecretumRCErrorTree gets a pointer to errortree parent error
func SecretumRCErrorTree(ctx floc.Context) (*error, error) {
	var e *error
	var rcerror error
	var ok bool

	obj := ctx.Value(secretumContextKeyRCError)
	if obj == nil {
		e = new(error)
		ctx.AddValue(secretumContextKeyRCError, e)
	} else if e, ok = obj.(*error); !ok {
		return nil, errortree.Add(rcerror, "RCErrorTree", fmt.Errorf("type mismatch with key %s", secretumContextKeyRCError))
	}

	return e, nil
}

func SecretumSetRCErrorTree(ctx floc.Context, key string, e error) error {
	var rcerror *error
	var err, rce error

	if rcerror, err = SecretumRCErrorTree(ctx); err == nil {
		*rcerror = errortree.Add(*rcerror, key, e)
	}

	return errortree.Add(rce, "SetRCErrorTree", err)
}

// SecretumCmdCtx gets a pointer to the command context
func SecretumCmdCtx(ctx floc.Context) (*common.Cmdctx, error) {
	var c *common.Cmdctx
	var ok bool
	var rcerror error

	obj := ctx.Value(secretumContextKeyCmdCtx)
	if obj == nil {
		c = new(common.Cmdctx)
		ctx.AddValue(secretumContextKeyCmdCtx, c)
	} else if c, ok = obj.(*common.Cmdctx); !ok {
		return nil, errortree.Add(rcerror, "NewApplications", fmt.Errorf("type mismatch with key %s", secretumContextKeyCmdCtx))
	}

	return c, nil
}

func SecretumSetCmdCtx(ctx floc.Context, p common.Cmdctx) error {
	var c *common.Cmdctx
	var err, rcerror error

	if c, err = SecretumCmdCtx(ctx); err == nil {
		*c = p
	}

	return errortree.Add(rcerror, "SetCmdCtx", err)
}
