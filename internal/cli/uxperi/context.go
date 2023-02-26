package uxperi

import (
	"fmt"

	"fry.org/cmo/cli/internal/cli/common"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
)

var (
	contextKeyCLI     = contextKey("cli")
	contextKeyRCError = contextKey("rcerror")
	contextKeyCmdCtx  = contextKey("cmdctx")
)

type contextKey string

func (c contextKey) String() string {
	return "uxperi." + string(c)
}

// Flags gets a pointer to CLI structure
func Flags(ctx floc.Context) (CLI, error) {
	var cli CLI
	var ok bool
	var rcerror error

	if cli, ok = ctx.Value(contextKeyCLI).(CLI); !ok {
		return CLI{}, errortree.Add(rcerror, "Flags", fmt.Errorf("type mismatch with key %s", contextKeyCLI))
	}

	return cli, nil
}

func SetFlags(ctx floc.Context, c CLI) error {

	ctx.AddValue(contextKeyCLI, c)

	return nil
}

// RCError gets a pointer to errortree parent error
func RCErrorTree(ctx floc.Context) (*error, error) {
	var e *error
	var rcerror error
	var ok bool

	obj := ctx.Value(contextKeyRCError)
	if obj == nil {
		e = new(error)
		ctx.AddValue(contextKeyRCError, e)
	} else if e, ok = obj.(*error); !ok {
		return nil, errortree.Add(rcerror, "RCErrorTree", fmt.Errorf("type mismatch with key %s", contextKeyRCError))
	}

	return e, nil
}

func SetRCErrorTree(ctx floc.Context, key string, e error) error {
	var rcerror *error
	var err, rce error

	if rcerror, err = RCErrorTree(ctx); err == nil {
		*rcerror = errortree.Add(*rcerror, key, e)
	}

	return errortree.Add(rce, "SetRCErrorTree", err)
}

// CmdCtx gets a pointer to the command context
func CmdCtx(ctx floc.Context) (*common.Cmdctx, error) {
	var c *common.Cmdctx
	var ok bool
	var rcerror error

	obj := ctx.Value(contextKeyCmdCtx)
	if obj == nil {
		c = new(common.Cmdctx)
		ctx.AddValue(contextKeyCmdCtx, c)
	} else if c, ok = obj.(*common.Cmdctx); !ok {
		return nil, errortree.Add(rcerror, "NewApplications", fmt.Errorf("type mismatch with key %s", contextKeyCmdCtx))
	}

	return c, nil
}

func SetCmdCtx(ctx floc.Context, p common.Cmdctx) error {
	var c *common.Cmdctx
	var err, rcerror error

	if c, err = CmdCtx(ctx); err == nil {
		*c = p
	}

	return errortree.Add(rcerror, "SetCmdCtx", err)
}
