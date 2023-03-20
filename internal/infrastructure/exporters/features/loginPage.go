package features

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"fry.org/cmo/cli/internal/infrastructure/exporters"
	"github.com/chromedp/chromedp"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/iancoleman/strcase"
	"github.com/speijnik/go-errortree"
)

// var (
// 	contextKeyScenarioName = contextKey("scenarioName")
// 	// contextKeyTargetUrl    = contextKey("targetUrl")
// )

// type contextKey string

// func (c contextKey) String() string {
// 	return "loginPage." + string(c)
// }

type loginPage struct {
	featureFolder string
	ctx           context.Context
	stats         exporters.CucumberStatsSet
	auth          struct {
		id       string
		password string
	}
}

func stringFromContext(ctx context.Context, key exporters.ContextKey) (string, error) {
	var value string
	var ok bool
	var rcerror error

	if value, ok = ctx.Value(key).(string); !ok {
		return "", errortree.Add(rcerror, "stringFromContext", fmt.Errorf("type mismatch with key %s", key))
	}

	return value, nil
}

func NewLoginPageFeature(path string, opts ...exporters.ExporterOption) (exporters.CucumberPlugin, error) {
	var rcerror error

	l := loginPage{
		featureFolder: path,
	}
	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(&l); err != nil {
			return nil, errortree.Add(rcerror, "NewLoginPageFeature", err)
		}
	}

	return &l, nil
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

		return errortree.Add(rcerror, "WithLoginPageAuth", errors.New("type mismatch, cucumberHandler expected"))
	})
}

func (l *loginPage) suiteInit(ctx *godog.TestSuiteContext) {

	ctx.BeforeSuite(func() {
		// This code will be executed once, before any scenarios are run
		l.stats = make(map[string][]exporters.CucumberStats)
	})
}

func (l *loginPage) scenarioInit(ctx *godog.ScenarioContext) {

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

		err := l.ctx.Err()
		if err != nil {
			// fmt.Printf("[DBG] '%v', context error: '%v')\n", err, st.Text)
			return c, errortree.Add(rcerror, "step.Before", err)
		}

		if name, err := stringFromContext(c, exporters.ContextKeyScenarioName); err != nil {
			return c, errortree.Add(rcerror, "step.Before", err)
		} else {
			l.stats[name] = append(l.stats[name], stat)
		}

		return c, nil
	})
	stepCtx.After(func(c context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
		var rcerror error

		if name, e := stringFromContext(c, exporters.ContextKeyScenarioName); e != nil {
			return c, errortree.Add(rcerror, "step.After", e)
		} else {
			stat := l.stats[name][len(l.stats[name])-1]
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
			l.stats[name][len(l.stats[name])-1] = stat
		}
		return c, nil
	})
	ctx.Step(`^I am on the login page$`, l.iAmOnTheLoginPage)
	ctx.Step(`^I enter my username and password$`, l.iEnterMyUsernameAndPassword)
	ctx.Step(`^I click the login button$`, l.iClickTheLoginButton)
	ctx.Step(`^I should be redirected to the dashboard page$`, l.iShouldBeRedirectedToTheDashboardPage)
}

func (l *loginPage) iAmOnTheLoginPage() error {
	var rcerror error

	// fmt.Println("I am on the login page")
	err := l.doAzureLogin()
	if err != nil {
		// fmt.Printf("[DBG] Error step: I am on the login page: '%v')\n", err)
		takeSnapshot(l.ctx, "iAmOnTheLoginPage")
		return errortree.Add(rcerror, "iAmOnTheLoginPage", err)
	}
	// fmt.Printf("[DBG]I am on the login page finished\n")

	return nil
}

func (l *loginPage) iEnterMyUsernameAndPassword() error {
	var rcerror error

	// fmt.Println("I enter my username and password")
	err := l.loadUserAndPasswordWindow()
	if err != nil {
		// fmt.Printf("[DBG] Error step: I enter my username and password: '%v')\n", err)
		takeSnapshot(l.ctx, "iEnterMyUsernameAndPassword")
		return errortree.Add(rcerror, "iEnterMyUsernameAndPassword", err)
	}
	// fmt.Printf("[DBG]I enter my username and password finished\n")

	return nil
}

func (l *loginPage) iClickTheLoginButton() error {
	var rcerror error

	// fmt.Println("I click the login button")
	err := l.loadConsentAzurePage()
	if err != nil {
		// fmt.Printf("[DBG] Error step: I click the login button: '%v')\n", err)
		takeSnapshot(l.ctx, "iClickTheLoginButton")
		return errortree.Add(rcerror, "iClickTheLoginButton", err)
	}
	// fmt.Printf("[DBG]I click the login button finished\n")

	return nil
}

func (l *loginPage) iShouldBeRedirectedToTheDashboardPage() error {
	var rcerror error

	// fmt.Println("I should be redirected to the dashboard page")
	err := l.isMainFELoad()
	if err != nil {
		// fmt.Printf("[DBG] Error step: I should be redirected to the dashboard page: '%v')\n", err)
		takeSnapshot(l.ctx, "iShouldBeRedirectedToTheDashboardPage")
		return errortree.Add(rcerror, "iShouldBeRedirectedToTheDashboardPage", err)
	}
	// fmt.Printf("[DBG]I should be redirected to the dashboard page finished\n")

	return nil
}

func (l *loginPage) Do(c context.Context, cancel context.CancelFunc) (exporters.CucumberStatsSet, error) {
	var rcerror error
	var rc int

	// //Initialize chromedp context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.111 Safari/537.36"),
	)
	actx, _ := chromedp.NewExecAllocator(c, opts...)
	l.ctx, _ = chromedp.NewContext(actx)
	godogOpts := godog.Options{
		//TODO: Remove colored output after debugging
		// Output: io.Discard,
		Output: colors.Colored(os.Stdout),
		Paths:  []string{l.featureFolder},
		//pretty, progress, cucumber, events and junit
		Format:        "pretty",
		StopOnFailure: true,
		//This is the context passed as argument to scenario hooks
		DefaultContext: l.ctx,
	}
	suite := godog.TestSuite{
		Name:                 "loginPage",
		TestSuiteInitializer: l.suiteInit,
		ScenarioInitializer:  l.scenarioInit,
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
		return l.stats, nil
	case 1:
		return l.stats, errortree.Add(rcerror, "loginPage.Do", fmt.Errorf("error  %d: failed test suite", rc))
	case 2:
		return l.stats, errortree.Add(rcerror, "loginPage.Do", fmt.Errorf("error %d:command line usage error running test suite", rc))
	default:
		return l.stats, errortree.Add(rcerror, "loginPage.Do", fmt.Errorf("error %d running test suite", rc))
	}

	//return l.stats, nil
}
