package routes

import (
	"fmt"

	"github.com/ggmolly/podcastify/orm"
	"github.com/ggmolly/podcastify/shared"
	"github.com/gofiber/fiber/v2"
)

func Podcastify(c *fiber.Ctx) error {
	podcast, err := orm.PodcastFromRequest(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if existingPodcast := shared.GetStream(podcast.Name()); existingPodcast != nil {
		existingPodcast.Bump()
		c.Response().Header.Set("HX-Redirect", fmt.Sprintf("/podcast/%s", podcast.Name()))
	}

	// If the podcast does not exist, download it
	if !podcast.Exists() {
		if _, err := podcast.Download(); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		if err := podcast.Postprocess(); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
	} else {
		podcast.Bump()
	}
	shared.UpdateStream(podcast.Name(), podcast)
	c.Response().Header.Set("HX-Redirect", fmt.Sprintf("/podcast/%s", podcast.Name()))
	return c.SendStatus(fiber.StatusOK)
}

func GetPodcast(c *fiber.Ctx) error {
	return nil
}
