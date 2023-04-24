package synthetos

import (
	"fmt"

	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
)

var (
	// synthetosContextKeyCLI     = synthetosContextKey("cli")
	synthetosContextKeyRCError = synthetosContextKey("rcerror")
)

type synthetosContextKey string

func (c synthetosContextKey) String() string {
	return "synthetos." + string(c)
}

// // SynthetosFlags gets a pointer to CLI structure
// func SynthetosFlags(ctx floc.Context) (CLI, error) {
// 	var cli CLI
// 	var ok bool
// 	var rcerror error

// 	if cli, ok = ctx.Value(synthetosContextKeyCLI).(CLI); !ok {
// 		return CLI{}, errortree.Add(rcerror, "Flags", fmt.Errorf("type mismatch with key %s", synthetosContextKeyCLI))
// 	}

// 	return cli, nil
// }

// func SynthetosSetFlags(ctx floc.Context, c CLI) error {

// 	ctx.AddValue(synthetosContextKeyCLI, c)

// 	return nil
// }

// SynthetosRCErrorTree gets a pointer to errortree parent error
func SynthetosRCErrorTree(ctx floc.Context) (*error, error) {
	var e *error
	var rcerror error
	var ok bool

	obj := ctx.Value(synthetosContextKeyRCError)
	if obj == nil {
		e = new(error)
		ctx.AddValue(synthetosContextKeyRCError, e)
	} else if e, ok = obj.(*error); !ok {
		return nil, errortree.Add(rcerror, "RCErrorTree", fmt.Errorf("type mismatch with key %s", synthetosContextKeyRCError))
	}

	return e, nil
}

func SynthetosSetRCErrorTree(ctx floc.Context, key string, e error) error {
	var rcerror *error
	var err, rce error

	if rcerror, err = SynthetosRCErrorTree(ctx); err == nil {
		*rcerror = errortree.Add(*rcerror, key, e)
	}

	return errortree.Add(rce, "SetRCErrorTree", err)
}
