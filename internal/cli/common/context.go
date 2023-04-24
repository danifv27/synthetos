package common

import (
	"fmt"

	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
)

var (
	commonContextKeyCmdCtx = commonContextKey("cmdctx")
)

type commonContextKey string

func (c commonContextKey) String() string {
	return "common." + string(c)
}

// CommonCmdCtx gets a pointer to the command context
func CommonCmdCtx(ctx floc.Context) (*Cmdctx, error) {
	var c *Cmdctx
	var ok bool
	var rcerror error

	obj := ctx.Value(commonContextKeyCmdCtx)
	if obj == nil {
		c = new(Cmdctx)
		ctx.AddValue(commonContextKeyCmdCtx, c)
	} else if c, ok = obj.(*Cmdctx); !ok {
		return nil, errortree.Add(rcerror, "CmdCtx", fmt.Errorf("type mismatch with key %s", commonContextKeyCmdCtx))
	}

	return c, nil
}

func CommonSetCmdCtx(ctx floc.Context, p Cmdctx) error {
	var c *Cmdctx
	var err, rcerror error

	if c, err = CommonCmdCtx(ctx); err == nil {
		*c = p
		return nil
	}

	return errortree.Add(rcerror, "SetCmdCtx", err)
}
