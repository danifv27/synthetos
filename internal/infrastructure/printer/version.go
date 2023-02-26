package printer

import (
	"encoding/json"
	"fmt"

	"fry.org/cmo/cli/internal/application/printer"
	"fry.org/cmo/cli/internal/application/version"
	"github.com/speijnik/go-errortree"
)

func (t *TablePrinterClient) PrintVersion(v version.Version, mode printer.PrinterMode) error {
	var err, rcerror error
	var info version.VersionInfo
	var out []byte

	if info, err = v.GetVersionInfo(); err != nil {
		return errortree.Add(rcerror, "PrintVersion", err)
	}

	switch mode {
	case printer.PrinterModeJSON:
		if out, err = json.MarshalIndent(info, "", "    "); err != nil {
			return errortree.Add(rcerror, "PrintVersion", err)
		}
		fmt.Println(string(out))
	default:
		fmt.Println(info)
	}

	return nil
}
