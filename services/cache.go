package services

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/azhai/gopaper/models"

	"log/slog"

	"github.com/VictoriaMetrics/fastcache"
)

type CacheVault struct {
	store     *fastcache.Cache
	mu        sync.RWMutex
	available bool
	logger    *slog.Logger
	snapshot  *CacheSnapshot
}

type CacheSnapshot struct {
	Articles map[string]*models.Article
	SiteTree *models.SiteTree
	DirMetas map[string]models.DirMeta
	BuiltAt  time.Time
}

func NewCacheVault(size int, logger *slog.Logger) *CacheVault {
	return &CacheVault{
		store:     fastcache.New(size),
		available: true,
		logger:    logger,
		snapshot: &CacheSnapshot{
			Articles: make(map[string]*models.Article),
			DirMetas: make(map[string]models.DirMeta),
		},
	}
}

func (cv *CacheVault) GetArticle(slug string) (*models.Article, bool) {
	cv.mu.RLock()
	defer cv.mu.RUnlock()

	if cv.snapshot == nil {
		return nil, false
	}
	article, ok := cv.snapshot.Articles[slug]
	return article, ok
}

func (cv *CacheVault) GetSiteTree() (*models.SiteTree, bool) {
	cv.mu.RLock()
	defer cv.mu.RUnlock()

	if cv.snapshot == nil {
		return nil, false
	}
	return cv.snapshot.SiteTree, cv.snapshot.SiteTree != nil
}

func (cv *CacheVault) GetArticleList(dirPath string) ([]*models.Article, bool) {
	cv.mu.RLock()
	defer cv.mu.RUnlock()

	if cv.snapshot == nil {
		return nil, false
	}
	var articles []*models.Article
	for _, a := range cv.snapshot.Articles {
		if a.DirPath == dirPath {
			articles = append(articles, a)
		}
	}
	return articles, true
}

func (cv *CacheVault) GetAllArticles() []*models.Article {
	cv.mu.RLock()
	defer cv.mu.RUnlock()

	if cv.snapshot == nil {
		return nil
	}
	articles := make([]*models.Article, 0, len(cv.snapshot.Articles))
	for _, a := range cv.snapshot.Articles {
		articles = append(articles, a)
	}
	return articles
}

func (cv *CacheVault) GetDirMeta(dirPath string) (models.DirMeta, bool) {
	cv.mu.RLock()
	defer cv.mu.RUnlock()

	if cv.snapshot == nil {
		return models.DirMeta{}, false
	}
	meta, ok := cv.snapshot.DirMetas[dirPath]
	return meta, ok
}

func (cv *CacheVault) GetDirType(dirPath string) string {
	meta, ok := cv.GetDirMeta(dirPath)
	if !ok {
		return ""
	}
	return meta.DIR_TYPE
}

func (cv *CacheVault) GetDirs() []models.DirInfo {
	cv.mu.RLock()
	defer cv.mu.RUnlock()

	if cv.snapshot == nil || cv.snapshot.SiteTree == nil {
		return nil
	}

	var dirs []models.DirInfo
	collectDirs(cv.snapshot.SiteTree, &dirs)
	return dirs
}

func collectDirs(tree *models.SiteTree, dirs *[]models.DirInfo) {
	if tree == nil {
		return
	}
	if tree.DirPath != "" && tree.Meta != nil {
		*dirs = append(*dirs, models.DirInfo{
			DirPath:      tree.DirPath,
			Title:        tree.Meta.SITE_TITLE,
			DirType:      tree.Meta.DIR_TYPE,
			Layout:       tree.Meta.LAYOUT,
			SortOrder:    tree.Meta.SORT_ORDER,
			NavOrder:     tree.Meta.NAV_ORDER,
			NavHide:      tree.Meta.NAV_HIDE,
			ArticleCount: len(tree.Articles),
		})
	}
	for _, child := range tree.Children {
		collectDirs(child, dirs)
	}
}

func (cv *CacheVault) Refresh(ctx context.Context, scanner *Scanner) error {
	siteTree, articles, dirMetas, err := scanner.ScanAll(ctx)
	if err != nil {
		cv.logger.Error("scan failed during cache refresh", "error", err)
		return err
	}

	articleMap := make(map[string]*models.Article, len(articles))
	for i := range articles {
		articleMap[articles[i].Slug] = &articles[i]
	}

	newSnapshot := &CacheSnapshot{
		Articles: articleMap,
		SiteTree: siteTree,
		DirMetas: dirMetas,
		BuiltAt:  time.Now(),
	}

	cv.mu.Lock()
	cv.snapshot = newSnapshot
	cv.available = true
	cv.mu.Unlock()

	cv.persistToFastCache(newSnapshot)

	cv.logger.Info("cache refreshed",
		"articleCount", len(articles),
		"builtAt", newSnapshot.BuiltAt,
	)
	return nil
}

func (cv *CacheVault) Invalidate() {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	cv.snapshot = &CacheSnapshot{
		Articles: make(map[string]*models.Article),
		DirMetas: make(map[string]models.DirMeta),
	}
	cv.store.Reset()
	cv.logger.Info("cache invalidated")
}

func (cv *CacheVault) IsAvailable() bool {
	cv.mu.RLock()
	defer cv.mu.RUnlock()
	return cv.available
}

func (cv *CacheVault) GetBuiltAt() time.Time {
	cv.mu.RLock()
	defer cv.mu.RUnlock()
	if cv.snapshot == nil {
		return time.Time{}
	}
	return cv.snapshot.BuiltAt
}

func (cv *CacheVault) persistToFastCache(snapshot *CacheSnapshot) {
	data, err := json.Marshal(snapshot)
	if err != nil {
		cv.logger.Error("marshal cache snapshot", "error", err)
		cv.available = false
		return
	}
	cv.store.Set([]byte("snapshot"), data)
}
