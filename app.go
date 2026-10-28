package main

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/azhai/gopaper/config"
	"github.com/azhai/gopaper/handlers"
	"github.com/azhai/gopaper/middlewares"
	"github.com/azhai/gopaper/services"
	"github.com/azhai/gopaper/web"

	"github.com/azhai/gobus/log"
	"github.com/labstack/echo/v5"
	midd "github.com/labstack/echo/v5/middleware"
)

type App struct {
	echo      *echo.Echo
	logger    *slog.Logger
	cache     *services.CacheVault
	scanner   *services.Scanner
	analytics *services.Analytics
	cfg       *config.AppConfig
}

func newApp(cfg *config.AppConfig, dev bool) (*App, error) {
	logger := initLogger("logs/app.log")

	cache := services.NewCacheVault(cfg.CACHE_SIZE, logger)
	scanner := services.NewScanner(cfg.CONTENT_DIR, logger)
	renderer := services.NewRenderer()
	articleForge := services.NewArticleForge(cfg.CONTENT_DIR, cache, scanner, logger)
	imageStore := services.NewImageStore(cfg.UPLOAD_DIR, cfg.MAX_UPLOAD_SIZE, cache, logger)
	sectionForge := services.NewSectionForge(cfg.CONTENT_DIR, cache, scanner, logger)
	themeForge := services.NewThemeForge(cfg.CONTENT_DIR, logger)
	analytics := services.NewAnalytics(cfg.CONTENT_DIR, logger)

	ttl, err := time.ParseDuration(cfg.JWT_TTL)
	if err != nil {
		ttl = 24 * time.Hour
	}
	authGuard := middlewares.NewAuthGuard(cfg.JWT_SECRET, ttl, cfg.Admin)

	authHandler := handlers.NewAuthHandler(authGuard)
	articleHandler := handlers.NewArticleHandler(articleForge, renderer)
	imageHandler := handlers.NewImageHandler(imageStore)
	cacheHandler := handlers.NewCacheHandler(cache, scanner, cfg.CONTENT_DIR)
	sectionHandler := handlers.NewSectionHandler(sectionForge)
	pageHandler := handlers.NewPageHandler(cache, renderer, cfg.SITE_URL)
	pageHandler.SetThemes(themeForge, logger)
	pageHandler.SetAnalytics(analytics)
	themeHandler := handlers.NewThemeHandler(themeForge)
	analyticsHandler := handlers.NewAnalyticsHandler(analytics)

	var sse *handlers.SSEServer
	if dev {
		sse = handlers.NewSSEServer(logger)
		pageHandler.SetDevMode(sse)
	}

	e := echo.New()
	e.Logger = logger
	e.HTTPErrorHandler = customErrorHandler

	e.Use(slogMiddleware(logger))
	e.Use(midd.Recover())
	allowedOrigins := cfg.ALLOWED_ORIGINS
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}
	e.Use(midd.CORSWithConfig(midd.CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	publicFS, _ := fs.Sub(web.PublicFS, "public")
	adminFS, _ := fs.Sub(publicFS, "admin")

	e.Group("/uploads", midd.StaticWithConfig(midd.StaticConfig{Root: cfg.UPLOAD_DIR}))
	e.Group("/static", midd.StaticWithConfig(midd.StaticConfig{Root: ".", Filesystem: publicFS, IgnoreBase: true}))
	e.Group("/admin", midd.StaticWithConfig(midd.StaticConfig{Root: ".", Filesystem: adminFS, IgnoreBase: true, HTML5: true}))

	e.GET("/", pageHandler.Index)
	e.GET("/:dir", pageHandler.DirList)
	e.GET("/:dir/page/:num", pageHandler.DirList)
	e.GET("/:dir/:slug", pageHandler.ArticleDetail)
	e.GET("/:d1/:d2/:d3", pageHandler.ResolvePath)
	e.GET("/:d1/:d2/:d3/:d4", pageHandler.ResolvePath)
	e.GET("/:d1/:d2/:d3/:d4/:d5", pageHandler.ResolvePath)
	e.GET("/index.xml", pageHandler.RSS)
	e.GET("/sitemap.xml", pageHandler.Sitemap)
	e.GET("/robots.txt", pageHandler.Robots)

	if dev {
		e.GET("/livereload", sse.Livereload)
		handlers.StartDevWatchers(pageHandler, cache, scanner, cfg.CONTENT_DIR, logger)
		logger.Info("dev mode enabled: template/content watchers started")
	}

	api := e.Group("/api")
	admin := api.Group("/admin")

	admin.POST("/login", authHandler.Login)

	adminAuth := admin.Group("", authGuard.EchoMiddleware())
	adminAuth.GET("/articles", articleHandler.List)
	adminAuth.GET("/articles/:slug", articleHandler.Get)
	adminAuth.POST("/articles", articleHandler.Create)
	adminAuth.PUT("/articles/:slug", articleHandler.Update)
	adminAuth.DELETE("/articles/:slug", articleHandler.Delete)
	adminAuth.POST("/articles/preview", articleHandler.Preview)

	adminAuth.GET("/images", imageHandler.List)
	adminAuth.POST("/images", imageHandler.Upload)
	adminAuth.DELETE("/images/:fileName", imageHandler.Delete)

	adminAuth.POST("/cache/refresh", cacheHandler.Refresh)
	adminAuth.GET("/dirs", cacheHandler.ListDirs)
	adminAuth.POST("/sections", sectionHandler.Create)
	adminAuth.PUT("/sections", sectionHandler.Update)
	adminAuth.GET("/settings", cacheHandler.GetSettings)
	adminAuth.PUT("/settings", cacheHandler.UpdateSettings)

	adminAuth.GET("/analytics", analyticsHandler.Stats)

	adminAuth.GET("/themes", themeHandler.List)
	adminAuth.POST("/themes", themeHandler.Create)
	adminAuth.POST("/themes/preview", themeHandler.Preview)
	adminAuth.PUT("/themes/:id", themeHandler.Update)
	adminAuth.DELETE("/themes/:id", themeHandler.Delete)
	adminAuth.POST("/themes/:id/activate", themeHandler.Activate)

	adminAuth.GET("/layouts", handlers.NewLayoutHandler().GetLayouts)
	adminAuth.PUT("/layouts", handlers.NewLayoutHandler().SaveLayouts)
	adminAuth.GET("/layouts/:name/regions", handlers.NewLayoutHandler().GetRegions)

	app := &App{
		echo:      e,
		logger:    logger,
		cache:     cache,
		scanner:   scanner,
		analytics: analytics,
		cfg:       cfg,
	}

	ctx := context.Background()
	if err := cache.Refresh(ctx, scanner); err != nil {
		logger.Warn("initial cache refresh failed", "error", err)
	}

	return app, nil
}

func (a *App) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%s", a.cfg.SERVER_PORT)
	a.logger.Info("server starting", "addr", addr)
	sc := echo.StartConfig{Address: addr, HideBanner: true, GracefulTimeout: 10 * time.Second}
	return sc.Start(ctx, a.echo)
}

