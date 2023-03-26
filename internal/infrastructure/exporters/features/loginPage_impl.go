package features

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"fry.org/cmo/cli/internal/infrastructure/exporters"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/speijnik/go-errortree"
)

type loginPageImpl struct{}

func (l *loginPageImpl) loadUserAndPasswordWindow(ctx context.Context, user string, pass string) error {
	var rcerror error

	// Wait for the email input field to become available
	emailInput := `//input[@type='email']`
	err := chromedp.Run(ctx, chromedp.Click(emailInput))
	if err != nil || errors.Is(err, context.Canceled) {
		return errortree.Add(rcerror, "loadUserAndPasswordWindow:getEmailbox", err)
	}

	// Fill in the email address
	err = chromedp.Run(ctx, chromedp.SendKeys(emailInput, user, chromedp.BySearch))
	if err != nil || errors.Is(err, context.Canceled) {
		return errortree.Add(rcerror, "loadUserAndPasswordWindow:fillEmail", err)
	}

	// Click the "Next" button to proceed to the password page
	nextButton := `//input[@value='Next']`
	err = chromedp.Run(ctx, chromedp.Click(nextButton))
	if err != nil || errors.Is(err, context.Canceled) {
		return errortree.Add(rcerror, "loadUserAndPasswordWindow:submitEmail", err)
	}

	// Wait for the password input field to become available
	passwordInput := `//input[@type='password']`
	err = chromedp.Run(ctx, chromedp.Click(passwordInput))
	if err != nil || errors.Is(err, context.Canceled) {
		return errortree.Add(rcerror, "loadUserAndPasswordWindow:getPasswordBox", err)
	}

	// Fill in the password
	err = chromedp.Run(ctx, chromedp.SendKeys(passwordInput, pass, chromedp.BySearch))
	if err != nil || errors.Is(err, context.Canceled) {
		return errortree.Add(rcerror, "loadUserAndPasswordWindow:fillPassword", err)
	}
	// Click the "Sign in" button to proceed to the OAuth2 consent page
	signInButton := `//input[@type='submit']`
	time.Sleep(3 * time.Second)
	err = chromedp.Run(ctx, chromedp.Click(signInButton))
	if err != nil || errors.Is(err, context.Canceled) {
		return errortree.Add(rcerror, "loadUserAndPasswordWindow:submitPassword", err)
	}

	return nil
}

func (l *loginPageImpl) loadConsentAzurePage(ctx context.Context) error {
	var rcerror error

	// Click the "Accept" button to finish the OAuth2 flow
	acceptButton := `//input[@type='submit']`
	err := chromedp.Run(ctx, chromedp.Click(acceptButton))
	if err != nil || errors.Is(err, context.Canceled) {
		return errortree.Add(rcerror, "loadConsentAzurePage:submitOauth2", err)
	}

	return nil
}

func (l *loginPageImpl) doAzureLogin(ctx context.Context) error {
	var rcerror, err error
	var redirectedURL, target string

	target, err = exporters.StringFromContext(ctx, exporters.ContextKeyTargetUrl)
	if err != nil || errors.Is(err, context.Canceled) {
		return errortree.Add(rcerror, "doAzureLogin:extractURL", err)
	}
	// Start by navigating to the login page
	err = chromedp.Run(ctx, chromedp.Navigate(target))
	if err != nil || errors.Is(err, context.Canceled) {
		return errortree.Add(rcerror, "doAzureLogin:navigateURL", err)
	}

	// Check if the page has been redirected
	err = chromedp.Run(ctx, chromedp.Evaluate(`window.location.href`, &redirectedURL))
	if err != nil || errors.Is(err, context.Canceled) {
		return errortree.Add(rcerror, "doAzureLogin:checkredirection", err)
	}
	if strings.Contains(redirectedURL, target) {
		return errortree.Add(rcerror, "doAzureLogin", errors.New("redirection failed"))
	}

	return nil
}

