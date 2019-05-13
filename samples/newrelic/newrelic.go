package newrelic

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"

	"github.com/liyy7/chromedp-samples/samples"
)

type newrelic struct {
	logger *log.Logger
}

func New(logger *log.Logger) *newrelic {
	return &newrelic{logger: logger}
}

func (n *newrelic) Login(ctx context.Context, email, password string) error {
	cs := samples.CookieSaver{Name: "newrelic", Logger: n.logger}
	if cs.RestoreCookies(ctx, "https://rpm.newrelic.com/applications") {
		return nil
	}

	n.logger.Println("Login to newrelic.com")
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctxt context.Context) error {
		ctxt, cancel := context.WithTimeout(ctxt, 5*time.Second)
		defer cancel()
		tasks := chromedp.Tasks{
			chromedp.Navigate("https://login.newrelic.com/login"),
			chromedp.SendKeys("#login_email", email, chromedp.NodeVisible),
			chromedp.SendKeys("#login_password", password),
			chromedp.Click("#login_submit"),
			chromedp.Sleep(time.Second),
			chromedp.ActionFunc(func(ctxt context.Context) error {
				cookies, err := network.GetAllCookies().Do(ctxt)
				if err != nil {
					return errors.WithStack(err)
				}
				n.logger.Println("Save cookies")
				if err := cs.SaveCookies(cookies); err != nil {
					return errors.WithStack(err)
				}
				return nil
			}),
		}
		return errors.WithStack(tasks.Do(ctxt))
	})); err != nil {
		n.logger.Printf("%+v", chromedp.Run(ctx, samples.CaptureAction))
		return errors.WithStack(err)
	}
	return nil
}

func (n *newrelic) GetErrorRate(ctx context.Context) (float64, error) {
	const setTimeWindowURL = "https://rpm.newrelic.com/set_time_window?back=https%3A%2F%2Frpm.newrelic.com%2Faccounts%2F459517%2Fapplications%2F11120465&tw%5Bfrom_local%5D=true&tw%5Bdur%5D=last_7_days"

	var errorRate string
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctxt context.Context) error {
		ctxt, cancel := context.WithTimeout(ctxt, 10*time.Second)
		defer cancel()
		tasks := chromedp.Tasks{
			chromedp.Navigate(setTimeWindowURL),
			chromedp.Text("#hosts_content_container > section > h3 > div > p", &errorRate, chromedp.NodeVisible),
		}
		return errors.WithStack(tasks.Do(ctxt))
	})); err != nil {
		n.logger.Printf("%+v", chromedp.Run(ctx, samples.CaptureAction))
		return 0, errors.WithStack(err)
	}
	val, _ := strconv.ParseFloat(strings.TrimRight(errorRate, " %"), 64)
	return val, nil
}

type AppPerformance struct {
	AppResponse   int
	AppHistogram  int
	AppPercentile int
}

func (n *newrelic) GetAppPerformance(ctx context.Context, transactionID string) (*AppPerformance, error) {
	var (
		setTimeWindowURL = "https://rpm.newrelic.com/set_time_window?back=https%3A%2F%2Frpm.newrelic.com%2Faccounts%2F459517%2Fapplications%2F11120465%2Ftransactions%23id%3D" +
			transactionID + "&tw%5Bfrom_local%5D=true&tw%5Bdur%5D=last_7_days"
	)

	var appPerformance AppPerformance
	if err := chromedp.Run(ctx,
		n.fapPrepareActions(setTimeWindowURL),
		n.fapAppResponseActions(transactionID, &appPerformance.AppResponse),
		n.fapAppHistogramActions(transactionID, &appPerformance.AppHistogram),
		n.fapAppPercentileActions(transactionID, &appPerformance.AppPercentile),
	); err != nil {
		n.logger.Printf("error occurred, last result: %+v", appPerformance)
		n.logger.Printf("%+v", chromedp.Run(ctx, samples.CaptureAction))
		return nil, errors.WithStack(err)
	}
	return &appPerformance, nil
}

func (n *newrelic) fapPrepareActions(setTimeWindowURL string) chromedp.Tasks {
	const (
		applicationURL       = "https://rpm.newrelic.com/accounts/459517/applications/11120465"
		saveSelectedChartsJS = `$.post("` + applicationURL + `/save_selected_charts", {selected: "app_response"})`
	)

	return chromedp.Tasks{chromedp.Navigate(applicationURL),
		chromedp.ActionFunc(func(ctxt context.Context) error {
			ctxt, cancel := context.WithTimeout(ctxt, 5*time.Second)
			defer cancel()
			var loaded bool
			action := chromedp.Evaluate("(typeof $ !== 'undefined' && typeof $.post === 'function')", &loaded)
			for {
				if err := action.Do(ctxt); err != nil {
					return errors.WithStack(err)
				}
				if loaded {
					return nil
				}
				time.Sleep(time.Second)
			}
		}),
		chromedp.Evaluate(saveSelectedChartsJS, &struct{}{}),
		chromedp.Navigate(setTimeWindowURL),
	}
}
func (n *newrelic) fapAppResponseActions(transactionID string, appResponse *int) chromedp.Action {
	return chromedp.ActionFunc(func(ctxt context.Context) error {
		ctxt, cancel := context.WithTimeout(ctxt, 10*time.Second)
		defer cancel()
		var text string
		if err := chromedp.Text("#tab_content_app_server_performance_breakdown > div.selectable_chart.response_chart > h2 > div:nth-child(1) > p.value",
			&text, chromedp.NodeVisible).Do(ctxt); err != nil {
			return errors.WithStack(err)
		}
		var err error
		*appResponse, err = strconv.Atoi(strings.TrimRight(text, " ms"))
		if err != nil {
			n.logger.Printf("Cannot extract app response: %+v", err)
		}
		return nil
	})
}

