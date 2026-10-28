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
---

## 管理后台重设计（2026-10-28）

- **信息架构**：工作台 / 内容 / 首页装修 / 媒体库 / 站点设置；旧路由（articles、pages、images、layout）重定向或删除，新增 `settings?tab=sections`。
- **新增页面**：`dashboard.ts`（待处理草稿、最近内容、网站体检、栏目概览）、`content.ts`（选项卡 + 搜索 + 栏目筛选）、`media.ts`（上传/网格/灯箱/分页）、`homepage.ts`（首页装修 + 结构预览 + 脏检查）、`settings.ts`（基本信息/联系方式/页脚/栏目/高级）。
- **编辑器**：Markdown 工具栏、媒体插入、预览、草稿/发布双栏、页面首页位置、移动端操作栏、⌘/Ctrl+S。
- **公共组件**：`ui.ts`（EmptyState/Field/Toggle/ContentStatus 等）、`mediaPicker.ts`、`dirtyGuard.ts`（应用内未保存提示）；`app.ts` 设置 `m.route.prefix='/admin'`。
- **后端**：`DirInfo.NavHide` 暴露栏目隐藏状态；栏目创建/更新 API；文章完整更新与 slug 生成；`article_test.go`。
- **修复**：`content/_meta.toml` 中联系/新闻等标量原本位于 `[[FEATURES]]` 之后被解析进最后一张卡片，已移到数组前；草稿深链选项卡未激活；媒体分页被上传后重置；应用内导航不拦截未保存修改。
- **验证**：端口 3100 实测登录（错误密码内联提示）、工作台 6/6 体检、草稿选项卡、编辑器脏检查弹窗、首页装修数据加载与保存、栏目设置、媒体空态/上传/灯箱/删除、文章创建/删除；`gofmt`（改动文件）、`go vet ./...`、`go test ./...`、`pnpm run build` 通过。

## 可选产品目录（2026-10-28）

- **数据模型**：新增产品目录类型 `DIR_TYPE=product`；前置元数据支持 `category`（最多两级，用 `/` 分级）、`cover`、`[[specs]]`（name/price/sku/jd_url/taobao_url）、`[[improvements]]`（date/content/visible）。
- **SKU 约束**：产品至少一个规格，每个规格必须有 SKU 且在单品内唯一（服务端 + 后台表单双重校验）。
- **分类**：两级 Nested Set（lft/rgt）由产品分类派生，父级筛选包含子级；后台一级/二级分类分开填写，保存时合并为 `一级/二级`。
- **前台**：`/products` 分类筛选（父级含子级）、产品卡片、详情页规格单选、SKU 展示、京东/淘宝外链、localStorage 购物车与批量结算；详情页「听劝记录」按日期倒序，只展示 `visible=true` 的记录。
- **后台**：编辑器商品信息卡（分类/封面/规格/听劝记录）、工作台新增「新增产品」入口与产品计数、栏目类型增加「产品目录」、内容列表产品选项卡。
- **示例内容**：企业 CMS、技术咨询、品牌官网模板（三配色 SKU）。
- **验证**：`gofmt`、`go vet ./...`、`go test ./...`（含新增 `TestValidateProductFields`）、`pnpm run build`、`go run . build` 通过；端口 3100 实测分类父子筛选、规格切换、购物车 SKU、听劝可见性、后台编辑回填，以及 API 创建/读取/删除产品。
- **构建修复**：`copyStaticAssets` 构建前清空 `dist/static`，避免增量构建残留旧哈希文件导致 HTML 随机引用旧 CSS/JS。

## 站点主题设计器（2026-10-28）

