package fabric

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"

	"github.com/li-go/chromedp-samples/samples"
)

type fablic struct {
	logger *log.Logger
}

func New(logger *log.Logger) *fablic {
	return &fablic{logger: logger}
}

func (f *fablic) Login(ctx context.Context, email, password string) error {
	cs := samples.CookieSaver{Name: "fabric", Logger: f.logger}
	if cs.RestoreCookies(ctx, "https://fabric.io/home") {
		return nil
	}

	f.logger.Println("Login to fabric.io")
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctxt context.Context) error {
		ctxt, cancel := context.WithTimeout(ctxt, 5*time.Second)
		defer cancel()
		tasks := chromedp.Tasks{
			chromedp.Navigate("https://fabric.io/login"),
			chromedp.SendKeys("#email", email, chromedp.NodeVisible),
			chromedp.SendKeys("#password", password),
			chromedp.Click("#l_sdk-chrome > div.relative.stage > div:nth-child(1) > div > form > button"),
			chromedp.Sleep(time.Second),
			chromedp.ActionFunc(func(ctxt context.Context) error {
				return errors.WithStack(cs.SaveCookies(ctxt))
			}),
		}
		return errors.WithStack(tasks.Do(ctxt))
	})); err != nil {
		f.logger.Printf("%+v", chromedp.Run(ctx, samples.CaptureAction))
		return errors.WithStack(err)
	}
	return nil
}

func (f *fablic) GetIOSCrashFreeUsers(ctx context.Context, organization, app string) (float32, error) {
	return f.getCrashFreeUsers(ctx,
		"https://fabric.io/"+organization+"/ios/apps/"+app+"/issues?time=last-seven-days&event_type=crash&subFilter=state&state=open&build%5B0%5D=top-builds")
}

func (f *fablic) GetAndroidCrashFreeUsers(ctx context.Context, organization, app string) (float32, error) {
	return f.getCrashFreeUsers(ctx,
		"https://fabric.io/"+organization+"/android/apps/"+app+"/issues?time=last-seven-days&event_type=crash&subFilter=state&state=open&build%5B0%5D=top-builds")
}

func (f *fablic) getCrashFreeUsers(ctx context.Context, url string) (float32, error) {
	const sel = "#l_dashboard > article > div.flex-1.flex-box > section > div > div > div.header-wrap > div.issues_metrics > div.stat-row.flex-box.top-bar > div > span > div:nth-child(1) > div > div.value > span:nth-child(1)"

	var crashFreeUsers string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctxt context.Context) error {
			ctxt, cancel := context.WithTimeout(ctxt, 10*time.Second)
			defer cancel()
			for {
				if err := chromedp.Text(sel, &crashFreeUsers, chromedp.NodeVisible).Do(ctxt); err != nil {
					return errors.WithStack(err)
				}
				if crashFreeUsers != "…" {
					return nil
				}
				time.Sleep(time.Second)
			}
		}),
	); err != nil {
		f.logger.Printf("%+v", chromedp.Run(ctx, samples.CaptureAction))
		return 0, errors.WithStack(err)
	}
	val, _ := strconv.ParseFloat(strings.TrimRight(crashFreeUsers, "%"), 32)
	return float32(val), nil
}
