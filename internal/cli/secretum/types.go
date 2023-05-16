package secretum

import (
	"fry.org/cmo/cli/internal/cli/common"
	"fry.org/cmo/cli/internal/cli/versio"
)

type CLI struct {
	Logging common.Log        `embed:"" prefix:"logging."`
	Config  common.Config     `embed:""`
	Version versio.VersionCmd `cmd:"" help:"Show version information"`
	Kms     KmsCmd            `cmd:"" help:"Manage KMS"`
}
