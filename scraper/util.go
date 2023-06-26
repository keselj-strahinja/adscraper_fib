package scraper

import (
	"context"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/chromedp/chromedp"
)

func FetchDataFromPage(ctx context.Context, url string, tasks ...chromedp.Action) error {
	tasks = append([]chromedp.Action{chromedp.Navigate(url)}, tasks...)
	return chromedp.Run(ctx, tasks...)
}

func CreateChromedpInstance() (context.Context, context.CancelFunc) {
	opts := append(
		// select all the elements after the third element
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	)
	// create chromedp's context
	parentCtx, cancelParent := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancelCtx := chromedp.NewContext(parentCtx)
	timeoutCtx, cancel := context.WithTimeout(ctx, 120*time.Second) // Adjust the time as necessary
	return timeoutCtx, func() {
		cancelCtx()
		cancelParent()
		cancel()
	}
}

func GetOnlyDigits(s string) (int64, error) {
	var result strings.Builder
	for _, rune := range s {
		if unicode.IsDigit(rune) {
			result.WriteRune(rune)
		}
	}
	return strconv.ParseInt(result.String(), 10, 64)
}
