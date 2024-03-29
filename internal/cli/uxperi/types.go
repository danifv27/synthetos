package uxperi

import "fry.org/cmo/cli/internal/cli/common"

type CLI struct {
	Logging common.Log `embed:"" prefix:"logging."`
	Version VersionCmd `cmd:"" help:"Show version information"`
	Test    TestCmd    `cmd:"" help:"Enter Prometheus mode"`
}
