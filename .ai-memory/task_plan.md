# 任务计划 (Task Plan)

> 本文件为用户契约，记录需求拆解与任务清单。agent 不得擅自修改，除非用户明确要求。
> 进度记录见 `progress.md`，研究发现见 `findings.md`。

---

## 项目目标

构建一个轻量级、易于部署的**静态站点生成系统**，适用于中小企业官网、个人博客或项目文档场景。

- **内容管理**：以 Markdown 文件为主要内容来源；产品等结构化数据可选使用 SQLite。
- **模板与主题**：全局 Layout 布局文件 + 固定 Theme 主题机制。
- **Front Matter**：每个 MD 文件头部含元数据，指定 layout、作者、发布时间、所属分类等。

当前状态：基础可用版本已完成（动态 Echo 服务器 + admin 后台），需进一步改进。

---

## 已确认决策（2026-08-19 用户拍板）

| # | 决策点 | 结论 | 理由 |
|---|--------|------|------|
| D1 | 系统定位 | **双模**：保留动态 + 新增 `build` 命令 | 兼顾 admin 易管理与静态易部署 |
| D2 | Front Matter 格式 | **保持 TOML** `+++` | 已实现，Go 生态自然，迁移无功能收益 |
| D3 | SQLite 需求 | **暂不引入**，Markdown 够用 | YAGNI，待结构化查询需求出现再加 |
| D4 | content/ 入库 | **保持忽略** | 内容作运行时数据，靠 admin 管理 |

---

## 任务拆解

### P0 — 核心差距（定位性，确认 D1 后启动）

| ID | 任务 | 关联决策 | 验收标准 |
|----|------|----------|----------|
| T01 | 新增 `build` 子命令：扫描 content → 渲染 → 输出静态 HTML/CSS/JS 到 `dist/` | D1 | `gopaper build` 生成 dist/，含 index.html 及各目录页，可直接用 nginx 托管 |
| T02 | 修正 LAYOUT 字段生效：`PageHandler` 根据 `DirMeta.LAYOUT` 选择模板，而非硬编码 | — | 在 `_meta.toml` 设 `LAYOUT="home"` 后该目录用 home 模板渲染 |
| T03 | 修正 Position/Region 分组：首页按 `Article.Position` 把页面分发到 hero/features/news/contact 区 | — | 页面设 position=hero 后出现在首页 hero 区 |
| T04 | 修正 `SORT_ORDER` 生效：`sortArticles` 读取 `DirMeta.SORT_ORDER` 决定升降序 | — | 设 SORT_ORDER=asc 后权重小者在前 |
| T05 | 修正嵌套目录路由：支持 `/:dir/:sub/:slug` 多段路径，或 flatten 后的两段路由 | — | content/a/b/c.md 可通过 /a/b/c 访问 |

### P1 — 重要功能补全

| ID | 任务 | 验收标准 |
|----|------|----------|
| T06 | 前台分页：`DirList` 按 `ARTICLES_PER_PAGE` 分页，生成 /news/page/2 等 | 列表页底部出现分页导航 |
| T07 | 主题目录规范化：`themes/default/{layouts,static,partials}` 结构，模板从主题加载 | 模板从 themes/default/ 加载，为未来多主题留结构 |
| T08 | RSS feed 生成：`/index.xml` 输出文章 RSS | 访问 /index.xml 得到合法 RSS |
| T09 | sitemap.xml + robots.txt 生成 | build 时生成 sitemap.xml |
| T10 | SEO meta 标签：每页输出 title/description/og 标签 | 页面源码含 og:title 等 |

### P2 — 安全与工程加固

| ID | 任务 | 验收标准 |
|----|------|----------|
| T11 | `config.toml` 加入 .gitignore，提供 `config.example.toml` | config.toml 不再入库，example 供参考 |
| T12 | CORS 收紧为可配置（`ALLOWED_ORIGINS`），默认仅本机 | 生产配置非 * |
| T13 | `customErrorHandler` 区分面向用户/内部错误，不泄露内部信息 | 500 错误返回通用文案，详情仅入日志 |
| T14 | 移除 `go.mod` 的 `replace ../gobus`，或发布 gobus 到可拉取的地址 | `go build` 在干净环境可通过 |
| T15 | 添加核心包单元测试（scanner/renderer/article forge），table-driven | `go test ./...` 通过 |
| T16 | 确认 Go 版本：1.26.4 是否真实存在，否则降到 1.23/1.24 | go.mod 版本与实际工具链匹配 |

### P3 — 体验优化

| ID | 任务 | 验收标准 |
|----|------|----------|
| T17 | dev 模式模板热重载（文件监听重新 ParseGlob） | 改模板不重启即生效 |
| T18 | dev 模式 livereload / 内容监听自动重建 | 改 md 后浏览器自动刷新 |
| T19 | 草稿/发布状态：Front Matter 加 `draft = true`，build 时排除 | draft 文章不出现在产物中 |
| T20 | 文章摘要（summary/excerpt）：Front Matter 或 `<!--more-->` 分隔 | 列表页显示摘要而非全文 |
| T21 | 静态资源指纹/hash（build 时对 CSS/JS 加 contenthash） | 静态资源文件名含 hash |

---

## 执行顺序

1. ~~确认 D1–D4~~（已完成 2026-08-19）。
2. P0 修复：T02–T05（bug 修复，立即做）+ T01（build 命令，双模核心）。
3. P2 安全加固：T11/T12/T13（可并行，不依赖其他）。
4. P1 功能补全（依赖 P0 的 build 基础）。
5. P3 体验优化。

---

## 变更记录

| 日期 | 变更 |
|------|------|
| 2026-08-19 | 初始创建：完成架构评估，拆解 21 项任务，识别 4 个待确认决策 |
| 2026-08-19 | 用户确认 D1-D4：双模 / 保持 TOML / 暂不引入 SQLite / content 保持忽略。启动 P0 执行。 |