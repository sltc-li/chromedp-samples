package main

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"

	"github.com/li-go/chromedp-samples/samples"
)

func main() {
	ctx, cancel := chromedp.NewContext(context.Background(), chromedp.WithLogf(log.Printf))
	defer cancel()

	if err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate("https://github.com/li-go"),
		samples.CaptureAction,
	}); err != nil {
		log.Fatal(err)
	}
}
