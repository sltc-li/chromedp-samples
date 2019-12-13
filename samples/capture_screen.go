package samples

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

var CaptureAction = chromedp.ActionFunc(func(ctxt context.Context) error {
	_, _, contentSize, err := page.GetLayoutMetrics().Do(ctxt)
	if err != nil {
		return fmt.Errorf("get layout metrics: %w", err)
	}
	if err := emulation.
		SetDeviceMetricsOverride(int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height)), 1, false).
		WithScreenOrientation(&emulation.ScreenOrientation{
			Type:  emulation.OrientationTypePortraitPrimary,
			Angle: 0,
		}).
		Do(ctxt); err != nil {
		return fmt.Errorf("set device metrics override: %w", err)
	}
	bs, err := page.CaptureScreenshot().WithQuality(100).WithClip(&page.Viewport{
		X:      contentSize.X,
		Y:      contentSize.Y,
		Width:  contentSize.Width,
		Height: contentSize.Height,
		Scale:  1,
	}).Do(ctxt)
	if err != nil {
		return fmt.Errorf("capture screenshot: %w", err)
	}
	if err := ioutil.WriteFile("error.png", bs, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
})
