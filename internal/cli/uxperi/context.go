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

	if cli, ok = ctx.Value(contextKeyCLI).(CLI); !ok {
		return CLI{}, fmt.Errorf("type mismatch with key %s", contextKeyCLI)
	}

	return cli, nil
}

func SetFlags(ctx floc.Context, c CLI) {

	ctx.AddValue(contextKeyCLI, c)
}

// RCError gets a pointer to errortree parent error
func RCErrorTree(ctx floc.Context) (*error, error) {
	var e *error
	var ok bool

	obj := ctx.Value(contextKeyRCError)
	if obj == nil {
		e = new(error)
		ctx.AddValue(contextKeyRCError, e)
	} else if e, ok = obj.(*error); !ok {
		return nil, fmt.Errorf("type mismatch with key %s", contextKeyRCError)
	}

	return e, nil
}

func SetRCErrorTree(ctx floc.Context, key string, e error) {
	var rcerror *error
	var err error

	if rcerror, err = RCErrorTree(ctx); err == nil {
		*rcerror = errortree.Add(*rcerror, key, e)
	}

}

// CmdCtx gets a pointer to the command context
func CmdCtx(ctx floc.Context) (*common.Cmdctx, error) {
	var c *common.Cmdctx
	var ok bool

	if c, ok = ctx.Value(contextKeyCmdCtx).(*common.Cmdctx); !ok {
		return nil, fmt.Errorf("type mismatch with key %s", contextKeyCmdCtx)
	}

	return c, nil
}

func SetCmdCtx(ctx floc.Context, p *common.Cmdctx) {

	ctx.AddValue(contextKeyCmdCtx, p)
}
