package app

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"time"

	"github.com/azhai/gopaper/internal/middleware"

	"github.com/azhai/gopaper/internal/service"

	"github.com/azhai/gopaper/internal/config"
	"github.com/azhai/gopaper/internal/handler"
	"github.com/azhai/gopaper/web"

	"github.com/azhai/gobus/log"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
)

type App struct {
	fiber   *fiber.App
	logger  *slog.Logger
	cache   *service.CacheVault
	scanner *service.Scanner
	cfg     *config.AppConfig
}

func New(cfg *config.AppConfig) (*App, error) {
	logger := initLogger("logs/app.log")

	cache := service.NewCacheVault(cfg.CACHE_SIZE, logger)
	scanner := service.NewScanner(cfg.CONTENT_DIR, logger)
	renderer := service.NewRenderer()
	articleForge := service.NewArticleForge(cfg.CONTENT_DIR, cache, scanner, logger)
	imageStore := service.NewImageStore(cfg.UPLOAD_DIR, cfg.MAX_UPLOAD_SIZE, cache, logger)

	ttl, err := time.ParseDuration(cfg.JWT_TTL)
	if err != nil {
		ttl = 24 * time.Hour
	}
	authGuard := middleware.NewAuthGuard(cfg.JWT_SECRET, ttl, cfg.Admin)

	authHandler := handler.NewAuthHandler(authGuard)
	articleHandler := handler.NewArticleHandler(articleForge, renderer)
	imageHandler := handler.NewImageHandler(imageStore)
	cacheHandler := handler.NewCacheHandler(cache, scanner, cfg.CONTENT_DIR)
	pageHandler := handler.NewPageHandler(cache, renderer)

	fiberApp := fiber.New(fiber.Config{
		AppName:      "StaticCMS",
		ServerHeader: "StaticCMS",
		ErrorHandler: customErrorHandler,
	})

	fiberApp.Use(slogMiddleware(logger))
	fiberApp.Use(recover.New())
	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	publicFS, _ := fs.Sub(web.PublicFS, "public")
	adminFS, _ := fs.Sub(publicFS, "admin")

	fiberApp.Use("/uploads", static.New(cfg.UPLOAD_DIR))
	fiberApp.Use("/static", static.New("", static.Config{FS: publicFS}))

	// Admin: serve static assets directly; SPA fallback for other routes
	fiberApp.Use("/admin", static.New("", static.Config{FS: adminFS}))
	fiberApp.Get("/admin/*", func(c fiber.Ctx) error {
		path := c.Params("*")
		// Try serving as a static file first
		if f, err := adminFS.Open(path); err == nil {
			f.Close()
			return c.SendFile(path, fiber.SendFile{FS: adminFS})
		}
		// SPA fallback: serve index.html
		return c.SendFile("index.html", fiber.SendFile{FS: adminFS})
	})

	fiberApp.Get("/", pageHandler.Index)
	fiberApp.Get("/:dir", pageHandler.DirList)
	fiberApp.Get("/:dir/:slug", pageHandler.ArticleDetail)

	api := fiberApp.Group("/api")
	admin := api.Group("/admin")

	admin.Post("/login", authHandler.Login)

	adminAuth := admin.Group("", authGuard.FiberMiddleware())
	adminAuth.Get("/articles", articleHandler.List)
	adminAuth.Get("/articles/:slug", articleHandler.Get)
	adminAuth.Post("/articles", articleHandler.Create)
	adminAuth.Put("/articles/:slug", articleHandler.Update)
	adminAuth.Delete("/articles/:slug", articleHandler.Delete)
	adminAuth.Post("/articles/preview", articleHandler.Preview)

	adminAuth.Get("/images", imageHandler.List)
	adminAuth.Post("/images", imageHandler.Upload)
	adminAuth.Delete("/images/:fileName", imageHandler.Delete)

	adminAuth.Post("/cache/refresh", cacheHandler.Refresh)
	adminAuth.Get("/dirs", cacheHandler.ListDirs)
	adminAuth.Get("/settings", cacheHandler.GetSettings)
	adminAuth.Put("/settings", cacheHandler.UpdateSettings)

	app := &App{
		fiber:   fiberApp,
		logger:  logger,
		cache:   cache,
		scanner: scanner,
		cfg:     cfg,
	}

	ctx := context.Background()
	if err := cache.Refresh(ctx, scanner); err != nil {
		logger.Warn("initial cache refresh failed", "error", err)
	}

	return app, nil
}

func (a *App) Start() error {
	addr := fmt.Sprintf(":%s", a.cfg.SERVER_PORT)
	a.logger.Info("server starting", "addr", addr)
	return a.fiber.Listen(addr)
}

func (a *App) Shutdown() error {
	a.logger.Info("server shutting down")
	return a.fiber.Shutdown()
}

func customErrorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"code":    code*100 + 1,
		"message": err.Error(),
	})
}

func initLogger(logFile string) *slog.Logger {
	l, err := log.NewDailyLogger(logFile, 7)
	if err != nil || l == nil {
		l = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		l.Warn("failed to create daily log writer, fallback to stderr", "err", err)
	}
	return l
}

func slogMiddleware(l *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		l.Info("request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"latency", latency.String(),
			"ip", c.IP(),
			"ua", c.Get("User-Agent"),
		)
		return err
	}
}
