package printer

import "fry.org/cmo/cli/internal/application/version"

type PrinterMode int

const (
	PrinterModeTable PrinterMode = iota //0
	PrinterModeJSON                     //1
	PrinterModeText                     //2
)

type Printer interface {
	PrintVersion(v version.Version, mode PrinterMode) error
}
