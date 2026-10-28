# 研究发现 (Findings)

> 本文件记录代码探索、架构评估、根因分析与设计权衡。按时间倒序追加。

---

## 2026-10-28 管理后台迁至 Alpine.js

- **现状**：后台是纯静态 Alpine.js 3.17.5 SPA，源码即产物，位于 `web/public/admin/`（`index.html` + `admin.css` + `js/` + `pages/`），由 `//go:embed all:public` 打包；无 Node、pnpm、Vite、TypeScript、Mithril。
- **移除**：`web/admin/`（TS 源码与构建配置）、`web/public/admin/assets/`（旧 Vite 产物）、`web/public/{app.js,index.js,icons.js,types.js,components/,routes/,services/}`（旧 tsc 输出）；`Makefile` 不再有 `front` 目标，`clean` 不再删除 `web/public/admin`。
- **代码高亮**：服务端 goldmark-highlighting/chroma 与 `github-light.css` 移除，前端用 `microlight.js` + `highlight.js`（`window.highlightCode`）在站点页面和后台预览统一高亮。

## 2026-08-19 架构评估（基线快照）

### 1. 技术栈与项目布局

| 维度 | 现状 |
|------|------|
| 语言 | Go 1.26.4（`go.mod`，版本号偏高，疑似预览/笔误，待确认） |
| Web 框架 | Echo v4 (`labstack/echo/v4`) |
| Markdown | goldmark（GFM、定义列表）；代码高亮由前端 microlight.js 完成 |
| Front Matter 解析 | `pelletier/go-toml/v2`，**TOML 格式 `+++` 分隔**，非 YAML |
| 模板引擎 | Go 标准库 `html/template`（`templates/*.html`） |
| 认证 | JWT (`golang-jwt/jwt/v5`) + bcrypt 密码 |
| 缓存 | `VictoriaMetrics/fastcache`（内存，启动全量扫描） |
| 管理后台 | Alpine.js 3.17.5 静态 SPA（`web/public/admin/`，无构建步骤），embed 进二进制 |
| 数据存储 | **纯文件系统，无 SQLite** |
| 日志 | `log/slog` + `azhai/gobus/log` 按天分卷 |

目录结构（`internal/`）：
```
app/        应用装配（路由、中间件、依赖注入）
config/     TOML 配置加载 + 环境变量覆盖
handler/    HTTP 处理器（article/auth/cache/image/layout/page）
service/    业务服务（article/cache/image/renderer/scanner）
model/      数据模型（article/auth/dirmeta/layout/sitetree）
middleware/  JWT 鉴权
common/     slug 生成 + 输入校验
```

### 2. 内容流转

- `Scanner.scanDir` 递归扫描 `content/`，读取每个目录的 `_meta.toml`（`DirMeta`），父子 meta 通过 `MergeDirMeta` 继承合并。
- 每个 `.md` 文件头部 `+++ ... +++` 之间是 TOML 元数据（`MetaData`：title/slug/author/date/tags/comments/weight/position），其余是 Markdown 正文。
- 扫描结果装入 `CacheVault`，`PageHandler` 从缓存取数据，用 `html/template` 渲染。
- 路由：`GET /`（首页）、`GET /:dir`（目录列表或单页）、`GET /:dir/:slug`（文章详情）。

### 3. 与用户需求的差距识别

#### 差距 A（根本性）：动态服务器 vs 静态站点生成器
- **需求目标**："轻量级、易于部署的**静态站点生成系统**"。
- **当前实现**：Echo **动态服务器**，每个请求实时从缓存取数据 + 模板渲染。
- **缺失**：没有 `build` 子命令把 content 编译成纯静态 HTML 输出到 `dist/`，部署应是"丢一堆 HTML 到 nginx"而非"跑一个 Go 进程"。
- **影响**：这是定位性差距，决定后续所有改进方向。需用户确认：保留动态模式 + 新增构建模式（双模），还是彻底转为纯 SSG？

