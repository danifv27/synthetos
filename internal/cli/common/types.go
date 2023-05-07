package common

import (
	"fmt"
	"reflect"
	"strings"

	"fry.org/cmo/cli/internal/application"
	"fry.org/cmo/cli/internal/infrastructure"
	"github.com/alecthomas/kong"
	"github.com/workanator/go-floc/v3"
)

type Log struct {
	Level        string `enum:"debug,info,warn,error,fatal" help:"Set the logging level (debug|info|warn|error|fatal)" default:"info" env:"SC_LOGGING_LEVEL" group:"logging"`
	DisableAudit bool   `env:"SC_LOGGING_DISABLE_AUDIT" help:"Disable auditing?." default:"true" hidden:"" group:"logging"`
	Json         bool   `env:"SC_LOGGING_OUTPUT_JSON" help:"If set the log output is formatted as a JSON." default:"false" group:"logging"`
}

type Probes struct {
	Enable     bool   `help:"enable actuator?." env:"SC_PROBES_ENABLE" default:"true" negatable:"" group:"probes"`
	Address    string `help:"actuator adress with port" default:":8081" env:"SC_PROBES_ADDRESS" optional:"" group:"probes"`
	RootPrefix string `help:"Prefix for the internal routes of web endpoints." env:"SC_PROBES_ROOT_PREFIX" default:"/actuator" optional:"" group:"probes"`
}

type Config struct {
	Config string `help:"Configuration file path." env:"SC_CONFIG_PATH" optional:"" type:"path"`
}

type Cmdctx struct {
	Cmd      string
	InitSeq  []floc.Job
	RunSeq   floc.Job
	Apps     application.Applications
	Adapters infrastructure.Adapters
	Ports    infrastructure.Ports
}

func (p Probes) AreProbesEnabled(ctx floc.Context) bool {

	return p.Enable
}

type K8sResource struct {
	Kind string
}

func (r K8sResource) Decode(ctx *kong.DecodeContext, target reflect.Value) error {

	values := ctx.Scan.PopWhile(func(t kong.Token) bool {
		return t.Type != kong.EOLToken && t.Type == kong.FlagValueToken
	})

	t := target.Type()
	//Reset the slice
	if t.Kind() == reflect.Slice {
		target.Set(reflect.MakeSlice(t, 0, 0))
	}
	for _, resource := range values {
		switch reflect.TypeOf(resource.Value).Kind() {
		case reflect.Slice:
			s := reflect.ValueOf((resource.Value))
			for i := 0; i < s.Len(); i++ {
				res := K8sResource{}
				res.Kind = fmt.Sprintf("%v", s.Index(i))
				target.Set(reflect.Append(target, reflect.ValueOf(res)))
			}
		case reflect.String:
			resourcesList := strings.Split(fmt.Sprintf("%v", resource.Value), ",")
			for _, r := range resourcesList {
				res := K8sResource{}
				res.Kind = r
				target.Set(reflect.Append(target, reflect.ValueOf(res)))
			}
		default:
			res := K8sResource{}
			res.Kind = fmt.Sprintf("%v", resource)
			target.Set(reflect.Append(target, reflect.ValueOf(res)))
		}
	}
	// If v represents a struct
	// v := target.FieldByName("Kind")
	// if v.IsValid() {
	// 	v.SetString(value)
	// }
	return nil
}
