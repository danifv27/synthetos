package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"fry.org/cmo/cli/internal/application/printer"
	"fry.org/cmo/cli/internal/application/provider"
	"github.com/alexeyco/simpletable"
	"github.com/gonejack/linesprinter"
	"github.com/speijnik/go-errortree"
	"github.com/tidwall/pretty"
)

func listKbomImagesJSON(ch <-chan provider.Image) error {
	var rcerror error
	var images []provider.Image

	for img := range ch {
		images = append(images, img)
	} //for
	//Sort by groupID
	sort.Slice(images, func(i int, j int) bool {
		rc := strings.Compare(images[i].Name, images[j].Name)

		return rc < 0
	})
	// Convert structs to JSON.
	if j, err := json.Marshal(images); err != nil {
		return errortree.Add(rcerror, "listKbomImagesJSON", err)
	} else {
		fmt.Printf("%s\n", pretty.Pretty(j))
	}

	return nil
}

func (t *PrinterClient) listKbomImagesTable(ch <-chan provider.Image) error {
	var images []provider.Image

	t.table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "Image"},
			{Align: simpletable.AlignCenter, Text: "Digest"},
		},
	}
	for r := range ch {
		images = append(images, r)
	} //for
	//Sort by name
	sort.Slice(images, func(i int, j int) bool {
		rc := strings.Compare(images[i].Name, images[j].Name)

		return rc < 0
	})

	for _, r := range images {
		name := new(bytes.Buffer)
		p := linesprinter.NewLinesPrinter(name, 128, []byte("\r\n"))
		if _, err := p.Write([]byte(r.Name)); err != nil {
			continue
		}
		p.Close()
		digest := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(digest, 96, []byte("\r\n"))
		if _, err := p.Write([]byte(r.Digest)); err != nil {
			continue
		}
		p.Close()

		row := []*simpletable.Cell{
			{Text: name.String()},
			{Text: digest.String()},
		}
		t.table.Body.Cells = append(t.table.Body.Cells, row)
	} //for

	t.table.Println()
	return nil
}

func listKbomImagesText(ch <-chan provider.Image) error {

	for r := range ch {
		fmt.Printf("%v", r)

	} //for

	return nil
}

func (t *PrinterClient) ListKbomImages(ch <-chan provider.Image, mode printer.PrinterMode) error {
	var rcerror error

	rcerror = errortree.Add(rcerror, "ListKbomImages", fmt.Errorf("printer mode %v not supported", mode))

	switch mode {
	case printer.PrinterModeJSON:
		rcerror = listKbomImagesJSON(ch)
	case printer.PrinterModeTable:
		rcerror = t.listKbomImagesTable(ch)
	case printer.PrinterModeText:
		rcerror = listKbomImagesText(ch)
	}

	return rcerror
}
