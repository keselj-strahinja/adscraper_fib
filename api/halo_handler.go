package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/keselj-strahinja/halo_scraper/db"
)

type HaloHandler struct {
	store db.HaloStore
}

func NewHaloHandler(haloStore db.HaloStore) *HaloHandler {
	return &HaloHandler{
		store: haloStore,
	}
}

func (h *HaloHandler) HandlePostUser(c *fiber.Ctx) error {

	return nil

}
