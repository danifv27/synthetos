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
type Cmdctx struct {
	Cmd      string
	InitSeq  []floc.Job
	RunSeq   floc.Job
	Apps     application.Applications
	Adapters infrastructure.Adapters
	Ports    infrastructure.Ports
}
