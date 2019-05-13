package pagerduty

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"

	"github.com/liyy7/chromedp-samples/samples"
)

type pagerduty struct {
	logger *log.Logger
}

func New(logger *log.Logger) *pagerduty {
	return &pagerduty{logger: logger}
}

func (p *pagerduty) Login(ctx context.Context, organization, email, password string) error {
	cs := samples.CookieSaver{Name: "pagerduty", Logger: p.logger}
	if cs.RestoreCookies(ctx, "https://"+organization+".pagerduty.com/incidents") {
		return nil
	}

	p.logger.Println("Login to pagerduty.com")
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctxt context.Context) error {
		ctxt, cancel := context.WithTimeout(ctxt, 5*time.Second)
		defer cancel()
		tasks := chromedp.Tasks{
			chromedp.Navigate("https://" + organization + ".pagerduty.com/sign_in"),
			chromedp.SendKeys("#user_email", email, chromedp.NodeVisible),
			chromedp.SendKeys("#user_password", password),
			chromedp.Click("#login_form > fieldset > div:nth-child(4) > div > div > input"),
			chromedp.Sleep(time.Second),
			chromedp.ActionFunc(func(ctxt context.Context) error {
				cookies, err := network.GetAllCookies().Do(ctxt)
				if err != nil {
					return errors.WithStack(err)
				}
				p.logger.Println("Save cookies")
				if err := cs.SaveCookies(cookies); err != nil {
					return errors.WithStack(err)
				}
				return nil
			}),
		}
		return errors.WithStack(tasks.Do(ctxt))
	})); err != nil {
		p.logger.Printf("%+v", chromedp.Run(ctx, samples.CaptureAction))
		return errors.WithStack(err)
	}
	return nil
}

func (p *pagerduty) GetOncallCount(ctx context.Context, organization string) (int, error) {
	t := time.Now().AddDate(0, 0, int(time.Sunday-time.Now().Weekday()))
	since := t.AddDate(0, 0, -7).Format("2006-01-02")
	until := t.Format("2006-01-02")
	var url = "https://" + organization + ".pagerduty.com/reports#team?since=" + since +
		"T00%3A00%3A00&until=" + until +
		"T00%3A00%3A00&filters[escalation_policy_ids][]=P45K6JU&grouped_by[]=service&grouped_by[]=daily&time_zone=Asia%2FTokyo"

	const sel = "#reports_index_page > div > div > div > div > div.pd-row.x-main-report > div > div:nth-child(3) > div.pd-col.pd-col-2 > div > div.report-comparison-widget > div > div > div.pd-widget-content.pd-widget-padded > div:nth-child(1) > p > span"

	var count string
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctxt context.Context) error {
		ctxt, cancel := context.WithTimeout(ctxt, 10*time.Second)
		defer cancel()
		tasks := chromedp.Tasks{
			chromedp.Navigate(url),
			chromedp.Text(sel, &count, chromedp.NodeVisible),
		}
		if err := tasks.Do(ctxt); err != nil {
			return errors.WithStack(err)
		}
		return nil
	})); err != nil {
		p.logger.Printf("%+v", chromedp.Run(ctx, samples.CaptureAction))
		return 0, errors.WithStack(err)
	}
	val, _ := strconv.Atoi(count)
	return val, nil
}
