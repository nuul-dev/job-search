package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"job-search/server/internal/config"
	"job-search/server/internal/controllers"
	"job-search/server/internal/handlers"
	"job-search/server/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := run(); err != nil {
		logrus.WithError(err).Error("server stopped")
		os.Exit(1)
	}
}
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	search := handlers.NewSearch(cfg.Timeout, handlers.CommandRunner(cfg.Root))
	defer search.Close()
	drafts := handlers.NewDrafts(repository.DraftFiles{Root: cfg.Root}, cfg.DraftTimeout, handlers.CodexDraftGenerator(cfg.DraftModel))
	defer drafts.Close()
	app := fiber.New(fiber.Config{DisableStartupMessage: true, BodyLimit: 9 * 1024 * 1024, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 30 * time.Second})
	app.Use(func(ctx *fiber.Ctx) error {
		if !(ctx.Method() == "POST" && ctx.Path() == "/api/resumes/upload") && len(ctx.Body()) > 128*1024 {
			return ctx.SendStatus(413)
		}
		return ctx.Next()
	})
	controllers.NewInbox(handlers.NewRuns(repository.Files{Root: cfg.Root}), search, cfg.Root, cfg.Port).RegisterRoutes(app)
	controllers.NewDrafts(drafts).RegisterRoutes(app)
	controllers.NewResumes(handlers.NewResumes(&repository.ResumeFiles{Root: cfg.Root})).RegisterRoutes(app)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	stopped := make(chan error, 1)
	go func() { stopped <- app.Listen(cfg.Address()) }()
	logrus.WithField("address", cfg.Address()).Info("job inbox listening")
	select {
	case err := <-stopped:
		return err
	case <-ctx.Done():
		drafts.Close()
		search.Close()
		return app.ShutdownWithTimeout(5 * time.Second)
	}
}
