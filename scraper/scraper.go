package scraper

import "github.com/gofiber/fiber/v2"

//var appartments []Appartment

type Scraper interface {
	ScrapelLinks(*fiber.Ctx)
	ScrapelBody(*fiber.Ctx)
}
