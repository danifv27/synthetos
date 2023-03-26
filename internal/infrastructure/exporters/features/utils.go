package features

import (
	"context"
	"os"

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
	err = os.WriteFile("/app/bin/features/snapshots/"+stepName+".png", buf, 0644)
	if err != nil {
		return errortree.Add(rcerror, "failed to save snapshot", err)
	}
	return nil
}

// func getSeasonFrom(season string) string {
// 	var seasonNumber string
// 	numStr := season[len(season)-2:]
// 	if strings.Contains(season, "SS") {
// 		seasonNumber = "20" + numStr + "1"
// 	} else {
// 		seasonNumber = "20" + numStr + "0"
// 	}
// 	return seasonNumber
// }

func waitUntilLoads(ctx context.Context, elementQuery string) error {
	var rcerror error
	for {
		// Reload the current page
		err := chromedp.Run(ctx, chromedp.Reload())
		if err != nil {
			return errortree.Add(rcerror, "reloadPage", err)
		}

		// Wait for the HTML object to become visible
		err = chromedp.Run(ctx, chromedp.WaitVisible(elementQuery, chromedp.ByQuery))
		if err == nil {
			// Object found, break out of loop
			break
		} else {
			// Object not found, continue looping
			continue
		}
	}
	return nil

}
