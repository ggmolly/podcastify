package utils

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func SendModal(c *fiber.Ctx, name, title string, bind fiber.Map) error {
	bindMap := fiber.Map{
		"ModalID":    strings.ToLower(name) + "-modal",
		"ModalTitle": title,
	}
	for k, v := range bind {
		bindMap[k] = v
	}
	c.Response().Header.Set("HX-Retarget", "#modal-container")
	c.Response().Header.Set("HX-Reswap", "innerhtml")
	return c.Render("partials/modals/"+name, bindMap, "layouts/modal")
}