#### 差距 B（需确认）：Front Matter 格式
- **需求描述**："YAML Front Matter 元数据"。
- **实际实现**：TOML Front Matter（`+++` 分隔）。
- **分析**：功能等价（都能指定 layout/作者/时间/分类），仅格式不同。TOML 在 Go 生态更自然，Hugo/Zola 均支持。改为 YAML 需引入 `gopkg.in/yaml.v3` 并重写 scanner 的 `parseMetaData` + article 的 `buildFileContent`，且所有现存 content 文件需迁移。
- **建议**：保持 TOML，修正需求描述。但需用户拍板。

#### 差距 C（未实现）：SQLite 结构化数据
- **需求描述**："产品之类的数据用 SQLite 存储"。
- **实际实现**：`content/products/` 下是 Markdown 文件（`enterprise-cms.md` 等），无任何数据库。
- **分析**：当前产品页就是 Markdown，对中小企业官网够用。SQLite 仅在需要**可查询/可过滤/关系型**的产品目录（如 SKU、价格、库存、多规格）时才有价值。
- **建议**：列为可选项，待用户确认是否有结构化产品目录需求。YAGNI 原则下不预先实现。

#### 差距 D（实现缺陷）：LAYOUT 字段未生效
- `DirMeta.LAYOUT` 字段存在，`_meta.toml` 可配 `LAYOUT = "home"`。
- 但 `PageHandler.renderTemplate` 的模板名是**硬编码** `"index"`/`"list"`/`"article"`，**未读取 LAYOUT 字段**。
- 即：配置了 LAYOUT 也不生效。这是 bug。

#### 差距 E（实现缺陷）：Position/Region 未分组渲染
- `Article.Position` 字段存在（hero/features/news/contact/main），意图是首页按区域分组展示页面。
- 但 `PageHandler.Index` 用 `collectArticles` 把所有文章拍平成一个列表，**未按 Position 分组**注入模板。
- 即：在页面管理里选了"布局位置=hero"，首页不会把它放到 hero 区。这是 bug。

#### 差距 F（未实现）：分页
- `DirMeta.ARTICLES_PER_PAGE` 存在（默认 10），但 `PageHandler.DirList` 一次性返回目录下全部文章，**无分页**。
- `ArticleForge.ListByDir` 有分页参数，但那是 admin API，前台未用。

### 4. 安全与工程问题

| # | 问题 | 位置 | 严重度 |
|---|------|------|--------|
| S1 | `config.toml` 含明文 `JWT_SECRET` 且**未被 .gitignore 忽略** | `config.toml:4` + `.gitignore` | 高 |
| S2 | CORS `AllowOrigins: ["*"]` 全开 | `internal/app/app.go:61` | 中 |
| S3 | `customErrorHandler` 把 `err.Error()` 原样返回客户端，可能泄露内部路径/堆栈 | `internal/app/app.go:132` | 中 |
| S4 | `go.mod` 有 `replace github.com/azhai/gobus => ../gobus`，他人 clone 后无法构建 | `go.mod:32` | 中 |
| S5 | `content/` 整体被 .gitignore 忽略（仅保留 about.md），内容资产不入库 | `.gitignore:34` | 视场景 |
| S6 | 无任何单元测试（无 `*_test.go`） | 全项目 | 中 |
| S7 | 模板启动时 `ParseGlob` 一次性加载，无热重载 | `internal/handler/page.go:24` | 低 |

### 5. 设计观察（非问题，仅记录）

- `MergeDirMeta` 用"非零即覆盖"策略，意味着子目录无法把父级设置的字段**清空**（只能覆盖为另一个非空值）。对站点配置继承够用。
- `sortArticles` 按 Weight 降序 → Date 降序 → Title 升序；`SORT_ORDER` 字段存在但 `sortArticles` **未读取**它，即 `SORT_ORDER=asc/desc` 配置不生效（又一个未接线字段）。
- `articleHref` 用 `dirPath + "/" + slug`，若 dirPath 含子目录会生成多段路径，但路由只注册了 `/:dir/:slug` 两段，**嵌套目录的文章 URL 404**。

### 6. 待用户确认的决策点

1. **定位**：纯 SSG（加 build 命令）vs 保留动态 + 新增 build（双模）？
2. **Front Matter**：保持 TOML（建议）还是迁移到 YAML？
3. **SQLite**：是否有结构化产品目录需求？还是 Markdown 够用？
4. **content/ 入库**：内容是否应纳入 git 管理（SSG 场景通常入库）？

