package main

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"

	"github.com/li-go/chromedp-samples/samples"
)

func main() {
	ctx, cancel := chromedp.NewContext(context.Background(), chromedp.WithLogf(log.Printf))
	defer cancel()

	if err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate("https://www.google.com/"),
		chromedp.Sleep(time.Second),
		samples.CaptureAction,
	}); err != nil {
		log.Fatal(err)
	}
}
