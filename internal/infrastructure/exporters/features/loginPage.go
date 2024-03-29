package features

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path"
	"time"

	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/infrastructure/exporters"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/iancoleman/strcase"
	"github.com/sethvargo/go-retry"
	"github.com/speijnik/go-errortree"
)

type loginPage struct {
	logger.Logger
	featureFolder string
	ctx           context.Context
	statsSet      exporters.CucumberStatsSet
	auth          struct {
		id       string
		password string
	}
	snapshotsFolder string
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

func WithLoginPageSnapshotFolder(path string) exporters.ExporterOption {

	return exporters.ExportOptionFn(func(i interface{}) error {
		var rcerror error
		var pl *loginPage
		var ok bool

		if pl, ok = i.(*loginPage); ok {
			pl.snapshotsFolder = path
			return nil
		}

		return errortree.Add(rcerror, "WithLoginPageSnapshotFolder", errors.New("type mismatch, loginPage expected"))
	})
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
		pl.statsSet = make(map[string]exporters.CucumberStatsItem)
	})
}

func (pl *loginPage) scenarioInit(ctx *godog.ScenarioContext) {

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
			item := pl.statsSet[name]
			item.Stats = append(item.Stats, stat)
			pl.statsSet[name] = item
		}

		return c, nil
	})
	stepCtx.After(func(c context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
		var rcerror error
		if name, e := exporters.StringFromContext(c, exporters.ContextKeyScenarioName); e != nil {
			return c, errortree.Add(rcerror, "step.After", e)
		} else {
			stat := pl.statsSet[name].Stats[len(pl.statsSet[name].Stats)-1]
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
			pl.statsSet[name].Stats[len(pl.statsSet[name].Stats)-1] = stat
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

func (pl *loginPage) GetScenarioName() (string, error) {

	return exporters.StringFromContext(pl.ctx, exporters.ContextKeyScenarioName)
}

func (pl *loginPage) Do(c context.Context) (exporters.CucumberStatsSet, error) {
	var rcerror error
	var rc int
	var godogOpts godog.Options

	pl.ctx = c
	buf := new(bytes.Buffer)
	if content, err := exporters.GetFeature(exporters.FeaturesFS, pl.featureFolder); err != nil {
		return pl.statsSet, errortree.Add(rcerror, "loginPage.Do", err)
	} else {

		godogOpts = godog.Options{

			// Output: io.Discard,
			// Output: colors.Colored(os.Stdout),
			Output: colors.Colored(buf),
			//pretty, progress, cucumber, events and junit
			Format:        "pretty",
			StopOnFailure: true,
			//This is the context passed as argument to scenario hooks
			DefaultContext:  pl.ctx,
			FeatureContents: content,
		}
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
	fmt.Println(buf.String())
	if name, err := pl.GetScenarioName(); err != nil {
		return pl.statsSet, errortree.Add(rcerror, "loginPage.Do", err)
	} else {
		item := pl.statsSet[name]
		item.Output = buf.String()
		pl.statsSet[name] = item
	}
	// We have to return l.stats always to return the partial errors in case of error
	switch rc {
	case 0:
		return pl.statsSet, nil
	case 1:
		return pl.statsSet, errortree.Add(rcerror, "loginPage.Do", fmt.Errorf("error  %d: failed test suite", rc))
	case 2:
		return pl.statsSet, errortree.Add(rcerror, "loginPage.Do", fmt.Errorf("error %d:command line usage error running test suite", rc))
	default:
		return pl.statsSet, errortree.Add(rcerror, "loginPage.Do", fmt.Errorf("error %d running test suite", rc))
	}

	//return l.stats, nil
}

func (pl *loginPage) iAmOnTheLoginPage() error {
	var rcerror error

	// pl.Logger.WithFields(logger.Fields{
	// 	"name": "I am on the login page",
	// }).Debug("Executing step")
	impl := loginPageImpl{
		snapshotsFolder: pl.snapshotsFolder,
	}
	if err := impl.doAzureLogin(pl.ctx); err != nil {
		takeSnapshot(pl.ctx, pl.snapshotsFolder, "iAmOnTheLoginPage")
		return errortree.Add(rcerror, "iAmOnTheLoginPage", err)
	}
	// takeSnapshot(pl.ctx, pl.snapshotsFolder, "iAmOnTheLoginPage_success")
	// pl.Logger.WithFields(logger.Fields{
	// 	"name": "I am on the login page",
	// }).Debug("Step done")

	return nil
}

func (pl *loginPage) iEnterMyUsernameAndPassword() error {
	var rcerror error

	impl := loginPageImpl{
		snapshotsFolder: pl.snapshotsFolder,
	}
	c := context.Background()
	b := retry.NewConstant(500 * time.Millisecond)
	b = retry.WithMaxDuration(7*time.Second, b)
	if err := retry.Do(c, b, func(ct context.Context) error {
		if err := impl.loadUserAndPasswordWindow(pl.ctx, pl.auth.id, pl.auth.password); err != nil {
			// fmt.Println("[DBG]retry loadUserAndPasswordWindow")
			takeSnapshot(pl.ctx, pl.snapshotsFolder, "iEnterMyUsernameAndPassword")
			// This marks the error as retryable
			return retry.RetryableError(err)
		}
		// fmt.Println("[DBG]success loadUserAndPasswordWindow")
		// takeSnapshot(pl.ctx, pl.snapshotsFolder, "loadUserAndPasswordWindow_success")
		return nil
	}); err != nil {
		return errortree.Add(rcerror, "iEnterMyUsernameAndPassword", err)
	}

	return nil
}

func (pl *loginPage) iClickTheLoginButton() error {
	var rcerror error

	impl := loginPageImpl{
		snapshotsFolder: pl.snapshotsFolder,
	}
	c := context.Background()
	b := retry.NewConstant(500 * time.Millisecond)
	b = retry.WithMaxDuration(7*time.Second, b)
	if err := retry.Do(c, b, func(ct context.Context) error {
		if err := impl.loadConsentAzurePage(pl.ctx); err != nil {
			// fmt.Println("[DBG]retry loadConsentAzurePage")
			takeSnapshot(pl.ctx, pl.snapshotsFolder, "iClickTheLoginButton")
			// This marks the error as retryable
			return retry.RetryableError(err)
		}
		// fmt.Println("[DBG]success loadConsentAzurePage")
		// takeSnapshot(pl.ctx, pl.snapshotsFolder, "loadConsentAzurePage_success")
		return nil
	}); err != nil {
		return errortree.Add(rcerror, "iClickTheLoginButton", err)
	}

	return nil
}

func (pl *loginPage) iShouldBeRedirectedToTheDashboardPage() error {
	var rcerror error

	impl := loginPageImpl{
		snapshotsFolder: pl.snapshotsFolder,
	}
	c := context.Background()
	b := retry.NewConstant(500 * time.Millisecond)
	b = retry.WithMaxDuration(7*time.Second, b)
	if err := retry.Do(c, b, func(ct context.Context) error {
		if err := impl.isMainFELoad(pl.ctx); err != nil {
			// fmt.Println("[DBG]retry isMainFELoad")
			takeSnapshot(pl.ctx, pl.snapshotsFolder, "iShouldBeRedirectedToTheDashboardPage")
			// This marks the error as retryable
			return retry.RetryableError(err)
		}
		// fmt.Println("[DBG]success isMainFELoad")
		// takeSnapshot(pl.ctx, pl.snapshotsFolder, "isMainFELoad_success")
		return nil
	}); err != nil {
		return errortree.Add(rcerror, "iShouldBeRedirectedToTheDashboardPage", err)
	}

	return nil
}
