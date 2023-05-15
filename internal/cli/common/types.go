package common

import (
	"fry.org/cmo/cli/internal/application"
	"fry.org/cmo/cli/internal/infrastructure"
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
	Path string `help:"Configuration file path." env:"SC_CONFIG_PATH" optional:"" type:"path"`
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
