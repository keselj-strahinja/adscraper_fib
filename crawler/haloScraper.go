package scraper

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"log"

	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly"
	"github.com/gofiber/fiber/v2"
	"github.com/keselj-strahinja/halo_scraper/db"
	models "github.com/keselj-strahinja/halo_scraper/type_models"
	"github.com/sirupsen/logrus"
)

const numWorkers = 10

var logger = logrus.New()

type HaloScraper struct {
	store db.HaloStore
	wg    sync.WaitGroup
}

func NewHaloScraper(haloStore db.HaloStore) *HaloScraper {
	return &HaloScraper{
		store: haloStore,
	}
}
func getLastPage() int {
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
func (h *HaloScraper) ScrapelLinks(fctx *fiber.Ctx) error {
	logger.Info("Setting all apartments to inactive")

	h.store.SetAllInactive(context.Background())

	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.MaxDepth(2),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 10, RandomDelay: 5 * time.Second})

	baseUrl := "https://www.halooglasi.com/nekretnine/izdavanje-stanova/beograd"
	numPages := getLastPage()

	c.OnHTML("div.product-item", func(e *colly.HTMLElement) {
		URL := e.Request.AbsoluteURL(e.ChildAttr("h3.product-title a", "href"))

		exists, err := h.store.URLExists(context.Background(), URL)
		if err != nil {
			logger.WithField("url", URL).Errorf("Failed to check if URL exists: %v", err)
		}
		if exists {
			h.store.SetActive(context.Background(), URL)
			return
		}

		apartment := models.Apartment{
			URL:     URL,
			Active:  true,
			Scraped: false,
		}

		if _, err := h.store.InsertListing(context.Background(), &apartment); err != nil {
			logger.WithField("url", URL).Errorf("Failed to insert listing: %v", err)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		logger.WithField("url", r.URL.String()).Info("Visiting")
	})

	c.OnError(func(_ *colly.Response, err error) {
		logger.Error("Something went wrong:", err)
	})

	c.OnScraped(func(r *colly.Response) {
		logger.WithField("url", r.Request.URL.String()).Info("Finished Scraping")
	})

	c.OnResponse(func(r *colly.Response) {

	})

	// Visit the website
	err := c.Visit(baseUrl)

	if err != nil {
		logger.WithField("url", baseUrl).Errorf("Error visiting base URL: %v", err)

		return nil
	}

	for i := 1; i <= numPages; i++ {
		url := fmt.Sprintf("%s?page=%d", baseUrl, i)
		err := c.Visit(url)
		if err != nil {
			log.Println("Error visiting page:", err)
			// decide whether to break out of the loop if there's an error
		}
	}

	// Wait until threads are finished
	c.Wait()

	h.ScrapeBody(fctx)
	return nil
}

func (h *HaloScraper) ScrapeBody(fctx *fiber.Ctx) error {

	// Create buffered channel with a maximum capacity of numWorkers
	jobs := make(chan string, numWorkers)

	// Add the number of URLs to the wait group
	urls, err := h.store.GetUnscrapedURLs(context.Background())

	if err != nil {
		return err
	}
	logger.WithField("UNSCRAPED URLS COUND", len(urls)).Info("Sending to workers")
	h.wg.Add(len(urls))

	// Start the workers
	for i := 0; i < 10; i++ {
		go func(workerID int) {
			logger.WithField("worker_id", workerID).Info("Starting worker")
			h.worker(jobs, fctx)
			logger.WithField("worker_id", workerID).Info("Finished worker")
		}(i)
	}

	// Send the jobs (URLs to be scraped) to the workers
	for _, url := range urls {
		jobs <- url
	}

	// Close the jobs channel
	close(jobs)

	// Wait until all the workers have finished
	h.wg.Wait()

	return nil
}

func (h *HaloScraper) worker(jobs <-chan string, fctx *fiber.Ctx) {
	for url := range jobs {

		logger.WithField("url", url).Info("Starting scraping for URL")
		delay := time.Duration(5+rand.Intn(6)) * time.Second
		time.Sleep(delay)

		err := h.scrapeSinglePage(url, fctx)
		if err != nil {
			logger.WithField("url", url).Errorf("Error scraping page: %v", err)

			h.store.SetScraped(context.Background(), url, false)

			continue
		}

		// Mark the page as scraped in the database
		err = h.store.SetScraped(context.Background(), url, true)
		if err != nil {
			logger.WithField("url", url).Errorf("Error marking page as scraped: %v", err)
		}

		logger.WithField("url", url).Info("Finished scraping for URL")

		defer h.wg.Done()

	}
}

