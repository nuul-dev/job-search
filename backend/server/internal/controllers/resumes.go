package controllers

import (
	"io"
	"mime"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"job-search/server/internal/handlers"
	"job-search/server/internal/repository"
)

type Resumes struct{ handler *handlers.Resumes }

func NewResumes(handler *handlers.Resumes) *Resumes { return &Resumes{handler: handler} }
func (c *Resumes) RegisterRoutes(router fiber.Router) {
	router.Post("/api/resumes", localJSON, c.create)
	router.Post("/api/resumes/upload", localOrigin, c.upload)
	router.Get("/api/resumes/:name/download", c.download)
	router.Get("/api/resumes/:name", c.get)
	router.Put("/api/resumes/:name", localJSON, c.update)
}
func resumeError(ctx *fiber.Ctx, err error) error {
	status := 400
	if errors.Is(err, repository.ErrResumeConflict) {
		status = 409
	}
	if errors.Is(err, repository.ErrResumeMissing) {
		status = 404
	}
	return ctx.Status(status).JSON(fiber.Map{"error": err.Error()})
}
func (c *Resumes) get(ctx *fiber.Ctx) error {
	view, err := c.handler.Get(ctx.UserContext(), ctx.Params("name"))
	if err != nil {
		return resumeError(ctx, err)
	}
	return ctx.JSON(view)
}
func (c *Resumes) create(ctx *fiber.Ctx) error {
	var request struct {
		Name string `json:"name"`
		Text string `json:"text"`
	}
	if !strictObject(ctx.Body(), &request) {
		return ctx.Status(400).JSON(fiber.Map{"error": "Некорректное резюме"})
	}
	view, err := c.handler.Write(request.Name, request.Text, "", true)
	if err != nil {
		return resumeError(ctx, err)
	}
	return ctx.Status(201).JSON(view)
}
func (c *Resumes) update(ctx *fiber.Ctx) error {
	var request struct {
		Text    string `json:"text"`
		Version string `json:"version"`
	}
	if !strictObject(ctx.Body(), &request) {
		return ctx.Status(400).JSON(fiber.Map{"error": "Некорректное резюме"})
	}
	view, err := c.handler.Write(ctx.Params("name"), request.Text, request.Version, false)
	if err != nil {
		return resumeError(ctx, err)
	}
	return ctx.JSON(view)
}
func (c *Resumes) upload(ctx *fiber.Ctx) error {
	media, _, err := mime.ParseMediaType(ctx.Get("Content-Type"))
	if err != nil || media != "multipart/form-data" {
		return ctx.SendStatus(415)
	}
	header, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Выберите файл PDF или Markdown"})
	}
	if header.Size > 8*1024*1024 {
		return ctx.SendStatus(413)
	}
	file, err := header.Open()
	if err != nil {
		return ctx.SendStatus(400)
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 8*1024*1024+1))
	if err != nil {
		return ctx.SendStatus(400)
	}
	view, err := c.handler.Upload(header.Filename, data)
	if err != nil {
		return resumeError(ctx, err)
	}
	return ctx.Status(201).JSON(view)
}
func (c *Resumes) download(ctx *fiber.Ctx) error {
	name := ctx.Params("name")
	data, err := c.handler.Download(name)
	if err != nil {
		return resumeError(ctx, err)
	}
	contentType := "text/markdown; charset=utf-8"
	if strings.EqualFold(filepath.Ext(name), ".pdf") {
		contentType = "application/pdf"
	}
	ctx.Set("Content-Type", contentType)
	ctx.Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	return ctx.Send(data)
}
