package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/keselj-strahinja/halo_scraper/api"
	"github.com/keselj-strahinja/halo_scraper/db"
	scraper "github.com/keselj-strahinja/halo_scraper/scraper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var config = fiber.Config{
	ErrorHandler: nil,
}

func main() {

	mongoEndpoint := os.Getenv("MONGO_DB_URL")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoEndpoint))
	if err != nil {
		log.Fatal(err)
	}
	var (
		app         = fiber.New(config)
		apiv1       = app.Group("/api/v1")
		haloStore   = db.NewMongoHaloStore(client)
		haloHandler = api.NewHaloHandler(haloStore)
		haloScraper = scraper.NewHaloScraper(haloStore)
	)

	apiv1.Get("/halo", haloHandler.HandlePostUser)
	apiv1.Get("/scrapeHaloLinks", haloScraper.ScrapelLinks)
	listenAddr := os.Getenv("HTTP_LISTEN_ADDRESS")

	log.Println("Starting server on", listenAddr)
	app.Listen(listenAddr)

}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
}
