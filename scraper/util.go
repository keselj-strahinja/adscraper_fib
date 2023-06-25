package scraper

import (
	"context"
	"log"
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

func GetLastPage() int {
	ctx, cancel := CreateChromedpInstance()
	defer cancel()

	// The URL to visit
	url := "https://www.halooglasi.com/nekretnine/izdavanje-stanova/beograd"

	var result string
	actions := []chromedp.Action{
		chromedp.WaitVisible(`div.light-theme.simple-pagination`),
		chromedp.Text(`div.light-theme.simple-pagination li.disabled + li a.page-link`, &result, chromedp.ByQuery),
	}

	if err := FetchDataFromPage(ctx, url, actions...); err != nil {
		logger.Errorf("error fetching data from page %s", err.Error())

	}
	// Clean up the result to get the page number
	result = strings.TrimSpace(result)
	lastPage, err := strconv.Atoi(result)

	if err != nil {
		log.Fatal(err)
	}

	return lastPage
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
	timeoutCtx, cancel := context.WithTimeout(ctx, 60*time.Second) // Adjust the time as necessary
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
