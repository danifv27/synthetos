package printer

import (
	"io"

	"github.com/alexeyco/simpletable"
	"github.com/speijnik/go-errortree"
)

// An PrinterOption applies optional changes to the Kong application.
type PrinterOption interface {
	Apply(t *PrinterClient) error
}

// PrinterOptionFunc is function that adheres to the Option interface.
type PrinterOptionFunc func(t *PrinterClient) error

func (o PrinterOptionFunc) Apply(t *PrinterClient) error {

	return o(t)
}

type PrinterClient struct {
	table *simpletable.Table
	wr    io.Writer
}

// NewPrinter Constructor
func NewPrinter(opts ...PrinterOption) (*PrinterClient, error) {
	var rcerror error

	t := PrinterClient{
		table: simpletable.New(),
	}
	t.table.SetStyle(simpletable.StyleMarkdown)
	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(&t); err != nil {
			return nil, errortree.Add(rcerror, "NewTablePrinter", err)
		}
	}

	return &t, nil
}

func WithWriter(w io.Writer) PrinterOption {

	return PrinterOptionFunc(func(c *PrinterClient) error {
		c.wr = w

		return nil
	})
}
