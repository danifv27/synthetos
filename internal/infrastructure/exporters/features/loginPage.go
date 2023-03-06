package features

import (
	"context"
	"fmt"

	"fry.org/cmo/cli/internal/infrastructure/exporters"
	"github.com/speijnik/go-errortree"
)

type loginPage struct{}

func NewLoginPageFeature(opts ...exporters.ExporterOption) (exporters.CucumberPlugin, error) {
	var rcerror error

	h := loginPage{}
	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(&h); err != nil {
			return nil, errortree.Add(rcerror, "NewLoginPageFeature", err)
		}
	}

	return &h, nil
}

func (l *loginPage) Do(ctx context.Context) error {

	fmt.Print("[DBG]this is loginPage.go")

	return nil
}
