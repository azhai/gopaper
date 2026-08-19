package handler

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/azhai/gopaper/internal/service"

	"github.com/labstack/echo/v4"
)

type SSEServer struct {
	mu      sync.Mutex
	clients map[chan struct{}]struct{}
	logger  *slog.Logger
}

func NewSSEServer(logger *slog.Logger) *SSEServer {
	return &SSEServer{clients: make(map[chan struct{}]struct{}), logger: logger}
}

func (s *SSEServer) addClient() chan struct{} {
	ch := make(chan struct{}, 1)
	s.mu.Lock()
	s.clients[ch] = struct{}{}
	s.mu.Unlock()
	return ch
}

func (s *SSEServer) removeClient(ch chan struct{}) {
	s.mu.Lock()
	delete(s.clients, ch)
	s.mu.Unlock()
	close(ch)
}

func (s *SSEServer) Broadcast() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for ch := range s.clients {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (s *SSEServer) Livereload(c echo.Context) error {
	w := c.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	ch := s.addClient()
	defer s.removeClient(ch)

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request().Context().Done():
			return nil
		case <-ch:
			_, _ = w.Write([]byte("data: reload\n\n"))
			w.Flush()
		case <-ticker.C:
			_, _ = w.Write([]byte(": ping\n\n"))
			w.Flush()
		}
	}
}

func StartDevWatchers(pageHandler *PageHandler, cache *service.CacheVault, scanner *service.Scanner, contentDir string, logger *slog.Logger) {
	go watchTemplates(pageHandler, logger)
	go watchContent(cache, scanner, contentDir, logger)
}

func watchTemplates(ph *PageHandler, logger *slog.Logger) {
	dir := "themes/default/layouts"
	var lastMtime time.Time
	for {
		time.Sleep(time.Second)
		mtime, err := dirMaxMtime(dir)
		if err != nil {
			continue
		}
		if !mtime.After(lastMtime) {
			continue
		}
		lastMtime = mtime
		if err := ph.reloadTemplates(); err != nil {
			logger.Warn("reload templates failed", "error", err)
			continue
		}
		logger.Info("templates reloaded")
		if ph.sse != nil {
			ph.sse.Broadcast()
		}
	}
}

func watchContent(cache *service.CacheVault, scanner *service.Scanner, contentDir string, logger *slog.Logger) {
	var lastMtime time.Time
	for {
		time.Sleep(time.Second)
		mtime, err := dirMaxMtime(contentDir)
		if err != nil {
			continue
		}
		if !mtime.After(lastMtime) {
			continue
		}
		lastMtime = mtime
		if err := cache.Refresh(context.Background(), scanner); err != nil {
			logger.Warn("cache refresh failed", "error", err)
			continue
		}
		logger.Info("content reloaded")
	}
}

func dirMaxMtime(dir string) (time.Time, error) {
	var max time.Time
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.ModTime().After(max) {
			max = info.ModTime()
		}
		return nil
	})
	if err != nil {
		return time.Time{}, err
	}
	return max, nil
}