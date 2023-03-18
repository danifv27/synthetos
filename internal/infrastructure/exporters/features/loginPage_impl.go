package features

import (
	"errors"
	"strings"
	"time"

	"fry.org/cmo/cli/internal/infrastructure/exporters"
	"github.com/chromedp/chromedp"
	"github.com/speijnik/go-errortree"
)

func (l *loginPage) loadUserAndPasswordWindow() error {
	var rcerror error

	// Wait for the email input field to become available
	emailInput := `//input[@type='email']`
	if err := chromedp.Run(l.ctx, chromedp.WaitVisible(emailInput)); err != nil {
		return errortree.Add(rcerror, "loadUserAndPasswordWindow", err)
	}

	// Fill in the email address
	if err := chromedp.Run(l.ctx, chromedp.SendKeys(emailInput, l.auth.id, chromedp.BySearch)); err != nil {
		return errortree.Add(rcerror, "loadUserAndPasswordWindow", err)
	}

	// Click the "Next" button to proceed to the password page
	nextButton := `//input[@value='Next']`
	if err := chromedp.Run(l.ctx, chromedp.Click(nextButton)); err != nil {
		return errortree.Add(rcerror, "loadUserAndPasswordWindow", err)
	}

	// Wait for the password input field to become available
	passwordInput := `//input[@type='password']`
	if err := chromedp.Run(l.ctx, chromedp.WaitVisible(passwordInput)); err != nil {
		return errortree.Add(rcerror, "loadUserAndPasswordWindow", err)
	}

	// Fill in the password
	if err := chromedp.Run(l.ctx, chromedp.SendKeys(passwordInput, l.auth.password, chromedp.BySearch)); err != nil {
		return errortree.Add(rcerror, "loadUserAndPasswordWindow", err)
	}
	// Click the "Sign in" button to proceed to the OAuth2 consent page
	signInButton := `//input[@type='submit']`
	time.Sleep(3 * time.Second)
	if err := chromedp.Run(l.ctx, chromedp.Click(signInButton)); err != nil {
		return errortree.Add(rcerror, "loadUserAndPasswordWindow", err)
	}

	return nil
}

func (l *loginPage) loadConsentAzurePage() error {
	var rcerror error

	// Wait for the consent checkbox to become available
	consentCheckbox := `//input[@type='checkbox']`
	if err := chromedp.Run(l.ctx, chromedp.WaitVisible(consentCheckbox)); err != nil {
		return errortree.Add(rcerror, "loadConsentAzurePage", err)
	}

	// Click the consent checkbox to give consent to the app
	if err := chromedp.Run(l.ctx, chromedp.Click(consentCheckbox)); err != nil {
		return errortree.Add(rcerror, "loadConsentAzurePage", err)
	}

	// Click the "Accept" button to finish the OAuth2 flow
	acceptButton := `//input[@type='submit']`
	if err := chromedp.Run(l.ctx, chromedp.Click(acceptButton)); err != nil {
		return errortree.Add(rcerror, "loadConsentAzurePage", err)
	}

	return nil
}

func (l *loginPage) doAzureLogin() error {
	var rcerror, err error
	var redirectedURL, target string

	if target, err = stringFromContext(l.ctx, exporters.ContextKeyTargetUrl); err != nil {
		return errortree.Add(rcerror, "doAzureLogin", err)
	}
	// Start by navigating to the login page
	if err = chromedp.Run(l.ctx, chromedp.Navigate(target)); err != nil {
		return errortree.Add(rcerror, "doAzureLogin", err)
	}

	// Check if the page has been redirected
	// redirectedURL := ""
	if err = chromedp.Run(l.ctx, chromedp.Evaluate(`window.location.href`, &redirectedURL)); err != nil {
		return errortree.Add(rcerror, "doAzureLogin", err)
	}
	if strings.Contains(redirectedURL, target) {
		return errortree.Add(rcerror, "doAzureLogin", errors.New("redirection failed"))
	}

	return nil
}

func (l *loginPage) isMainFELoad() error {
	var rcerror error

	time.Sleep(5 * time.Second)
	// check if main.css has been loaded
	cssLoaded := false
	err := chromedp.Run(l.ctx, chromedp.EvaluateAsDevTools(`
		Array.from(document.querySelectorAll('link[rel="stylesheet"]'))
			.some(link => link.href.includes('main.min.css'))
	`, &cssLoaded))
	if err != nil {
		return errortree.Add(rcerror, "isMainFELoad", err)
	}
	// log.Printf("main.css loaded: %v", cssLoaded)

	// check if main.js has been loaded
	jsLoaded := false
	err = chromedp.Run(l.ctx, chromedp.EvaluateAsDevTools(`
		Array.from(document.querySelectorAll('script[src]'))
			.some(script => script.src.includes('main.min.js'))`, &jsLoaded))
	if err != nil {
		return errortree.Add(rcerror, "isMainFELoad", err)
	}
	// log.Printf("main.js loaded: %v", jsLoaded)

	//Last, but not least, check if CREATION-PORTAL title is part of the html
	htmlLoaded := ""
	var expectedText = "CREATION PORTAL"
	err = chromedp.Run(l.ctx, chromedp.Evaluate(`document.querySelector('h3').textContent`, &htmlLoaded))
	if err != nil {
		return errortree.Add(rcerror, "isMainFELoad", err)
	} else {
		if expectedText != htmlLoaded {
			errortree.Add(rcerror, "isMainFELoad", errors.New("element from main page didn't match which creation portal expression"))
		}
	}
	// log.Printf("Creation Portal html loaded: %v", htmlLoaded)

	return nil
}
