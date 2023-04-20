package printer

import (
	"bytes"
	"encoding/json"
	"fmt"

	"fry.org/cmo/cli/internal/application/kms"
	"fry.org/cmo/cli/internal/application/printer"
	"github.com/alexeyco/simpletable"
	"github.com/gonejack/linesprinter"
	"github.com/speijnik/go-errortree"
	"github.com/tidwall/pretty"
)

func (t *PrinterClient) printListGroupsTable(groups []kms.Group) error {
	var rcerror error

	t.table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "Group ID"},
			{Align: simpletable.AlignCenter, Text: "Name"},
			{Align: simpletable.AlignCenter, Text: "Description"},
			{Align: simpletable.AlignCenter, Text: "Created At"},
		},
	}

	for _, g := range groups {
		// bufGvk := new(bytes.Buffer)
		// p := linesprinter.NewLinesPrinter(bufGvk, 16, []byte("\r\n"))
		// if _, err := p.Write([]byte(key)); err != nil {
		// 	return errortree.Add(rcerror, "printListGroupsTable", err)
		// }
		// p.Close()
		groupID := new(bytes.Buffer)
		p := linesprinter.NewLinesPrinter(groupID, 36, []byte("\r\n"))
		if _, err := p.Write([]byte(g.GroupID)); err != nil {
			return errortree.Add(rcerror, "printListGroupsTable", err)
		}
		p.Close()
		name := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(name, 48, []byte("\r\n"))
		if _, err := p.Write([]byte(g.Name)); err != nil {
			return errortree.Add(rcerror, "printListGroupsTable", err)
		}
		p.Close()
		description := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(description, 64, []byte("\r\n"))
		if _, err := p.Write([]byte(*g.Description)); err != nil {
			return errortree.Add(rcerror, "printListGroupsTable", err)
		}
		p.Close()
		created := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(created, 64, []byte("\r\n"))
		if _, err := p.Write([]byte(g.CreatedAt)); err != nil {
			return errortree.Add(rcerror, "printListGroupsTable", err)
		}
		p.Close()
		r := []*simpletable.Cell{
			{Text: groupID.String()},
			{Text: name.String()},
			{Text: description.String()},
			{Text: created.String()},
		}
		t.table.Body.Cells = append(t.table.Body.Cells, r)
	}
	t.table.Println()

	return nil
}

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
		return t.printListGroupsTable(groups)
	default:
		fmt.Printf("%v", groups)
	}

	return nil
}
