package versio

import (
	"fmt"

	"fry.org/cmo/cli/internal/cli/common"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
)

var (
	versioContextKeyCLI        = versioContextKey("cli")
	versioContextKeyRCError    = versioContextKey("rcerror")
	versioContextKeyCmdCtx     = versioContextKey("cmdctx")
	versioContextKeyVersionCmd = versioContextKey("versioncmd")
)

type versioContextKey string

func (c versioContextKey) String() string {
	return "versio." + string(c)
}

// VersioVersionCmd gets a pointer to versio.VersionCmd structure
func VersioVersionCmd(ctx floc.Context) (VersionCmd, error) {
	var cmd VersionCmd
	var ok bool
	var rcerror error

	if cmd, ok = ctx.Value(versioContextKeyVersionCmd).(VersionCmd); !ok {
		return VersionCmd{}, errortree.Add(rcerror, "VersionCmd", fmt.Errorf("type mismatch with key %s", versioContextKeyVersionCmd))
	}

	return cmd, nil
}

func VersioSetVersionCmd(ctx floc.Context, c VersionCmd) error {

	ctx.AddValue(versioContextKeyVersionCmd, c)

	return nil
}

// VersioFlags gets a pointer to CLI structure
func VersioFlags(ctx floc.Context) (CLI, error) {
	var cli CLI
	var ok bool
	var rcerror error

	if cli, ok = ctx.Value(versioContextKeyCLI).(CLI); !ok {
		return CLI{}, errortree.Add(rcerror, "Flags", fmt.Errorf("type mismatch with key %s", versioContextKeyCLI))
	}

	return cli, nil
}

func VersioSetFlags(ctx floc.Context, c CLI) error {

	ctx.AddValue(versioContextKeyCLI, c)

	return nil
}

// VersioRCErrorTree gets a pointer to errortree parent error
func VersioRCErrorTree(ctx floc.Context) (*error, error) {
	var e *error
	var rcerror error
	var ok bool

	obj := ctx.Value(versioContextKeyRCError)
	if obj == nil {
		e = new(error)
		ctx.AddValue(versioContextKeyRCError, e)
	} else if e, ok = obj.(*error); !ok {
		return nil, errortree.Add(rcerror, "RCErrorTree", fmt.Errorf("type mismatch with key %s", versioContextKeyRCError))
	}

	return e, nil
}

func VersioSetRCErrorTree(ctx floc.Context, key string, e error) error {
	var rcerror *error
	var err, rce error

	if rcerror, err = VersioRCErrorTree(ctx); err == nil {
		*rcerror = errortree.Add(*rcerror, key, e)
	}

	return errortree.Add(rce, "SetRCErrorTree", err)
}

// VersioCmdCtx gets a pointer to the command context
func VersioCmdCtx(ctx floc.Context) (*common.Cmdctx, error) {
	var c *common.Cmdctx
	var ok bool
	var rcerror error

	obj := ctx.Value(versioContextKeyCmdCtx)
	if obj == nil {
		c = new(common.Cmdctx)
		ctx.AddValue(versioContextKeyCmdCtx, c)
	} else if c, ok = obj.(*common.Cmdctx); !ok {
		return nil, errortree.Add(rcerror, "NewApplications", fmt.Errorf("type mismatch with key %s", versioContextKeyCmdCtx))
	}

	return c, nil
}

func VersioSetCmdCtx(ctx floc.Context, p common.Cmdctx) error {
	var c *common.Cmdctx
	var err, rcerror error

	if c, err = VersioCmdCtx(ctx); err == nil {
		*c = p
	}

	return errortree.Add(rcerror, "SetCmdCtx", err)
}
