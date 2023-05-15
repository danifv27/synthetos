package printer

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"fry.org/cmo/cli/internal/application/kms"
	"fry.org/cmo/cli/internal/application/printer"
	"github.com/alexeyco/simpletable"
	"github.com/aquilax/truncate"
	"github.com/gonejack/linesprinter"
	"github.com/speijnik/go-errortree"
	"github.com/tidwall/pretty"
)

func (t *PrinterClient) listKmsGroupsTable(groups []kms.Group) error {
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

func listKmsGroupsJSON(groups []kms.Group) error {
	var rcerror error

	//Sort by groupID
	sort.Slice(groups, func(i int, j int) bool {
		rc := strings.Compare(groups[i].GroupID, groups[j].GroupID)

		return rc < 0
	})
	// Convert structs to JSON.
	if j, err := json.Marshal(groups); err != nil {
		return errortree.Add(rcerror, "listKmsGroupsJSON", err)
	} else {
		fmt.Printf("%s\n", pretty.Pretty(j))
	}

	return nil
}

func listKmsGroupsText(groups []kms.Group) error {

	//Sort by groupID
	sort.Slice(groups, func(i int, j int) bool {
		rc := strings.Compare(groups[i].GroupID, groups[j].GroupID)

		return rc < 0
	})
	fmt.Printf("%v\n", groups)

	return nil
}

func (t *PrinterClient) ListKmsGroups(groups []kms.Group, mode printer.PrinterMode) error {
	var rcerror error

	rcerror = errortree.Add(rcerror, "ListKmsGroups", fmt.Errorf("printer mode %v not supported", mode))

	switch mode {
	case printer.PrinterModeJSON:
		rcerror = listKmsGroupsJSON(groups)
	case printer.PrinterModeTable:
		rcerror = t.listKmsGroupsTable(groups)
	case printer.PrinterModeText:
		rcerror = listKmsGroupsText(groups)
	}

	return rcerror
}

func decode(input string) string {
	var err error
	var decoded []byte
	var b strings.Builder

	if decoded, err = base64.StdEncoding.DecodeString(input); err != nil {
		fmt.Printf("[DBG]clear text\n")
		//If there is a key value string, try to decode only the value
		scanner := bufio.NewScanner(strings.NewReader(input))
		for scanner.Scan() {
			line := scanner.Text()
			split := strings.Split(line, ":")
			if len(split) == 2 {
				if value, e := base64.StdEncoding.DecodeString(strings.TrimSpace(split[1])); e != nil {
					fmt.Fprintf(&b, "%s\n", string(line))
				} else {
					fmt.Fprintf(&b, "%s: %s\n", split[0], string(value))
				}
			}
		}

	} else {
		fmt.Fprintf(&b, "%s", string(decoded))
	}

	return b.String()
}

func listKmsSecretsJSON(ch <-chan kms.Secret) error {
	var rcerror error
	var secrets []kms.Secret

	for r := range ch {
		r.Value = decode(string(*r.Blob))
		secrets = append(secrets, r)
	} //for
	//Sort by groupID
	sort.Slice(secrets, func(i int, j int) bool {
		rc := strings.Compare(*secrets[i].GroupID, *secrets[j].GroupID)

		return rc < 0
	})
	// Convert structs to JSON.
	if j, err := json.Marshal(secrets); err != nil {
		return errortree.Add(rcerror, "listKmsSecretsJSON", err)
	} else {
		fmt.Printf("%s\n", pretty.Pretty(j))
	}

	return nil
}

func (t *PrinterClient) listKmsSecretsTable(ch <-chan kms.Secret) error {
	var secrets []kms.Secret

	t.table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "Group ID"},
			{Align: simpletable.AlignCenter, Text: "Secret ID"},
			{Align: simpletable.AlignCenter, Text: "Name"},
			{Align: simpletable.AlignCenter, Text: "Value"},
			{Align: simpletable.AlignCenter, Text: "Description"},
			{Align: simpletable.AlignCenter, Text: "Created At"},
			{Align: simpletable.AlignCenter, Text: "Last Used At"},
		},
	}
	for r := range ch {
		r.Value = decode(string(*r.Blob))
		secrets = append(secrets, r)
	} //for
	//Sort by groupID
	sort.Slice(secrets, func(i int, j int) bool {
		rc := strings.Compare(*secrets[i].GroupID, *secrets[j].GroupID)

		return rc < 0
	})

	for _, r := range secrets {
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
		value := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(value, 96, []byte("\r\n"))
		truncated := truncate.Truncate(r.Value, 512, "...", truncate.PositionEnd)
		if _, err := p.Write([]byte(truncated)); err != nil {
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
		if _, err := p.Write([]byte(r.CreatedAt)); err != nil {
			continue
		}
		p.Close()
		lastUsedAt := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(lastUsedAt, 96, []byte("\r\n"))
		if _, err := p.Write([]byte(r.LastusedAt)); err != nil {
			continue
		}
		p.Close()
		row := []*simpletable.Cell{
			{Text: groupID.String()},
			{Text: secretID.String()},
			{Text: name.String()},
			{Text: value.String()},
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
		if decoded, err := base64.StdEncoding.DecodeString(string(*r.Blob)); err == nil {
			r.Value = string(decoded)
		} else {
			r.Value = string(*r.Blob)
		}
		secrets = append(secrets, r)
	} //for
	//Sort by groupID
	sort.Slice(secrets, func(i int, j int) bool {
		rc := strings.Compare(*secrets[i].GroupID, *secrets[j].GroupID)

		return rc < 0
	})
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
