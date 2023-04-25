package secretum

import (
	"fmt"

	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
)

var (
	// secretumContextKeyCLI     = secretumContextKey("cli")
	secretumContextKeyKmsCmd  = secretumContextKey("testcmd")
	secretumContextKeyRCError = secretumContextKey("rcerror")
)

type secretumContextKey string

func (c secretumContextKey) String() string {
	return "secretum." + string(c)
}

// SecretumKmsCmd gets a pointer to secretum.KmsCmd structure
func SecretumKmsCmd(ctx floc.Context) (KmsCmd, error) {
	var cmd KmsCmd
	var ok bool
	var rcerror error

	if cmd, ok = ctx.Value(secretumContextKeyKmsCmd).(KmsCmd); !ok {
		return KmsCmd{}, errortree.Add(rcerror, "KmsCmd", fmt.Errorf("type mismatch with key %s", secretumContextKeyKmsCmd))
	}

	return cmd, nil
}

func SecretumSetKmsCmd(ctx floc.Context, c KmsCmd) error {

	ctx.AddValue(secretumContextKeyKmsCmd, c)

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
