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
	Enable     bool   `help:"enable actuator?." default:"true" prefix:"probes." env:"SC_TEST_PROBES_ENABLE" negatable:""`
	Address    string `help:"actuator adress with port" prefix:"probes." default:":8081" env:"SC_TEST_PROBES_ADDRESS" optional:""`
	RootPrefix string `help:"Prefix for the internal routes of web endpoints." prefix:"probes." env:"SC_TEST_PROBES_ROOT_PREFIX" default:"/actuator" optional:""`
	// Root           string  `help:"endpoint root" default:"/health" env:"SC_TEST_PROBES_ROOT" optional:"" group:"probes"`
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