- **存储**：`content/_themes.toml`（`model.ThemeFile`），最多 20 套；每套 = 布局（容器宽度/圆角/导航与横幅风格/吸顶/阴影）+ 7 个配色 + 字体（正文/标题/字号 14–20px）。扫描器跳过 `_themes.toml`，不进内容索引。
- **生效规则**：`enabled_at`（分钟精度，本地时区）≤ 现在 的最后一套为当前主题；> 现在 的最早一套是「下一次切换」。立即启用＝把时间设为现在并挪到文件末尾，同分钟连点也能确定性地换装。
- **渲染**：`Theme.CSS()` 只输出白名单枚举与校验过的 hex 派生的 CSS 变量，注入 `<style>` 到 index/list/article 三套布局的 head（`PageData.ThemeCSS`，`template.CSS` 免转义）；动态服务与 `gopaper build` 走同一条 `buildBaseData`，静态产物一致。
- **样式令牌化**：`web/public/css/style.css` 的 `:root` 增加约 40 个默认变量，取值复刻原金色外观，未保存主题时视觉零变化（构建产物 `<style>:root` 计数 0）；阴影改为 `--shadow-rgb` × `--shadow-a`。
- **API**：`GET/POST /api/admin/themes`、`POST /themes/preview`（草稿实时预览，避免前端重算混色）、`PUT/DELETE /themes/:id`、`POST /themes/:id/activate`；422/409/404 错误码与字段级提示。
- **后台**：`routes/themes.ts` 设计器＝状态卡（现在生效/下一次切换）+ 编辑器（布局下拉、取色器＋hex 输入、字体与字号）+ 主题卡片（缩略色带、生效中/定时标记、编辑/立即启用/复制/删除）+ 预设套用 + iframe `srcdoc` 真站预览（脏检查与离开拦截）。
- **验证**：`TestPickThemeByTime`、`TestImmediateActivationWinsSameMinute`、`TestThemeForgeLifecycle`、`TestThemeForgeValidation`、`TestThemeCSS*`；`gofmt`/`go vet ./...`/`go test ./...`/`pnpm run build`/`go run . build` 全通过。端口 3100（临时 content 副本）实测：五类页面注入同一主题变量、未来定时不提前生效、同分钟二次启用立即换色、422/409 校验、预览接口返回草稿 CSS、iframe 内 `--c-primary` 生效且 `/static/css/style.css` 正常加载。

## 前台改用 Alpine.js 3.17.5（2026-10-28）

- **本地化**：`npm pack alpinejs@3.17.5` 取官方 `dist/cdn.min.js`（55 KB，`Alpine.version` 3.17.5）存为 `web/public/js/alpine.min.js`；三套布局删除 jsDelivr `3.x.x` CDN，改为本地 `/static/js/alpine.min.js`，构建哈希与引用重写正常（`alpine.min.<hash>.js`）。
- **加载顺序**：`shop.js` 与 `alpine.min.js` 都在 head 用 `defer`，且 shop.js 在前——只注册 `alpine:init` 监听，确保 store/组件在 Alpine 启动前完成注册（CDN 构建是 `window.Alpine` + `queueMicrotask(start)`）。
- **重写 `shop.js`**：`Alpine.store('cart')`（items/selected/open/status、count/fabLabel、add/remove/toggle、platformCount、checkout、copyChecklist、localStorage 持久化与 URL 白名单校验不变）；`Alpine.data` 三个组件——`cartDrawer`（焦点还原/首焦点关闭键/Tab 循环/Escape/滚动锁）、`productBuy`（规格切换、价格、加入购物车 1.5s 反馈）、`productFilter`（父级含子级、`?category=` 深链、空态计数）。
- **模板**：购物车抽屉 `x-for` 渲染明细与勾选、下单链接计数/禁用、状态文本；详情页规格面板 `x-show` + `x-cloak`（首屏仅第一块、无 JS 时仍由 `[x-cloak]` 隐藏非首块），radio `@change` 切换；列表筛选用 `x-show` + `:class`/`:aria-pressed`，分类值仍从 `data-*` 读取，避免向 JS 表达式注入模板内容。
- **样式**：`style.css` 增加 `[x-cloak]{display:none!important}`。
- **验证**：`gofmt`（改动文件）/`go vet ./...`/`go test ./...` 通过；端口 3100（真实 content）浏览器实测：Alpine 3.17.5 本地加载 200、规格切换价格 ¥1,280→¥1,580、加购 SKU 与持久化、抽屉焦点/Shift+Tab 首尾循环/Escape 关闭并解锁滚动、勾选联动京东/淘宝计数与禁用、复制失败回退提示、空购物车状态、父级/子级分类筛选与深链回填、移动菜单；`gopaper build`（临时目录）产物哈希引用正确、无 CDN 引用。

