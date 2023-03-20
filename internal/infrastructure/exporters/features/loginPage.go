package features

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/infrastructure/exporters"
	"github.com/chromedp/chromedp"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/iancoleman/strcase"
	"github.com/speijnik/go-errortree"
)

type loginPage struct {
	logger.Logger
	featureFolder string
	ctx           context.Context
	stats         exporters.CucumberStatsSet
	auth          struct {
		id       string
		password string
	}
}

func NewLoginPageFeature(p string, opts ...exporters.ExporterOption) (exporters.CucumberPlugin, error) {
	var rcerror error

	l := loginPage{
		featureFolder: path.Join(p, "loginPage.feature"),
	}
	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(&l); err != nil {
			return nil, errortree.Add(rcerror, "NewLoginPageFeature", err)
		}
	}

	return &l, nil
}

func WithLoginPageLogger(l logger.Logger) exporters.ExporterOption {

	return exporters.ExportOptionFn(func(i interface{}) error {
		var rcerror error
		var pl *loginPage
		var ok bool

		if pl, ok = i.(*loginPage); ok {
			pl.Logger = l
			return nil
		}

		return errortree.Add(rcerror, "WithLoginPageLogger", errors.New("type mismatch, loginPage expected"))
	})
}

func WithLoginPageAuth(id string, p string) exporters.ExporterOption {

	return exporters.ExportOptionFn(func(i interface{}) error {
		var rcerror error
		var l *loginPage
		var ok bool

		if l, ok = i.(*loginPage); ok {
			l.auth.id = id
			l.auth.password = p
			return nil
		}

		return errortree.Add(rcerror, "WithLoginPageAuth", errors.New("type mismatch, loginPage expected"))
	})
}

func (pl *loginPage) suiteInit(ctx *godog.TestSuiteContext) {

	ctx.BeforeSuite(func() {
		// This code will be executed once, before any scenarios are run
		pl.stats = make(map[string][]exporters.CucumberStats)
	})
}

func (pl *loginPage) scenarioInit(ctx *godog.ScenarioContext) {

	ctx.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		// This code will be executed once, before any scenarios are run

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

func (pl *loginPage) registerSteps(ctx *godog.ScenarioContext) {

	ctx.Step(`^I am on the login page$`, pl.iAmOnTheLoginPage)
	ctx.Step(`^I enter my username and password$`, pl.iEnterMyUsernameAndPassword)
	ctx.Step(`^I click the login button$`, pl.iClickTheLoginButton)
	ctx.Step(`^I should be redirected to the dashboard page$`, pl.iShouldBeRedirectedToTheDashboardPage)
}

func (pl *loginPage) Do(c context.Context, cancel context.CancelFunc) (exporters.CucumberStatsSet, error) {
	var rcerror error
	var rc int

	//Initialize chromedp context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.111 Safari/537.36"),
	)
	actx, _ := chromedp.NewExecAllocator(c, opts...)
	pl.ctx, _ = chromedp.NewContext(actx)
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
		Name:                 "loginPage",
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
		return pl.stats, errortree.Add(rcerror, "loginPage.Do", fmt.Errorf("error  %d: failed test suite", rc))
	case 2:
		return pl.stats, errortree.Add(rcerror, "loginPage.Do", fmt.Errorf("error %d:command line usage error running test suite", rc))
	default:
		return pl.stats, errortree.Add(rcerror, "loginPage.Do", fmt.Errorf("error %d running test suite", rc))
	}

	//return l.stats, nil
}

func (pl *loginPage) iAmOnTheLoginPage() error {
	var rcerror error

	pl.Logger.WithFields(logger.Fields{
		"name": "I am on the login page",
	}).Debug("Executing step")
	impl := loginPageImpl{}
	if err := impl.doAzureLogin(pl.ctx); err != nil {
		takeSnapshot(pl.ctx, "iAmOnTheLoginPage")
		return errortree.Add(rcerror, "iAmOnTheLoginPage", err)
	}
	pl.Logger.WithFields(logger.Fields{
		"name": "I am on the login page",
	}).Debug("Step done")

	return nil
}

func (pl *loginPage) iEnterMyUsernameAndPassword() error {
	var rcerror error

	impl := loginPageImpl{}
	if err := impl.loadUserAndPasswordWindow(pl.ctx, pl.auth.id, pl.auth.password); err != nil {
		takeSnapshot(pl.ctx, "iEnterMyUsernameAndPassword")
		return errortree.Add(rcerror, "iEnterMyUsernameAndPassword", err)
	}

	return nil
}

func (pl *loginPage) iClickTheLoginButton() error {
	var rcerror error

	impl := loginPageImpl{}
	if err := impl.loadConsentAzurePage(pl.ctx); err != nil {
		takeSnapshot(pl.ctx, "iClickTheLoginButton")
		return errortree.Add(rcerror, "iClickTheLoginButton", err)
	}
	return nil
}

func (pl *loginPage) iShouldBeRedirectedToTheDashboardPage() error {
	var rcerror error

	impl := loginPageImpl{}
	if err := impl.isMainFELoad(pl.ctx); err != nil {
		takeSnapshot(pl.ctx, "iShouldBeRedirectedToTheDashboardPage")
		return errortree.Add(rcerror, "iShouldBeRedirectedToTheDashboardPage", err)
	}
	return nil
}
