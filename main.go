package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ggmolly/podcastify/middlewares"
	"github.com/ggmolly/podcastify/orm"

	"github.com/ggmolly/podcastify/routes"

	"github.com/gofiber/fiber/v2"

	"encoding/json"

	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

var (
	BindAddress    = "127.0.0.1"
	Port           = "8000"
	ExpirationTime = 6 * time.Hour
)

// janitor is a background task that runs every 5 minutes to clean up old files
func janitor() {
	for {
		time.Sleep(5 * time.Minute)
		var podcast []orm.Podcast
		if err := orm.GormDB.
			Find(&podcast).
			Where("updated_at < ?", time.Now().Add(-ExpirationTime)).
			Error; err != nil {
			log.Printf("failed to fetch podcasts: %v\n", err)
			continue
		}
		log.Printf("cleaning up %d podcasts\n", len(podcast))
		// no transaction because we don't rollback if something fails in the middle
		// of the loop
		for _, p := range podcast {
			if err := os.Remove(filepath.Join(os.Getenv("PODCAST_ROOT_DIR"), p.Name())); err != nil {
				log.Printf("failed to remove podcast file: %v\n", err)
			}
			if err := orm.GormDB.Delete(&p).Error; err != nil {
				log.Printf("failed to delete podcast: %v\n", err)
			}
		}
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load .env file: %v\n", err)
	}
	Port = os.Getenv("PORT")
	BindAddress = os.Getenv("BIND_ADDRESS")
	orm.InitDatabase()
	// parse the expiration time from the .env file, set to 6 hours if not found
	timeInt, err := strconv.Atoi(os.Getenv("EXPIRATION_TIME"))
	if err != nil {
		log.Printf("failed to parse EXPIRATION_TIME variable: %v\n", err)
		log.Println("setting expiration time to 6 hours")
	} else {
		ExpirationTime = time.Duration(timeInt) * time.Hour
	}
	timeInt, err = strconv.Atoi(os.Getenv("MAX_VIDEO_LENGTH"))
	if err != nil {
		log.Printf("failed to parse MAX_VIDEO_LENGTH variable: %v\n", err)
		log.Println("setting max video length to 4 hours")
	} else {
		orm.MaxLength = time.Duration(timeInt) * time.Second
	}
}

func main() {
	engine := html.New("./views", ".html")

	if os.Getenv("MODE") != "production" {
		log.Println("dev mode enabled, reloading templates on each request")
		engine.Reload(true)
	}
	app := fiber.New(fiber.Config{
		Views:        engine,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		AppName:      "Podcastify",
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
		ProxyHeader:  "CF-Connecting-IP",
	})

	app.Use(middlewares.FiberCompress)

	// Static resources
	app.Static("/static", "./static")

	// Routes
	app.Get("/", routes.Index)

	api := app.Group("/api/v1")
	{
		api.Post("/podcastify", routes.Podcastify)
	}

	app.Get("/podcast/:name", func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(os.Getenv("PODCAST_ROOT_DIR"), c.Params("name")))
	})

	go janitor()

	// Listen on port 8000
	if err := app.Listen(fmt.Sprintf("%s:%s", BindAddress, Port)); err != nil {
		log.Fatalf("Failed to listen on port %s: %v\n", Port, err)
	}
}
