package services

import (
	"context"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

type Renderer struct {
	md goldmark.Markdown
}

func NewRenderer() *Renderer {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.DefinitionList,
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	return &Renderer{md: md}
}

func (r *Renderer) Render(ctx context.Context, markdown []byte) ([]byte, error) {
	var buf []byte
	if err := r.md.Convert(markdown, &writeBuffer{data: &buf}); err != nil {
		return nil, fmt.Errorf("render markdown: %w", err)
	}
	return buf, nil
}

func (r *Renderer) RenderString(ctx context.Context, markdown string) (string, error) {
	result, err := r.Render(ctx, []byte(markdown))
	if err != nil {
		return "", err
	}
	return string(result), nil
}

type writeBuffer struct {
	data *[]byte
}

func (w *writeBuffer) Write(p []byte) (int, error) {
	*w.data = append(*w.data, p...)
	return len(p), nil
}

func (w *writeBuffer) WriteString(s string) (int, error) {
	*w.data = append(*w.data, s...)
	return len(s), nil
}