func (n *newrelic) fapAppHistogramActions(transactionID string, appHistogram *int) chromedp.Action {
	return chromedp.ActionFunc(func(ctxt context.Context) error {
		ctxt, cancel := context.WithTimeout(ctxt, 10*time.Second)
		defer cancel()
		tasks := chromedp.Tasks{
			chromedp.Click("#tab_content_app_server_performance_breakdown > div.selectable_chart.response_chart > h2 > ul > li > ul > li.app_histogram"),
			chromedp.ActionFunc(func(ctxt context.Context) error {
				gSel := "#current_app_transaction_histogram_" + transactionID + " g.highcharts-tracker > g"
				var rectNodes []*cdp.Node
				if err := chromedp.Run(ctxt,
					chromedp.WaitReady(gSel),
					chromedp.Nodes(gSel+" > rect", &rectNodes),
				); err != nil {
					return errors.WithStack(err)
				}
				var maxIdx, maxHeight int
				for i, node := range rectNodes {
					height, _ := strconv.Atoi(node.AttributeValue("height"))
					if height > maxHeight {
						maxIdx, maxHeight = i, height
					}
				}
				low := 25 * (len(rectNodes) - maxIdx - 1)
				*appHistogram = low + 25/2
				return nil
			}),
		}
		return errors.WithStack(tasks.Do(ctxt))
	})
}

func (n *newrelic) fapAppPercentileActions(transactionID string, appPercentile *int) chromedp.Action {
	return chromedp.ActionFunc(func(ctxt context.Context) error {
		ctxt, cancel := context.WithTimeout(ctxt, 10*time.Second)
		defer cancel()
		action := chromedp.ActionFunc(func(ctxt context.Context) error {
			var textNodes []*cdp.Node
			var text string
			var paths []*cdp.Node
			sel := "#app_server_percentiles_" + transactionID + " g.highcharts-axis-labels > text[x='0']"
			tasks := chromedp.Tasks{
				chromedp.WaitReady(sel),
				chromedp.Nodes(sel, &textNodes),
				chromedp.Evaluate(fmt.Sprintf(`$("%s").text()`, sel), &text),
				chromedp.Nodes("#app_server_percentiles_"+transactionID+" g.highcharts-tracker > g > path", &paths, chromedp.NodeReady),
			}
			if err := tasks.Do(ctxt); err != nil {
				return errors.WithStack(err)
			}
			type axisInfo struct {
				y   float64
				val int
			}
			axisInfos := make([]axisInfo, len(textNodes)-1)
			vals := strings.Split(text, " ms")[1:]
			for i, node := range textNodes[1:] {
				ys := node.AttributeValue("y")
				y, _ := strconv.ParseFloat(ys, 64)
				val, _ := strconv.Atoi(strings.Replace(vals[i], "k", "000", 1))
				axisInfos[i] = axisInfo{y: y, val: val}
			}
			// Calculate min Y
			d := strings.TrimSpace(paths[0].AttributeValue("d"))
			minY := float64(1000)
			for _, s := range strings.Split(d[strings.Index(d, " L "):], " L ")[1:] {
				y, _ := strconv.ParseFloat(strings.Split(s, " ")[1], 64)
				if y < minY {
					minY = y
				}
			}
			// Calculate app percentiles by min Y
			*appPercentile = axisInfos[0].val +
				int(float64(axisInfos[len(axisInfos)-1].val-axisInfos[0].val)/(axisInfos[0].y-axisInfos[len(axisInfos)-1].y)*(axisInfos[0].y-minY))
			return nil
		})
		tasks := chromedp.Tasks{
			chromedp.Click("#app_transaction_histogram > h2 > ul > li > ul > li.app_percentiles"),
			chromedp.ActionFunc(func(ctxt context.Context) error {
				for {
					if err := action.Do(ctxt); err != nil {
						return errors.WithStack(err)
					}
					if *appPercentile > 0 {
						return nil
					}
					time.Sleep(time.Second)
				}
			}),
		}
		return errors.WithStack(tasks.Do(ctxt))
	})
}
