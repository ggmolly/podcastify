package orm

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/ggmolly/podcastify/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var (
	PATH       = os.Getenv("PODCAST_ROOT_DIR")
	Validator  = validator.New()
	ytIdRegex  = regexp.MustCompile("[a-zA-Z0-9_-]{11}")
	ffmpegLock = sync.Mutex{}
	MaxLength  = 4 * time.Hour
)

type Podcast struct {
	URL           string    `gorm:"-" form:"url" validate:"required,url"`
	YoutubeID     string    `gorm:"primaryKey"`
	Sponsors      bool      `filename:"s" form:"sponsors" validate:"omitempty,boolean" arg:"sponsor"`
	SelfPromotion bool      `filename:"sp" form:"self-promotion" validate:"omitempty,boolean" arg:"selfpromo"`
	Intermissions bool      `filename:"i" form:"intermissions" validate:"omitempty,boolean" arg:"intro"`
	Reminders     bool      `filename:"r" form:"reminders" validate:"omitempty,boolean" arg:"interaction"`
	Credits       bool      `filename:"c" form:"credits" validate:"omitempty,boolean" arg:"outro"`
	Recaps        bool      `filename:"r" form:"recaps" validate:"omitempty,boolean" arg:"preview"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime;index,sort:asc" form:"-" validate:"-"`
}

// PodcastFromRequest returns a podcast struct from the request body
func PodcastFromRequest(c *fiber.Ctx) (*Podcast, error) {
	p := new(Podcast)
	if err := c.BodyParser(p); err != nil {
		return nil, err
	}
	if err := Validator.Struct(p); err != nil {
		return nil, err
	}
	// Find the youtube id from the URL
	if matches := ytIdRegex.FindStringSubmatch(p.URL); len(matches) > 0 {
		p.YoutubeID = matches[0]
	} else {
		return nil, fmt.Errorf("invalid youtube url")
	}
	return p, nil
}

// Returns the complete path of the podcast file based on the podcast struct
func (p *Podcast) Path() string {
	filename := p.Name()
	completePath := filepath.Join(PATH, filename)
	return completePath
}

// Returns whether the podcast file exists
func (p *Podcast) Exists() bool {
	if _, err := os.Stat(p.Path()); os.IsNotExist(err) {
		return false
	}
	return true
}

// Runs a bit of postprocessing on the podcast file
// 1. Converts to low bitrate ogg (64kbps) using ffmpeg
// 2. Removes the original file
func (p *Podcast) Postprocess() error {
	// Lock to avoid multiple ffmpeg processes running at the same time
	ffmpegLock.Lock()
	defer ffmpegLock.Unlock()
	input := p.Path()
	input = input[:len(input)-3] + "m4a"
	output := input[:len(input)-3] + "ogg"
	cmd := exec.Command("ffmpeg", "-i", input, "-c:a", "libvorbis", "-b:a", "64k", output)
	if err := cmd.Run(); err != nil {
		log.Println("failed to convert podcast: ", err)
		return err
	}
	if err := os.Remove(input); err != nil {
		log.Println("failed to remove original podcast: ", err)
		return err
	}
	return nil
}

// Returns a comma-separated list of the podcast's blacklisted categories
func (p *Podcast) BlacklistedCategories() string {
	var categories []string
	v := reflect.ValueOf(p).Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		arg := field.Tag.Get("arg")
		if arg != "" {
			if v.Field(i).Bool() {
				categories = append(categories, arg)
			}
		}
	}
	return strings.Join(categories, ",")
}

// Removes the podcast file and the podcast from the database
func (p *Podcast) Remove() {
	if err := os.Remove(p.Path()); err != nil {
		log.Println("failed to remove podcast file: ", err)
	}
	if err := GormDB.Delete(p).Error; err != nil {
		log.Println("failed to remove podcast from database: ", err)
	}
}

// Bump updates the UpdatedAt field of the podcast to now
// this to avoid the podcast from being deleted by the cleanup routine
func (p *Podcast) Bump() {
	p.UpdatedAt = time.Now()
	if err := GormDB.Save(p).Error; err != nil {
		log.Println("failed to update podcast: ", err)
	}
}

// Returns the length of the video in seconds (using yt-dlp)
func (p *Podcast) Length() (uint32, error) {
	args := []string{"--dump-json", "--simulate", fmt.Sprintf("https://www.youtube.com/watch?v=%s", p.YoutubeID)}
	cmd := exec.Command("yt-dlp", args...)
	out, err := cmd.Output()
	if err != nil {
		log.Println("failed to get video metadata: ", err)
		return 0, err
	}
	jsonDecoder := json.NewDecoder(strings.NewReader(string(out)))
	// Iterate over the json object to find the duration_string key
	for {
		t, err := jsonDecoder.Token()
		if err != nil {
			break
		}
		if t == "duration_string" {
			t, err = jsonDecoder.Token()
			if err != nil {
				break
			}
			return utils.ParseDuration(t.(string))
		}
	}
	return 0, nil
}

// Download downloads the podcast from youtube, must manually check if the podcast already exists
func (p *Podcast) Download() (string, error) {
	length, err := p.Length()
	if err != nil {
		return "", err
	}
	if length > uint32(MaxLength.Seconds()) {
		return "", fmt.Errorf("video is too long, max length is %s", MaxLength.String())
	}
	output := p.Path()
	blacklisted := p.BlacklistedCategories()
	args := []string{"-f", "140"}
	if blacklisted != "" {
		args = append(args, "--sponsorblock-remove", blacklisted)
	}
	args = append(args, fmt.Sprintf("https://www.youtube.com/watch?v=%s", p.YoutubeID), "-o", output[:len(output)-3]+"m4a")
	cmd := exec.Command("yt-dlp", args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Println("failed to download podcast: ", err)
		return "", err
	}
	return output, nil
}

func (p *Podcast) Name() string {
	name := p.YoutubeID + "_"
	v := reflect.ValueOf(p).Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		suffix := field.Tag.Get("filename")
		if suffix != "" {
			if v.Field(i).Bool() {
				name += suffix
			}
		}
	}
	return name + ".ogg"
}
