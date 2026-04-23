package scraper

import (
	"context"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type Result struct {
	URL         string
	Title       string
	Description string
	H1s         []string
	H2s         []string
	Links       []string
	BodyText    string
	HasSSL      bool
	ImageCount  int
}

func Scrape(targetURL string) (*Result, error) {
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
		)...,
	)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var title, bodyText, metaDesc string
	var h1s, h2s, links []string
	var imageCount int

	err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body"),
		chromedp.Title(&title),
		chromedp.InnerHTML("body", &bodyText, chromedp.ByQuery),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('h1')).map(e => e.innerText.trim()).filter(Boolean)`, &h1s),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('h2')).map(e => e.innerText.trim()).filter(Boolean)`, &h2s),
		chromedp.Evaluate(`document.querySelector('meta[name="description"]')?.getAttribute('content') ?? ""`, &metaDesc),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a[href]')).map(e => e.href)`, &links),
		chromedp.Evaluate(`document.querySelectorAll('img').length`, &imageCount),
	)
	if err != nil {
		return nil, err
	}

	// Strip HTML tags from body and truncate to avoid blowing the token budget
	bodyText = stripTags(bodyText)
	if len(bodyText) > 4000 {
		bodyText = bodyText[:4000]
	}

	return &Result{
		URL:         targetURL,
		Title:       title,
		Description: metaDesc,
		H1s:         h1s,
		H2s:         h2s,
		Links:       links,
		BodyText:    bodyText,
		HasSSL:      strings.HasPrefix(targetURL, "https://"),
		ImageCount:  imageCount,
	}, nil
}

func stripTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
			b.WriteRune(' ')
		case !inTag:
			b.WriteRune(r)
		}
	}
	// Collapse whitespace
	return strings.Join(strings.Fields(b.String()), " ")
}
