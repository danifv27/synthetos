package features

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"time"

	"fry.org/cmo/cli/internal/infrastructure/exporters"
	"github.com/cucumber/godog"
	"github.com/iancoleman/strcase"
	"github.com/speijnik/go-errortree"
)

var (
	contextKeyScenarioName = contextKey("scenarioName")
)

type contextKey string

func (c contextKey) String() string {
	return "loginPage." + string(c)
}

type loginPage struct {
	featureFolder string
	ctx           context.Context
	stats         exporters.CucumberStatsSet
}

func scenarioNameFromContext(ctx context.Context) (string, error) {
	var name string
	var ok bool
	var rcerror error

	if name, ok = ctx.Value(contextKeyScenarioName).(string); !ok {
		return "", errortree.Add(rcerror, "scenarioNameFromContext", fmt.Errorf("type mismatch with key %s", contextKeyScenarioName))
	}

	return name, nil
}

func NewLoginPageFeature(path string, opts ...exporters.ExporterOption) (exporters.CucumberPlugin, error) {
	var rcerror error

	//TODO: remove seed when implementing steps
	rand.Seed(time.Now().UnixNano())
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

func (l *loginPage) suiteInit(ctx *godog.TestSuiteContext) {

	ctx.BeforeSuite(func() {

		l.stats = make(map[string][]exporters.CucumberStats)
		// This code will be executed once, before any scenarios are run
	})
}

func (l *loginPage) scenarioInit(ctx *godog.ScenarioContext) {

	ctx.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {

		// This code will be executed once, before any scenarios are run
		return context.WithValue(c, contextKeyScenarioName, strcase.ToCamel(sc.Name)), nil
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
		if name, err := scenarioNameFromContext(c); err != nil {
			return c, errortree.Add(rcerror, "step.Before", err)
		} else {
			l.stats[name] = append(l.stats[name], stat)
		}

		return c, nil
	})
	stepCtx.After(func(c context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
		var rcerror error

		if name, err := scenarioNameFromContext(c); err != nil {
			return c, errortree.Add(rcerror, "step.Before", err)
		} else {
			stat := l.stats[name][len(l.stats[name])-1]
			stat.Duration = time.Since(stat.Start)
			if status == godog.StepPassed {
				stat.Result = exporters.CucumberSuccess
			} else {
				stat.Result = exporters.CucumberFailure
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

func (l *loginPage) iAmOnTheLoginPage(ctx context.Context) error {

	d := time.Duration(5+rand.Intn(3)) * time.Second
	fmt.Printf("[DBG]I am on the login page (sleeping %v)\n", d)
	err := l.ctx.Err()
	if err != nil {
		fmt.Printf("[DBG]I am on the login page, context error: '%v')\n", err)
		return err
	}
	// Do work
	time.Sleep(d)
	fmt.Printf("[DBG]I am on the login page finished\n")

	return nil
}

func (l *loginPage) iEnterMyUsernameAndPassword() error {

	d := time.Duration(1+rand.Intn(3)) * time.Second
	fmt.Printf("[DBG]I enter my username and password (sleeping %v\n", d)
	err := l.ctx.Err()
	if err != nil {
		fmt.Printf("[DBG]I enter my username and password, context error: '%v')\n", err)
		return err
	}
	// Do work
	time.Sleep(d)
	fmt.Printf("[DBG]I enter my username and password finished\n")

	return nil
}

func (l *loginPage) iClickTheLoginButton() error {

	d := time.Duration(1+rand.Intn(3)) * time.Second
	err := l.ctx.Err()
	fmt.Printf("[DBG]I click the login button (sleeping %v)\n", d)
	if err != nil {
		fmt.Printf("[DBG]I click the login button, context error: '%v')\n", err)
		return err
	}
	// Do work
	time.Sleep(d)
	fmt.Printf("[DBG]I click the login button finished\n")

	return nil
}

func (l *loginPage) iShouldBeRedirectedToTheDashboardPage() error {

	d := time.Duration(1+rand.Intn(3)) * time.Second
	err := l.ctx.Err()
	fmt.Printf("[DBG]I should be redirected to the dashboard page (sleeping %v)\n", d)
	if err != nil {
		fmt.Printf("[DBG]I should be redirected to the dashboard page, context error: '%v')\n", err)
		return err
	}
	// Do work
	time.Sleep(d)
	fmt.Printf("[DBG]I should be redirected to the dashboard page finished\n")

	return nil
}

func (l *loginPage) Do(c context.Context, cancel context.CancelFunc) (exporters.CucumberStatsSet, error) {
	var rcerror error
	var rc int

	l.ctx = c
	godogOpts := godog.Options{
		Output: io.Discard,
		Paths:  []string{l.featureFolder},
		//pretty, progress, cucumber, events and junit
		Format:        "junit",
		StopOnFailure: true,
		//This is the context passed as argument to scenario hooks
		DefaultContext: c,
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
	fmt.Printf("[DBG]Waiting for context done\n")
	<-done
	switch rc {
	case 0:
		return l.stats, nil
	case 1:
		return exporters.CucumberStatsSet{}, errortree.Add(rcerror, "loginPage.Do", fmt.Errorf("error  %d: failed test suite", rc))
	case 2:
		return exporters.CucumberStatsSet{}, errortree.Add(rcerror, "loginPage.Do", fmt.Errorf("error %d:command line usage error running test suite", rc))
	default:
		return exporters.CucumberStatsSet{}, errortree.Add(rcerror, "loginPage.Do", fmt.Errorf("error %d running test suite", rc))
	}

	// return exporters.CucumberStatsSet{}, nil
}
