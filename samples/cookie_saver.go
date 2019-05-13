package samples

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
)

type CookieSaver struct {
	Name   string
	Logger *log.Logger
}

func (cs *CookieSaver) cookieFile() string {
	return "/tmp/" + cs.Name + "_cookies.json"
}

func (cs *CookieSaver) getCookies() ([]*network.Cookie, error) {
	buf, err := ioutil.ReadFile(cs.cookieFile())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var cookies []*network.Cookie
	if err := json.Unmarshal(buf, &cookies); err != nil {
		return nil, errors.WithStack(err)
	}
	return cookies, nil
}

func (cs *CookieSaver) RestoreCookies(ctx context.Context, checkURL string) bool {
	cookies, err := cs.getCookies()
	if err != nil {
		cs.Logger.Printf("Failed to restore cookies of %s: %+v", cs.Name, err)
		return false
	}
	if cookies == nil {
		return false
	}

	cs.Logger.Printf("Restore cookies of %s", cs.Name)
	var url string
	if err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.ActionFunc(func(ctxt context.Context) error {
			params := make([]*network.CookieParam, len(cookies))
			for i, c := range cookies {
				params[i] = &network.CookieParam{
					Name:   c.Name,
					Value:  c.Value,
					Domain: c.Domain,
				}
			}
			return network.SetCookies(params).Do(ctxt)
		}),
		chromedp.Navigate(checkURL),
		chromedp.Location(&url),
	}); err != nil {
		cs.Logger.Printf("%+v", err)
		return false
	}
	return url == checkURL
}

func (cs *CookieSaver) SaveCookies(ctx context.Context) error {
	cookies, err := network.GetAllCookies().Do(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	cs.Logger.Println("Save cookies")
	buf, err := json.Marshal(cookies)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := ioutil.WriteFile(cs.cookieFile(), buf, 0644); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
