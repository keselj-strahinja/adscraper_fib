package scraper

import (
	"context"
	"fmt"

	"log"
	"strconv"
	"strings"

	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly"
	"github.com/gofiber/fiber/v2"
	"github.com/keselj-strahinja/halo_scraper/db"
	models "github.com/keselj-strahinja/halo_scraper/type_models"
)

type HaloScraper struct {
	store db.HaloStore
}

func NewHaloScraper(haloStore db.HaloStore) *HaloScraper {
	return &HaloScraper{
		store: haloStore,
	}
}

func (h *HaloScraper) ScrapelLinks(fctx *fiber.Ctx) error {

	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.MaxDepth(2),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 10})

	baseUrl := "https://www.halooglasi.com/nekretnine/izdavanje-stanova/beograd"
	numPages := 250

	c.OnHTML("div.product-item", func(e *colly.HTMLElement) {
		apartment := models.Apartment{}

		apartment.Name = e.ChildText("h3.product-title a")
		apartment.PostingDate = e.ChildText("span.publish-date")
		apartment.URL = e.Request.AbsoluteURL(e.ChildAttr("h3.product-title a", "href"))

		price := e.ChildAttr("div.central-feature-wrapper div.central-feature span", "data-value")
		priceStr := strings.ReplaceAll(strings.ReplaceAll(price, ".", ""), ",", "")
		priceConv, err := strconv.ParseInt(strings.TrimSpace(priceStr), 10, 64)
		if err != nil {
			log.Printf("Error while parsing price: %v", err)
		} else {
			apartment.Price = priceConv
		}

		apartment.ImageURL = e.ChildAttr("figure.pi-img-wrapper a img", "src")

		e.ForEach("ul.subtitle-places li", func(_ int, el *colly.HTMLElement) {
			apartment.Location += el.Text + ", "
		})

		e.ForEach("ul.product-features li", func(_ int, el *colly.HTMLElement) {
			feature := el.ChildText("span.legend")
			value := strings.TrimSpace(el.ChildText("div.value-wrapper"))

			if feature == "Kvadratura" {
				// Removing "m2Kvadratura" from the end and parsing to int64
				sizeStr := strings.TrimSuffix(strings.TrimSpace(value), "\u00a0m2Kvadratura")
				size, err := strconv.ParseInt(sizeStr, 10, 64)
				if err != nil {
					log.Printf("Error while parsing size: %v", err)
				} else {
					apartment.Size = size
				}
			} else if feature == "Broj soba" {
				// Removing "\u00a0Broj soba" from the end and parsing to float64
				roomsStr := strings.TrimSuffix(strings.TrimSpace(value), "\u00a0Broj soba")
				rooms, err := strconv.ParseFloat(roomsStr, 64)
				if err != nil {
					log.Printf("Error while parsing rooms: %v", err)
				} else {
					apartment.Rooms = rooms
				}
			} else if feature == "Spratnost" {
				apartment.Floor = strings.TrimSpace(value)
			}
		})

		apartment.Agency = e.ChildText("span.basic-info span")
		apartment.Description = e.ChildText("p.text-description-list.product-description.short-desc")

		fmt.Printf("%+v\n", apartment)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished Scraping", r.Request.URL)
	})

	c.OnResponse(func(r *colly.Response) {

	})

	// Visit the website
	err := c.Visit(baseUrl)

	if err != nil {
		panic(err)
	}

	for i := 1; i <= numPages; i++ {
		url := fmt.Sprintf("%s?page=%d", baseUrl, i)
		err := c.Visit(url)
		if err != nil {
			log.Println("Error visiting page:", err)
			// decide whether to break out of the loop if there's an error
		}
		c.Wait() // wait for all responses before proceeding
	}

	// Wait until threads are finished
	c.Wait()
	return nil
}

func (h *HaloScraper) ScrapeBody() {
	var html string
	baseUrl := "https://www.halooglasi.com/nekretnine/izdavanje-stanova/izdajem-stan-sakura-park/5425643297480?sid=1687534803853"
	opts := append(
		// select all the elements after the third element
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	)
	// create chromedp's context
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	//ctx, cancel := chromedp.NewContext(parentCtx)
	err := chromedp.Run(ctx,
		chromedp.Navigate(baseUrl),
		chromedp.WaitVisible(`body`, chromedp.ByQuery), // adjust this to something on your page that you know will be there once it's fully loaded
		chromedp.OuterHTML(`html`, &html, chromedp.ByQuery),
	)

	if err != nil {
		panic(err)
	}

	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.MaxDepth(2),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 10})

	c.OnHTML("div.product-page", func(e *colly.HTMLElement) {

		fmt.Println("karina")

		apartment := models.Apartment{}

		apartment.Name = e.ChildText("#plh11")
		fmt.Println(apartment.Name)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished Scraping", r.Request.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		r.Body = []byte(html)
	})

	// Visit the website
	err = c.Visit(baseUrl)

	if err != nil {
		panic(err)
	}

}
