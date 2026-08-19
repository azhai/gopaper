# AGENTS.md

(中文版本仅供人类阅读! 最后更新时间 2026-07-05)

本文件是 AI 编码助手在本仓库工作时的行为准则。所有 agent 必须先读完本文件再动手。

核心理念：**高效但不偷懒，精简但不草率**。最好的代码是没写出来的代码，但该写的检查一行都不能少。

每次梳理出来的知识、上下文等等，都必须记录到文档！记录到文档！进度记录到 .ai-memory/progress.md。

> **来源**：核心理念及第 1–5 章内容来自 [ponytail](https://github.com/DietrichGebert/ponytail) 项目的 "lazy senior dev" 工作哲学，结合本仓库实践进行了扩展。如需更新，请对照上游仓库最新版本。
>
> **依赖 Skills**： 目录位于 `~/.agents/skills`。 —

---

## 内容来源索引

| 章节 | 内容分组 | 来源 | 依赖 Skills |
|------|----------|------|-------------|
| 核心理念、1–5 | 工作原则 | 外部项目：[ponytail](https://github.com/DietrichGebert/ponytail) | — |
| 6 | Loop Engineering 规范 | 外部项目：[loop-engineering](https://github.com/cobusgreyling/loop-engineering) | `loop-triage`、`minimal-fix`、`loop-verifier`、`loop-audit`、`loop-constraints`、`loop-budget` |
| 7 | Planning-with-Files 规范 | 外部项目：[planning-with-files](https://github.com/OthmanAdi/planning-with-files) | `planning-with-files` |
| 8 | 工作流 | **本仓库原创** | — |
| 9 | 上下文管理 | **本仓库原创** | — |
| 10 | 知识沉淀 | **本仓库原创** | — |
| 11 | 检验标准 | **本仓库原创** | — |
| 12 | 安全规范 | **本仓库原创** | — |
| 13 | Git 规范 | **本仓库原创** | — |
| 14 | 通用编码规范 | **本仓库原创**（部分思想与 ponytail 一致） | — |
| 15 | 前端规范（HTML/CSS/JS/TS） | **本仓库原创**（参考 Chakra UI、ui-ux-pro-max、xi-yun.top 理念） | — |
| 16 | Go 语言规范 | **本仓库原创** | — |

> **维护提示**：上表中标记为"外部项目"的章节（1–5、6、7）应定期对照上游仓库更新；标记为"本仓库原创"的章节（8–16）由本项目维护。更新时请检查对应 skill 的可用性和路径是否有变化。

---

## 1. 动手前先想

> **来源**：本章节内容来自 [ponytail](https://github.com/DietrichGebert/ponytail) 项目的"先充分理解问题、再选择实现路径"原则，结合本仓库实践进行了扩展。如需更新，请对照上游仓库最新版本。
>
> **依赖 Skills**：—

**不要假设，不要隐藏困惑，把权衡摆到台面上。**

实现之前：
- 明确说出你的假设。不确定就问。
- 存在多种理解时，列出来让用户选，不要默默替他决定。
- 有更简单的方案时，直说。该 push back 就 push back。
- 不清楚的地方，停下来，指出哪里不清楚，然后问。

---

## 2. 简洁优先

> **来源**：本章节内容来自 [ponytail](https://github.com/DietrichGebert/ponytail) 项目的 "No abstractions that weren't explicitly requested / No new dependency / No boilerplate / Deletion over addition / Fewest files possible" 原则，结合本仓库实践进行了扩展。如需更新，请对照上游仓库最新版本。
>
> **依赖 Skills**：—

**解决问题的最少代码。不做投机性设计。**

- 不做没被要求的功能。
- 单次使用的代码不抽象。
- 不加没被要求的"灵活性"或"可配置性"。
- 不为不可能发生的场景写错误处理。
- 200 行能压到 50 行的，重写。

自问："一个资深工程师会不会觉得这过度复杂了？" 会的话，简化。

---

## 3. 外科手术式改动

> **来源**：本章节内容来自 [ponytail](https://github.com/DietrichGebert/ponytail) 项目的"Bug fix = root cause, not symptom"与"Shortest working diff wins, but only once you understand the problem"原则，结合本仓库实践进行了扩展。如需更新，请对照上游仓库最新版本。
>
> **依赖 Skills**：—

**只动该动的。只清理自己制造的烂摊子。**

编辑已有代码时：
- 不"顺手改进"相邻的代码、注释、格式。
- 不重构没坏的东西。
- 匹配现有风格，哪怕你觉得有更好的写法。
- 发现无关的死代码，提一句，不要删。

你的改动产生的孤儿：
- 删掉因你的改动而变 unused 的 import / 变量 / 函数。
- 不删早就存在的死代码，除非被要求。

检验标准：每一行改动都能直接追溯到用户的请求。

---

## 4. 目标驱动执行

> **来源**：本章节内容来自 [ponytail](https://github.com/DietrichGebert/ponytail) 项目的"non-trivial logic leaves ONE runnable check behind"原则，结合本仓库实践进行了扩展。如需更新，请对照上游仓库最新版本。
>
> **依赖 Skills**：—

**定义成功标准，循环验证直到通过。**

把任务转成可验证的目标：
- "加校验" → "为非法输入写测试，然后让测试通过"
- "修 bug" → "写一个能复现 bug 的测试，然后修到它通过"
- "重构 X" → "确保改前改后测试都通过"

多步任务先给简短计划：
```
1. [步骤] → 验证：[检查点]
2. [步骤] → 验证：[检查点]
```

强成功标准让你能独立循环。弱标准（"让它能跑"）会不断需要澄清。

---

## 5. 不偷懒的地方

> **来源**：本章节内容来自 [ponytail](https://github.com/DietrichGebert/ponytail) 项目的"Not lazy about"清单，结合本仓库实践进行了扩展。如需更新，请对照上游仓库最新版本。
>
> **依赖 Skills**：—

**输入校验、防数据丢失的错误处理、安全、可访问性——这些不能省。**

- 信任边界（用户输入、外部 API）的校验必须做。
- 防止数据丢失的错误处理必须做。
- 安全相关不能省。
- 非平凡逻辑留一个可运行的检查：最小的能验证逻辑没坏的断言或小测试。平凡的一行代码不需要测试。

---

## 6. Loop Engineering 规范

> **来源**：本章节内容来自 [loop-engineering](https://github.com/cobusgreyling/loop-engineering) 项目的 AGENTS.md，作为本仓库 loop 操作的参考规范。如需更新，请对照上游仓库最新版本。
>
> **依赖 Skills**：`loop-triage`、`minimal-fix`、`loop-verifier`、`loop-audit`、`loop-constraints`、`loop-budget`

### 6.1 构建与验证

```
# Loop readiness audit CLI
cd tools/loop-audit && npm ci && npm run build
node dist/cli.js ../.. # audit repo root
node dist/cli.js ../.. --suggest # show copy commands for gaps

# Before/after demo (scores an empty dir → starter → L2)
bash scripts/before-after-demo.sh
```

CI 在每次 push/PR 上运行 `validate-patterns` 和 `audit`（参见 `.github/workflows/`）。

### 6.2 审查规范

- 模式和启动器必须保持**工具无关的意图**；特定工具的路径位于 `examples/` 和各工具启动器下。
- 未经人工审查，不得自动合并对 `docs/primitives*.md`、`tools/loop-audit/src/` 或展示资源的更改。
- `stories/` 中的失败故事应包括 token 成本、根本原因和补救措施——而不仅仅是成功案例。
- 新模式需要在 `patterns/registry.yaml` 中有一个条目。

### 6.3 Loop 操作（本仓库）

- **每日分诊**：`loop-triage` skill → `STATE.md`（仅报告，L1）。
- **修复**：仅通过 PR 并经人工审查；`minimal-fix` + `loop-verifier` 用于辅助更改（L2）。
- **隔离**：对于任何无人值守的代码更改实验，使用 git worktrees（参见 `LOOP.md`）。

### 6.4 测试命令

本仓库没有应用测试套件。质量门禁：

```
cd tools/loop-audit && npm run build && node dist/cli.js ../../
bash scripts/before-after-demo.sh
```

---

## 7. Planning-with-Files 规范

> **来源**：本章节内容来自 [planning-with-files](https://github.com/OthmanAdi/planning-with-files) 项目的 AGENTS.md，作为本仓库文件式规划与发布流程的参考规范。如需更新，请对照上游仓库最新版本。
>
> **依赖 Skills**：`planning-with-files`

### 7.1 文件式规划方法

**核心理念**：将规划、进度、发现持久化到磁盘文件中，而非依赖对话上下文，确保工作跨 session 可恢复。

**三个核心文件**（均位于 `.ai-memory/`）：

- **task_plan.md** — 任务计划（用户所有的契约文件，agent 不直接编辑，除非用户明确要求）
- **progress.md** — 进度跟踪。子 agent 的返回结果、完成的步骤、遇到的问题，都记在这里。**不要把子 agent 返回直接塞进 task_plan.md**。
- **findings.md** — 发现与调研。代码探索、根因分析、架构理解，记在这里。

**使用规则**：
- 复杂任务（≥3 个 tool calls）开始前，先写 task_plan.md。
- 每完成一个子任务，更新 progress.md，不修改 task_plan.md。
- 调研过程中的发现（代码位置、根因、设计权衡）写入 findings.md。
- session 切换或 /clear 后，先读这三个文件恢复上下文。
- task_plan.md 是"用户合同"，不得擅自修改其内容；如需调整计划，先向用户提出。

### 7.2 提交规范（增强版）

**格式**：Conventional Commits — `fix:`、`feat:`、`release:`、`docs:`、`refactor:`、`test:`、`chore:`、`perf:`、`ci:` 前缀。

**原则**：
- 一个提交一个目的，不混合无关改动。
- 不自动提交，除非用户明确要求。
- 不使用 `--no-verify`。
- 不对 main/master 做 force push（标签引用更新除外）。
- 贡献者致谢放在 CHANGELOG 和 CONTRIBUTORS.md 中，不放在 commit trailer 里。

### 7.3 发布检查清单（参考）

适用于有正式版本发布的场景：

1. 完整阅读相关 issue 和 PR。
2. 验证 bug 真实存在：精确定位文件/行号，确认报告者正确。
3. 运行全部测试，确保改动前基线通过。
4. Squash merge 为单个提交。
5. 更新 CHANGELOG —— 新版本条目在顶部，分 `### Fixed` / `### Added` / `### Changed`。
6. 更新 CONTRIBUTORS.md —— 添加报告者/贡献者，更新总数和日期。
7. 所有涉及版本号的文件同步更新。
8. 更新 README 中的版本徽章和发布记录表。
9. 提交、打 tag、推送。
10. 创建 GitHub Release。
11. 在 issue/PR 下发布评论。
12. 关闭相关 issue。

### 7.4 CHANGELOG 格式

```
## [X.Y.Z] - YYYY-MM-DD

### Fixed
- 简短描述问题是什么、如何修复的。

### Thanks
- @handle — 贡献内容（issue #N / PR #N）
```

**规则**：
- 客观、事实陈述，不加情绪化语言。
- 贡献者行：用户名或 @handle，一句话，附 issue/PR 引用。

### 7.5 禁止事项速查

- 不要在任何提交中添加 `Co-Authored-By:`。
- 不要直接编辑 `task_plan.md`（用户所有的契约文件）。
- 不要把子 agent 返回记录到 `task_plan.md` —— 用 `progress.md`。
- 不要跳过测试就提交修复。
- 不要在未验证 bug 真实性的情况下就开始修。
- 不要用"顺手改一下"的心态改无关代码（见第 3 章外科手术式改动）。

---

## 8. 工作流

1. **理解** — 读相关代码，理解现状，不假设。
2. **计划** — 复杂任务先给计划，简单任务直接做。
3. **实现** — 最小改动，外科手术式。
4. **验证** — 编译通过、测试通过、vet 通过。
5. **清理** — 删自己产生的孤儿，不留垃圾。
6. **说明** — 简要说明改了什么、为什么，不啰嗦。

---

## 9. 上下文管理

- 避免在上下文窗口最后 20% 做大型重构和多文件特性开发。
- 单文件编辑、文档、简单修复对上下文利用率容忍度更高。
- 不确定时，用搜索工具（Grep/Glob/SearchCodebase）而非假设。

---

## 10. 知识沉淀

- 个人调试笔记、偏好、临时上下文 → 记在脑子里或临时文件，不污染仓库。
- 团队/项目知识（架构决策、API 变更、runbook）→ 放项目已有文档结构里。
- 当前任务已经产出了相关文档或代码注释，就别在别处重复一遍。
- 没有明显的文档位置时，先问，别擅自建顶层文件。

---

## 11. 检验标准

这些准则生效的标志：
- diff 里没有不必要的改动。
- 没有因过度复杂而返工。
- 澄清问题出现在实现之前，而不是出错之后。
- 每一行改动都能追溯到用户请求。
- 编译、vet、测试通过。

---

## 12. 安全规范

**提交前必查：**
- 无硬编码密钥（API key、密码、token）。
- 用户输入全部校验。
- SQL 用参数化查询。
- HTML 输出转义防 XSS。
- 错误信息不泄漏敏感数据。

**发现安全问题：** 停下 → 标记 → 修 CRITICAL → 轮换已暴露密钥 → 排查同类问题。

---

## 13. Git 规范

### 13.1 提交格式
Conventional Commits：`<type>: <description>`

类型：`feat`、`fix`、`refactor`、`docs`、`test`、`chore`、`perf`、`ci`

### 13.2 提交原则
- 一个提交一个目的，不混合无关改动。
- 不自动提交，除非用户明确要求。
- 不提交密钥、`.env`、`credentials.json` 等敏感文件。
- 暂存用具体文件名，不用 `git add -A` / `git add .`。

### 13.3 分支
- 不 force push 到 main/master。
- 不直接 push 到 main/master，除非用户明确要求。

---

## 14. 通用编码规范

> **来源**：本章节为本仓库原创，部分思想与 [ponytail](https://github.com/DietrichGebert/ponytail) 的简洁原则一致。如需更新，请对照上游仓库最新版本。
>
> **依赖 Skills**：—

### 14.1 文件组织
- 多个小文件优于少数大文件。单文件典型 200-400 行，上限 800。
- 按功能/领域组织，不按类型组织。高内聚低耦合。
- 函数小（<50 行），文件聚焦（<800 行）。
- 嵌套不超过 4 层。

### 14.2 控制流与早返回（Guard Clauses）
- **优先判断退出条件，早返回，避免过度嵌套。** 适用于 Go 和 TypeScript。
- 用 guard clause 把异常/边界情况提到函数顶部，主逻辑保持扁平。
  ```go
  // ✅ 好：guard clause，主逻辑扁平
  func process(r *Request) error {
      if r == nil { return ErrInvalid }
      if !r.Valid() { return ErrInvalid }
      // 主逻辑，无额外缩进
      return nil
  }
  // ❌ 坏：箭头型嵌套
  func process(r *Request) error {
      if r != nil {
          if r.Valid() {
              // 主逻辑被深埋
          }
      }
      return nil
  }
  ```
- 嵌套超过 3 层时，考虑提取子函数或用 early return/continue/break 拍平。
- 条件判断优先用正向判断（`if ok`），避免 `if !bad` 套 `else`。
- 避免空 `else` 分支，必要时用 `if !cond { return }` 提前退出。

### 14.3 错误处理
- 每一层都处理错误。
- UI 代码给用户友好的提示。
- 服务端记录详细上下文。
- 永不静默吞掉错误。

### 14.4 不可变性
- 优先创建新对象而非修改已有对象。
- 返回应用了变更的新副本。

### 14.5 命名
- 可读、有意义、自解释的标识符。
- 不用缩写，除非是领域通用术语。

### 14.6 函数设计
- 单一职责：一个函数只做一件事。
- 参数少（≤4 个），多了用结构体/对象封装。
- 无副作用优先：纯函数易测试，副作用集中在边界层。
- 布尔参数考虑拆成两个函数或用枚举，避免 `process(data, true, false, true)`。
- 避免布尔标志参数控制函数行为分支。

### 14.7 注释
- 解释"为什么"，不解释"是什么"（代码本身已说明）。
- TODO/FIXME 带署名或 issue 号：`// TODO(azhai): ...`。
- 公共 API 注释说明用法和边界条件。
- 删掉的代码不留注释痕迹（如 `// 旧逻辑已删除`）。

---

## 15. 前端规范（HTML / CSS / JS / TS）

### 15.1 设计原则（参考 Chakra UI 与简洁实用主义）

**约束式设计（Constraint-Based Design）：**
- 所有视觉值用设计令牌（Design Tokens），禁止硬编码 hex/px。
- 颜色阶梯 50-900：每个色相提供 10 个层级（50 最浅、900 最深），覆盖从背景到强调的全场景。
- 4pt 网格系统：间距基础单位 4px（`1=0.25rem`、`2=0.5rem`、`4=1rem`、`8=2rem`），所有 padding/margin/gap 取 4 的倍数。
- 字号阶梯：`xs/sm/md/lg/xl/2xl...`，字重 `normal/medium/semibold/bold`，行高 `short/moderate/tall`。

**语义化令牌（Semantic Tokens）：**
- 用语义名而非原始色值：`bg.subtle`、`bg.muted`、`fg.muted`、`border.emphasized`、`fg.error`。
- 语义令牌自动适配深色模式（light/dark 双值），业务代码不写 `dark:` 前缀。
- 状态色统一：`error=red`、`warning=orange`、`success=green`、`info=blue`。

**可访问性优先（WAI-ARIA）：**
- 所有交互元素提供 `aria-label` / `role`，图标按钮必须有 `aria-label`。
- 键盘可达：`tabindex` 顺序合理，焦点可见（`focus-visible` 样式）。
- 表单 `<label for>` 关联，错误提示用 `aria-describedby`。
- 颜色对比度满足 WCAG AA（正常文本 4.5:1，大文本 3:1）。

**简洁实用主义（xi-yun.top 理念）：**
- "代码是写给人看的，只是恰好机器能运行"——交互恰到好处，不堆砌功能。
- 轻盈：一行命令开始，不引入沉重配置和依赖。
- 简单、可靠、优雅：衡量 UI 质量的标尺。

**UX 规则优先级（参考 ui-ux-pro-max 体系）：**
按影响从高到低决策，冲突时高优先级胜出。

| 优先级 | 类别 | 影响等级 | 关键检查项 | 反模式 |
|--------|------|----------|------------|--------|
| 1 | 可访问性 | CRITICAL | 对比度 4.5:1、alt 文本、键盘导航、aria-label | 移除焦点环、无 label 的图标按钮 |
| 2 | 触摸与交互 | CRITICAL | 点击区 ≥44×44px、间距 ≥8px、加载反馈 | 仅依赖 hover、状态瞬变（0ms） |
| 3 | 性能 | HIGH | WebP/AVIF、懒加载、预留空间（CLS<0.1） | 布局抖动、累计布局偏移 |
| 4 | 风格选择 | HIGH | 匹配产品类型、一致性、SVG 图标（非 emoji） | 随机混搭 flat 与拟物、emoji 当图标 |
| 5 | 布局与响应式 | HIGH | mobile-first 断点、viewport meta、无横向滚动 | 横向滚动、固定 px 容器宽、禁用缩放 |
| 6 | 排版与颜色 | MEDIUM | 正文 16px、行高 1.5、语义颜色令牌 | 正文 <12px、灰底灰字、组件内裸 hex |
| 7 | 动画 | MEDIUM | 时长 150–300ms、动效传达含义、空间连续性 | 纯装饰动画、动画 width/height、无 reduced-motion |
| 8 | 表单与反馈 | MEDIUM | 可见 label、错误就近、helper 文本、渐进披露 | 仅占位符当 label、错误只在顶部 |
| 9 | 导航模式 | HIGH | 可预测的返回、底部导航 ≤5 项、深链接 | 导航过载、返回行为错乱、无深链接 |
| 10 | 图表与数据 | LOW | 图例、tooltip、可访问配色 | 仅靠颜色传达含义 |

**关键交互规则（带具体阈值）：**
- 点击区：交互元素可视区 ≥44×44px（图标按钮用 padding 扩大命中区，`hitSlop` 思路）。
- 动画时长：微交互 150–300ms，复杂过渡 ≤400ms，禁用 >500ms；进入用 `ease-out`，退出用 `ease-in`，退出比进入快（约 60–70% 时长）。
- 动画性能：只动画 `transform`/`opacity`，禁止动画 `width`/`height`/`top`/`left`；动画必须可中断、不阻塞用户输入。
- 加载反馈：操作 >300ms 显示骨架屏或进度，>1s 用骨架屏替代阻塞式 spinner。
- 触摸间距：相邻可点击元素间距 ≥8px，避免误触。
- 焦点状态：交互元素必须可见 `focus-visible`（2–4px 焦点环），禁止 `outline: none` 不补替代样式。
- 对比度：正文 ≥4.5:1，大文本 ≥3:1，交互状态色（error/success）也须满足对比度。
- 深色模式：用降饱和/提亮的色调变体，非简单反色；分隔线与交互状态在双主题下均需可辨；模态遮罩 40–60% 黑。
- 图标：用 SVG 矢量图标（Lucide/Heroicons），同一图标族统一线宽与圆角；禁止 emoji 当结构化图标；图标尺寸作为设计令牌（如 `icon-sm`/`icon-md=24px`）。
- 表单：可见 label（非仅占位符）、错误信息位于字段下方、必填用星号、失焦校验而非按键即校验、提交后自动聚焦首个错误字段。
- 数值排版：数据列、价格、计时器用 `tabular-nums` 等宽数字，防布局抖动。

**交付前 UI 检查清单：**
- [ ] 无 emoji 当图标，图标族与线宽统一。
- [ ] 按下/悬停/禁用状态视觉清晰，不引起布局位移。
- [ ] 语义令牌一致使用，无 ad-hoc 硬编码颜色。
- [ ] 点击区 ≥44×44px，微交互时长 150–300ms。
- [ ] 禁用态视觉清晰且不可交互。
- [ ] 屏幕阅读器焦点顺序匹配视觉顺序，交互元素有描述性 label。
- [ ] 浅色与深色模式正文对比度均 ≥4.5:1，次级文本 ≥3:1。
- [ ] 分隔线/边框/交互状态在双主题下均可辨。
- [ ] 移动端 375px 与横屏下验证无横向滚动、无内容被固定栏遮挡。
- [ ] 4/8dp 间距节奏在组件、区块、页面三级一致。
- [ ] 颜色非唯一信息载体（附加图标/文本）。
- [ ] `prefers-reduced-motion` 与系统字号放大下不破版。

### 15.2 HTML
- 语义化标签优先（`<header>`/`<nav>`/`<main>`/`<article>`/`<section>`），不用 `<div>` 堆砌。
- 可访问性：`alt`、`aria-label`、`role`、`tabindex` 不能省。
- 表单元素配 `<label>`，`for` 关联。
- 不用内联 `style`，除非是动态计算的值。

### 15.3 CSS
- 不 reinvent 轮子，项目已有 Tailwind 就用 Tailwind class。
- 自定义 CSS 放对应组件的样式文件，不全局散落。
- 颜色、间距用设计 token（CSS 变量 / Tailwind config），不硬编码。
- 响应式：mobile-first，`sm:`/`md:`/`lg:` 递进。
- 不用 `!important`，除非覆盖第三方库且无其他办法。
- 间距取 4 的倍数（4pt 网格），对齐 Chakra UI 约定。

### 15.4 JavaScript / TypeScript
- **TypeScript 优先**，能用 TS 就不用 JS。
- 严格模式：`strict: true`，不用 `any`，必要时用 `unknown` + 类型守卫。
- 类型定义在 `types.ts` 或紧邻使用处，不全局散落。
- 函数式优先：纯函数、不可变数据、`map`/`filter`/`reduce` 优于 for 循环修改数组。
- 异步用 `async/await`，不用裸 `.then()` 链（除非必要）。
- 组件：单一职责，props 接口明确，不透传。
- 命名：组件 PascalCase，变量/函数 camelCase，常量 UPPER_SNAKE_CASE。
- 不用 `var`，`const` 优先，需要重赋值才 `let`。

### 15.5 前端测试
- 关键交互逻辑写单元测试。
- 组件测试用 Testing Library 思路：测用户行为，不测实现细节。
- E2E 仅覆盖关键用户流程。

### 15.6 构建与依赖
- 不随意加 npm 依赖，加之前自问能不能用已有依赖或原生实现。
- `package.json` 保持干净，`npm install` 后能直接 `npm run build`。
- 构建产物不入库（`.gitignore` 已排除 `dist/`）。

---

## 16. Go 语言规范

### 16.1 风格
- 遵循 `gofmt` / `goimports`，不争论格式。
- `go vet` 必须通过。
- 包注释、导出标识符注释遵循 Go convention（`// FuncName does X.`）。
- 错误变量命名 `err`，不 invent 新名字。

### 16.2 现代 Go 写法（Go 1.21+，优先使用新语法）

**迭代与循环：**
- 整数 range（Go 1.22）：`for i := range 10` 代替 `for i := 0; i < 10; i++`。
- 循环变量作用域已修复（Go 1.22+），无需再写 `i := i` 闭包捕获。
- range-over-func 迭代器（Go 1.23+ 稳定）：自定义容器用 `iter.Seq[T]` / `iter.Seq2[K, V]`，避免暴露内部结构或用 channel 模拟。
  ```go
  // 自定义迭代器
  func (s *Set[T]) All() iter.Seq[T] {
      return func(yield func(T) bool) {
          for v := range s.m {
              if !yield(v) { return }
          }
      }
  }
  // 使用：for v := range s.All() { ... }
  ```

**泛型（Go 1.18+，1.24 完整支持类型别名）：**
- 类型别名泛型（Go 1.24）：`type Set[T comparable] = map[T]struct{}`，无需包装 struct。
- 用 `slices` / `maps` 标准库代替手写工具函数。
- 约束：`comparable` 仅 `==`/`!=`，`cmp.Ordered` 支持比较。

**内建函数（Go 1.21+）：**
- 用内建 `min` / `max` 代替手写比较函数。
- 用 `clear(m)` 清空 map / slice。

**错误处理（Go 1.20+）：**
- `errors.Join(errs...)` 聚合多错误，代替第三方 multierror。
- 哨兵错误用 `var ErrNotFound = errors.New(...)` + `errors.Is`，禁止 `nil, nil` 表示"未找到"。

**context（Go 1.21+）：**
- `context.WithoutCancel(ctx)` 用于 fire-and-forget（webhook、后台任务），保留 parent values。
- `context.AfterFunc(ctx, fn)` 注册取消回调，代替手动 goroutine 监听 `ctx.Done()`。

**slog（Go 1.21+）：**
- 按操作富化 logger：`log := slog.With("job_id", id)`，避免每次调用重复字段。
- 库代码禁止 `slog.SetDefault()`，那是 `main.go` 的职责。

**测试（Go 1.24+）：**
- `t.Context()` 获取测试 context，自动在测试结束时取消，代替手动 `context.WithCancel`。
- `testing/synctest`（Go 1.24 实验）测试时间相关逻辑，代替 `time.Sleep`。
- Benchmark 用 `b.Loop()`（Go 1.24）代替 `for range b.N`，setup 只执行一次。

**工具依赖（Go 1.24+）：**
- `go.mod` 用 `tool` 指令管理可执行依赖，代替 `tools.go` blank import hack。
- `go get -tool <pkg>` 添加工具依赖，`go get tool` 升级全部。

### 16.3 错误处理
- 错误必须显式处理，禁止 `_ =` 忽略（除非有明确理由并加注释）。
- `if err != nil` 立即返回，不深层嵌套。
- wrap 错误时用 `fmt.Errorf("doing X: %w", err)` 保留链。
- 自定义错误类型仅在需要类型断言时定义，否则用普通 error。

### 16.4 并发
- 优先用 channel 而非共享内存。
- 必须共享内存时，用 `sync.Mutex` / `sync.RWMutex` 保护。
- goroutine 必须有退出路径（`ctx.Done()` 或关闭 channel），禁止泄漏。
- `sync.WaitGroup` 用于等待，`defer wg.Done()` 紧跟 `wg.Add(1)`。

### 16.5 接口与抽象
- 接口定义在消费方，不预先在实现方定义。
- 接口越小越好（Go proverb: "The bigger the interface, the weaker the abstraction"）。
- 不为单次使用建接口。
- 不为"未来可能需要"建抽象。

### 16.6 测试
- 测试文件 `xxx_test.go` 与被测文件同包。
- 表驱动测试优先（`[]struct{ name string; ... }`）。
- `t.Run(name, func(t *testing.T) { ... })` 子测试。
- 测试名 `TestXxx_Yyy`。
- 不为 trivial getter/setter 写测试。

### 16.7 依赖
- 不引入新依赖，除非确实解决不了。
- 引入前自问：标准库能做吗？已安装的依赖能做吗？
- `go mod tidy` 保持干净。

### 16.8 平台相关代码
- 用 build tags 分隔：`//go:build unix` / `//go:build !unix`。
- 非 Unix 平台提供空实现，不让调用方写平台判断。
- 系统调用封装在 `*_unix.go` / `*_nounix.go`，不泄漏到业务层。

### 16.9 日志
- 用 `log/slog`（标准库），不引入第三方日志库。
- 结构化日志：`slog.Info("msg", "key", value)`。
- Debug 级别放高频/诊断信息，Info 放关键生命周期事件。

---

## 17. 人机配合要求

- 前三步（需求规格设计、实现方案创建、编码任务规划）应尽量简化，中间不逐一确认，仅在"任务执行"前确认一次。
- 除非有关键决策问题需要用户判断，否则不在中间步骤暂停等待确认。
- 关注用户补充的信息，及时响应并调整方案。