## Alpine 内嵌二进制与构建产物刷新（2026-10-28）

- **版本核对**：`npm view alpinejs version` 与 jsDelivr 均为 3.17.5；重新下载官方 `dist/cdn.min.js`（SHA-256 `b4036e9e…9b522e`）与 `web/public/js/alpine.min.js` 完全一致。
- **内嵌方式**：`web/embed.go` 的 `//go:embed all:public` 已将 `shop.js`、`alpine.min.js`、CSS、后台构建包一起打进二进制；`/static` 与 `/admin` 由 `web.PublicFS` 提供。新增 `web/embed_test.go` 锁定两个站点脚本存在且 Alpine 为 3.17.5。
- **内嵌验证**：在只有 `content`、`themes` 的临时目录运行编译后的二进制，`/static/js/alpine.min.js`（SHA 一致）、`/static/js/shop.js`、`/admin/` 均 200，说明运行期不依赖 `web/public` 目录。
- **全量替换 CDN**：用当前模板重新 `gopaper build`，`dist/` 内 12 个仍引用 jsDelivr 的旧页面全部消失，所有页面只引用 `/static/js/alpine.min.<hash>.js` 与 `/static/js/shop.<hash>.js`；动态服务页面引用未哈希的 `/static/js/alpine.min.js` + `shop.js`。
- **遗留提示**：`web/public/{app.js,index.js,icons.js,types.js,components/,routes/,services/}` 是 `tsc`（tsconfig `outDir: ../public`）的旧输出，未被任何页面引用；如需彻底清理，可将 tsconfig 改为 `noEmit` 并删除这些文件。

## 管理后台迁至 Alpine.js 与旧工具链清理（2026-10-28）

- **后台静态化**：8 个页面（登录/工作台/内容/编辑器/首页装修/媒体库/主题设计/站点设置）全部改为 `web/public/admin/pages/*.html` + `web/public/admin/js/pages/*.js` 的 Alpine 组件；外壳、路由、图标、store、API 全在 `web/public/admin/{index.html,admin.css,js/}`，无构建步骤，`//go:embed all:public` 直接打包。
- **删除**：`web/admin/`（TS 源码、package.json、pnpm-lock/workspace、tsconfig、vite 配置与 node_modules）、`web/public/admin/assets/`（旧 Vite 产物）、`web/public/{app.js,index.js,icons.js,types.js,components/,routes/,services/}`（旧 tsc 输出），以及根目录误生成的 `package.json`/`pnpm-lock.yaml`/空 `node_modules`。
- **Makefile**：移除 `ADMIN_DIR`、`front` 目标与 `one/build/dev` 的 pnpm 步骤；`clean` 只删 `bin/`、`tmp/`，不再删 `web/public/admin`。`.air.toml` 去掉已不存在的 `web/admin/node_modules`。
- **代码高亮**：站点与后台预览统一用 `microlight.js`（`window.highlightCode`），服务端不再依赖 highlighting 扩展。
- **测试**：`web/embed_test.go` 增加后台入口/脚本存在性断言；`go vet ./...`、`go test ./...`、`make one`、`./bin/gopaper build` 通过（`gofmt -l` 仅剩历史遗留的 `internal/handler/dev.go` 缺尾换行）。
- **运行验证（端口 3100，独立二进制）**：`/admin/`、`admin.css`、`js/app.js`、`js/pages/settings.js`、`pages/settings.html`、`static/js/alpine.min.js` 均 200；HTML 无 `admin/assets`/Mithril 残留；登录后 `/api/admin/{articles,settings,dirs,themes}` 正常返回；`make clean` 后后台静态文件仍在。测试服务已停止。

## 示例数据替换为真实数码配件（2026-10-28）

