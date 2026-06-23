package model

type DirInfo struct {
	DirPath      string `json:"dirPath"`
	Title        string `json:"title"`
	DirType      string `json:"dirType"` // page, news, docs
	Layout       string `json:"layout"`
	SortOrder    string `json:"sortOrder"`
	NavOrder     int    `json:"navOrder"`
	ArticleCount int    `json:"articleCount"`
}

type SiteTree struct {
	Title    string      `json:"title"`
	Slug     string      `json:"slug"`
	DirPath  string      `json:"dirPath"`
	Children []*SiteTree `json:"children,omitempty"`
	Articles []*Article  `json:"articles,omitempty"`
	Meta     *DirMeta    `json:"meta,omitempty"`
}
