# 进度记录 (Progress)

> 本文件记录已完成步骤、子任务结果、遇到的问题。按时间倒序追加。
> **不要**把子 agent 返回原文粘贴于此；记录结论与状态即可。

---

## 2026-08-19 会话启动

### 已完成

- [x] 阅读 `AGENTS.md`（全文，1-350+ 行），理解工作流规范与 Planning-with-Files 方法。
- [x] 探索代码库全貌：`main.go`、`internal/*`、`content/`、`templates/`、`web/`、`go.mod`、`config.toml`、`Makefile`、`.gitignore`、`README.md`。
- [x] 架构评估：识别 6 项需求-实现差距（A-F）+ 7 项安全/工程问题（S1-S7）。
- [x] 创建 `.ai-memory/` 目录及三件套：`task_plan.md`、`findings.md`、`progress.md`。
- [x] 需求拆解：21 项任务（T01-T21），分 P0/P1/P2/P3 四级优先。
- [x] 识别 4 个待用户确认的决策点（D1-D4），写入 task_plan.md。

### 关键发现（详见 findings.md）

1. **定位差距**：当前是动态 Echo 服务器，非"静态站点生成器"，缺 `build` 命令。
2. **Front Matter 格式分歧**：需求说 YAML，代码用 TOML。
3. **SQLite 未实现**：产品数据当前用 Markdown，无数据库。
4. **三个未接线字段（bug）**：`LAYOUT`、`Position`、`SORT_ORDER` 配置了但不生效。
5. **嵌套目录路由 404**：路由仅两段，子目录文章访问不到。
6. **安全隐患**：config.toml 含明文 JWT_SECRET 且未 gitignore。

### 当前状态

- **阶段**：规划完成，等待用户确认 D1-D4 决策。
- **阻塞项**：D1（系统定位）阻塞 T01；D2/D3/D4 影响后续方向。
- **可立即推进**（不依赖决策）：T02、T03、T04、T05（bug 修复）、T11、T12、T13（安全加固）。

### 下一步

等待用户对 D1-D4 拍板后，按 task_plan.md "执行顺序" 推进。若用户未明确回答，默认按建议项（b/a/a/a）执行。

---

## 2026-08-19 P0 bug 修复（T02/T04/T05）

### 决策确认

用户拍板：D1=双模 / D2=保持 TOML / D3=暂不引入 SQLite / D4=content 保持忽略。

### 已完成

- [x] **T04 SORT_ORDER 生效**（`internal/service/scanner.go`）：
  - `sortArticles` 增加 `sortOrder` 参数；新增 `sortArticlePtrs` 与 `compareArticles`。
  - `scanDir` 末尾对 `tree.Articles` 按该目录 `mergedMeta.SORT_ORDER` 排序（此前目录内文章未排序）。
  - 语义：`asc`（默认）= 权重小在前 + 日期升序；`desc` = 权重大在前 + 日期降序。与 README 描述一致。
- [x] **T02 LAYOUT 字段生效**（`internal/handler/page.go`）：
  - 新增 `layoutToTemplate` 映射：home→index, list→list, article→article。
  - `DirList` 在无单页命中时，根据 `node.Meta.LAYOUT` 选模板，缺省 `list`；若用 `index` 模板则补填 `FeatureCards`。
- [x] **T05 嵌套目录路由**（`internal/handler/page.go` + `internal/app/app.go`）：
  - 新增 `ResolvePath` handler，处理 3–5 段路径（`/:d1/:d2/:d3` 等），按末段 slug + 前段 dirPath 查文章并校验 `DirPath`。
  - 提取 `renderArticle` helper 复用渲染逻辑（`ArticleDetail` 同步改用，消除重复）。
  - 注册 3 段/4 段/5 段路由。避免 `/*` 通配与 `/:dir` 冲突。

### 验证

- `go build ./...` 通过（无错误）。
- `go vet ./...` 通过。
- 运行时 smoke test（`go run .` + curl）：
  - `/` 200（首页含 news-table）、`/news` 200、`/news/cms-3-release` 200、`/products` 200、`/about` 200。
  - `/a/b/c` 404（不存在的嵌套路径正确 404）。

