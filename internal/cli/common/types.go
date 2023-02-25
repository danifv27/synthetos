package common

import (
	"fry.org/cmo/cli/internal/application"
	"fry.org/cmo/cli/internal/infrastructure"
	"github.com/workanator/go-floc/v3"
)

type Log struct {
	Level string `enum:"debug,info,warn,error,fatal" help:"Set the logging level (debug|info|warn|error|fatal)" default:"info" env:"OC_LOGGING_LEVEL" group:"logging"`
	// Format       string `help:"The log target and format. Example: logger:syslog?appname=bob&local=7 or logger:stdout?json=true" default:"logger:hooked?json=false" env:"OC_LOGGING_FORMAT" group:"logging"`
	DisableAudit bool `env:"OC_LOGGING_DISABLE_AUDIT" help:"Disable auditing?." default:"true" group:"logging"`
	Json         bool `env:"OC_LOGGING_OUTPUT_JSON" help:"The log output is formatted as a JSON string." default:"false" group:"logging"`
}
type Cmdctx struct {
	Cmd      string
	InitSeq  []floc.Job
	RunSeq   floc.Job
	Apps     application.Applications
	Adapters infrastructure.Adapters
	Ports    infrastructure.Ports
}
