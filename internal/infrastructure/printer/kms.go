package printer

import (
	"encoding/json"
	"errors"
	"fmt"

	"fry.org/cmo/cli/internal/application/kms"
	"fry.org/cmo/cli/internal/application/printer"
	"github.com/speijnik/go-errortree"
	"github.com/tidwall/pretty"
)

func (t *PrinterClient) ListKmsGroups(groups []kms.Group, mode printer.PrinterMode) error {
	var rcerror error

	switch mode {
	case printer.PrinterModeJSON:
		// Convert structs to JSON.
		if j, err := json.Marshal(groups); err != nil {
			return errortree.Add(rcerror, "ListKmsGroups", err)
		} else {
			fmt.Printf("%s\n", pretty.Pretty(j))
		}
	case printer.PrinterModeTable:
		return errortree.Add(rcerror, "ListKmsGroups", errors.New("table mode not implemented"))
	default:
		fmt.Printf("%v", groups)
	}

	return nil
}