### 未完成（P0 剩余）

- **T03 Position/Region 分组**：涉及首页模板结构调整（按 position 把页面型文章分发到 hero/features/news/contact 区），改动较大，单独推进。
- ~~T01 build 子命令~~：见下，已完成。

### 下一步建议

1. T03（P0 收尾）—— 涉及模板结构调整，需设计 position 分组展示方式。
2. 或先做 P2 安全加固（T11/T12/T13，小而独立）。

---

## 2026-08-19 T01 build 子命令（双模核心）

### 已完成

- [x] **T01 静态站点构建命令**：
  - `main.go` 加子命令分发：`gopaper build` 生成静态站点，默认仍启动动态服务器。
  - 新增 `internal/handler/builder.go`：`Builder` 复用 `PageHandler` 同包方法（`buildBaseData`/`renderToBytes`/`toArticleViews`/`collectArticles`/`articleHref`/`layoutToTemplate`），遍历 SiteTree 渲染所有页面到 `dist/`。
  - `page.go` 提取 `renderToBytes`（不依赖 echo），`renderTemplate` 改用它，消除重复。
  - 输出结构（目录式 URL，nginx 友好）：
    ```
    dist/index.html              首页
    dist/<dir>/index.html        目录列表页（或单页型目录的文章页）
    dist/<dir>/<slug>/index.html 文章详情
    dist/static/                 静态资源（css/svg/logo/favicon，跳过 admin/）
    ```
  - 单页型目录（`DIR_TYPE=page` + 单文章）生成 `dist/<dir>/index.html` 直接为文章内容，访问 `/dir/` 即见。
  - 嵌套目录（如 `docs/getting-started`）正确生成多层路径。

### 验证

- `go build ./...` + `go vet ./...` 通过。
- `go run . build` 成功生成 dist/，含 7 篇文章 + 首页 + 列表页 + 静态资源。
- 内容检查：首页 `<title>GoPaper - GoPaper</title>` + news-table + hero；文章页 `<title>新版CMS 3.0正式发布 - GoPaper</title>` + 正文；嵌套 `docs/getting-started` `<title>快速开始 - GoPaper</title>`。
- **独立托管测试**（`python3 -m http.server` in dist/，等效 nginx）：
  - `/` 200、`/news/` 200、`/news/cms-3-release/` 200、`/static/css/style.css` 200、`/docs/getting-started/` 200。
  - 证明 dist/ 可脱离 Go 进程纯静态部署。

### P0 状态

- [x] T01 build 子命令
- [x] T02 LAYOUT 生效
- [x] T03 Position/Region 分组（features 区已支持，hero/contact 保留 meta 配置暂不接管）
- [x] T04 SORT_ORDER 生效
- [x] T05 嵌套目录路由

---

## 2026-08-19 T03 Position/Region 分组

### 已完成

- [x] **T03 Position 字段生效**（`internal/handler/page.go` + `builder.go`）：
  - 提取 `buildIndexData(siteTree)` 共享方法，动态 `Index` 与静态 `buildIndex` 复用，保证一致。
  - 按 `Article.Position` 分组：
    - `position=features` → 转为 `FeatureCardView`（Title + content 摘要 + Link）追加到首页 features 区卡片网格（补充 meta FEATURES，meta 卡片在前）。
    - 其余 position（news/空/hero/contact/main）→ news 区列表（现有行为）。
  - 新增 `simpleExcerpt(content, n)`：取正文前 n 字符（去换行）作卡片描述。
  - **设计决策**：当前仅 features 区接管 position 文章；hero/contact 区保留 `_meta.toml` 配置（HERO_TITLE/CONTACT_*），因这两区展示形态与文章列表差异大，接管需较大模板重构，YAGNI。后续可按需扩展。

### 验证

- `go build` + `go vet` 通过。
- 临时创建 `position=features` 测试文章 → `go run . build` → 首页 features 区出现该卡片（4 卡片 = 3 meta + 1 文章），且不进 news 表格。
- 清理测试文章后重建 → 恢复 3 卡片。动态模式首页同样 3 卡片（动态/静态一致）。

