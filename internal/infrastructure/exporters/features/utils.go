package features

import (
	"context"
	"io/ioutil"
	"strings"

	"github.com/chromedp/chromedp"
	"github.com/speijnik/go-errortree"
)

func takeSnapshot(ctx context.Context, stepName string) error {

	var rcerror error
	// take screenshot
	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.CaptureScreenshot(&buf),
	)
	if err != nil {
		return errortree.Add(rcerror, "failed to take snapshot", err)
	}

	// save screenshot to file
	err = ioutil.WriteFile("/app/bin/features/snapshots/"+stepName+".png", buf, 0644)
	if err != nil {
		return errortree.Add(rcerror, "failed to save snapshot", err)
	}
	return nil
}

func getSeasonFrom(season string) string {
	var seasonNumber string
	numStr := season[len(season)-2:]
	if strings.Contains(season, "SS") {
		seasonNumber = "20" + numStr + "1"
	} else {
		seasonNumber = "20" + numStr + "0"
	}
	return seasonNumber
}
