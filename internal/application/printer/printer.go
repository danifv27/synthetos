package printer

import (
	"fry.org/cmo/cli/internal/application/kms"
	"fry.org/cmo/cli/internal/application/version"
)

type PrinterMode int

const (
	PrinterModeNone  PrinterMode = iota //0
	PrinterModeJSON                     //1
	PrinterModeText                     //2
	PrinterModeTable                    //3
)

type Printer interface {
	PrintVersion(v version.Version, mode PrinterMode) error
	ListKmsGroups(groups []kms.Group, mode PrinterMode) error
}
