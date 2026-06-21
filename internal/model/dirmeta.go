package model

type FeatureCard struct {
	Title string `toml:"title" json:"title"`
	Desc  string `toml:"desc" json:"desc"`
	Icon  string `toml:"icon" json:"icon"`
	Link  string `toml:"link" json:"link"`
}

type DirMeta struct {
	SITE_TITLE        string        `toml:"SITE_TITLE" json:"SITE_TITLE"`
	SITE_DESC         string        `toml:"SITE_DESC" json:"SITE_DESC"`
	SITE_THEME        string        `toml:"SITE_THEME" json:"SITE_THEME"`
	NAV_ORDER         int           `toml:"NAV_ORDER" json:"NAV_ORDER"`
	NAV_HIDE          bool          `toml:"NAV_HIDE" json:"NAV_HIDE"`
	DIR_TYPE          string        `toml:"DIR_TYPE" json:"DIR_TYPE"` // page, news, docs
	ARTICLES_PER_PAGE int           `toml:"ARTICLES_PER_PAGE" json:"ARTICLES_PER_PAGE"`
	HERO_TITLE        string        `toml:"HERO_TITLE" json:"HERO_TITLE"`
	HERO_SUBTITLE     string        `toml:"HERO_SUBTITLE" json:"HERO_SUBTITLE"`
	HERO_IMAGE        string        `toml:"HERO_IMAGE" json:"HERO_IMAGE"`
	HERO_CTA_TEXT     string        `toml:"HERO_CTA_TEXT" json:"HERO_CTA_TEXT"`
	HERO_CTA_LINK     string        `toml:"HERO_CTA_LINK" json:"HERO_CTA_LINK"`
	FEATURE_TITLE     string        `toml:"FEATURE_TITLE" json:"FEATURE_TITLE"`
	FEATURES          []FeatureCard `toml:"FEATURES" json:"FEATURES"`
	NEWS_TITLE        string        `toml:"NEWS_TITLE" json:"NEWS_TITLE"`
	NEWS_LINK         string        `toml:"NEWS_LINK" json:"NEWS_LINK"`
	CONTACT_EMAIL     string        `toml:"CONTACT_EMAIL" json:"CONTACT_EMAIL"`
	CONTACT_PHONE     string        `toml:"CONTACT_PHONE" json:"CONTACT_PHONE"`
	CONTACT_ADDRESS   string        `toml:"CONTACT_ADDRESS" json:"CONTACT_ADDRESS"`
	FOOTER_TEXT       string        `toml:"FOOTER_TEXT" json:"FOOTER_TEXT"`
	ICP               string        `toml:"ICP" json:"ICP"`
}

func DefaultDirMeta() DirMeta {
	return DirMeta{
		ARTICLES_PER_PAGE: 10,
	}
}

func MergeDirMeta(parent, child DirMeta) DirMeta {
	result := parent
	if child.SITE_TITLE != "" {
		result.SITE_TITLE = child.SITE_TITLE
	}
	if child.SITE_DESC != "" {
		result.SITE_DESC = child.SITE_DESC
	}
	if child.SITE_THEME != "" {
		result.SITE_THEME = child.SITE_THEME
	}
	if child.NAV_ORDER != 0 {
		result.NAV_ORDER = child.NAV_ORDER
	}
	if child.NAV_HIDE {
		result.NAV_HIDE = child.NAV_HIDE
	}
	if child.DIR_TYPE != "" {
		result.DIR_TYPE = child.DIR_TYPE
	}
	if child.ARTICLES_PER_PAGE != 0 {
		result.ARTICLES_PER_PAGE = child.ARTICLES_PER_PAGE
	}
	if child.HERO_TITLE != "" {
		result.HERO_TITLE = child.HERO_TITLE
	}
	if child.HERO_SUBTITLE != "" {
		result.HERO_SUBTITLE = child.HERO_SUBTITLE
	}
	if child.HERO_IMAGE != "" {
		result.HERO_IMAGE = child.HERO_IMAGE
	}
	if child.HERO_CTA_TEXT != "" {
		result.HERO_CTA_TEXT = child.HERO_CTA_TEXT
	}
	if child.HERO_CTA_LINK != "" {
		result.HERO_CTA_LINK = child.HERO_CTA_LINK
	}
	if child.FEATURE_TITLE != "" {
		result.FEATURE_TITLE = child.FEATURE_TITLE
	}
	if len(child.FEATURES) > 0 {
		result.FEATURES = child.FEATURES
	}
	if child.NEWS_TITLE != "" {
		result.NEWS_TITLE = child.NEWS_TITLE
	}
	if child.NEWS_LINK != "" {
		result.NEWS_LINK = child.NEWS_LINK
	}
	if child.CONTACT_EMAIL != "" {
		result.CONTACT_EMAIL = child.CONTACT_EMAIL
	}
	if child.CONTACT_PHONE != "" {
		result.CONTACT_PHONE = child.CONTACT_PHONE
	}
	if child.CONTACT_ADDRESS != "" {
		result.CONTACT_ADDRESS = child.CONTACT_ADDRESS
	}
	if child.FOOTER_TEXT != "" {
		result.FOOTER_TEXT = child.FOOTER_TEXT
	}
	if child.ICP != "" {
		result.ICP = child.ICP
	}
	return result
}