- **产品（4 款，含真实参数/SKU/参考价）**：闪迪 CZ48 U盘（32/64/128/256GB，¥34/57/107/239）、小米移动电源3（10000mAh ¥79 / 20000mAh ¥109）、索尼 WH-CH520 头戴耳机（五色 ¥279）、绿联 CD244 65W 充电器（单头 ¥99 / 套装 ¥128）。每款都有分类、封面、规格 SKU、京东/淘宝搜索链接和听劝改进记录。
- **图片**：四张真实商品图经 `sips`/PIL 转码裁剪后放入 `uploads/`，产品列表与详情页全部 200 加载。
- **栏目与首页**：`_meta.toml` 改为数码配件店（站名、Hero、三大品类、联系方式）；产品/新闻 `SORT_ORDER=desc`（新品在前）。首页“上新动态”新增 `indexArticles`，只取 `NEWS_LINK`（/news）子树，避免产品/指南/关于混进新闻表；`TestIndexArticles` 覆盖子树命中与回退。
- **内容**：重写 `about.md` 为配件店介绍，`docs/getting-started.md` 改为《选购与售后指南》（容量换算、Wh 乘机规则、充电协议、下单与保修）；新增 3 篇真实上新/补货新闻（CZ48 补货、WH-CH520 五色、CD244 上架）；移除 2016 年 PHP 示例文（备份于 `/tmp/gopaper-content-backup-20261028/`）。
- **验证（3100 独立二进制，未触碰 3000）**：产品列表 4 卡、分类父子筛选数据齐全；四张详情页 SKU/价格/京东/淘宝链接/改进记录正常；3 篇新闻与表格渲染正常；首页 3 条新闻、3 张品类卡、导航“数码配件/选购指南/上新动态”；`/about`、`/docs/getting-started` 200；登录后台 articles 返回 9 条、dirs 显示 4/1/3 篇；`go vet ./...`、`go test ./...` 通过。测试服务已停止。
- **未提交**：按约定不自动 commit。

## 内置主题改为 MD3 / 马卡龙 / 莫奈 / 节庆红（2026-10-28）

- **四套预设**（`PresetThemes`）：Material 3（#6750a4 + #625b71，默认/soft/sans）、马卡龙（#a5536f + #85405a，wide/pill）、莫奈（#3f6f70 + #32595a，narrow/serif）、节庆红（#b23b30 + #922a22，wide/round/serif）。每套只保留一个主色系，其余为同色系深浅或灰白；红色主题替换原「夜墨灰」。
- **渲染规则**：`heroSurface` 不再把 accent 原色塞进渐变中间，改为「主色浅调 → 更浅 → surface」；header 渐变起点混入 20% surface，保证深色文字的对比度（四套实测栏目标题 ≥5.2:1）。`--c-footer-text` 改由提亮后的 primary 派生，深色页脚不再出现暗紫/暗红字；删除未使用的 `--c-accent-light`。
- **样式兜底**：`style.css` 的 `:root` 同步为 Material 3 默认值；`--c-price` 改为跟随 `--c-accent`，页脚分隔线改为白色 15% 透明，移动端菜单背景改用 `--header-bg` 以沿用可读的 header 文字色。
- **数据现状**：`content/_themes.toml` 当前为 `themes = []`（未改动），因此前台直接使用 `style.css` 的 Material 3 兜底；四套预设可在后台主题设计器「套用」。
- **验证**：`gofmt`、`go vet ./...`、`go test ./...`（新增 `TestPresetThemes`，更新默认色断言）通过；3100 实测 `/api/admin/themes` 返回四套预设与配色/布局，浏览器计算样式确认 `--c-primary:#6750a4`、页头/横幅渐变与深色文字生效；四套配色对比度脚本检查，正文/主色/页脚均 ≥4.5:1（栏目标题最浅端 ≥5.2:1）。测试服务已停止。
- **未提交**：按约定不自动 commit。


## 字体层级压缩 + SanDisk 经销商样例站（2026-10-28）

