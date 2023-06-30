package scraper

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gocolly/colly"
	"github.com/gofiber/fiber/v2"
	models "github.com/keselj-strahinja/halo_scraper/type_models"
	"go.mongodb.org/mongo-driver/bson"
)

func (h *HaloScraper) ScrapeLinks(fctx *fiber.Ctx) error {
	logger.Info("Setting all apartments to inactive")

	// set all store s to inactive
	h.store.UpdateProperty(
		context.Background(), 
		bson.M{}, 
		bson.M{"$set": bson.M{"active": false}})

	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.MaxDepth(2),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 10, RandomDelay: 5 * time.Second})

	//numPages := h.getLastPage()
	numPages := 0

	c.OnHTML("div.product-item", func(e *colly.HTMLElement) {
		URL := e.Request.AbsoluteURL(e.ChildAttr("h3.product-title a", "href"))

		exists, err := h.store.URLExists(context.Background(), URL)
		if err != nil {
			logger.WithField("url", URL).Errorf("Failed to check if URL exists: %v", err)
		}
		if exists {
			// set as listing active
			h.store.UpdateProperty(
				context.Background(),
				bson.M{"url": URL},
				bson.M{"$set": bson.M{"active": true}},
			)
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
	err := c.Visit(h.baseUrl)

	if err != nil {
		logger.WithField("url", h.baseUrl).Errorf("Error visiting base URL: %v", err)

		return nil
	}

	for i := 1; i <= numPages; i++ {
		url := fmt.Sprintf("%s?page=%d", h.baseUrl, i)
		err := c.Visit(url)
		if err != nil {
			log.Println("Error visiting page:", err)
			// decide whether to break out of the loop if there's an error
		}
	}

	// Wait until threads are finished
	c.Wait()
	
	return fctx.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Scraping links finished",
	})
}