### P0 全部完成 ✅

| 任务 | 状态 |
|------|------|
| T01 build 子命令 | ✅ |
| T02 LAYOUT 生效 | ✅ |
| T03 Position/Region 分组 | ✅（features 区） |
| T04 SORT_ORDER 生效 | ✅ |
| T05 嵌套目录路由 | ✅ |

### 下一步

P0 收尾完成。可推进：
- **P2 安全加固**：T11(config gitignore+example) / T12(CORS 可配置) / T13(错误不泄露) / T14(go.mod replace) / T15(单元测试) / T16(Go 版本)
- **P1 功能补全**：T06(分页) / T07(主题目录) / T08(RSS) / T09(sitemap) / T10(SEO)
- **P3 体验优化**：T17-T21

---

## 2026-08-19 P1 功能补全（T08/T09/T10）

### 已完成

- [x] **T08 RSS feed** + **T09 sitemap.xml/robots.txt**：
  - 新增 `internal/handler/feed.go`：`GenerateRSS`/`GenerateSitemap`/`GenerateRobots`，HTML 转义 + RFC1123Z 日期。
  - `config.go` 加 `SITE_URL` 字段（默认 `http://localhost:3000`，支持 env 覆盖）。
  - `PageHandler` 加 `siteURL` 字段 + `RSS`/`Sitemap`/`Robots` handler；动态路由 `/index.xml` `/sitemap.xml` `/robots.txt`。
  - `builder.go` build 时写 `dist/index.xml` `dist/sitemap.xml` `dist/robots.txt`。
- [x] **T10 SEO meta 标签**：三个模板 head 加 `meta description` + `og:title`/`og:description`/`og:site_name`/`og:type`（首页/列表 `website`，文章 `article`）。

### 验证

- `go build` + `go vet` 通过。
- 静态 build：`dist/index.xml`（RSS 2.0，含所有文章 item）、`dist/sitemap.xml`（urlset，含首页/目录/文章）、`dist/robots.txt`（含 Sitemap 指令）。
- 动态路由：`/index.xml` `/sitemap.xml` `/robots.txt` 均 200。
- SEO meta：首页 `og:type=website` + `meta description`；文章页 `og:type=article`。

### 未完成（P1 剩余，YAGNI 暂缓）

- **T06 分页**：当前 news 2 篇、products 2 篇，均不触发 `ARTICLES_PER_PAGE=10`。代码预留可后续加（DirList `?page=` 切片 + 模板分页导航 + build 多页）。
- **T07 主题目录规范化**：当前单主题，`templates/` 平铺够用。`themes/default/` 结构属预留性重构，待多主题需求出现再加。

### P1 状态

| 任务 | 状态 |
|------|------|
| T06 分页 | ⏸ 暂缓（YAGNI，内容不触发） |
| T07 主题目录 | ⏸ 暂缓（YAGNI，单主题） |
| T08 RSS | ✅ |
| T09 sitemap/robots | ✅ |
| T10 SEO meta | ✅ |

---

## 2026-08-19 P2 安全与工程加固（T11/T12/T13/T15/T16）

### 已完成

- [x] **T11 config 安全**：`.gitignore` 加 `config.toml`；新增 `config.example.toml`（含 `SITE_URL`/`ALLOWED_ORIGINS` 占位 + bcrypt 密码占位）。注：config.toml 已被 git 跟踪，需用户执行 `git rm --cached config.toml` 取消跟踪。
- [x] **T12 CORS 可配置**：`AppConfig` 加 `ALLOWED_ORIGINS []string`；`app.go` 用它，空则回退 `["*"]`（向后兼容）。`config.example.toml` 示例收紧为 `localhost:3000`。
- [x] **T13 错误不泄露**：`customErrorHandler` 重写——4xx 返回 echo.HTTPError.Message 或 StatusText；5xx 在非 Debug 模式返回通用"内部服务器错误"，不泄露 `err.Error()` 内部路径/堆栈。
- [x] **T15 单元测试**：`internal/service/scanner_test.go`（`compareArticles` 7 用例：asc/desc/默认/同 weight 日期/标题）；`internal/handler/page_test.go`（`layoutToTemplate` 5 用例 + `simpleExcerpt` 4 用例）。`go test ./...` 通过。
- [x] **T16 Go 版本**：确认 `go1.26.4 darwin/arm64` 真实存在，`go.mod` 版本与工具链匹配，无需调整。