- **前台字号**：`:root` 增加 `--font-size: 15px`（原来硬编码 16px）；Hero 标题 42→34px、副标题 20→16px、CTA 16→15px；区块标题 30→24px；页面横幅标题 34→28px、副标题 17→15px；文章页 h1 28→24px、h2 20px、h3 17px；卡片标题 20/18→18/16px，正文 15→14px，元信息 13→12px。保留 1.7 行高与 4pt 间距。
- **品类图标**：首页特色卡不再用 emoji，改为字母徽标（SD/USB/SSD/PRO），`--c-tint-soft` 底 + 主色文字，46×46 圆角块。
- **后台字号**：正文 14px、表单/表格 14px、按钮 13px、辅助文字 12px、小按钮 12px；可点击控件最小高度 44px（小按钮 36px 且带内边距），`focus-visible` 2px 焦点环。
- **产品列表**：服务端分类筛选 + 分页。`/products?category=存储卡`（父级含子级）、`/products/page/N`，每页 9 条；`ARTICLES_PER_PAGE=9`；分类 chip 保留 `?category=` 深链并标 `aria-current="page"`；分页链接保留当前分类（`productPageHref`）。`internal/handler/page.go` 新增 `renderProductList` / `filterProductsByCategory` / `pageNumber` / `productPageHref`，`page_test.go` 覆盖筛选与链接构造。
- **样例目录（16 款，6 个叶子分类）**：microSD 5（Extreme / Extreme PRO / High Endurance / GamePlay / microSD Express）、SD 3（Ultra / Extreme PRO UHS-I / Extreme PRO UHS-II V60）、USB-A 3（CZ48 / CZ73 / CZ880）、双接口 2（DDC4 Luxe / iXpand Luxe）、便携 SSD 2（E61 / E81）、CFexpress 1（Type B）。每款含真实容量档位、顺序写读速度、速度等级、SKU、参考价、京东/淘宝搜索链接、容量对照表、兼容性说明和 1–3 条“听劝记录”。
- **栏目与首页**：`content/_meta.toml` 改为“存储优选”授权经销站（Hero/四大场景入口/到货补货/联系方式）；产品目录 `_meta.toml` 设 `DIR_TYPE=product`、`SORT_ORDER=desc`、每页 9 条。
- **文档/新闻/关于**：`about.md` 改为经销商定位与售后说明；`docs/getting-started.md` 重写为存储卡选购指南（设备兼容表、速度等级、容量估算、三大误区）；新增 `docs/warranty.md`（保修期限表、非保情形、送保前准备）；3 篇到货/补货/价格趋势新闻。
- **图片**：`uploads/` 保留 16 张产品图（WebP/PNG/JPG），均来自闪迪官方商店/CDN 或电商商品图；已知瑕疵：Extreme PRO microSD 图为 Extreme 款、GamePlay 为促销横幅、iXpand Luxe 带中文页头横幅，待后续替换。
- **验证**：`gofmt -l` 无输出；`go vet ./...`；`go test ./internal/...` 全通过；`go build` 通过。端口 3100（独立二进制，未触碰 3000）实测：`/products` 9 卡、`/products/page/2` 7 卡、`?category=存储卡` 8 卡、`?category=U盘` 5 卡；16 个详情页、`/docs/getting-started`、`/docs/warranty`、`/news/microsd-express-stock` 均 200；后台 `/admin/content` 列表 22 条，选项卡为文章 6 / 产品 16 / 页面 0 / 草稿 0，切到“产品”后 16 行；后台全部静态资源 200、无控制台报错。
- **未提交**：按约定不自动 commit。

### 图片修正（2026-10-28）

- **至尊超极速 microSD**：换成官网 128GB 版本产品图（Extreme PRO、microSDXC、V30/A2），白底 1680²，转 JPEG 1200²（94KB）。
- **GamePlay microSD**：换成官网 128GB 正面产品图（去促销横幅），白底 1280²，转 JPEG 1200²（228KB）。
- **iXpand Luxe**：换成官网 USB-C/Lightning 双接口产品图（仅保留 Made for iPhone 角标，无中文页头横幅），1280² WebP（28KB）。
- **验证**：三张 `/uploads/*` 均 200 且 MIME 正确；产品列表/详情引用路径未变；`gofmt`、`go vet ./...`、`go test ./internal/...` 通过。

## 后台新增“统计”与访问埋点（2026-10-28）