func (a *App) Close() error {
	a.logger.Info("server shutting down")
	return a.analytics.Close()
}

func customErrorHandler(c *echo.Context, err error) {
	code := echo.StatusCode(err)
	if code == 0 {
		code = http.StatusInternalServerError
	}
	msg := http.StatusText(code)
	if he, ok := err.(*echo.HTTPError); ok && he.Message != "" {
		msg = he.Message
	}
	if code >= 500 {
		msg = "内部服务器错误"
	}
	_ = c.JSON(code, map[string]any{
		"code":    code*100 + 1,
		"message": msg,
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

func slogMiddleware(l *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			err := next(c)
			latency := time.Since(start)

			_, status := echo.ResolveResponseStatus(c.Response(), err)
			l.Info("request",
				"method", c.Request().Method,
				"path", c.Path(),
				"status", status,
				"latency", latency.String(),
				"ip", c.RealIP(),
				"ua", c.Request().UserAgent(),
			)
			return err
		}
	}
}

func mimeByExt(path string) string {
	switch {
	case len(path) >= 4 && path[len(path)-4:] == ".css":
		return "text/css; charset=utf-8"
	case len(path) >= 3 && path[len(path)-3:] == ".js":
		return "application/javascript"
	case len(path) >= 5 && path[len(path)-5:] == ".html":
		return "text/html; charset=utf-8"
	case len(path) >= 4 && path[len(path)-4:] == ".svg":
		return "image/svg+xml"
	case len(path) >= 4 && path[len(path)-4:] == ".png":
		return "image/png"
	case len(path) >= 4 && path[len(path)-4:] == ".jpg":
		return "image/jpeg"
	case len(path) >= 5 && path[len(path)-5:] == ".json":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}
