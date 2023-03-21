package features

import (
	"errors"
	"fmt"
	"strconv"

	"fry.org/cmo/cli/internal/infrastructure/exporters"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/speijnik/go-errortree"
)

var (
	rowNodes []*cdp.Node
	target   string
)

type productTabsImpl struct{}

func (pl *productsTab) loadModelProductsPage() error {
	var rcerror, err error
	var destinationPath = "/products"

	if target, err = exporters.StringFromContext(pl.ctx, exporters.ContextKeyTargetUrl); err != nil {
		return errortree.Add(rcerror, "loadProducts:composeURL", err)
	}
	// Start by navigating to the products - article page
	if err = chromedp.Run(pl.ctx, chromedp.Navigate(target+destinationPath)); err != nil {
		return errortree.Add(rcerror, "loadProducts:navigate", err)
	}

	return nil
}

func (pl *productsTab) loadModelDataInTable() error {
	var rcerror, err error
	// select all the rows in the table
	var rowCount = ""
	err = chromedp.Run(pl.ctx, chromedp.Evaluate(`document.querySelector('.ag-root').getAttribute('aria-rowcount') !== null ? document.querySelector('.ag-root').getAttribute('aria-rowcount') : null`, &rowCount))
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

func (pl *productsTab) loadArticleDataInfoFromTable() error {

	var rcerror, err error
	// find the first element in the "%s" column
	var modelNumber = ""
	var season = ""
	err = chromedp.Run(pl.ctx, chromedp.Evaluate(`document.querySelectorAll('[col-id="mdl.modelNumber"]')[1].textContent !== null ? document.querySelectorAll('[col-id="mdl.modelNumber"]')[1].textContent : null`, &modelNumber))
	if err != nil {
		if modelNumber == "" {
			return errortree.Add(rcerror, "loadArticleDataInfoFromTable:getModelNumberElement", errors.New("modelNumber not found in page"))
		}
	}
	err = chromedp.Run(pl.ctx, chromedp.Evaluate(`document.querySelectorAll('[col-id="mdl.season"]')[1].textContent !== null ? document.querySelectorAll('[col-id="mdl.season"]')[1].textContent : null`, &season))
	if err != nil {
		if season == "" {
			return errortree.Add(rcerror, "loadArticleDataInfoFromTable:getArticleNumberElement", errors.New("season not found in page"))
		}
	}

	//Last, navigate into details page
	path := fmt.Sprintf("/product/%s/%s", modelNumber, getSeasonFrom(season))
	if err = chromedp.Run(pl.ctx, chromedp.Navigate(target+path)); err != nil {
		return errortree.Add(rcerror, "loadProductDetails:navigate", err)
	}
	return nil
}

func (pl *productsTab) checkProductDetailsPage() error {
	var rcerror, err error
	var modelDetails = ""
	err = chromedp.Run(pl.ctx, chromedp.Evaluate(`document.querySelector('.product-details-header h4').textContent !== null ? document.querySelector('.product-details-header h4').textContent : null`, &modelDetails))
	if err != nil {
		if modelDetails == "" {
			return errortree.Add(rcerror, "checkProductDetailsPage:getDetails", errors.New("model information not found in page"))
		}
	}
	return nil
}
