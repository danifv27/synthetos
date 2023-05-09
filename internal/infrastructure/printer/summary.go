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
			continue
		}
		p.Close()
		kind := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(kind, 48, []byte("\r\n"))
		if _, err := p.Write([]byte(r.Kind)); err != nil {
			continue
		}
		p.Close()
		name := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(name, 96, []byte("\r\n"))
		if _, err := p.Write([]byte(r.Name)); err != nil {
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

func printResourcesSummaryJSON(ch <-chan provider.Summary) error {
	var rcerror error
	var summaries []provider.Summary

	for r := range ch {
		summaries = append(summaries, r)
	} //for
	// Convert structs to JSON.
	if j, err := json.Marshal(summaries); err != nil {
		return errortree.Add(rcerror, "printResourcesSummaryJSON", err)
	} else {
		fmt.Printf("%s\n", pretty.Pretty(j))
	}

	return nil
}

func printResourcesSummaryText(ch <-chan provider.Summary) error {
	var summaries []provider.Summary

	for r := range ch {
		summaries = append(summaries, r)
	} //for
	fmt.Printf("%v\n", summaries)

	return nil
}

func (t *PrinterClient) PrintResourceSummary(ch <-chan provider.Summary, mode printer.PrinterMode) error {
	var rcerror error

	rcerror = errortree.Add(rcerror, "PrintResourceSummary", fmt.Errorf("printer mode %v not supported", mode))

	switch mode {
	case printer.PrinterModeJSON:
		rcerror = printResourcesSummaryJSON(ch)
	case printer.PrinterModeTable:
		rcerror = t.printResourcesSummaryTable(ch)
	case printer.PrinterModeText:
		rcerror = printResourcesSummaryText(ch)
	}

	return rcerror
}
