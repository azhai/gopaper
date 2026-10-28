package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/azhai/gopaper/config"
	"github.com/azhai/gopaper/handlers"
	"github.com/azhai/gopaper/services"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "build":
			if err := runBuild(); err != nil {
				fmt.Fprintf(os.Stderr, "build: %v\n", err)
				os.Exit(1)
			}
			return
		case "dev":
			runServer(true)
			return
		}
	}
	runServer(false)
}

func runServer(dev bool) {
	cfg, err := config.Load("config.toml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	application, err := newApp(cfg, dev)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init app: %v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := application.Start(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Fprintf(os.Stderr, "start server: %v\n", err)
		os.Exit(1)
	}
	if err := application.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "shutdown: %v\n", err)
	}
}

func runBuild() error {
	cfg, err := config.Load("config.toml")
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cache := services.NewCacheVault(cfg.CACHE_SIZE, logger)
	scanner := services.NewScanner(cfg.CONTENT_DIR, logger)
	renderer := services.NewRenderer()

	if err := cache.Refresh(context.Background(), scanner); err != nil {
		return fmt.Errorf("scan: %w", err)
	}

	pageHandler := handlers.NewPageHandler(cache, renderer, cfg.SITE_URL)
	pageHandler.SetThemes(services.NewThemeForge(cfg.CONTENT_DIR, logger), logger)
	builder := handlers.NewBuilder(pageHandler, "dist", logger)
	return builder.Build(context.Background())
}
