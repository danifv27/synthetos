package printer

import (
	"bytes"
	"fmt"

	"fry.org/cmo/cli/internal/application/printer"
	"fry.org/cmo/cli/internal/application/provider"
	"github.com/alexeyco/simpletable"
	"github.com/gonejack/linesprinter"
	"github.com/speijnik/go-errortree"
)

func (t *PrinterClient) printResourcesSummaryTable(ch <-chan provider.Summary) error {
	// var rcerror error

	t.table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "API Version"},
			{Align: simpletable.AlignCenter, Text: "Kind"},
			{Align: simpletable.AlignCenter, Text: "Name"},
		},
	}
	for r := range ch {
		apiVersion := new(bytes.Buffer)
		p := linesprinter.NewLinesPrinter(apiVersion, 48, []byte("\r\n"))
		if _, err := p.Write([]byte(r.APIVersion)); err != nil {
			// return errortree.Add(rcerror, "printResourcesSummaryTable", err)
			continue
		}
		p.Close()
		kind := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(kind, 48, []byte("\r\n"))
		if _, err := p.Write([]byte(r.Kind)); err != nil {
			// return errortree.Add(rcerror, "printResourcesSummaryTable", err)
			continue
		}
		p.Close()
		name := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(name, 96, []byte("\r\n"))
		if _, err := p.Write([]byte(r.Name)); err != nil {
			// return errortree.Add(rcerror, "printResourcesSummaryTable", err)
			continue
		}
		p.Close()
		row := []*simpletable.Cell{
			{Text: apiVersion.String()},
			{Text: kind.String()},
			{Text: name.String()},
		}
		t.table.Body.Cells = append(t.table.Body.Cells, row)
	} //for

	t.table.Println()
	return nil
}

func (t *PrinterClient) PrintResourceSummary(ch <-chan provider.Summary, mode printer.PrinterMode) error {
	var rcerror error

	rcerror = errortree.Add(rcerror, "PrintResourceSummary", fmt.Errorf("printer mode %v not supported", mode))

	switch mode {
	case printer.PrinterModeJSON:
		// Convert structs to JSON.
		// if j, err := json.Marshal(resources); err != nil {
		// 	return errortree.Add(rcerror, "PrintResourceSummary", err)
		// } else {
		// 	fmt.Printf("%s\n", pretty.Pretty(j))
		// }
		// return nil
	case printer.PrinterModeTable:
		rcerror = t.printResourcesSummaryTable(ch)
		// default:
		// fmt.Printf("%v", resources)
	}

	return rcerror
}
