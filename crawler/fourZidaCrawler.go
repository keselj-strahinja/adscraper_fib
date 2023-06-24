package scraper

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly"
	"github.com/gofiber/fiber/v2"
	models "github.com/keselj-strahinja/halo_scraper/type_models"
)

type FourZidaCrawler struct {
}

func (f *FourZidaCrawler) ScrapelLinks(fctx *fiber.Ctx) error {
	var html string

	baseURL := "https://www.4zida.rs/izdavanje-stanova/beograd?sortiranje=najnoviji"
	//baseURL := "https://www.4zida.rs/izdavanje-stanova/blok-37-novi-beograd-beograd/dvosoban-stan/6495e5e3f3f0f157990746c8"
	opts := append(
		// select all the elements after the third element
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	)
	// create chromedp's context
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	// defer cancel()

	// ctx, cancel := chromedp.NewContext(parentCtx)
	// defer cancel()
	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.WaitVisible(`body`, chromedp.ByQuery), // adjust this to something on your page that you know will be there once it's fully loaded
		chromedp.OuterHTML(`html`, &html, chromedp.ByQuery),
	)

	if err != nil {
		panic(err)
	}

	cancel()

	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.MaxDepth(2),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 10})

	c.OnHTML("div.ed-card-details a", func(e *colly.HTMLElement) {
		fmt.Println("penis")
		apartment := models.Apartment{}

		apartment.URL = e.Request.AbsoluteURL(e.Attr("href"))

		fmt.Println(apartment.URL)
		// fmt.Printf("%v", apartment)
		// apartment.PostingDate = e.ChildText("span.publish-date")
		// apartment.URL = e.Request.AbsoluteURL(e.ChildAttr("h3.product-title a", "href"))
		// apartment := Apartment{}

		// apartment.Name = e.ChildText("span.mb-2")
		// fmt.Println(apartment.Name)
		// apartment.Rooms, err = strconv.ParseFloat(e.ChildText("strong.text-base"), 64)
		// apartment.URL = e.Request.AbsoluteURL(e.ChildAttr("h3.product-title a", "href"))
	})

	c.OnResponse(func(r *colly.Response) {
		r.Body = []byte(html)
		//fmt.Println(string(r.Body))
	})

	c.OnRequest(func(r *colly.Request) {
		time.Sleep(time.Second * 5)
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished Scraping", r.Request.URL)
	})

	c.SetRequestTimeout(30 * time.Second) // Increase timeout to 30 seconds

	err = c.Visit(baseURL)
	if err != nil {
		return err
	}

	return nil
}

func (f *FourZidaCrawler) ScrapeBody() {

}