func (l *loginPageImpl) isMainFELoad(ctx context.Context) error {
	var rcerror error
	// check if main.css has been loaded

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventExceptionThrown:
			// Since ts.URL uses a random port, replace it.
			s := ev.ExceptionDetails.Error()
			fmt.Printf("* %s\n", s)
		}
	})

	cssLoaded := false
	err := chromedp.Run(ctx, chromedp.EvaluateAsDevTools(`
		Array.from(document.querySelectorAll('link[rel="stylesheet"]'))
			.some(link => link.href.includes('main.min.css'))
	`, &cssLoaded))
	if err != nil {
		return errortree.Add(rcerror, "isMainFELoad:loadcss", err)
	}
	// log.Printf("main.css loaded: %v", cssLoaded)

	// check if main.js has been loaded
	jsLoaded := false
	err = chromedp.Run(ctx, chromedp.EvaluateAsDevTools(`
		Array.from(document.querySelectorAll('script[src]'))
			.some(script => script.src.includes('main.min.js'))`, &jsLoaded))
	if err != nil {
		return errortree.Add(rcerror, "isMainFELoad:loadjs", err)
	}
	// log.Printf("main.js loaded: %v", jsLoaded)
	err = waitUntilLoads(ctx, "h3")
	if err != nil {
		return errortree.Add(rcerror, "isMainFELoad", errors.New("failed to load h3 element in main page"))
	}
	//Last, but not least, check if CREATION-PORTAL title is part of the html
	htmlLoaded := ""
	htmlContent := ""
	expectedText := "CREATION PORTAL"

	chromedp.Run(ctx, chromedp.Evaluate(`document.documentElement.outerHTML`, &htmlContent))
	err = chromedp.Run(ctx, chromedp.Evaluate(`document.querySelector('h3') !== null ? document.querySelector('h3').textContent : null`, &htmlLoaded))
	if err != nil {
		if htmlLoaded == "" {
			return errortree.Add(rcerror, "isMainFELoad", errors.New("creation portal element not found in html main page"))
		} else {
			if expectedText != htmlLoaded {
				return errortree.Add(rcerror, "isMainFELoad", errors.New("element from main page didn't match which creation portal expression"))
			}
		}

	}
	return nil
}

func (l *loginPageImpl) doFeature(ctx context.Context, user string, pass string) error {
	var rcerror, err error

	impl := loginPageImpl{}

	time.Sleep(5 * time.Second)
	if err = impl.doAzureLogin(ctx); err != nil {
		return errortree.Add(rcerror, "doFeature.iEnterMyUsernameAndPassword", err)
	}

	time.Sleep(5 * time.Second)
	if err = impl.loadUserAndPasswordWindow(ctx, user, pass); err != nil {
		return errortree.Add(rcerror, "doFeature.iEnterMyUsernameAndPassword", err)
	}
	time.Sleep(5 * time.Second)
	if err = impl.loadConsentAzurePage(ctx); err != nil {
		return errortree.Add(rcerror, "doFeature.iClickTheLoginButton", err)
	}
	time.Sleep(5 * time.Second)
	if err := impl.isMainFELoad(ctx); err != nil {
		return errortree.Add(rcerror, "iShouldBeRedirectedToTheDashboardPage", err)
	}

	return nil
}

// func navigate(ctx context.Context) error {
// 	var target string
// 	var rcerror, err error

// 	target, err = exporters.StringFromContext(ctx, exporters.ContextKeyTargetUrl)
// 	if err != nil || errors.Is(err, context.Canceled) {
// 		return errortree.Add(rcerror, "doAzureLogin:extractURL", err)
// 	}
// 	// Start by navigating to the login page
// 	err = chromedp.Run(ctx, chromedp.Navigate(target))
// 	if err != nil || errors.Is(err, context.Canceled) {
// 		return errortree.Add(rcerror, "doAzureLogin:navigateURL", err)
// 	}
// 	return nil
// }