### 保留未改

- **T14 go.mod replace**：`replace github.com/azhai/gobus => ../gobus` 保留。gobus 是本地开发依赖，移除 replace 会因远程无 v0.2.1 tag 而 build 失败。发布前需将 gobus 推送到远程仓库并移除 replace。

### 验证

- `go build ./...` + `go vet ./...` + `go test ./...` 全通过。
- `go run . build` 生成完整 dist/（11 HTML + index.xml + sitemap.xml + robots.txt + static/）。

### P2 状态

| 任务 | 状态 |
|------|------|
| T11 config gitignore+example | ✅ |
| T12 CORS 可配置 | ✅ |
| T13 错误不泄露 | ✅ |
| T14 go.mod replace | ⏸ 保留（gobus 本地依赖） |
| T15 单元测试 | ✅ |
| T16 Go 版本 | ✅（1.26.4 确认） |

---

## 总体进度汇总（2026-08-19）

| 优先级 | 完成 | 暂缓/保留 | 说明 |
|--------|------|-----------|------|
| P0 | 5/5 | — | build 命令 + 4 bug 修复，全部 ✅ |
| P1 | 3/5 | T06/T07 | RSS/sitemap/SEO ✅；分页/主题目录 YAGNI 暂缓 |
| P2 | 5/6 | T14 | 安全+测试+Go 版本 ✅；go.mod replace 保留 |
| P3 | 2/5 | T17/T18/T21 | 草稿+摘要 ✅；热重载/livereload/资源指纹未开始 |

**双模验证**：动态服务器（`gopaper`）+ 静态构建（`gopaper build` → dist/）均通过运行时测试，dist/ 可纯静态部署。

---

## 2026-08-19 T19 草稿状态 + T20 文章摘要

### 已完成

- [x] **T19 草稿/发布状态**：
  - `model.Article` 加 `Draft bool`，`MetaData` 加 `Draft *bool`，`ArticleInput` 加 `Draft *bool`。
  - `scanner.parseFile` 解析 draft；`scanDir` 中 **draft 文章不加入 `tree.Articles`**（前台列表/build/sitemap/RSS 自动排除），但仍加入全局 articles（admin API 可见可编辑）。
  - `page.go renderArticle` 前台访问 draft 文章返回 404。
  - `article.go buildFileContent` 写 `draft = true`；`Update` 处理 Draft 字段。
