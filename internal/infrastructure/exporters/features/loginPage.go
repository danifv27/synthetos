package features

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"time"

	"fry.org/cmo/cli/internal/infrastructure/exporters"
	"github.com/cucumber/godog"
	"github.com/speijnik/go-errortree"
)

type loginPage struct {
	scenarioID string
	stats      exporters.CucumberStatsSet
}

func NewLoginPageFeature(opts ...exporters.ExporterOption) (exporters.CucumberPlugin, error) {
	var rcerror error

	rand.Seed(time.Now().UnixNano())
	l := loginPage{}
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
		// fmt.Printf("[DBG]Before scenario hook %s\n", sc.Id)

		// This code will be executed once, before any scenarios are run
		l.scenarioID = sc.Id

		return c, nil
	})

	ctx.After(func(c context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		// This code will be executed once, after all scenarios have been run
		l.scenarioID = ""

		return c, nil
	})
	stepCtx := ctx.StepContext()
	stepCtx.Before(func(c context.Context, st *godog.Step) (context.Context, error) {

		stat := exporters.CucumberStats{
			Id:     st.Id,
			Start:  time.Now(),
			Result: exporters.CucumberNotExecuted,
		}
		l.stats[l.scenarioID] = append(l.stats[l.scenarioID], stat)
		// fmt.Printf("[DBG]Before step hook %s (sc.Id %s)\n", st.Id, l.scenarioID)

		return c, nil
	})
	stepCtx.After(func(ctx context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {

		// fmt.Printf("[DBG]After step hook %s\n", st.Id)
		stat := l.stats[l.scenarioID][len(l.stats[l.scenarioID])-1]

		stat.Duration = time.Since(stat.Start)
		if status == godog.StepPassed {
			stat.Result = exporters.CucumberSuccess
		} else {
			stat.Result = exporters.CucumberFailure
		}
		l.stats[l.scenarioID][len(l.stats[l.scenarioID])-1] = stat

		return ctx, nil
	})
	ctx.Step(`^I am on the login page$`, l.iAmOnTheLoginPage)
	ctx.Step(`^I enter my username and password$`, l.iEnterMyUsernameAndPassword)
	ctx.Step(`^I click the login button$`, l.iClickTheLoginButton)
	ctx.Step(`^I should be redirected to the dashboard page$`, l.iShouldBeRedirectedToTheDashboardPage)
}

func (l *loginPage) iAmOnTheLoginPage() error {

	time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
	fmt.Println("[DBG]I am on the login page")

	return nil
}

func (l *loginPage) iEnterMyUsernameAndPassword() error {

	time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
	fmt.Println("[DBG]I enter my username and password")

	return nil
}

func (l *loginPage) iClickTheLoginButton() error {

	time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
	fmt.Println("[DBG]I click the login button")

	return nil
}

func (l *loginPage) iShouldBeRedirectedToTheDashboardPage() error {

	time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
	fmt.Println("[DBG]I should be redirected to the dashboard page")

	return nil
}

func (l *loginPage) Do(ctx context.Context) (exporters.CucumberStatsSet, error) {

	godogOpts := godog.Options{
		Output: io.Discard,
		Paths:  []string{"/Users/fraildan/Proyectos/packagesGit/github/danifv27_donotdelete/synthetos/internal/infrastructure/exporters/features"},
		//pretty, progress, cucumber, events and junit
		Format: "pretty",
	}
	suite := godog.TestSuite{
		Name:                 "loginPage",
		TestSuiteInitializer: l.suiteInit,
		ScenarioInitializer:  l.scenarioInit,
		Options:              &godogOpts,
	}
	suite.Run()

	return l.stats, nil
}
