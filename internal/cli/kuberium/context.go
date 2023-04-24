package kuberium

import (
	"fmt"

	"fry.org/cmo/cli/internal/cli/common"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
)

var (
	kuberiumContextKeyCLI     = kuberiumContextKey("cli")
	kuberiumContextKeyRCError = kuberiumContextKey("rcerror")
	kuberiumContextKeyCmdCtx  = kuberiumContextKey("cmdctx")
)

type kuberiumContextKey string

func (c kuberiumContextKey) String() string {
	return "kuberium." + string(c)
}

// KuberiumFlags gets a pointer to CLI structure
func KuberiumFlags(ctx floc.Context) (CLI, error) {
	var cli CLI
	var ok bool
	var rcerror error

	if cli, ok = ctx.Value(kuberiumContextKeyCLI).(CLI); !ok {
		return CLI{}, errortree.Add(rcerror, "Flags", fmt.Errorf("type mismatch with key %s", kuberiumContextKeyCLI))
	}

	return cli, nil
}

func KuberiumSetFlags(ctx floc.Context, c CLI) error {

	ctx.AddValue(kuberiumContextKeyCLI, c)

	return nil
}

// KuberiumRCErrorTree gets a pointer to errortree parent error
func KuberiumRCErrorTree(ctx floc.Context) (*error, error) {
	var e *error
	var rcerror error
	var ok bool

	obj := ctx.Value(kuberiumContextKeyRCError)
	if obj == nil {
		e = new(error)
		ctx.AddValue(kuberiumContextKeyRCError, e)
	} else if e, ok = obj.(*error); !ok {
		return nil, errortree.Add(rcerror, "RCErrorTree", fmt.Errorf("type mismatch with key %s", kuberiumContextKeyRCError))
	}

	return e, nil
}

func KuberiumSetRCErrorTree(ctx floc.Context, key string, e error) error {
	var rcerror *error
	var err, rce error

	if rcerror, err = KuberiumRCErrorTree(ctx); err == nil {
		*rcerror = errortree.Add(*rcerror, key, e)
	}

	return errortree.Add(rce, "SetRCErrorTree", err)
}

// KuberiumCmdCtx gets a pointer to the command context
func KuberiumCmdCtx(ctx floc.Context) (*common.Cmdctx, error) {
	var c *common.Cmdctx
	var ok bool
	var rcerror error

	obj := ctx.Value(kuberiumContextKeyCmdCtx)
	if obj == nil {
		c = new(common.Cmdctx)
		ctx.AddValue(kuberiumContextKeyCmdCtx, c)
	} else if c, ok = obj.(*common.Cmdctx); !ok {
		return nil, errortree.Add(rcerror, "NewApplications", fmt.Errorf("type mismatch with key %s", kuberiumContextKeyCmdCtx))
	}

	return c, nil
}

func KuberiumSetCmdCtx(ctx floc.Context, p common.Cmdctx) error {
	var c *common.Cmdctx
	var err, rcerror error

	if c, err = KuberiumCmdCtx(ctx); err == nil {
		*c = p
	}

	return errortree.Add(rcerror, "SetCmdCtx", err)
}
