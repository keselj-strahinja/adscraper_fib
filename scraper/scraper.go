package scraper

import "github.com/gofiber/fiber/v2"

type Scraper interface {
	ScrapeLinks(ctx *fiber.Ctx) error
    ScrapeBody(ctx *fiber.Ctx) error
}
