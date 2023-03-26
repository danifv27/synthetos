package features

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"fry.org/cmo/cli/internal/infrastructure/exporters"
	"github.com/chromedp/chromedp"
	"github.com/speijnik/go-errortree"
)

type productTabsImpl struct{}

func (pl *productTabsImpl) loadModelProductsPage(ctx context.Context) error {
	var rcerror, err error
	var target string
	destinationPath := "/products"
	if target, err = exporters.StringFromContext(ctx, exporters.ContextKeyTargetUrl); err != nil {
		return errortree.Add(rcerror, "loadProducts:composeURL", err)
	}
	// Start by navigating to the products - article page
	if err = chromedp.Run(ctx, chromedp.Navigate(target+destinationPath)); err != nil {
		return errortree.Add(rcerror, "loadProducts:navigate", err)
	}

	return nil
}

func (pl *productTabsImpl) loadModelDataInTable(ctx context.Context) error {
	var rcerror, err error
	// select all the rows in the table
	var rowCount = ""
	err = waitUntilLoads(ctx, `.ag-root`)
	if err != nil {
		return errortree.Add(rcerror, "isMainFELoad", errors.New("failed to load main table element in main page"))
	}
	err = chromedp.Run(ctx, chromedp.Evaluate(`document.querySelector('.ag-root').getAttribute('aria-rowcount') !== null ? document.querySelector('.ag-root').getAttribute('aria-rowcount') : null`, &rowCount))
	if err != nil {
		if rowCount == "" {
			return errortree.Add(rcerror, "loadArticleDataInTable:loadTableInfo", errors.New("table products element not found in page"))
		}
	} else {
		if num, err := strconv.Atoi(rowCount); err == nil && num == 0 {
			return errortree.Add(rcerror, "loadArticleDataInTable:checkRowsInTable", err)
		}
	}
	return nil
}

func (pl *productTabsImpl) loadArticleDataInfoFromTable(ctx context.Context) error {
	var target string
	var rcerror, err error

	modelNumber := ""
	season := ""

	err = waitUntilLoads(ctx, `[col-id="mdl.modelNumber"]`)
	if err != nil {
		return errortree.Add(rcerror, "isMainFELoad", errors.New("failed to load first element table in main page"))
	}
	err = chromedp.Run(ctx, chromedp.Evaluate(`document.querySelectorAll('[col-id="mdl.modelNumber"]')[1].textContent !== null ? document.querySelectorAll('[col-id="mdl.modelNumber"]')[1].textContent : null`, &modelNumber))
	if err != nil {
		if modelNumber == "" {
			return errortree.Add(rcerror, "loadArticleDataInfoFromTable:getModelNumberElement", errors.New("modelNumber not found in page"))
		}
	}
	err = chromedp.Run(ctx, chromedp.Evaluate(`document.querySelectorAll('[col-id="mdl.season"]')[1].textContent !== null ? document.querySelectorAll('[col-id="mdl.season"]')[1].textContent : null`, &season))
	if err != nil {
		if season == "" {
			return errortree.Add(rcerror, "loadArticleDataInfoFromTable:getArticleNumberElement", errors.New("season not found in page"))
		}
	}

	//Last, navigate into details page
	path := fmt.Sprintf("/product/%s/%s", modelNumber, getSeasonFrom(season))
	if err = chromedp.Run(ctx, chromedp.Navigate(target+path)); err != nil {
		return errortree.Add(rcerror, "loadProductDetails:navigate", err)
	}

	return nil
}

func (pl *productTabsImpl) checkProductDetailsPage(ctx context.Context) error {
	var rcerror, err error
	var modelDetails = ""
	err = waitUntilLoads(ctx, `.product-details-header h4`)
	if err != nil {
		return errortree.Add(rcerror, "isMainFELoad", errors.New("failed to load product details page"))
	}
	err = chromedp.Run(ctx, chromedp.Evaluate(`document.querySelector('.product-details-header h4').textContent !== null ? document.querySelector('.product-details-header h4').textContent : null`, &modelDetails))
	if err != nil {
		if modelDetails == "" {
			return errortree.Add(rcerror, "checkProductDetailsPage:getDetails", errors.New("model information not found in page"))
		}
	}
	return nil
}
