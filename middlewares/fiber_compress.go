package middlewares

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
)

var (
	FiberCompress = compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
		Next: func(c *fiber.Ctx) bool { // Only compress static files
			return !strings.HasPrefix(c.Path(), "/static")
		},
	})
)
