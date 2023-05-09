package uxperi

import (
	"fmt"

	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
)

var (
	uxperiContextKeyRCError = uxperiContextKey("rcerror")
	uxperiContextKeyTestCmd = uxperiContextKey("testcmd")
)

type uxperiContextKey string

func (c uxperiContextKey) String() string {
	return "uxperi." + string(c)
}

// UxperiTestCmd gets a pointer to uxperi.TestCmd structure
func UxperiTestCmd(ctx floc.Context) (TestCmd, error) {
	var cmd TestCmd
	var ok bool
	var rcerror error

	if cmd, ok = ctx.Value(uxperiContextKeyTestCmd).(TestCmd); !ok {
		return TestCmd{}, errortree.Add(rcerror, "TestCmd", fmt.Errorf("type mismatch with key %s", uxperiContextKeyTestCmd))
	}

	return cmd, nil
}

func UxperiSetTestCmd(ctx floc.Context, c TestCmd) error {

	ctx.AddValue(uxperiContextKeyTestCmd, c)

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
