package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/pressly/goose/v3"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/school-management/pos/internal/auth"
	"github.com/school-management/pos/internal/config"
	"github.com/school-management/pos/internal/database"
	"github.com/school-management/pos/internal/handler"
	"github.com/school-management/pos/internal/middleware"
	"github.com/school-management/pos/internal/repository"
	"github.com/school-management/pos/internal/service"
	"github.com/school-management/pos/internal/storage"
	"github.com/school-management/pos/internal/validator"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	logger := newLogger(cfg.LogLevel)
	ctx := context.Background()

	pool, err := database.NewPool(ctx, cfg.Database.URL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := runMigrations(cfg.Database.URL); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	repos := repository.New(pool)
	tokens := auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL, cfg.App.Name)
	services := service.NewServices(repos, cfg, tokens, logger)

	if err := services.Seed.EnsureAdmin(ctx); err != nil {
		logger.Error("admin seed failed", "error", err)
		os.Exit(1)
	}

	if err := services.Backup.EnsureDir(); err != nil {
		logger.Warn("backup directory init failed", "error", err)
	}
	schedCtx, schedCancel := context.WithCancel(ctx)
	defer schedCancel()
	services.Backup.StartScheduler(schedCtx)

	store, err := storage.New(cfg.R2)
	if err != nil {
		logger.Error("storage init failed", "error", err)
		os.Exit(1)
	}

	validate := validator.New()
	mw := middleware.New(tokens, services.Auth, cfg.CSRF.Secret, logger)
	h := handler.New(services, validate, cfg, store, logger)

	app := fiber.New(fiber.Config{
		AppName:        cfg.App.Name,
		ServerHeader:   "School-POS",
		ReadBufferSize: 16384,
	})

	app.Use(recover.New())
	app.Use(mw.RequestLogger())
	app.Use(limiter.New(limiter.Config{
		Max:        cfg.Rate.Max,
		Expiration: cfg.Rate.Expiration,
	}))

	app.Use("/static", static.New("./web/static", static.Config{
		MaxAge: 86400,
	}))

	h.Register(app, mw)

	go func() {
		addr := ":" + cfg.App.Port
		logger.Info("server starting", "addr", addr, "env", cfg.App.Env)
		if err := app.Listen(addr); err != nil {
			logger.Error("server stopped", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = app.ShutdownWithContext(shutdownCtx)
	logger.Info("server shutdown complete")
}

func runMigrations(databaseURL string) error {
	db, err := goose.OpenDBWithDriver("pgx", databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	return goose.Up(db, "migrations")
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}
