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

func listKmsSecretsJSON(ch <-chan kms.Secret) error {
	var rcerror error
	var secrets []kms.Secret

	for r := range ch {
		secrets = append(secrets, r)
	} //for
	// Convert structs to JSON.
	if j, err := json.Marshal(secrets); err != nil {
		return errortree.Add(rcerror, "listKmsSecretsJSON", err)
	} else {
		fmt.Printf("%s\n", pretty.Pretty(j))
	}

	return nil
}

func (t *PrinterClient) listKmsSecretsTable(ch <-chan kms.Secret) error {
	// var rcerror error

	t.table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "Group ID"},
			{Align: simpletable.AlignCenter, Text: "Secret ID"},
			{Align: simpletable.AlignCenter, Text: "Name"},
			{Align: simpletable.AlignCenter, Text: "Description"},
			{Align: simpletable.AlignCenter, Text: "Created At"},
			{Align: simpletable.AlignCenter, Text: "Last Used At"},
		},
	}
	for r := range ch {
		groupID := new(bytes.Buffer)
		p := linesprinter.NewLinesPrinter(groupID, 48, []byte("\r\n"))
		if _, err := p.Write([]byte(*r.GroupID)); err != nil {
			continue
		}
		p.Close()
		secretID := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(secretID, 48, []byte("\r\n"))
		if _, err := p.Write([]byte(*r.SecretID)); err != nil {
			continue
		}
		p.Close()
		name := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(name, 96, []byte("\r\n"))
		if _, err := p.Write([]byte(*r.Name)); err != nil {
			continue
		}
		p.Close()
		description := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(description, 96, []byte("\r\n"))
		if _, err := p.Write([]byte(*r.Description)); err != nil {
			continue
		}
		p.Close()
		createAt := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(createAt, 96, []byte("\r\n"))
		if _, err := p.Write([]byte(*&r.CreatedAt)); err != nil {
			continue
		}
		p.Close()
		lastUsedAt := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(lastUsedAt, 96, []byte("\r\n"))
		if _, err := p.Write([]byte(*&r.LastusedAt)); err != nil {
			continue
		}
		p.Close()
		row := []*simpletable.Cell{
			{Text: groupID.String()},
			{Text: secretID.String()},
			{Text: name.String()},
			{Text: description.String()},
			{Text: createAt.String()},
			{Text: lastUsedAt.String()},
		}
		t.table.Body.Cells = append(t.table.Body.Cells, row)
	} //for

	t.table.Println()
	return nil
}

func listKmsSecretsText(ch <-chan kms.Secret) error {
	var secrets []kms.Secret

	for r := range ch {
		secrets = append(secrets, r)
	} //for
	fmt.Printf("%v\n", secrets)

	return nil
}

func (t *PrinterClient) ListKmsSecrets(ch <-chan kms.Secret, mode printer.PrinterMode) error {
	var rcerror error

	rcerror = errortree.Add(rcerror, "ListKmsSecrets", fmt.Errorf("printer mode %v not supported", mode))

	switch mode {
	case printer.PrinterModeJSON:
		rcerror = listKmsSecretsJSON(ch)
	case printer.PrinterModeTable:
		rcerror = t.listKmsSecretsTable(ch)
	case printer.PrinterModeText:
		rcerror = listKmsSecretsText(ch)
	}

	return rcerror
}
