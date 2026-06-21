package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/azhai/gopaper/internal/app"
	"github.com/azhai/gopaper/internal/config"
)

func main() {
	cfg, err := config.Load("config.toml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	app, err := app.New(cfg)
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