- **统计服务**：新增 `internal/service/analytics.go`。访客事件写入 `content/_analytics.jsonl`（`at/path/title/kind/visitor`），启动加载并按 31 天保留期压缩；内存上限 10 万条。报表按 5m/1h/1d/1w/1m 五档分桶（1 分钟/5 分钟/1 小时/1 天/1 天），输出总访问、总访客、内容数、产品/文章/新闻三项的访问数与访客数、趋势序列和内容明细。
- **分类口径**：产品目录 → 产品；`news` 目录（含子路径）→ 新闻；`docs`、根目录文章 → 文章；`DIR_TYPE=page` 的单页不统计。访客按 `gp_visitor` 匿名 Cookie（128 位随机、HttpOnly、SameSite=Lax、TLS 下 Secure、1 年）去重。
- **埋点位置**：`renderArticle` 和 `DirList` 两处直接渲染文章的分支；Cookie 在渲染前写入、响应成功后再记账。列表页、首页、草稿、单页、静态构建均不计入。
- **API 与装配**：`GET /api/admin/analytics?range=5m|1h|1d|1w|1m`（需登录），未知范围返回 400；`app.go` 创建服务、注入 `PageHandler`、注册路由，`Shutdown` 一并关闭日志文件。
- **后台页面**：侧边栏“内容”下新增“统计”（activity 图标）；`/admin/analytics` 页面含时间段切换、总访问/访客/内容数、产品/文章/新闻三张分类卡、访问趋势双柱图（无第三方图表库）、按类型/关键词筛选的内容明细表和保留期说明；新增 `pages/analytics.html`、`js/pages/analytics.js`，app 路由与 API 客户端同步。
- **测试**：`analytics_test.go` 覆盖分类口径、五档范围的总量/分桶/分类去重、最新标题、持久化重载、非法范围；`web/embed_test.go` 增加统计页/脚本内嵌断言。
- **验证（3100，未触碰 3000）**：`gofmt`、`go vet ./...`、`go test ./...`、`go build` 通过。用 3 个 Cookie 罐制造 7 次详情访问（产品 3、文章 2、新闻 2），五档报表均为 7 访问 / 3 访客 / 4 内容，分类去重正确；重启后数据与访客数不变；列表、首页、404 未计数；后台无控制台报错，时间切换、类型筛选、关键词搜索、图表 5/12/24/7/30 个分桶正常。
- **未提交**：按约定不自动 commit。

### 统计页最终回归（2026-10-28）

- 重新 `go build` 打包最新 CSS/JS（分段按钮 40px、零值柱不占高、静默刷新失败保留旧数据并弹 toast），重启 3100 测试服务。
- API 五档均返回 7 访问 / 3 访客 / 4 内容，分桶数 5/12/24/7/30，非法范围 400。
- 浏览器实测 `/admin/analytics`：侧边栏顺序为 工作台/内容/统计/首页装修/主题设计/媒体库/站点设置；KPI 7/3/4；24 小时 48 根柱、最后一桶 100%/43%；切 5 分钟得 5 个桶（21:38–21:42）；类型筛选新闻 1 行、关键词 switch 1 行；无控制台报错，全部资源 200。
- 未提交。

### 后台侧边栏修复（2026-10-28）

- **左上角多余 X**：`.sidebar-close` 的 `display:none` 被后面的 `.icon-btn { display:inline-flex }` 同权重覆盖，导致桌面端也显示；改为 `.icon-btn.sidebar-close` 提高权重，仅 ≤900px 抽屉模式显示。
- **折叠按钮无效**：`.menu-toggle { display:none }` 同样被 `.icon-btn` 覆盖，桌面端按钮可见但只改移动端 `sidebarOpen`。新增 `sidebarCollapsed` 状态：桌面点击折叠为 76px 图标栏（隐藏品牌文字、分组标题、导航文字和用户信息，保留 title/aria-label），移动端仍为抽屉开关；状态写入 `localStorage`。
- **验证**：1440px 下折叠 248→76px、主区 76px，X 隐藏、菜单按钮可切换，展开恢复；775px 下菜单打开抽屉、X 关闭，按钮 aria-label/aria-expanded 随状态更新；无控制台报错；`gofmt`、`go vet ./...`、`go test ./...` 通过。
- 未提交。

