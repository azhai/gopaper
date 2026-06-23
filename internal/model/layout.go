package model

type Region struct {
	Name  string `toml:"name" json:"name"`
	Title string `toml:"title" json:"title"`
}

type Template struct {
	Name    string   `toml:"name" json:"name"`
	Title   string   `toml:"title" json:"title"`
	File    string   `toml:"file" json:"file"`
	Desc    string   `toml:"desc" json:"desc"`
	Regions []Region `toml:"regions" json:"regions"`
}

type LayoutConfig struct {
	Templates []Template `toml:"templates" json:"templates"`
}

func DefaultLayoutConfig() LayoutConfig {
	return LayoutConfig{
		Templates: []Template{
			{
				Name:  "home",
				Title: "首页模板",
				File:  "index.html",
				Desc:  "Hero + 功能 + 新闻 + 联系方式",
				Regions: []Region{
					{Name: "hero", Title: "横幅区"},
					{Name: "features", Title: "功能区"},
					{Name: "news", Title: "新闻区"},
					{Name: "contact", Title: "联系区"},
				},
			},
			{
				Name:  "list",
				Title: "列表模板",
				File:  "list.html",
				Desc:  "文章/新闻列表页",
				Regions: []Region{
					{Name: "main", Title: "主内容区"},
				},
			},
			{
				Name:  "article",
				Title: "文章模板",
				File:  "article.html",
				Desc:  "单篇文章详情页",
				Regions: []Region{
					{Name: "main", Title: "主内容区"},
				},
			},
		},
	}
}
