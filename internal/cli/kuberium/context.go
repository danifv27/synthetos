package kuberium

import (
	"fmt"

	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
)

var (
	kuberiumContextKeyK8sCmd  = kuberiumContextKey("k8scmd")
	kuberiumContextKeyRCError = kuberiumContextKey("rcerror")
	kuberiumContextKeyKmzCmd  = kuberiumContextKey("kmzcmd")
)

type kuberiumContextKey string

func (c kuberiumContextKey) String() string {
	return "kuberium." + string(c)
}

// KuberiumKubeCmd gets a pointer to kuberium.K8sCmd structure
func KuberiumKubeCmd(ctx floc.Context) (KubeCmd, error) {
	var cmd KubeCmd
	var ok bool
	var rcerror error

	if cmd, ok = ctx.Value(kuberiumContextKeyK8sCmd).(KubeCmd); !ok {
		return KubeCmd{}, errortree.Add(rcerror, "K8sCmd", fmt.Errorf("type mismatch with key %s", kuberiumContextKeyK8sCmd))
	}

	return cmd, nil
}

func KuberiumSetKubeCmd(ctx floc.Context, c KubeCmd) error {

	ctx.AddValue(kuberiumContextKeyK8sCmd, c)

	return nil
}

// KuberiumKmzCmd gets a pointer to kuberium.K8sCmd structure
func KuberiumKmzCmd(ctx floc.Context) (KmzCmd, error) {
	var cmd KmzCmd
	var ok bool
	var rcerror error

	if cmd, ok = ctx.Value(kuberiumContextKeyKmzCmd).(KmzCmd); !ok {
		return KmzCmd{}, errortree.Add(rcerror, "KmzCmd", fmt.Errorf("type mismatch with key %s", kuberiumContextKeyKmzCmd))
	}

	return cmd, nil
}

func KuberiumSetKmzCmd(ctx floc.Context, c KmzCmd) error {

	ctx.AddValue(kuberiumContextKeyKmzCmd, c)

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
