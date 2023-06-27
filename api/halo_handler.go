package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/keselj-strahinja/halo_scraper/db"
	"go.mongodb.org/mongo-driver/bson"
)

type HaloHandler struct {
	store db.HaloStore
}

func NewHaloHandler(haloStore db.HaloStore) *HaloHandler {
	return &HaloHandler{
		store: haloStore,
	}
}

func (h *HaloHandler) HandleGetActiveListings(c *fiber.Ctx) error {
	listings, err := h.store.GetActiveListings(c.Context(), bson.M{
		"active": true,
	})

	if err != nil {
		return err
	}

	return c.JSON(listings)

}