### 产品图库与详情页收尾（2026-10-28）

- **封面统一**：16 款产品封面全部换成白底官方图（原 CZ880/E61/E81 等为场景图、microSD/High Endurance 等为绿底 PNG），列表卡片 `.product-cover` 改白底 + `object-fit: contain`（12px 内边距），避免 4:3 裁切导致产品被切边。
- **图库补图**：55 张电商/官网图继续保留；至尊超极速 SD UHS-II 追加官方 256GB 右视/正视图（`sandisk-extreme-pro-sd-uhs2-g2/g3.webp`，均标注 280/150，与正文一致），该款由 1 图增至 3 图。
- **详情页信息密度**：规格选项改 2 列紧凑胶囊（原 4 行通栏），京东/淘宝按钮并排、加入购物车通栏；右栏高度由约 1100px 降到约 900px，缓解左栏主图下方的空白带。
- **展示区**：`.showcase-item` 由渐变底改白底，方图在 4:3 容器内不再出现色块边。
- **内容去重**：删除 16 个产品正文末尾重复的 `## 听劝记录` 段（Front Matter 的 `[[improvements]]` 已由模板渲染），每个详情页恰好出现 1 处“听劝记录”。
- **验证（3100，未触碰 3000）**：`gofmt -l` 无输出、`go build`、`go test ./...` 通过。16 个详情页 200，缩略图数量 3–5（UHS-II 3），引用图片全部存在且 200；`/products` 分类筛选数量与顺序不变；390px 用 CDP `Emulation.setDeviceMetricsOverride` 实测 `scrollWidth=390`（列表/详情均无横向溢出），移动端固定购买条正常。
- 未提交。

### 观感优化：边框收敛（2026-10-28）

- **统一 1px**：把前台所有 `2px` 卡片/面板/控件边框降为 `1px`，颜色由 `--c-border-strong` 收到 `--c-tint-soft` 或 `--c-border`；涉及新闻卡、列表卡、文章容器、分类 chip、产品卡、主图框、缩略图、购买面板、规格选项、购物车复制按钮、详情图框。
- **减少线条**：文章/产品正文表格改为只保留水平分隔线（表头线略深），去掉纵向网格；产品分组标题、筛选分隔线、虚线分隔全部改用 `--c-tint-soft`；产品详情统计分隔、购物车头尾同步淡化。
- **阴影替代边框**：顶栏阴影从 `.3` 降到 `.16`；卡片 hover 边框由 `--c-border-strong` 改为更浅的 `--c-primary-light`；缩略图选中态改为 1px 主色边 + `rgba(primary,.14)` 淡光晕，不再加粗边框。
- **保留项**：导航当前项 2px 下划线、引用块 3px 左色条、所有 `focus-visible` 3px 焦点环（可访问性需要）。
- **补充**：规格单选项补 `input:focus-visible + .spec-option-body` 焦点环，键盘可见。
- **验证（3100）**：`gofmt`、`go vet ./...`、`go build`、`go test ./...` 通过；首页/产品列表/详情/文档页 1440px 与 390px 截图复查，`scrollWidth=390` 无横向溢出，选中/悬停状态清晰。
- 未提交。

### 修复站点标题不可见（2026-10-28）

- **问题**：页头 `.logo` 只有一个立方体 SVG 图标，站点名 `存储优选` 仅存在于 `<title>` 和 `img alt`，页面上看不到品牌名。
- **修复**：四个模板（`themes/default/layouts/{index,list,article}.html`、`templates/list.html`）的 logo 链接补 `<span class="logo-text">{{.SiteTitle}}</span>`，图标 `alt=""`（文字已提供可访问名称）；`style.css` 给 `.logo` 加 `gap:10px`、继承 `--header-text` 颜色，`.logo-text` 18px/700。
- **顺带**：`<title>` 改为 `{{if and .Title (ne .Title .SiteTitle)}}{{.Title}} - {{end}}{{.SiteTitle}}`，首页不再出现 `存储优选 - 存储优选` 重复。
- **验证（3100）**：`gofmt`、`go vet`、`go test ./...`、`go build` 通过；首页 `<title>存储优选</title>`，产品/文档页为 `页面名 - 存储优选`；1440px 与 390px 截图页头显示正常，`scrollWidth=390` 无溢出。
- 未提交。

