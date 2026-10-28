package models

import "time"

type ProductSpec struct {
	Name      string `json:"name" toml:"name"`
	Price     string `json:"price,omitempty" toml:"price"`
	SKU       string `json:"sku,omitempty" toml:"sku"`
	JDURL     string `json:"jdUrl,omitempty" toml:"jd_url"`
	TaobaoURL string `json:"taobaoUrl,omitempty" toml:"taobao_url"`
}

type ProductImprovement struct {
	Date    string `json:"date" toml:"date"`
	Content string `json:"content" toml:"content"`
	Visible bool   `json:"visible" toml:"visible"`
}

type Article struct {
	Slug         string               `json:"slug"`
	Title        string               `json:"title"`
	Author       string               `json:"author,omitempty"`
	Date         string               `json:"date,omitempty"`
	Tags         []string             `json:"tags,omitempty"`
	Comments     bool                 `json:"comments"`
	Weight       int                  `json:"weight,omitempty"`
	Position     string               `json:"position,omitempty"`
	Draft        bool                 `json:"draft,omitempty"`
	Summary      string               `json:"summary,omitempty"`
	Category     string               `json:"category,omitempty"`
	Cover        string               `json:"cover,omitempty"`
	Gallery      []string             `json:"gallery,omitempty"`
	Specs        []ProductSpec        `json:"specs,omitempty"`
	Improvements []ProductImprovement `json:"improvements,omitempty"`
	Content      string               `json:"content"`
	DirPath      string               `json:"dirPath"`
	FilePath     string               `json:"-"`
}

type ArticleInput struct {
	DirPath      string               `json:"dirPath"`
	Title        string               `json:"title"`
	Slug         string               `json:"slug,omitempty"`
	Author       string               `json:"author,omitempty"`
	Date         string               `json:"date,omitempty"`
	Tags         []string             `json:"tags,omitempty"`
	Comments     *bool                `json:"comments,omitempty"`
	Weight       int                  `json:"weight,omitempty"`
	Position     string               `json:"position,omitempty"`
	Draft        *bool                `json:"draft,omitempty"`
	Summary      string               `json:"summary,omitempty"`
	Category     string               `json:"category,omitempty"`
	Cover        string               `json:"cover,omitempty"`
	Gallery      []string             `json:"gallery,omitempty"`
	Specs        []ProductSpec        `json:"specs,omitempty"`
	Improvements []ProductImprovement `json:"improvements,omitempty"`
	Content      string               `json:"content"`
}

type ArticleSummary struct {
	Slug     string        `json:"slug"`
	Title    string        `json:"title"`
	Author   string        `json:"author,omitempty"`
	Date     string        `json:"date,omitempty"`
	Tags     []string      `json:"tags,omitempty"`
	Weight   int           `json:"weight,omitempty"`
	Position string        `json:"position,omitempty"`
	Draft    bool          `json:"draft,omitempty"`
	Summary  string        `json:"summary,omitempty"`
	Category string        `json:"category,omitempty"`
	Cover    string        `json:"cover,omitempty"`
	Gallery  []string      `json:"gallery,omitempty"`
	Specs    []ProductSpec `json:"specs,omitempty"`
	DirPath  string        `json:"dirPath"`
}

func ToSummary(a *Article) *ArticleSummary {
	return &ArticleSummary{
		Slug:     a.Slug,
		Title:    a.Title,
		Author:   a.Author,
		Date:     a.Date,
		Tags:     a.Tags,
		Weight:   a.Weight,
		Position: a.Position,
		Draft:    a.Draft,
		Summary:  a.Summary,
		Category: a.Category,
		Cover:    a.Cover,
		Gallery:  a.Gallery,
		Specs:    a.Specs,
		DirPath:  a.DirPath,
	}
}

type MetaData struct {
	Title        string               `toml:"title"`
	Slug         string               `toml:"slug"`
	Author       string               `toml:"author"`
	Date         string               `toml:"date"`
	Tags         []string             `toml:"tags"`
	Comments     *bool                `toml:"comments"`
	Weight       int                  `toml:"weight"`
	Position     string               `toml:"position"`
	Draft        *bool                `toml:"draft"`
	Summary      string               `toml:"summary"`
	Category     string               `toml:"category"`
	Cover        string               `toml:"cover"`
	Gallery      []string             `toml:"gallery"`
	Specs        []ProductSpec        `toml:"specs"`
	Improvements []ProductImprovement `toml:"improvements"`
}

type ImageInfo struct {
	FileName     string    `json:"fileName"`
	OriginalName string    `json:"originalName"`
	FileSize     int64     `json:"fileSize"`
	FileType     string    `json:"fileType"`
	UploadPath   string    `json:"uploadPath"`
	UploadTime   time.Time `json:"uploadTime"`
}
