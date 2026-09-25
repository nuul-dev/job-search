package controllers

import (
	"bytes"
	"encoding/json"
	"io"

	"job-search/server/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

type Drafts struct{ handler *handlers.Drafts }

func NewDrafts(handler *handlers.Drafts) *Drafts { return &Drafts{handler: handler} }
func (c *Drafts) RegisterRoutes(router fiber.Router) {
	router.Get("/api/resumes", c.resumes)
	router.Post("/api/drafts", localJSON, c.start)
	router.Get("/api/drafts/:id", c.get)
	router.Patch("/api/drafts/:id", localJSON, c.save)
}
func strictObject(data []byte, value any) bool {
	if len(bytes.TrimSpace(data)) == 0 || bytes.TrimSpace(data)[0] != '{' {
		return false
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if d.Decode(value) != nil {
		return false
	}
	return d.Decode(new(any)) == io.EOF
}
func draftError(ctx *fiber.Ctx, err error) error {
	status := 400
	switch err {
	case handlers.ErrDraftNotFound:
		status = 404
	case handlers.ErrDraftBusy, handlers.ErrDraftNotReady:
		status = 409
	case handlers.ErrClosed:
		status = 503
	}
	return ctx.Status(status).JSON(fiber.Map{"error": err.Error()})
}
func (c *Drafts) resumes(ctx *fiber.Ctx) error {
	resumes, err := c.handler.Resumes()
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Не удалось открыть папку resumes"})
	}
	return ctx.JSON(fiber.Map{"resumes": resumes})
}
func (c *Drafts) start(ctx *fiber.Ctx) error {
	var request handlers.DraftRequest
	if !strictObject(ctx.Body(), &request) {
		return ctx.Status(400).JSON(fiber.Map{"error": "Некорректные параметры письма"})
	}
	job, err := c.handler.Start(request)
	if err != nil {
		return draftError(ctx, err)
	}
	return ctx.Status(202).JSON(job)
}
func (c *Drafts) get(ctx *fiber.Ctx) error {
	job, err := c.handler.Get(ctx.Params("id"))
	if err != nil {
		return draftError(ctx, err)
	}
	return ctx.JSON(job)
}
func (c *Drafts) save(ctx *fiber.Ctx) error {
	var request struct {
		Text string `json:"text"`
	}
	if !strictObject(ctx.Body(), &request) {
		return ctx.Status(400).JSON(fiber.Map{"error": "Некорректный текст письма"})
	}
	job, err := c.handler.Save(ctx.Params("id"), request.Text)
	if err != nil {
		return draftError(ctx, err)
	}
	return ctx.JSON(job)
}
