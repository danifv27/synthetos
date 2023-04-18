package secretum

import "fry.org/cmo/cli/internal/cli/common"

type CLI struct {
	Logging common.Log `embed:"" prefix:"logging."`
	Version VersionCmd `cmd:"" help:"Show version information"`
	Kms     KmsCmd     `cmd:"" help:"Manage KMS"`
}
