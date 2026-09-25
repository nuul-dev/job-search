package controllers

import (
	"encoding/json"
	"mime"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"job-search/fetch/searchoptions"
	"job-search/server/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type Inbox struct {
	runs   *handlers.Runs
	search *handlers.Search
	root   string
	port   int
}

func NewInbox(runs *handlers.Runs, search *handlers.Search, root string, port int) *Inbox {
	return &Inbox{runs: runs, search: search, root: root, port: port}
}
func (c *Inbox) RegisterRoutes(router fiber.Router) {
	router.Use(func(ctx *fiber.Ctx) error {
		ctx.Set("Cache-Control", "no-store")
		ctx.Set("X-Content-Type-Options", "nosniff")
		if !c.localHost(string(ctx.Request().Host())) {
			return ctx.SendStatus(fiber.StatusForbidden)
		}
		return ctx.Next()
	})
	router.Get("/api/runs", c.list)
	router.Get("/api/search", func(ctx *fiber.Ctx) error { return ctx.JSON(c.search.Status()) })
	router.Post("/api/search", localJSON, c.start)
	router.Post("/api/search/cancel", localJSON, c.cancel)
	for route, asset := range map[string]struct{ name, mime string }{
		"/":           {"index.html", "text/html; charset=utf-8"},
		"/app.js":     {"app.js", "text/javascript; charset=utf-8"},
		"/letters.js": {"letters.js", "text/javascript; charset=utf-8"},
		"/resumes.js": {"resumes.js", "text/javascript; charset=utf-8"},
		"/style.css":  {"style.css", "text/css; charset=utf-8"},
	} {
		router.Get(route, func(ctx *fiber.Ctx) error {
			data, err := os.ReadFile(filepath.Join(c.root, "frontend", asset.name))
			if err != nil {
				logrus.WithError(err).Error("read asset")
				return ctx.SendStatus(500)
			}
			ctx.Set("Content-Type", asset.mime)
			return ctx.Send(data)
		})
	}
}
func (c *Inbox) localHost(host string) bool {
	name, port, err := net.SplitHostPort(host)
	if err != nil {
		return c.port == 80 && (host == "localhost" || host == "127.0.0.1")
	}
	return port == strconv.Itoa(c.port) && (name == "localhost" || name == "127.0.0.1")
}
func (c *Inbox) list(ctx *fiber.Ctx) error {
	result, err := c.runs.List()
	if err != nil {
		logrus.WithError(err).Error("list runs")
		return ctx.Status(500).JSON(fiber.Map{"error": "Could not read exports"})
	}
	return ctx.JSON(result)
}
func checkOrigin(ctx *fiber.Ctx) error {
	if origin := ctx.Get("Origin"); origin != "" {
		u, err := url.Parse(origin)
		if err != nil || u.Scheme != "http" || u.Host != string(ctx.Request().Host()) || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
			return fiber.NewError(403, "Foreign origin")
		}
	}
	if site := ctx.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "none" {
		return fiber.NewError(403, "Foreign origin")
	}
	return nil
}
func localOrigin(ctx *fiber.Ctx) error {
	if err := checkOrigin(ctx); err != nil {
		return err
	}
	return ctx.Next()
}
func localJSON(ctx *fiber.Ctx) error {
	if err := checkOrigin(ctx); err != nil {
		return err
	}
	media, _, err := mime.ParseMediaType(ctx.Get("Content-Type"))
	if err != nil || media != "application/json" {
		return ctx.Status(415).JSON(fiber.Map{"error": "Content-Type must be application/json"})
	}
	return ctx.Next()
}
func (c *Inbox) cancel(ctx *fiber.Ctx) error {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(ctx.Body(), &body); err != nil || body == nil || len(body) != 0 {
		return ctx.Status(400).JSON(fiber.Map{"error": "Ожидается пустой JSON-объект"})
	}
	return ctx.JSON(c.search.Cancel())
}
func (c *Inbox) start(ctx *fiber.Ctx) error {
	options, err := searchoptions.Decode(ctx.Body())
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	state, err := c.search.Start(options)
	if err == handlers.ErrRunning {
		return ctx.Status(409).JSON(state)
	}
	if err != nil {
		return ctx.Status(503).JSON(fiber.Map{"error": "Server shutting down"})
	}
	return ctx.Status(202).JSON(state)
}
