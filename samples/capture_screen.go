package samples

import (
	"context"
	"io/ioutil"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
)

var CaptureAction = chromedp.ActionFunc(func(ctxt context.Context) error {
	var buf []byte
	tasks := chromedp.Tasks{
		emulation.SetDeviceMetricsOverride(1680, 2048, 0, false),
		chromedp.CaptureScreenshot(&buf),
	}
	if err := tasks.Do(ctxt); err != nil {
		return errors.WithStack(err)
	}
	if err := ioutil.WriteFile("error.png", buf, 0644); err != nil {
		return errors.WithStack(err)
	}
	return nil
})