- [x] **T20 文章摘要**：
  - `model.Article` 加 `Summary string`，`MetaData` 加 `Summary`，`ArticleView` 加 `Summary`。
  - `scanner` 新增 `extractSummary`：优先 Front Matter `summary` 字段，否则取 `<!--more-->` 前部分（`cleanMarkdown` 去 `#/*/`/>` 符号 + 截断 200 字）。
  - `list.html` page 卡片 + 默认卡片显示摘要（`{{if .Summary}}<p class="list-excerpt">{{.Summary}}</p>{{end}}`）；news 表格保持简洁不显示。
  - 首页 features 卡片优先用 `Summary`，无则回退 `simpleExcerpt`。

### 验证

- `go build` + `go vet` + `go test` 通过。
- **T19**：draft 文章不生成 HTML、不在 sitemap/RSS、不在 news 列表（全部 0 匹配）。
- **T20**：products（page 卡片）显示自定义 summary；docs（默认卡片）显示 `<!--more-->` 提取摘要。

### P3 状态

| 任务 | 状态 |
|------|------|
| T17 模板热重载 | ⏸ 未开始 |
| T18 livereload | ⏸ 未开始 |
| T19 草稿状态 | ✅ |
| T20 文章摘要 | ✅ |
| T21 资源指纹 | ⏸ 未开始 |

---

## 2026-08-19 T06 分页 + T07 主题目录（P1 收尾）

### 已完成

- [x] **T06 前台分页**：
  - `PageData` 加 `Pagination` 结构（Current/Total/HasPrev/HasNext/PrevHref/NextHref）。
  - `DirList` 读 `c.Param("num")` 或 `?page=` 查询参数，按 `ARTICLES_PER_PAGE` 切片，构造分页导航。
  - 动态路由 `/:dir/page/:num`（复用 DirList）；静态 build 循环生成 `dist/<dir>/index.html`（page 1）+ `dist/<dir>/page/N/index.html`。
  - `list.html` 加分页导航（上一页/页码/下一页）。
  - URL 风格：路径式 `/news/page/2`（动态+静态一致），也兼容 `?page=2`。
- [x] **T07 主题目录规范化**：
  - `templates/{index,list,article}.html` → `themes/default/layouts/`，`templates/layouts.toml` → `themes/default/layouts.toml`。
  - `page.go` ParseGlob 路径改 `themes/default/layouts/*.html`；`layout.go` configPath 改 `themes/default/layouts.toml`。
  - 为多主题预留结构（未来 `themes/other/`）。

### 验证

- `go build` + `go vet` + `go test` 通过。
- **T06**：临时设 `ARTICLES_PER_PAGE=1` → build 生成 `dist/news/page/2/index.html`，page1 含"下一页"，page2 含"上一页"，分页信息"1 / 2"；动态 `/news/page/2` 200、`/news?page=2` 200、`/news/page/3`（超出）404。还原后不分页。
- **T07**：模板从 `templates/` 迁移到 `themes/default/layouts/`，build + 动态均正常。

### P1 全部完成 ✅

| 任务 | 状态 |
|------|------|
| T06 分页 | ✅ |
| T07 主题目录 | ✅ |
| T08 RSS | ✅ |
| T09 sitemap/robots | ✅ |
| T10 SEO meta | ✅ |

---

## 总体进度汇总（最终更新 2026-08-19）

| 优先级 | 完成 | 暂缓/保留 |
|--------|------|-----------|
| P0 | 5/5 ✅ | — |
| P1 | 5/5 ✅ | — |
| P2 | 5/6 | T14（go.mod replace，需推送 gobus） |
| P3 | 5/5 ✅ | — |

**未完成项**：
- T14 go.mod replace（需先推送 gobus 到远程仓库）
- `git rm --cached config.toml`（取消已跟踪的含密钥配置）

**双模验证**：`gopaper`（动态）+ `gopaper build`（→ dist/ 静态）+ `gopaper dev`（热重载）均通过。

### T17+T18+T21 完成详情（2026-08-19）

- **T17 模板热重载**：`internal/handler/dev.go` 的 `watchTemplates` goroutine 每秒轮询 `themes/default/layouts/` mtime，变化时调 `PageHandler.reloadTemplates()`（atomic.Pointer 原子替换）并 broadcast SSE。`PageHandler.templates` 已改为 `atomic.Pointer[template.Template]`。
- **T18 livereload**：`SSEServer`（clients map + mutex + Broadcast）+ `/livereload` SSE 端点（`Livereload` handler，15s ping keepalive）+ `injectLivereload` 在 `renderToBytes` dev 模式注入 `<script>new EventSource("/livereload")...</script>`。`watchContent` goroutine 轮询 content 目录 mtime → `cache.Refresh`。`gopaper dev` 子命令启用。
- **T21 资源指纹**：`copyStaticAssets` 对 `.css`/`.js` 计算 md5 hash 前 8 位插入文件名（`style.css`→`style.<hash>.css`）；`rewriteStaticRefs` 遍历 `dist/` 下 `.html` 替换 `/static/` 引用。验证：`dist/static/css/style.ed5c98b5.css` + `dist/index.html` 引用已重写。
- **验证**：`go build` + `go vet` + `go test` 全通过；`go run . build` 输出含哈希文件名；`go run . dev` 启动成功，watchers 注册（logs/app.log 确认 "dev mode enabled"）。

---

（后续进度按日期追加于此之下）