### 移除非真实“听劝记录”样例（2026-10-28）

- **原则**：听劝记录应来自真实用户反馈，没有反馈时不硬写。
- **改动**：删除 16 个产品 Front Matter 中全部 `[[improvements]]` 样例（共 31 条，均为建站时虚构的“已采纳”记录）；模板 `article.html` 原有 `{{if .Product.Improvements}}` 条件保留，无记录时整段不渲染，价格栏也不再显示“已听劝 N 次”。
- **保留能力**：后台产品编辑器的听劝记录增删/校验逻辑不变，后续有真实反馈可正常录入。
- **验证（3100）**：`content/` 已无 `听劝`/`[[improvements]]`；产品详情页 0 处“听劝”，详情页 1440px/390px 截图无空段或断层；`go vet ./...`、`go test ./...` 通过。
- 未提交。

### 目录结构调整：移出 internal，组件包复数化（2026-10-29）

- **新结构**（根目录直接可见）：`app/`、`common/`、`config/`、`handlers/`、`middlewares/`、`models/`、`services/`；`main.go` 仍按单二进制惯例放在根目录；`internal/` 已删除。
- **包名与导入**：`handler→handlers`、`model→models`、`service→services`、`middleware→middlewares`，`package` 声明、`github.com/azhai/gopaper/...` 导入路径及所有包限定符同步替换；`app`/`common`/`config` 包名不变，仅去掉 `internal/` 前缀。
- **历史保留**：全部使用 `git mv`，文件移动在 Git 中记录为 rename。
- **验证**：`gofmt -l .` 无输出、`go build ./...`、`go vet ./...`、`go test ./...` 全部通过；重建 `/tmp/gopaper-test` 并重启 3100，`/`、`/products`、产品详情、`/docs/getting-started`、`/news`、`/admin/` 均 200。
- 未提交。

### Echo v5 + app.go 根目录 + 模板/dist 清理（2026-10-29）

- **Echo v5.4.2**（v4 依赖已移除）：
  - `echo.Context` → `*echo.Context`（45 处）；导入路径统一 `github.com/labstack/echo/v5`。
  - 错误处理改为 v5 签名 `func(c *echo.Context, err error)`，用 `echo.StatusCode(err)` 兼容 `ErrNotFound` 等哨兵错误；`HTTPError.Message` 现为 string。
  - 启动/退出改用 `echo.StartConfig{Address, HideBanner, GracefulTimeout}` + `signal.NotifyContext`；v5 已无 `Echo.Shutdown`，`App.Close()` 只负责关闭 analytics。
  - 静态文件 `StaticConfig.Filesystem` 改为 `fs.FS`，`Root: "."` 直接使用 `fs.Sub` 结果；`/uploads` 仍用磁盘目录。
  - SSE 的 `w.Flush()` 改用 `http.NewResponseController(w).Flush()`（v5 `Response()` 返回 `http.ResponseWriter`）。
  - 请求日志状态码改用 `echo.ResolveResponseStatus`，404/500 也能正确记录。
- **app.go**：`app/app.go` 移到根目录，包名改为 `main`，构造函数 `New` → `newApp`；`main.go` 直接调用，`app/` 目录删除。
- **模板合并**：`templates/list.html` 与 `themes/default/layouts/list.html` 重复且从未被加载，删除 `templates/`；模板唯一来源为 `themes/default/layouts/`，README 路径同步修正。
- **dist**：是 `gopaper build` 的生成物且已被 `.gitignore` 忽略，删除本地目录；实测 `gopaper build` 可重新生成 index.html/products/static 等。
- **验证（3100）**：`gofmt`、`go vet`、`go test ./...`、`go build` 通过；首页/产品列表/详情/文档/新闻、`/static`、`/admin/`、`/uploads` 均 200；API 401、静态 404 JSON、CORS 头、SIGTERM 优雅退出正常。
- 未提交。