func (h *HaloScraper) scrapeSinglePage(url string, fctx *fiber.Ctx) error {
	logger.WithField("url", url).Info("Starting to scrape page")
	const (
		clickButtonSelector = `.show-phone-numbers`
		phoneSelector       = `a[href^="tel:"]`
		publishSelector     = `strong#plh81`
		contactSelector     = `div#plh65`
		htmlSelector        = `html`
		body                = `body`
	)
	var (
		html              string
		phoneNumberString string
		postingDateString string
		contactName       string
	)
	ctx, cancel := CreateChromedpInstance()
	defer cancel()
	// TODO maybe make it so the fetch data func gets the selector and action with which it auto creates
	// the chromedp run
	waitBodyAction := chromedp.WaitVisible(body, chromedp.ByQuery)
	clickButtonAction := chromedp.Click(clickButtonSelector)
	phoneNumberWaitAction := chromedp.WaitVisible(phoneSelector, chromedp.BySearch)
	phoneNumberAction := chromedp.TextContent(phoneSelector, &phoneNumberString, chromedp.AtLeast(0))
	postingDateAction := chromedp.Text(publishSelector, &postingDateString, chromedp.NodeVisible, chromedp.ByQuery)
	contactNameAction := chromedp.Text(contactSelector, &contactName, chromedp.NodeVisible, chromedp.ByQuery)
	outerHtmlAction := chromedp.OuterHTML(htmlSelector, &html, chromedp.ByQuery)

	actions := []chromedp.Action{
		waitBodyAction,
		clickButtonAction,
		phoneNumberWaitAction,
		phoneNumberAction,
		postingDateAction,
		contactNameAction,
		outerHtmlAction,
	}

	if err := FetchDataFromPage(ctx, url, actions...); err != nil {
		logger.Errorf("error fetching data from page %s url: %s", err.Error(), url)
		return err

	}
	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.MaxDepth(2),
	)

	apartment, err := h.store.GetApartmant(context.Background(), url)
	if err != nil {
		logger.Errorf("there was an error while trying to get the Apt from the database %s", err.Error())
	}
	t, err := time.Parse("02.01.2006. u 15:04", postingDateString)

	if err != nil {
		logger.Errorf("there was an error while trying to parse the date %s", err.Error())

	}
	unixTimestamp := t.Unix()

	apartment.PostingDate = unixTimestamp

	apartment.Phone = phoneNumberString

	apartment.ContactName = contactName

	c.OnHTML("div.product-page", func(e *colly.HTMLElement) {
		var (
			labels1      = []string{}
			labels2      = []string{}
			descriptions = []string{}
			images       = []string{}
		)

		if err != nil {
			log.Fatalf("Failed to get apt: %v", err)
		}

		apartment.Name = e.ChildText("#plh1")
		apartment.City = e.ChildText("#plh2")
		apartment.Location = e.ChildText("#plh3")
		apartment.Microlocation = e.ChildText("#plh4")
		apartment.Street = e.ChildText("#plh5")
		apartment.RealEstateType = e.ChildText("#plh10")

		squareSize := e.ChildText("#plh11")
		parts := strings.Split(squareSize, " ")
		if len(parts) > 0 {
			value, err := strconv.ParseInt(parts[0], 10, 64)
			if err == nil {
				apartment.SquareSize = value
			} else {
				fmt.Println("Failed to convert to integer:", err)
			}
		}

		rooms, err := strconv.ParseFloat(e.ChildText("#plh12"), 64)
		if err != nil {
			log.Printf("Failed to parse rooms: %v", err)
		}
		apartment.Rooms = rooms
		apartment.Owner = e.ChildText("#plh13")
		apartment.BuildingType = e.ChildText("#plh14")
		apartment.State = e.ChildText("#plh15")
		apartment.Furnishing = e.ChildText("#plh16")
		apartment.Heating = e.ChildText("#plh17")
		apartment.Floor = e.ChildText("#plh18")

		buildingFloors, err := strconv.ParseInt(e.ChildText("#plh19"), 10, 64)
		if err != nil {
			log.Printf("Failed to parse objectFloors: %v", err)
		}
		apartment.BuildingFloors = buildingFloors

		monthlyBills, err := GetOnlyDigits(e.ChildText("#plh20"))
		if err != nil {
			log.Printf("Failed to parse monthly bills: %v", err)
		}

		apartment.MonthlyBills = monthlyBills

		apartment.PaymentType = e.ChildText("#plh21")

		e.ForEach("div#tabTopHeader1 span.flag-attribute", func(_ int, el *colly.HTMLElement) {
			label := el.ChildText("label")
			labels1 = append(labels1, label)
		})

		e.ForEach("div#tabTopHeader2 span.flag-attribute", func(_ int, el *colly.HTMLElement) {
			label := el.ChildText("label")
			labels2 = append(labels2, label)
		})

		apartment.AdditionalPerks = labels1

		apartment.OtherPerks = labels2

		e.ForEach("div#tabTopHeader3 span#plh51", func(_ int, el *colly.HTMLElement) {
			el.ForEach("p", func(_ int, p *colly.HTMLElement) {
				descriptions = append(descriptions, p.Text)
			})
			if len(descriptions) == 0 {
				descriptions = append(descriptions, el.Text)
			}
		})

		apartment.Description = descriptions

		priceValue := e.ChildText("span#plh6 .offer-price-value")
		priceValueNumber, err := GetOnlyDigits(priceValue)

		if err != nil {
			log.Printf("error during conversion of price")
		}
		apartment.PriceValue = priceValueNumber
		priceUnit := e.ChildText("span#plh6 .offer-price-unit")
		apartment.PriceUnit = priceUnit

		e.ForEach(".fotorama__nav__frame", func(_ int, el *colly.HTMLElement) {
			image := el.ChildAttr("img", "src")
			images = append(images, image)
		})
		apartment.ImageURL = images

		h.store.UpdateListing(context.Background(), apartment)

	})
	c.OnRequest(func(r *colly.Request) {
		logger.WithField("url", r.URL.String()).Info("Starting a new request")
	})
	c.OnError(func(_ *colly.Response, err error) {
		logger.Error("Something went wrong:", err)
	})

	c.OnScraped(func(r *colly.Response) {

	})

	c.OnResponse(func(r *colly.Response) {
		r.Body = []byte(html)
	})

	// Visit the website
	err = c.Visit(url)

	if err != nil {
		logger.WithField("url", url).Errorf("Error visiting page: %v", err)
	}

	return nil
}
