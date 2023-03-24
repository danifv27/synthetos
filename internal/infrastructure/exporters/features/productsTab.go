package features

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"fry.org/cmo/cli/internal/infrastructure/exporters"
	"github.com/chromedp/chromedp"
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
		pl.ctx = context.WithValue(pl.ctx, exporters.ContextKeyScenarioName, strcase.ToCamel(sc.Name))
		return context.WithValue(c, exporters.ContextKeyScenarioName, strcase.ToCamel(sc.Name)), nil
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
		if name, e := exporters.StringFromContext(c, exporters.ContextKeyScenarioName); e != nil {
			return c, errortree.Add(rcerror, "step.After", e)
		} else {
			stat := pl.stats[name][len(pl.stats[name])-1]
			stat.Duration = time.Since(stat.Start)
            if err != nil {
				stat.Result = exporters.CucumberFailure
			} else {
			    switch status {
				case 0:
					stat.Result = exporters.CucumberSuccess
                case 1:
                    stat.Result = exporters.CucumberFailure
				case 2:
					stat.Result = exporters.CucumberNotExecuted
				}
			}
			pl.stats[name][len(pl.stats[name])-1] = stat
		}
		return c, nil
	})
	pl.registerSteps(ctx)
}

func (pl *productsTab) registerSteps(ctx *godog.ScenarioContext) {

	ctx.Step(`^I am logged in to creation portal$`, pl.iAmLoggedInToCreationPortal)
	ctx.Step(`^the user switches to the "model" view with basic filter$`, pl.theUserSwitchesToTheModelViewWithBasicFilter)
	ctx.Step(`^the model info for the APP product should be displayed$`, pl.theModelInfoForTheAPPProductShouldBeDisplayed)
	ctx.Step(`^the user clicks on the first product in the "table" view on Product Page$`, pl.theUserClicksOnTheFirstProductInTheTableViewOnProductPage)
	ctx.Step(`^the Product Details Page should be loaded$`, pl.theProductDetailsPageShouldBeLoaded)
}

func (pl *productsTab) GetScenarioName() (string, error) {

	return exporters.StringFromContext(pl.ctx, exporters.ContextKeyScenarioName)
}

func (pl *productsTab) Do(c context.Context, cancel context.CancelFunc) (exporters.CucumberStatsSet, error) {
	var rcerror error
	var rc int
	var godogOpts godog.Options

	pl.ctx = c
	if content, err := exporters.GetFeature(exporters.FeaturesFS, pl.featureFolder); err != nil {
		return pl.stats, errortree.Add(rcerror, "loginPage.Do", err)
	} else {
	    godogOpts = godog.Options{
		    //TODO: Remove colored output after debugging
		    // Output: io.Discard,
		    Output: colors.Colored(os.Stdout),
		    //pretty, progress, cucumber, events and junit
		    Format:        "pretty",
		    StopOnFailure: true,
		    //This is the context passed as argument to scenario hooks
		    DefaultContext: pl.ctx,
			FeatureContents: content,
        }
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

func (pl *productsTab) theUserSwitchesToTheModelViewWithBasicFilter() error {
	var rcerror error
	time.Sleep(5 * time.Second)

	err := pl.loadModelProductsPage()
	if err != nil {
		return errortree.Add(rcerror, "theUserSwitchesToTheModelViewWithBasicFilter", err)
	}

	return nil
}

func (pl *productsTab) theModelInfoForTheAPPProductShouldBeDisplayed() error {
	var rcerror error

	time.Sleep(5 * time.Second)

	err := pl.loadModelDataInTable()
	if err != nil {
		return errortree.Add(rcerror, "theModelInfoForTheAPPProductShouldBeDisplayed", err)
	}

	return nil
}

func (pl *productsTab) theUserClicksOnTheFirstProductInTheTableViewOnProductPage() error {
	var rcerror error

	time.Sleep(5 * time.Second)
	err := pl.loadArticleDataInfoFromTable()
	if err != nil {
		return errortree.Add(rcerror, "theUserClicksOnTheFirstProductInTheTableViewOnProductPage", err)
	}

	return nil
}

func (pl *productsTab) theProductDetailsPageShouldBeLoaded() error {
	var rcerror error
	time.Sleep(5 * time.Second)

	err := pl.checkProductDetailsPage()
	if err != nil {
		return errortree.Add(rcerror, "theProductDetailsPageShouldBeLoaded", err)
	}

	return nil
}
