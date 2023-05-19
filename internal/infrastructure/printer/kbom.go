package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
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

func listKbomResourcesJSON(ch <-chan provider.ResourceList) error {
	var rcerror error
	var resourceLists []provider.ResourceList

	for res := range ch {
		resourceLists = append(resourceLists, res)
	} //for
	//Sort by groupID
	sort.Slice(resourceLists, func(i int, j int) bool {
		rc := strings.Compare(resourceLists[i].Kind, resourceLists[j].Kind)

		return rc < 0
	})
	// Convert structs to JSON.
	if j, err := json.Marshal(resourceLists); err != nil {
		return errortree.Add(rcerror, "listKbomResourcesJSON", err)
	} else {
		fmt.Printf("%s\n", pretty.Pretty(j))
	}

	return nil
}

func (t *PrinterClient) listKbomResourcesTable(ch <-chan provider.ResourceList) error {
	var resourceLists []provider.ResourceList

	t.table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "Kind"},
			{Align: simpletable.AlignCenter, Text: "APIVersion"},
			{Align: simpletable.AlignCenter, Text: "Namespaced"},
			{Align: simpletable.AlignCenter, Text: "ResourcesCount"},
		},
	}
	for r := range ch {
		resourceLists = append(resourceLists, r)
	} //for
	//Sort by name
	sort.Slice(resourceLists, func(i int, j int) bool {
		rc := strings.Compare(resourceLists[i].Kind, resourceLists[j].Kind)

		return rc < 0
	})

	for _, r := range resourceLists {
		kind := new(bytes.Buffer)
		p := linesprinter.NewLinesPrinter(kind, 128, []byte("\r\n"))
		if _, err := p.Write([]byte(r.Kind)); err != nil {
			continue
		}
		p.Close()
		apiversion := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(apiversion, 96, []byte("\r\n"))
		if _, err := p.Write([]byte(r.APIVersion)); err != nil {
			continue
		}
		p.Close()
		namespaced := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(namespaced, 16, []byte("\r\n"))
		if _, err := p.Write([]byte(strconv.FormatBool(r.Namespaced))); err != nil {
			continue
		}
		p.Close()
		count := new(bytes.Buffer)
		p = linesprinter.NewLinesPrinter(count, 96, []byte("\r\n"))
		if _, err := p.Write([]byte(strconv.Itoa(r.ResourcesCount))); err != nil {
			continue
		}
		p.Close()
		row := []*simpletable.Cell{
			{Text: kind.String()},
			{Text: apiversion.String()},
			{Text: namespaced.String()},
			{Text: count.String()},
		}
		t.table.Body.Cells = append(t.table.Body.Cells, row)
	} //for

	t.table.Println()
	return nil
}

func listKbomResourcesText(ch <-chan provider.ResourceList) error {

	for r := range ch {
		fmt.Printf("%v", r)
	} //for

	return nil
}

func (t *PrinterClient) ListKbomResources(receiveCh <-chan provider.ResourceList, mode printer.PrinterMode) error {
	var rcerror error

	rcerror = errortree.Add(rcerror, "ListKbomImages", fmt.Errorf("printer mode %v not supported", mode))

	switch mode {
	case printer.PrinterModeJSON:
		rcerror = listKbomResourcesJSON(receiveCh)
	case printer.PrinterModeTable:
		rcerror = t.listKbomResourcesTable(receiveCh)
	case printer.PrinterModeText:
		rcerror = listKbomResourcesText(receiveCh)
	}

	return rcerror
}
