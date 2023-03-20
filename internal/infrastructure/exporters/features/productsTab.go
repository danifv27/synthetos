package features

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"fry.org/cmo/cli/internal/infrastructure/exporters"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/iancoleman/strcase"
	"github.com/speijnik/go-errortree"
)

type productsTab struct {
	featureFolder string
	ctx           context.Context
	stats         exporters.CucumberStatsSet
	auth          struct {
		id       string
		password string
	}
}

func NewProductsTabFeature(p string, opts ...exporters.ExporterOption) (exporters.CucumberPlugin, error) {
	var rcerror error

	l := productsTab{
		featureFolder: path.Join(p, "productsTab.feature"),
	}
	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(&l); err != nil {
			return nil, errortree.Add(rcerror, "NewProductsTabFeature", err)
		}
	}

	return &l, nil
}

func WithProductsTabAuth(id string, p string) exporters.ExporterOption {

	return exporters.ExportOptionFn(func(i interface{}) error {
		var rcerror error
		var pl *productsTab
		var ok bool

		if pl, ok = i.(*productsTab); ok {
			pl.auth.id = id
			pl.auth.password = p
			return nil
		}

		return errortree.Add(rcerror, "WithProductTabsAuth", errors.New("type mismatch, productsTab expected"))
	})
}

func (pl *productsTab) suiteInit(ctx *godog.TestSuiteContext) {

	ctx.BeforeSuite(func() {
		// This code will be executed once, before any scenarios are run
		pl.stats = make(map[string][]exporters.CucumberStats)
	})
}

func (pl *productsTab) scenarioInit(ctx *godog.ScenarioContext) {

	ctx.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		// This code will be executed once, before any scenarios are run

		return context.WithValue(c, exporters.ContextKeyScenarioName, strcase.ToCamel(sc.Name)), nil
	})

	ctx.After(func(c context.Context, sc *godog.Scenario, err error) (context.Context, error) {

		// This code will be executed once, after all scenarios have been run
		// v, ok := c.Value(contextKeyScenarioName).(string)

		return c, nil
	})
	stepCtx := ctx.StepContext()
	stepCtx.Before(func(c context.Context, st *godog.Step) (context.Context, error) {
		var rcerror error

		stat := exporters.CucumberStats{
			Id:     strcase.ToCamel(st.Text),
			Start:  time.Now(),
			Result: exporters.CucumberNotExecuted,
		}

		err := pl.ctx.Err()
		if err != nil {
			// fmt.Printf("[DBG] '%v', context error: '%v')\n", err, st.Text)
			return c, errortree.Add(rcerror, "step.Before", err)
		}

		if name, err := exporters.StringFromContext(c, exporters.ContextKeyScenarioName); err != nil {
			return c, errortree.Add(rcerror, "step.Before", err)
		} else {
			pl.stats[name] = append(pl.stats[name], stat)
		}

		return c, nil
	})
	stepCtx.After(func(c context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
		var rcerror error
		//FIXME check if err is not nil indicating an error in the Before step
		if name, e := exporters.StringFromContext(c, exporters.ContextKeyScenarioName); err != nil {
			return c, errortree.Add(rcerror, "step.After", e)
		} else {
			stat := pl.stats[name][len(pl.stats[name])-1]
			stat.Duration = time.Since(stat.Start)
			if status == godog.StepPassed {
				stat.Result = exporters.CucumberSuccess
			} else {
				stat.Result = exporters.CucumberFailure
			}
			pl.stats[name][len(pl.stats[name])-1] = stat
		}
		return c, nil
	})
	pl.registerSteps(ctx)
}

func (pl *productsTab) registerSteps(ctx *godog.ScenarioContext) {

	ctx.Step(`^I am logged in to creation portal$`, pl.iAmLoggedInToCreationPortal)
	ctx.Step(`^I can see factory info for APP product$`, pl.iCanSeeFactoryInfoForAPPProduct)
	ctx.Step(`^I switch to "([^"]*)" view on Product Page$`, pl.iSwitchToViewOnProductPage)
	ctx.Step(`^user clicks on first product in "([^"]*)" view on Product Page$`, pl.userClicksOnFirstProductInViewOnProductPage)
	ctx.Step(`^user sees Factory Details Page is loaded$`, pl.userSeesFactoryDetailsPageIsLoaded)
}

func (pl *productsTab) Do(c context.Context, cancel context.CancelFunc) (exporters.CucumberStatsSet, error) {
	var rcerror error
	var rc int

	pl.ctx = c
	godogOpts := godog.Options{
		//TODO: Remove colored output after debugging
		// Output: io.Discard,
		Output: colors.Colored(os.Stdout),
		Paths:  []string{pl.featureFolder},
		//pretty, progress, cucumber, events and junit
		Format:        "pretty",
		StopOnFailure: true,
		//This is the context passed as argument to scenario hooks
		DefaultContext: pl.ctx,
	}
	suite := godog.TestSuite{
		Name:                 "productsTab",
		TestSuiteInitializer: pl.suiteInit,
		ScenarioInitializer:  pl.scenarioInit,
		Options:              &godogOpts,
	}

	done := make(chan bool)
	go func() {
		rc = suite.Run()
		done <- true
	}()
	// fmt.Printf("[DBG]Waiting for context done\n")
	<-done
	// We have to return l.stats always to return the partial errors in case of error
	switch rc {
	case 0:
		return pl.stats, nil
	case 1:
		return pl.stats, errortree.Add(rcerror, "productsTab.Do", fmt.Errorf("error  %d: failed test suite", rc))
	case 2:
		return pl.stats, errortree.Add(rcerror, "productsTab.Do", fmt.Errorf("error %d:command line usage error running test suite", rc))
	default:
		return pl.stats, errortree.Add(rcerror, "productsTab.Do", fmt.Errorf("error %d running test suite", rc))
	}

	//return l.stats, nil
}

func (pl *productsTab) iAmLoggedInToCreationPortal() error {
	var rcerror error

	impl := loginPageImpl{}

	err := impl.doFeature(pl.ctx, pl.auth.id, pl.auth.password)
	if err != nil {
		return errortree.Add(rcerror, "iAmLoggedInToCreationPortal", err)
	}

	return nil
}

func (pl *productsTab) iCanSeeFactoryInfoForAPPProduct() error {
	return godog.ErrPending
}

func (pl *productsTab) iSwitchToViewOnProductPage(arg1 string) error {
	return godog.ErrPending
}

func (pl *productsTab) userClicksOnFirstProductInViewOnProductPage(arg1 string) error {
	return godog.ErrPending
}

func (pl *productsTab) userSeesFactoryDetailsPageIsLoaded() error {
	return godog.ErrPending
}
