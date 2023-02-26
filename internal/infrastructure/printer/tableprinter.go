package printer

import (
	"github.com/alexeyco/simpletable"
	"github.com/speijnik/go-errortree"
)

// An TablePrinterOption applies optional changes to the Kong application.
type TablePrinterOption interface {
	Apply(t *TablePrinterClient) error
}

// TablePrinterOptionFunc is function that adheres to the Option interface.
type TablePrinterOptionFunc func(t *TablePrinterClient) error

func (o TablePrinterOptionFunc) Apply(t *TablePrinterClient) error {

	return o(t)
}

type TablePrinterClient struct {
	table *simpletable.Table
}

// NewTablePrinter Constructor
func NewTablePrinter(opts ...TablePrinterOption) (*TablePrinterClient, error) {
	var rcerror error

	t := TablePrinterClient{
		table: simpletable.New(),
	}
	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(&t); err != nil {
			return nil, errortree.Add(rcerror, "NewTablePrinter", err)
		}
	}

	return &t, nil
}
