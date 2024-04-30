package routes

import (
	"github.com/ggmolly/podcastify/utils"
	"github.com/gofiber/fiber/v2"
)

// GET /
func Index(c *fiber.Ctx) error {
	return c.Render("pages/index", fiber.Map{
		"page":       "index",
		"Annoyances": utils.Annoyances,
	}, "layouts/main")
}
