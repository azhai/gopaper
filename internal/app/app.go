package app

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/azhai/gopaper/internal/middleware"

	"github.com/azhai/gopaper/internal/service"

	"github.com/azhai/gopaper/internal/config"
	"github.com/azhai/gopaper/internal/handler"
	"github.com/azhai/gopaper/web"

	"github.com/azhai/gobus/log"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

type App struct {
	echo    *echo.Echo
	logger  *slog.Logger
	cache   *service.CacheVault
	scanner *service.Scanner
	cfg     *config.AppConfig
}

func New(cfg *config.AppConfig, dev bool) (*App, error) {
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
	pageHandler := handler.NewPageHandler(cache, renderer, cfg.SITE_URL)

	var sse *handler.SSEServer
	if dev {
		sse = handler.NewSSEServer(logger)
		pageHandler.SetDevMode(sse)
	}

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = customErrorHandler

	e.Use(slogMiddleware(logger))
	e.Use(echoMiddleware.Recover())
	allowedOrigins := cfg.ALLOWED_ORIGINS
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	publicFS, _ := fs.Sub(web.PublicFS, "public")
	adminFS, _ := fs.Sub(publicFS, "admin")

	e.Group("/uploads", echoMiddleware.StaticWithConfig(echoMiddleware.StaticConfig{Root: cfg.UPLOAD_DIR}))
	e.Group("/static", echoMiddleware.StaticWithConfig(echoMiddleware.StaticConfig{Root: "", Filesystem: http.FS(publicFS), IgnoreBase: true}))
	e.Group("/admin", echoMiddleware.StaticWithConfig(echoMiddleware.StaticConfig{Root: "", Filesystem: http.FS(adminFS), IgnoreBase: true, HTML5: true}))

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
		handler.StartDevWatchers(pageHandler, cache, scanner, cfg.CONTENT_DIR, logger)
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
	adminAuth.GET("/settings", cacheHandler.GetSettings)
	adminAuth.PUT("/settings", cacheHandler.UpdateSettings)

	adminAuth.GET("/layouts", handler.NewLayoutHandler().GetLayouts)
	adminAuth.PUT("/layouts", handler.NewLayoutHandler().SaveLayouts)
	adminAuth.GET("/layouts/:name/regions", handler.NewLayoutHandler().GetRegions)

	app := &App{
		echo:    e,
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
	return a.echo.Start(addr)
}

func (a *App) Shutdown() error {
	a.logger.Info("server shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return a.echo.Shutdown(ctx)
}

func customErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	msg := "内部服务器错误"
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if he.Message != nil {
			msg = fmt.Sprintf("%v", he.Message)
		} else {
			msg = http.StatusText(code)
		}
	} else if code < 500 {
		msg = http.StatusText(code)
	}
	if code >= 500 && !c.Echo().Debug {
		msg = "内部服务器错误"
	}
	c.JSON(code, map[string]any{
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
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			latency := time.Since(start)

			l.Info("request",
				"method", c.Request().Method,
				"path", c.Path(),
				"status", c.Response().Status,
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
