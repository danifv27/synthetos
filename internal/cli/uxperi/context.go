package uxperi

import (
	"fmt"

	"fry.org/cmo/cli/internal/cli/common"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
)

var (
	uxperiContextKeyCLI     = uxperiContextKey("cli")
	uxperiContextKeyRCError = uxperiContextKey("rcerror")
	uxperiContextKeyCmdCtx  = uxperiContextKey("cmdctx")
)

type uxperiContextKey string

func (c uxperiContextKey) String() string {
	return "uxperi." + string(c)
}

// UxperiFlags gets a pointer to CLI structure
func UxperiFlags(ctx floc.Context) (CLI, error) {
	var cli CLI
	var ok bool
	var rcerror error

	if cli, ok = ctx.Value(uxperiContextKeyCLI).(CLI); !ok {
		return CLI{}, errortree.Add(rcerror, "UxperiFlags", fmt.Errorf("type mismatch with key %s", uxperiContextKeyCLI))
	}

	return cli, nil
}

func UxperiSetFlags(ctx floc.Context, c CLI) error {

	ctx.AddValue(uxperiContextKeyCLI, c)

	return nil
}

// UxperiRCErrorTree gets a pointer to errortree parent error
func UxperiRCErrorTree(ctx floc.Context) (*error, error) {
	var e *error
	var rcerror error
	var ok bool

	obj := ctx.Value(uxperiContextKeyRCError)
	if obj == nil {
		e = new(error)
		ctx.AddValue(uxperiContextKeyRCError, e)
	} else if e, ok = obj.(*error); !ok {
		return nil, errortree.Add(rcerror, "UxperiRCErrorTree", fmt.Errorf("type mismatch with key %s", uxperiContextKeyRCError))
	}

	return e, nil
}

func UxperiSetRCErrorTree(ctx floc.Context, key string, e error) error {
	var rcerror *error
	var err, rce error

	if rcerror, err = UxperiRCErrorTree(ctx); err == nil {
		*rcerror = errortree.Add(*rcerror, key, e)
	}

	return errortree.Add(rce, "SetRCErrorTree", err)
}

// UxperiCmdCtx gets a pointer to the command context
func UxperiCmdCtx(ctx floc.Context) (*common.Cmdctx, error) {
	var c *common.Cmdctx
	var ok bool
	var rcerror error

	obj := ctx.Value(uxperiContextKeyCmdCtx)
	if obj == nil {
		c = new(common.Cmdctx)
		ctx.AddValue(uxperiContextKeyCmdCtx, c)
	} else if c, ok = obj.(*common.Cmdctx); !ok {
		return nil, errortree.Add(rcerror, "UxperiCmdCtx", fmt.Errorf("type mismatch with key %s", uxperiContextKeyCmdCtx))
	}

	return c, nil
}

func UxperiSetCmdCtx(ctx floc.Context, p common.Cmdctx) error {
	var c *common.Cmdctx
	var err, rcerror error

	if c, err = UxperiCmdCtx(ctx); err == nil {
		*c = p
	}

	return errortree.Add(rcerror, "UxperiSetCmdCtx", err)
}
