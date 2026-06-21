package model

import "time"

type Article struct {
	Slug     string   `json:"slug"`
	Title    string   `json:"title"`
	Author   string   `json:"author,omitempty"`
	Date     string   `json:"date,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Comments bool     `json:"comments"`
	Weight   int      `json:"weight,omitempty"`
	Content  string   `json:"content"`
	DirPath  string   `json:"dirPath"`
	FilePath string   `json:"-"`
}

type ArticleInput struct {
	DirPath  string   `json:"dirPath"`
	Title    string   `json:"title"`
	Slug     string   `json:"slug,omitempty"`
	Author   string   `json:"author,omitempty"`
	Date     string   `json:"date,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Comments *bool    `json:"comments,omitempty"`
	Weight   int      `json:"weight,omitempty"`
	Content  string   `json:"content"`
}

type ArticleSummary struct {
	Slug    string   `json:"slug"`
	Title   string   `json:"title"`
	Author  string   `json:"author,omitempty"`
	Date    string   `json:"date,omitempty"`
	Tags    []string `json:"tags,omitempty"`
	Weight  int      `json:"weight,omitempty"`
	DirPath string   `json:"dirPath"`
}

func ToSummary(a *Article) *ArticleSummary {
	return &ArticleSummary{
		Slug:    a.Slug,
		Title:   a.Title,
		Author:  a.Author,
		Date:    a.Date,
		Tags:    a.Tags,
		Weight:  a.Weight,
		DirPath: a.DirPath,
	}
}

type MetaData struct {
	Title    string   `toml:"title"`
	Slug     string   `toml:"slug"`
	Author   string   `toml:"author"`
	Date     string   `toml:"date"`
	Tags     []string `toml:"tags"`
	Comments *bool    `toml:"comments"`
	Weight   int      `toml:"weight"`
}

type ImageInfo struct {
	FileName     string    `json:"fileName"`
	OriginalName string    `json:"originalName"`
	FileSize     int64     `json:"fileSize"`
	FileType     string    `json:"fileType"`
	UploadPath   string    `json:"uploadPath"`
	UploadTime   time.Time `json:"uploadTime"`
}
