package scraper

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/chromedp/chromedp"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	dockerUrl = "wss://localhost:9222"
)

func FetchDataFromPage(ctx context.Context, url string, tasks ...chromedp.Action) error {
	tasks = append([]chromedp.Action{chromedp.Navigate(url)}, tasks...)
	return chromedp.Run(ctx, tasks...)
}

func(h *HaloScraper) getLastPage() int {
	ctx, cancel := CreateChromedpInstance()
	defer cancel()

	// The URL to visit

	var result string
	actions := []chromedp.Action{
		chromedp.WaitVisible(`div.light-theme.simple-pagination`),
		chromedp.Text(`div.light-theme.simple-pagination li.disabled + li a.page-link`, &result, chromedp.ByQuery),
	}

	if err := FetchDataFromPage(ctx, h.baseUrl, actions...); err != nil {
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
	// opts := append(
	// 	// select all the elements after the third element
	// 	chromedp.DefaultExecAllocatorOptions[:],
	// 	chromedp.NoFirstRun,
	// 	chromedp.NoDefaultBrowserCheck,
	// )
	// create chromedp's context
	//parentCtx, cancelParent := chromedp.NewExecAllocator(context.Background(), opts...)
	parentCtx, cancelParent := chromedp.NewRemoteAllocator(context.Background(), dockerUrl)
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


func (h *HaloScraper) worker(jobs <-chan string, fctx *fiber.Ctx) {
	for url := range jobs {
		defer h.wg.Done()
		logger.WithField("url", url).Info("Starting scraping for URL")
		delay := time.Duration(5+rand.Intn(6)) * time.Second
		time.Sleep(delay)

		err := h.scrapeSinglePage(url, fctx)
		if err != nil {
			logger.WithField("url", url).Errorf("Error scraping page: %v", err)
			logger.Infof("setting url to unavailable. url: %s", url)
			// TODO this could be better, error handling could be better. Shit works at least.
			if strings.Contains(err.Error(), "wsURL"){
				logger.WithField("url", url).Errorf("Error scraping page: %v", err)
				continue
			}
			// set listing to unavailable
			if err := h.store.UpdateProperty(
				context.Background(),
				bson.M{"url": url},
				bson.M{"$set": bson.M{"unavailable": true}},
			); err != nil {
				logger.Error("Failed to set listing as unavailable: ", err)
			}
			continue
		}

		// Mark the page as scraped in the database
		err = h.store.UpdateProperty(
			context.Background(),
			bson.M{"url": url},
			bson.M{"$set": bson.M{"scraped": true}},
		)
		if err != nil {
		//TODO Decide what to do based on the error
		// e.g., you might want to stop the worker if the database is unreachable
		// or you might want to skip this URL but continue with others if it's a problem with just this URL
		// For now, just continue to the next URL
		continue
		}

		logger.WithField("url", url).Info("Finished scraping for URL")
	}
}