---

## 数码配件示例数据与真实来源（2026-10-28）

- **产品数据**：闪迪 CZ48（32/64/128/256GB，¥34/57/107/239，SDCZ48-032G/064G/128G/256G）、小米移动电源3（10000mAh 22.5W ¥79 PLM12ZM；20000mAh 18W ¥109 PLM18ZM）、索尼 WH-CH520（五色 ¥279，WH-CH520/B、W、L、C、P）、绿联 CD244 65W（单头 ¥99 CD244；套装 ¥128 CD244-KIT）。参数与参考价交叉核对自 ZOL、PConline、苏宁、什么值得买、京东商品页与绿联官网，时间为 2026-10-28。
- **图片**：CZ48 用 homedo 白底图（避开 ZOL 水印），其余取京东/品牌官方图；统一用 `sips` 转格式、PIL 裁剪后存 `uploads/`，文件名与产品 slug 对应。
- **首页新闻区**：原 `buildIndexData` 用 `collectArticles` 把产品、指南、关于全塞进“上新动态”。新增 `indexArticles`：`NEWS_LINK` 指向的子树优先，目录不存在时回退全站，避免其他站点因空配置丢内容。`TestIndexArticles` 覆盖三条分支。
- **价格口径**：全部写“参考价，以下单页为准”，本站不收款、不配送，京东/淘宝链接统一用搜索词而非可能失效的商品 ID。
- **内容备份**：旧示例内容在 `/tmp/gopaper-content-backup-20261028/`；2016 年的 PHP 示例文 `last-month-day.md` 已移入该备份，不随站点发布。

## 访问统计实现要点（2026-10-28）

- **记录时机**：Echo 的响应一旦写出（`c.HTMLBlob`），之后 `c.SetCookie` 不再生效。首次实现把 `Set-Cookie` 放在渲染后，导致每次访问都生成新访客、用户数等于访问数；正确做法是在渲染前解析/写入 `gp_visitor` Cookie，渲染成功后再记账。
- **记录范围**：仅已发布的详情页（产品/文章/新闻）。`renderArticle` 与 `DirList` 中两处直接渲染文章的分支都要记账；列表页、首页、单页（`DIR_TYPE=page`）和 `gopaper build` 不计入，避免内容浏览被重复放大。
- **分类口径**：`DIR_TYPE=product` → 产品；`DIR_TYPE=news` 或路径在 `news/` 下 → 新闻；其余（含 `docs`、根目录文章）→ 文章；单页跳过。与后台“文章”筛选口径一致，同时把新闻单独拆出。
- **存储**：`content/_analytics.jsonl` 追加写，扫描器只认 `.md`，不会把日志当内容。启动时加载并按 31 天保留期压缩；内存上限 10 万条，超限丢最旧 20% 并压缩重开文件。`content/` 已在 `.gitignore` 中，日志不会入库。
- **分桶边界**：`sort.Search` 的谓词必须是“第一个 start 晚于事件时间”（`starts[i].After(at)`），写成 `!After` 会把当前桶的事件判成 -1；5m/1h/1d 用整分/5 分/整点对齐，7d/30d 用本地零点对齐。
- **去重**：访客按 128 位随机 Cookie 去重；总用户、分类用户、单内容用户分别维护集合，不能把内容行用户数相加当分类用户数。

## 本地视觉验证要点（2026-10-28）

- **headless Chrome 最小窗口宽度 500px**：`--window-size=390,1600` 的实际 `innerWidth` 仍是 500，移动端截图会被裁成 390，看起来像横向溢出。判断是否真的溢出必须用 CDP：`Emulation.setDeviceMetricsOverride {width:390, mobile:true}` 后再读 `document.documentElement.scrollWidth`。
- **懒加载与全页截图**：`captureBeyondViewport` 不会触发 `loading="lazy"`，列表远端卡片的封面会是空白；截图前先把 `img[loading=lazy]` 改成 `eager` 并等待加载。
- **透明 PNG 的调色板陷阱**：PIL 对带透明度的 P 模式 PNG 直接 `convert("RGB")` 会用调色板 0 号色（常见为绿色）填底，做联系表时应先 `convert("RGBA")` 再与白底 `alpha_composite`。
