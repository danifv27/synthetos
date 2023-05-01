package printer

import (
	"bytes"
	"encoding/json"
	"fmt"

	"fry.org/cmo/cli/internal/application/printer"
	"fry.org/cmo/cli/internal/application/provider"
	"github.com/alexeyco/simpletable"
	"github.com/gonejack/linesprinter"
	"github.com/speijnik/go-errortree"
	"github.com/tidwall/pretty"
)

func (t *PrinterClient) printResourcesSummaryTable(resources []provider.Summary) error {
	var rcerror error

	t.table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "API Version"},
			{Align: simpletable.AlignCenter, Text: "Kind"},
			{Align: simpletable.AlignCenter, Text: "Name"},
		},
	}

	for _, r := range resources {
		apiVersion := new(bytes.Buffer)
		p := linesprinter.NewLinesPrinter(apiVersion, 48, []byte("\r\n"))
		if _, err := p.Write([]byte(r.APIVersion)); err != nil {
			return errortree.Add(rcerror, "printResourcesSummaryTable", err)
		}
		p.Close()
		kind := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(kind, 48, []byte("\r\n"))
		if _, err := p.Write([]byte(r.Kind)); err != nil {
			return errortree.Add(rcerror, "printResourcesSummaryTable", err)
		}
		p.Close()
		name := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(name, 96, []byte("\r\n"))
		if _, err := p.Write([]byte(r.Name)); err != nil {
			return errortree.Add(rcerror, "printResourcesSummaryTable", err)
		}
		p.Close()

		r := []*simpletable.Cell{
			{Text: apiVersion.String()},
			{Text: kind.String()},
			{Text: name.String()},
		}
		t.table.Body.Cells = append(t.table.Body.Cells, r)
	}
	t.table.Println()

	return nil
}

func (t *PrinterClient) PrintResourceSummary(resources []provider.Summary, mode printer.PrinterMode) error {
	var rcerror error

	switch mode {
	case printer.PrinterModeJSON:
		// Convert structs to JSON.
		if j, err := json.Marshal(resources); err != nil {
			return errortree.Add(rcerror, "PrintResourceSummary", err)
		} else {
			fmt.Printf("%s\n", pretty.Pretty(j))
		}
	case printer.PrinterModeTable:
		return t.printResourcesSummaryTable(resources)
	default:
		fmt.Printf("%v", resources)
	}

	return nil
}
