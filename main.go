package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/azhai/gopaper/internal/app"
	"github.com/azhai/gopaper/internal/config"
	"github.com/azhai/gopaper/internal/handler"
	"github.com/azhai/gopaper/internal/service"
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

	app, err := app.New(cfg, dev)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init app: %v\n", err)
		os.Exit(1)
	}

	go func() {
		if err := app.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "start server: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := app.Shutdown(); err != nil {
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

	cache := service.NewCacheVault(cfg.CACHE_SIZE, logger)
	scanner := service.NewScanner(cfg.CONTENT_DIR, logger)
	renderer := service.NewRenderer()

	if err := cache.Refresh(context.Background(), scanner); err != nil {
		return fmt.Errorf("scan: %w", err)
	}

	pageHandler := handler.NewPageHandler(cache, renderer, cfg.SITE_URL)
	builder := handler.NewBuilder(pageHandler, "dist", logger)
	return builder.Build(context.Background())
}
