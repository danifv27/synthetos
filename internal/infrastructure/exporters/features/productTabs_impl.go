package features

import (
	"fmt"

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

func (pl *productsTab) loadArticleProductsPage() error {
	var rcerror, err error
	var destinationPath = "/papp/product-search/article-season/items?page=0&size=20&sort=art.articleNumber,asc"

	if target, err = exporters.StringFromContext(pl.ctx, exporters.ContextKeyTargetUrl); err != nil {
		return errortree.Add(rcerror, "loadProducts:composeURL", err)
	}
	// Start by navigating to the products - article page
	if err = chromedp.Run(pl.ctx, chromedp.Navigate(target+destinationPath)); err != nil {
		return errortree.Add(rcerror, "loadProducts:navigate", err)
	}

	return nil
}

func (pl *productsTab) loadArticleDataInTable() error {
	var rcerror, err error
	// select all the rows in the table
	var rowNodes []*cdp.Node
	err = chromedp.Run(pl.ctx, chromedp.Nodes(`table tr`, &rowNodes))
	if err != nil {
		return errortree.Add(rcerror, "loadArticleDataInTable:loadTable", err)
	}
	return nil
}

func (pl *productsTab) loadArticleDataInfoFromTable() error {
	var rcerror, err error
	// find the first element in the "%s" column
	var modelNumber string
	var articleNumber string
	var season string
	path := fmt.Sprintf("/product/%s/%d/article/%e", modelNumber, season, articleNumber)
	err = chromedp.Run(pl.ctx, chromedp.Text("#mdl.modelNumber td:first-child", &modelNumber))
	if err != nil {
		return errortree.Add(rcerror, "loadArticleDataInfoFromTable:getModelNumberElement", err)
	}
	err = chromedp.Run(pl.ctx, chromedp.Text("#art.articleNumber td:first-child", &articleNumber))
	if err != nil {
		return errortree.Add(rcerror, "loadArticleDataInfoFromTable:getArticleNumberElement", err)
	}
	err = chromedp.Run(pl.ctx, chromedp.Text("#mdl.season td:first-child", &season))
	if err != nil {
		return errortree.Add(rcerror, "loadArticleDataInfoFromTable:getArticleNumberElement", err)
	}

	//Last, navigate into details page
	if err = chromedp.Run(pl.ctx, chromedp.Navigate(target+path)); err != nil {
		return errortree.Add(rcerror, "loadProductDetails:navigate", err)
	}
	return nil
}
