package scraper

//var appartments []Appartment

type Scraper interface {
	ScrapelLinks()
	ScrapelBody()
}
