# AGENTS.md

(Last updated 2026-07-05)

This file defines the behavioral guidelines for AI coding agents working in this repository. Every agent must read this file in full before taking any action.

Core principle: **Efficient but not lazy, concise but not careless.** The best code is the code never written — but every check that should be there, must be.

Always document your findings! Record progress in `.ai-memory/progress.md`.

> **Note**: A Chinese version is available at `AGENTS-ZH.md` for human reference only. This English file is the authoritative reference for all agents.

> **Source**: Core principle and Chapters 1–5 are derived from the [ponytail](https://github.com/DietrichGebert/ponytail) project's "lazy senior dev" philosophy, extended with this repository's practices. Check the upstream repository for updates.
>
> **Dependency Skills**: —

---

## Content Source Index

| Chapter | Group | Source | Dependency Skills |
|---------|-------|--------|-------------------|
| Core, 1–5 | Work Principles | External: [ponytail](https://github.com/DietrichGebert/ponytail) | — |
| 6 | Loop Engineering | External: [loop-engineering](https://github.com/cobusgreyling/loop-engineering) | `loop-triage`, `minimal-fix`, `loop-verifier`, `loop-audit`, `loop-constraints`, `loop-budget` |
| 7 | Planning-with-Files | External: [planning-with-files](https://github.com/OthmanAdi/planning-with-files) | `planning-with-files` |
| 8 | Workflow | **Original** | — |
| 9 | Context Management | **Original** | — |
| 10 | Knowledge Persistence | **Original** | — |
| 11 | Verification Criteria | **Original** | — |
| 12 | Security | **Original** | — |
| 13 | Git Conventions | **Original** | — |
| 14 | General Coding Standards | **Original** (partially aligned with ponytail) | — |
| 15 | Frontend (HTML/CSS/JS/TS) | **Original** (references Chakra UI, ui-ux-pro-max, xi-yun.top) | — |
| 16 | Go Language | **Original** | — |
| 17 | Human-Agent Collaboration | **Original** | — |

> **Maintenance note**: Chapters marked "External" (1–5, 6, 7) should be periodically checked against their upstream repositories for updates. Chapters marked "Original" (8–17) are maintained by this project. When updating, verify that the corresponding skill is still available and its path hasn't changed.

---

## 1. Think Before You Act

> **Source**: Derived from [ponytail](https://github.com/DietrichGebert/ponytail)'s "read the task and the code it touches, trace the real flow end to end, then climb" principle, extended with this repository's practices. Check the upstream repository for updates.
>
> **Dependency Skills**: —

The ladder runs after you understand the problem, not instead of it: read the task and the code it touches, trace the real flow end to end, then climb. A small diff you don't understand is just laziness dressed up as efficiency.

**No assumptions, no hidden confusion, put trade-offs on the table.**

Before implementing:
- State your assumptions explicitly. If unsure, ask.
- When multiple interpretations exist, list them and let the user choose — never decide silently.
- If a simpler approach exists, say so. Push back when you should.
- When something is unclear, stop, point out what's unclear, and ask.

---

## 2. Simplicity First

> **Source**: Derived from [ponytail](https://github.com/DietrichGebert/ponytail)'s "No abstractions that weren't explicitly requested / No new dependency / No boilerplate / Deletion over addition / Fewest files possible" principles, extended with this repository's practices. Check the upstream repository for updates.
>
> **Dependency Skills**: —

**The least code that solves the problem. No speculative design.**

Before writing any code, stop at the first rung that holds:

1. Does this need to be built at all? (YAGNI)
2. Does it already exist in this codebase? Reuse it, don't re-write it.
3. Does the standard library already do this? Use it.
4. Does a native platform feature cover it? Use it.
5. Does an already-installed dependency solve it? Use it.
6. Can this be one line? Make it one line.
7. Only then: write the minimum code that works.

Rules:
- No abstractions that weren't explicitly requested.
- No new dependency if it can be avoided.
- No boilerplate nobody asked for.
- Deletion over addition. Boring over clever. Fewest files possible.
- Question complex requests: "Do you actually need X, or does Y cover it?"
- Pick the edge-case-correct option when two stdlib approaches are the same size — lazy means less code, not the flimsier algorithm.
- Mark intentional simplifications with a `ponytail:` comment. If the shortcut has a known ceiling (global lock, O(n²) scan, naive heuristic), the comment names the ceiling and the upgrade path.

Ask yourself: "Would a senior engineer find this over-engineered?" If yes, simplify.

---

## 3. Surgical Changes

> **Source**: Derived from [ponytail](https://github.com/DietrichGebert/ponytail)'s "Bug fix = root cause, not symptom" and "Shortest working diff wins, but only once you understand the problem" principles, extended with this repository's practices. Check the upstream repository for updates.
>
> **Dependency Skills**: —

**Only touch what needs touching. Only clean up your own mess.**

Bug fix = root cause, not symptom: a report names a symptom. Grep every caller of the function you touch and fix the shared function once — one guard there is a smaller diff than one per caller, and patching only the path the ticket names leaves a sibling caller still broken.

Shortest working diff wins, but only once you understand the problem. The smallest change in the wrong place isn't lazy, it's a second bug.

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting while you're there.
- Don't refactor what isn't broken.
- Match the existing style, even if you think there's a better way.
- If you spot unrelated dead code, mention it — don't delete it.

Orphans from your changes:
- Delete imports, variables, and functions that became unused because of your change.
- Don't delete long-standing dead code unless asked.

Verification: every changed line traces directly back to the user's request.

---

## 4. Goal-Driven Execution

> **Source**: Derived from [ponytail](https://github.com/DietrichGebert/ponytail)'s "non-trivial logic leaves ONE runnable check behind" principle, extended with this repository's practices. Check the upstream repository for updates.
>
> **Dependency Skills**: —

**Define success criteria, loop until they pass.**

Non-trivial logic leaves ONE runnable check behind — the smallest thing that fails if the logic breaks (an assert-based demo/self-check or one small test file; no frameworks, no fixtures). Trivial one-liners need no test.

Turn tasks into verifiable goals:
- "Add validation" → "Write a test for invalid input, then make it pass"
- "Fix bug" → "Write a test that reproduces the bug, then fix until it passes"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, give a brief plan first:
```
1. [Step] → Verify: [checkpoint]
2. [Step] → Verify: [checkpoint]
```

Strong success criteria let you loop independently. Weak ones ("make it work") require constant clarification.

---

## 5. Where Not to Be Lazy

> **Source**: Derived from [ponytail](https://github.com/DietrichGebert/ponytail)'s "Not lazy about" checklist, extended with this repository's practices. Check the upstream repository for updates.
>
> **Dependency Skills**: —

**Input validation, data-loss-preventing error handling, security, accessibility — these are never optional.**

Not lazy about: understanding the problem (read it fully and trace the real flow before picking a rung), input validation at trust boundaries, error handling that prevents data loss, security, accessibility, the calibration real hardware needs (the platform is never the spec ideal, a clock drifts, a sensor reads off), anything explicitly requested. Lazy code without its check is unfinished.

- Validation at trust boundaries (user input, external APIs) is mandatory.
- Error handling that prevents data loss is mandatory.
- Security-related checks are never optional.
- Non-trivial logic leaves one runnable check: the smallest assertion or small test that verifies the logic hasn't broken. Trivial one-liners need no test.

---

## 6. Loop Engineering

> **Source**: Content from the [loop-engineering](https://github.com/cobusgreyling/loop-engineering) project's AGENTS.md, used as the reference specification for loop operations in this repository. Check the upstream repository for updates.
>
> **Dependency Skills**: `loop-triage`, `minimal-fix`, `loop-verifier`, `loop-audit`, `loop-constraints`, `loop-budget`

### 6.1 Build & Verify

```bash
# Loop readiness audit CLI
cd tools/loop-audit && npm ci && npm run build
node dist/cli.js ../..              # audit repo root
node dist/cli.js ../.. --suggest    # show copy commands for gaps

# Before/after demo (scores an empty dir → starter → L2)
bash scripts/before-after-demo.sh
```

CI runs `validate-patterns` and `audit` on every push/PR (see `.github/workflows/`).

### 6.2 Review Norms

- Patterns and starters must stay **tool-agnostic in intent**; tool-specific paths live under `examples/` and per-tool starters.
- Never auto-merge changes to `docs/primitives*.md`, `tools/loop-audit/src/`, or showcase assets without human review.
- Failure stories in `stories/` should include token cost, root cause, and remediation — not just wins.
- New patterns require an entry in `patterns/registry.yaml`.

### 6.3 Loop Operation (This Repo)

- **Daily triage**: `loop-triage` skill → `STATE.md` (report-only, L1).
- **Fixes**: only via PR with human review; `minimal-fix` + `loop-verifier` for assisted changes (L2).
- **Isolation**: use git worktrees for any unattended code-change experiments (see `LOOP.md`).

### 6.4 Test Commands

This repo has no application test suite. Quality gates:

```bash
cd tools/loop-audit && npm run build && node dist/cli.js ../../
bash scripts/before-after-demo.sh
```

---

## 7. Planning-with-Files

> **Source**: Content from the [planning-with-files](https://github.com/OthmanAdi/planning-with-files) project's AGENTS.md, used as the reference specification for file-based planning and release workflows in this repository. Check the upstream repository for updates.
>
> **Dependency Skills**: `planning-with-files`

### 7.1 File-Based Planning Method

**Core idea**: Persist plans, progress, and findings to disk files rather than relying on conversation context, ensuring work is recoverable across sessions.

**Three core files** (all in `.ai-memory/`):

- **task_plan.md** — Task plan (user-owned contract file; agents must not edit directly unless explicitly requested)
- **progress.md** — Progress tracking. Sub-agent results, completed steps, and issues go here. **Never paste sub-agent returns into task_plan.md**.
- **findings.md** — Discoveries and research. Code exploration, root-cause analysis, architectural understanding.

**Usage rules**:
- Before starting a complex task (≥3 tool calls), write task_plan.md first.
- After completing each sub-task, update progress.md — do not modify task_plan.md.
- Research findings (code locations, root causes, design trade-offs) go into findings.md.
- After a session switch or /clear, read these three files to restore context.
- task_plan.md is a "user contract" — never modify its content without asking the user first.

### 7.2 Commit Rules

- **Format**: Conventional Commits — `fix:`, `feat:`, `release:`, `docs:`, `refactor:`, `test:`, `chore:`, `perf:`, `ci:` prefixes.
- One commit, one purpose. No mixing unrelated changes.
- No auto-commit unless the user explicitly requests it.
- No `--no-verify`.
- No force push to master except tag ref updates.
- Contributors are credited in CHANGELOG `### Thanks` and `CONTRIBUTORS.md`, never in commit trailers.

### 7.3 Release Checklist (Reference)

For projects with formal version releases:

1. Read the relevant issue and PR in full.
2. Verify the bug is real: find the exact file/line, confirm the reporter is correct.
3. Run all tests — baseline must pass before touching anything.
4. Merge preserving contributor authorship: `git fetch origin pull/N/head:pr-N && git cherry-pick <pr-head-sha>`, or `gh pr merge --rebase`. Do NOT use `git merge --squash`.
5. CHANGELOG — new version entry at top, `### Fixed`/`### Added`/`### Changed`.
6. CONTRIBUTORS.md — add reporter/contributor, bump count and date.
7. Version bump across all relevant files.
8. README — update version badge and releases table.
9. `git commit`, `git tag vX.Y.Z`, `git push origin master`, `git push origin vX.Y.Z`.
10. `gh release create vX.Y.Z --title "vX.Y.Z - <short description>" --notes "<release notes>"`.
11. Post comment on PR and/or issue.
12. Close the related issue.

### 7.4 CHANGELOG Format

```
## [X.Y.Z] - YYYY-MM-DD

### Fixed
- Short description of what was wrong and how it was fixed.

### Thanks
- @handle — what they contributed (issue #N / PR #N)
```

Rules:
- Sachlich (matter-of-fact). No hype.
- Contributor line: first name or @handle, one sentence, issue/PR reference.

### 7.5 Quick Reference: What NOT to Do

- Do not add `Co-Authored-By:` to any commit.
- Do not edit `task_plan.md` directly (user-owned contract).
- Do not log sub-agent returns into `task_plan.md` — use `progress.md`.
- Do not skip tests before committing a fix.
- Do not start fixing before verifying the bug is real.
- Do not make "quick improvements" to unrelated code (see Chapter 3: Surgical Changes).

---

## 8. Workflow

1. **Understand** — Read the relevant code, understand the current state, no assumptions.
2. **Plan** — Give a plan for complex tasks; simple tasks, just do them.
3. **Implement** — Minimal changes, surgical.
4. **Verify** — Compile passes, tests pass, vet passes.
5. **Clean up** — Delete your own orphans, leave no garbage.
6. **Explain** — Briefly state what changed and why, no verbosity.

---

## 9. Context Management

- Avoid large-scale refactoring and multi-file feature development in the last 20% of the context window.
- Single-file edits, documentation, and simple fixes tolerate lower context utilization better.
- When uncertain, use search tools (Grep/Glob/SearchCodebase) rather than assumptions.

---

## 10. Knowledge Persistence

- Personal debug notes, preferences, temporary context → keep in your head or temp files; don't pollute the repo.
- Team/project knowledge (architecture decisions, API changes, runbooks) → put in the project's existing documentation structure.
- If a document or code comment already captures something, don't duplicate it elsewhere.
- When there's no obvious documentation location, ask first — don't create top-level files on your own.

---

## 11. Verification Criteria

Signs these guidelines are working:
- No unnecessary changes in the diff.
- No rework caused by over-complexity.
- Clarification questions come before implementation, not after errors.
- Every changed line traces back to the user's request.
- Compile, vet, and tests pass.

---

## 12. Security

**Pre-commit checklist:**
- No hardcoded secrets (API keys, passwords, tokens).
- All user input is validated.
- SQL uses parameterized queries.
- HTML output is escaped to prevent XSS.
- Error messages don't leak sensitive data.

**When you find a security issue:** Stop → Flag → Fix CRITICAL → Rotate exposed keys → Audit for similar issues.

---

## 13. Git Conventions

### 13.1 Commit Format
Conventional Commits: `<type>: <description>`

Types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `perf`, `ci`

### 13.2 Commit Principles
- One commit, one purpose. No mixing unrelated changes.
- No auto-commit unless the user explicitly requests it.
- Never commit secrets, `.env`, `credentials.json`, or similar sensitive files.
- Stage specific filenames; avoid `git add -A` / `git add .`.

### 13.3 Branches
- No force push to main/master.
- No direct push to main/master unless the user explicitly requests it.

---

## 14. General Coding Standards

> **Source**: Original to this repository. Some principles align with [ponytail](https://github.com/DietrichGebert/ponytail)'s simplicity philosophy. Check the upstream repository for updates.
>
> **Dependency Skills**: —

### 14.1 File Organization
- Many small files over few large ones. Typical single file: 200–400 lines, max 800.
- Organize by feature/domain, not by type. High cohesion, low coupling.
- Small functions (<50 lines), focused files (<800 lines).
- No more than 4 levels of nesting.

### 14.2 Control Flow & Early Returns (Guard Clauses)
- **Prefer exit conditions first, return early, avoid deep nesting.** Applies to Go and TypeScript.
- Use guard clauses to move edge cases to the top of the function; keep the main logic flat.
  ```go
  // ✅ Good: guard clause, flat main logic
  func process(r *Request) error {
      if r == nil { return ErrInvalid }
      if !r.Valid() { return ErrInvalid }
      // Main logic, no extra indentation
      return nil
  }
  // ❌ Bad: arrow-shaped nesting
  func process(r *Request) error {
      if r != nil {
          if r.Valid() {
              // Main logic buried deep
          }
      }
      return nil
  }
  ```
- When nesting exceeds 3 levels, consider extracting a sub-function or using early return/continue/break to flatten.
- Prefer positive conditions (`if ok`) over `if !bad` with `else`.
- Avoid empty `else` blocks; use `if !cond { return }` to exit early instead.

### 14.3 Error Handling
- Handle errors at every level.
- UI code: user-friendly messages.
- Server code: log detailed context.
- Never silently swallow errors.

### 14.4 Immutability
- Prefer creating new objects over modifying existing ones.
- Return new copies with changes applied.

### 14.5 Naming
- Readable, meaningful, self-documenting identifiers.
- No abbreviations unless they're domain-standard terminology.

### 14.6 Function Design
- Single responsibility: one function does one thing.
- Few parameters (≤4); use structs/objects to encapsulate more.
- Prefer no side effects: pure functions are easy to test; concentrate side effects at boundary layers.
- Consider splitting boolean parameters into two functions or using enums; avoid `process(data, true, false, true)`.
- Avoid boolean flag parameters that control function behavior branches.

### 14.7 Comments
- Explain "why", not "what" (the code already says what).
- TODO/FIXME with attribution or issue number: `// TODO(azhai): ...`.
- Public API comments: describe usage and edge cases.
- Deleted code leaves no comment traces (e.g., `// old logic removed`).

---

## 15. Frontend Standards (HTML / CSS / JS / TS)

### 15.1 Design Principles (referencing Chakra UI & Pragmatic Minimalism)

**Constraint-Based Design:**
- All visual values use Design Tokens; no hardcoded hex/px.
- Color scale 50–900: each hue provides 10 levels (50 lightest, 900 darkest), covering background-to-accent scenarios.
- 4pt grid system: spacing base unit 4px (`1=0.25rem`, `2=0.5rem`, `4=1rem`, `8=2rem`); all padding/margin/gap are multiples of 4.
- Type scale: `xs/sm/md/lg/xl/2xl...`, weight `normal/medium/semibold/bold`, line height `short/moderate/tall`.

**Semantic Tokens:**
- Use semantic names over raw color values: `bg.subtle`, `bg.muted`, `fg.muted`, `border.emphasized`, `fg.error`.
- Semantic tokens auto-adapt to dark mode (light/dark dual values); business code never writes `dark:` prefixes.
- Status colors unified: `error=red`, `warning=orange`, `success=green`, `info=blue`.

**Accessibility First (WAI-ARIA):**
- All interactive elements provide `aria-label` / `role`; icon buttons must have `aria-label`.
- Keyboard reachable: `tabindex` order is logical, focus is visible (`focus-visible` style).
- Form `<label for>` association; error messages use `aria-describedby`.
- Color contrast meets WCAG AA (normal text 4.5:1, large text 3:1).

**Pragmatic Minimalism (xi-yun.top philosophy):**
- "Code is written for humans, it just happens to run on machines" — interactions should be恰到好处, not feature-stuffed.
- Lightweight: one command to start, no heavy configs or dependencies.
- Simple, reliable, elegant: the yardstick for UI quality.

**UX Rule Priority (referencing ui-ux-pro-max system):**
Decide by impact from high to low; higher priority wins on conflict.

| Priority | Category | Impact | Key Checks | Anti-patterns |
|----------|----------|--------|------------|---------------|
| 1 | Accessibility | CRITICAL | Contrast 4.5:1, alt text, keyboard nav, aria-label | Removing focus rings, unlabeled icon buttons |
| 2 | Touch & Interaction | CRITICAL | Tap target ≥44×44px, spacing ≥8px, loading feedback | Hover-only, 0ms state transitions |
| 3 | Performance | HIGH | WebP/AVIF, lazy loading, space reservation (CLS<0.1) | Layout thrashing, cumulative layout shift |
| 4 | Style Choices | HIGH | Match product type, consistency, SVG icons (not emoji) | Random flat/skeuomorphic mix, emoji as icons |
| 5 | Layout & Responsive | HIGH | Mobile-first breakpoints, viewport meta, no horizontal scroll | Horizontal scroll, fixed px container width, disabled zoom |
| 6 | Typography & Color | MEDIUM | Body 16px, line-height 1.5, semantic color tokens | Body <12px, gray-on-gray, bare hex in components |
| 7 | Animation | MEDIUM | Duration 150–300ms, meaningful motion, spatial continuity | Decorative-only animation, animating width/height, no reduced-motion |
| 8 | Forms & Feedback | MEDIUM | Visible labels, inline errors, helper text, progressive disclosure | Placeholder-only labels, top-only errors |
| 9 | Navigation Patterns | HIGH | Predictable back, bottom nav ≤5 items, deep linking | Nav overload, broken back behavior, no deep links |
| 10 | Charts & Data | LOW | Legends, tooltips, accessible color palettes | Color as the only information carrier |

**Key Interaction Rules (with specific thresholds):**
- Tap targets: interactive element visible area ≥44×44px (use padding to expand hit area for icon buttons, `hitSlop` approach).
- Animation duration: micro-interactions 150–300ms, complex transitions ≤400ms, disable >500ms; use `ease-out` for enter, `ease-in` for exit, exit ~60–70% of enter duration.
- Animation performance: only animate `transform`/`opacity`; never animate `width`/`height`/`top`/`left`; animations must be interruptible and not block user input.
- Loading feedback: show skeleton or progress after >300ms; use skeleton over blocking spinner after >1s.
- Touch spacing: ≥8px between adjacent clickable elements to prevent mis-taps.
- Focus state: interactive elements must show visible `focus-visible` (2–4px focus ring); never `outline: none` without a replacement style.
- Contrast: body text ≥4.5:1, large text ≥3:1; interactive state colors (error/success) must also meet contrast.
- Dark mode: use desaturated/lightened hue variants, not simple inversion; dividers and interactive states must be distinguishable in both themes; modal overlay 40–60% black.
- Icons: use SVG vector icons (Lucide/Heroicons); unified stroke width and border radius within an icon family; no emoji as structural icons; icon sizes as design tokens (e.g., `icon-sm`/`icon-md=24px`).
- Forms: visible labels (not placeholder-only), error messages below the field, required fields with asterisk, validate on blur not on keystroke, auto-focus first error field after submit.
- Numeric typography: data columns, prices, timers use `tabular-nums` for equal-width digits, preventing layout shifts.

**Pre-delivery UI Checklist:**
- [ ] No emoji as icons; icon family and stroke width are unified.
- [ ] Pressed/hovered/disabled states are visually clear with no layout shift.
- [ ] Semantic tokens used consistently; no ad-hoc hardcoded colors.
- [ ] Tap targets ≥44×44px; micro-interaction duration 150–300ms.
- [ ] Disabled state is visually clear and non-interactive.
- [ ] Screen reader focus order matches visual order; interactive elements have descriptive labels.
- [ ] Light and dark mode body text contrast ≥4.5:1; secondary text ≥3:1.
- [ ] Dividers/borders/interactive states are distinguishable in both themes.
- [ ] Mobile 375px and landscape: no horizontal scroll, no content hidden behind fixed bars.
- [ ] 4/8dp spacing rhythm is consistent at component, block, and page levels.
- [ ] Color is not the sole information carrier (supplemental icons/text present).
- [ ] `prefers-reduced-motion` and system font scaling don't break the layout.

### 15.2 HTML
- Semantic tags first (`<header>`/`<nav>`/`<main>`/`<article>`/`<section>`); no `<div>` soup.
- Accessibility: `alt`, `aria-label`, `role`, `tabindex` are mandatory.
- Form elements paired with `<label>`, `for` association.
- No inline `style` unless the value is dynamically computed.

### 15.3 CSS
- Don't reinvent the wheel; if the project has Tailwind, use Tailwind classes.
- Custom CSS goes in the corresponding component's style file, not scattered globally.
- Colors and spacing use design tokens (CSS variables / Tailwind config); no hardcoding.
- Responsive: mobile-first, `sm:`/`md:`/`lg:` progressive.
- No `!important` unless overriding a third-party library with no alternative.
- Spacing in multiples of 4 (4pt grid), aligned with Chakra UI conventions.

### 15.4 JavaScript / TypeScript
- **TypeScript first**; use TS over JS whenever possible.
- Strict mode: `strict: true`; no `any`; use `unknown` + type guards when necessary.
- Type definitions in `types.ts` or adjacent to usage; not scattered globally.
- Functional style first: pure functions, immutable data, `map`/`filter`/`reduce` over for-loop array mutation.
- Async uses `async/await`; no bare `.then()` chains (unless necessary).
- Components: single responsibility, explicit props interface, no pass-through.
- Naming: components PascalCase, variables/functions camelCase, constants UPPER_SNAKE_CASE.
- No `var`; `const` by default, `let` only when reassignment is needed.

### 15.5 Frontend Testing
- Write unit tests for critical interaction logic.
- Component tests follow Testing Library philosophy: test user behavior, not implementation details.
- E2E covers only key user flows.

### 15.6 Build & Dependencies
- Don't add npm dependencies casually; ask whether an existing dependency or native implementation works first.
- Keep `package.json` clean; `npm install` then `npm run build` should work directly.
- Build artifacts don't go into the repo (`.gitignore` excludes `dist/`).

---

## 16. Go Language

### 16.1 Style
- Follow `gofmt` / `goimports`; no format debates.
- `go vet` must pass.
- Package comments and exported identifier comments follow Go convention (`// FuncName does X.`).
- Error variable naming: `err`; don't invent new names.

### 16.2 Modern Go (Go 1.21+, prefer new syntax)

**Iteration & Loops:**
- Integer range (Go 1.22): `for i := range 10` instead of `for i := 0; i < 10; i++`.
- Loop variable scoping fixed (Go 1.22+); no more `i := i` closure captures.
- Range-over-func iterators (Go 1.23+ stable): custom containers use `iter.Seq[T]` / `iter.Seq2[K, V]`; avoid exposing internal structure or using channels to simulate.
  ```go
  // Custom iterator
  func (s *Set[T]) All() iter.Seq[T] {
      return func(yield func(T) bool) {
          for v := range s.m {
              if !yield(v) { return }
          }
      }
  }
  // Usage: for v := range s.All() { ... }
  ```

**Generics (Go 1.18+, 1.24 full type alias support):**
- Type alias generics (Go 1.24): `type Set[T comparable] = map[T]struct{}`, no wrapper struct needed.
- Use `slices` / `maps` standard library instead of hand-written utility functions.
- Constraints: `comparable` for `==`/`!=` only; `cmp.Ordered` supports comparison.

**Built-in Functions (Go 1.21+):**
- Use built-in `min` / `max` instead of hand-written comparison functions.
- Use `clear(m)` to empty a map / slice.

**Error Handling (Go 1.20+):**
- `errors.Join(errs...)` to aggregate multiple errors; replaces third-party multierror.
- Sentinel errors: `var ErrNotFound = errors.New(...)` + `errors.Is`; no `nil, nil` for "not found".

**Context (Go 1.21+):**
- `context.WithoutCancel(ctx)` for fire-and-forget (webhooks, background tasks); preserves parent values.
- `context.AfterFunc(ctx, fn)` to register cancellation callbacks; replaces manual goroutine listening on `ctx.Done()`.

**slog (Go 1.21+):**
- Enrich logger per operation: `log := slog.With("job_id", id)`; avoid repeating fields on every call.
- Library code must not call `slog.SetDefault()`; that's `main.go`'s job.

**Testing (Go 1.24+):**
- `t.Context()` for test context; auto-cancels when the test ends; replaces manual `context.WithCancel`.
- `testing/synctest` (Go 1.24 experimental) for time-related logic; replaces `time.Sleep`.
- Benchmarks use `b.Loop()` (Go 1.24) instead of `for range b.N`; setup runs only once.

**Tool Dependencies (Go 1.24+):**
- `go.mod` uses `tool` directive for executable dependencies; replaces `tools.go` blank import hack.
- `go get -tool <pkg>` to add tool dependencies; `go get tool` to upgrade all.

### 16.3 Error Handling
- Errors must be handled explicitly; no `_ =` ignoring (unless there's a clear reason with a comment).
- `if err != nil` return immediately; no deep nesting.
- Wrap errors with `fmt.Errorf("doing X: %w", err)` to preserve the chain.
- Custom error types only when type assertion is needed; otherwise use plain error.

### 16.4 Concurrency
- Prefer channels over shared memory.
- When shared memory is necessary, protect with `sync.Mutex` / `sync.RWMutex`.
- Goroutines must have an exit path (`ctx.Done()` or closed channel); no leaks.
- `sync.WaitGroup` for waiting; `defer wg.Done()` immediately after `wg.Add(1)`.

### 16.5 Interfaces & Abstraction
- Define interfaces at the consumer side; don't pre-define at the implementation side.
- Smaller interfaces are better (Go proverb: "The bigger the interface, the weaker the abstraction").
- Don't build interfaces for single-use cases.
- Don't build abstractions for "might need it someday."

### 16.6 Testing
- Test files `xxx_test.go` in the same package as the file under test.
- Table-driven tests preferred (`[]struct{ name string; ... }`).
- `t.Run(name, func(t *testing.T) { ... })` for sub-tests.
- Test names: `TestXxx_Yyy`.
- No tests for trivial getter/setter.

### 16.7 Dependencies
- Don't introduce new dependencies unless truly unavoidable.
- Before adding: can the standard library do it? Can an existing dependency do it?
- `go mod tidy` keeps it clean.

### 16.8 Platform-Specific Code
- Use build tags to separate: `//go:build unix` / `//go:build !unix`.
- Non-Unix platforms provide empty implementations; callers don't write platform checks.
- System calls wrapped in `*_unix.go` / `*_nounix.go`; don't leak to business layer.

### 16.9 Logging
- Use `log/slog` (standard library); no third-party logging libraries.
- Structured logging: `slog.Info("msg", "key", value)`.
- Debug level for high-frequency/diagnostic info; Info for key lifecycle events.

---

## 17. Human-Agent Collaboration

- The first three phases (requirements specification, design, task planning) should be streamlined — no intermediate confirmations; confirm once before "task execution" at most.
- Unless there is a key decision point requiring user judgment, do not pause for confirmation between steps.
- Pay attention to supplementary information from the user and respond promptly.
