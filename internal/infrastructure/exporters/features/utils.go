package features

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/chromedp/chromedp"
	"github.com/speijnik/go-errortree"
)

func takeSnapshot(ctx context.Context, folder string, stepName string) error {

	var rcerror, err error
	// take screenshot
	var buf []byte
	if err = chromedp.Run(ctx, chromedp.CaptureScreenshot(&buf)); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return errortree.Add(rcerror, "failed to take snapshot", err)
	}
	// save screenshot to file
	if err = os.WriteFile(fmt.Sprintf("%s.png", path.Join(folder, stepName)), buf, 0644); err != nil {
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

// FIXME: this wait would last forever
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
