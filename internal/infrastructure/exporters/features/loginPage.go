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
	// scenarioName  string
	featureFolder string
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

func (l *loginPage) iAmOnTheLoginPage() error {

	d := time.Duration(1+rand.Intn(3)) * time.Second
	time.Sleep(d)
	fmt.Printf("[DBG]I am on the login page (sleeping %v)\n", d)

	return nil
}

func (l *loginPage) iEnterMyUsernameAndPassword() error {

	d := time.Duration(1+rand.Intn(3)) * time.Second
	time.Sleep(d)
	fmt.Printf("[DBG]I enter my username and password (sleeping %v)\n", d)

	return nil
}

func (l *loginPage) iClickTheLoginButton() error {

	d := time.Duration(1+rand.Intn(3)) * time.Second
	time.Sleep(d)
	fmt.Printf("[DBG]I click the login button (sleeping %v)\n", d)

	return nil
}

func (l *loginPage) iShouldBeRedirectedToTheDashboardPage() error {

	d := time.Duration(1+rand.Intn(3)) * time.Second
	time.Sleep(d)
	fmt.Printf("[DBG]I should be redirected to the dashboard page (sleeping %v)\n", d)

	return nil
}

func (l *loginPage) Do(ctx context.Context) (exporters.CucumberStatsSet, error) {
	var rcerror error

	godogOpts := godog.Options{
		Output: io.Discard,
		Paths:  []string{l.featureFolder},
		//pretty, progress, cucumber, events and junit
		Format: "junit",
	}
	suite := godog.TestSuite{
		Name:                 "loginPage",
		TestSuiteInitializer: l.suiteInit,
		ScenarioInitializer:  l.scenarioInit,
		Options:              &godogOpts,
	}
	rc := suite.Run()
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
