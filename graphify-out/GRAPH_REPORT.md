# Graph Report - ai-dev-tools  (2026-11-22)

## Corpus Check
- Corpus is ~9,323 words - fits in a single context window. You may not need a graph.

## Summary
- 160 nodes · 183 edges · 17 communities (10 shown, 7 thin omitted)
- Extraction: 79% EXTRACTED · 21% INFERRED · 0% AMBIGUOUS · INFERRED: 38 edges (avg confidence: 0.88)
- Token cost: 62,659 input · 20,886 output

## Community Hubs (Navigation)
- LLM Provider Abstraction
- CLI Commands (Cobra)
- Implementation Tasks
- Linting Config (golangci)
- Project Overview & Architecture
- Code Review & Changelog
- Git & GitHub Integration
- Linting Rules
- Main Entry & Testing
- Configuration System
- Project Milestones & Metrics
- Go Version & Dependencies
- Conventional Commits
- Task Execution Order
- Review Focus Areas
- Internal Package

## God Nodes (most connected - your core abstractions)
1. `linters-settings` - 8 edges
2. `init()` - 8 edges
3. `pkg/llm.Provider Interface` - 7 edges
4. `Task 6: Configuration (pkg/config)` - 7 edges
5. `LLM Provider Interface (Complete, Stream, Name, MaxTokens, SupportsTools, SupportsJSONMode)` - 7 edges
6. `AI-Powered Developer Tool Suite (CLI)` - 6 edges
7. `Modular Go Monorepo Architecture` - 6 edges
8. `Task 4: Git Operations (pkg/git)` - 5 edges
9. `Task 16: E2E Tests & Final Integration` - 5 edges
10. `ai-dev-tools Project` - 5 edges

## Surprising Connections (you probably didn't know these)
- `Provider-Agnostic LLM Abstraction` --semantically_similar_to--> `LLM Provider Interface (Complete, Stream, Name, MaxTokens, SupportsTools, SupportsJSONMode)`  [INFERRED] [semantically similar]
  docs/superpowers/plans/2026-11-22-ai-dev-tools-plan.md → docs/superpowers/specs/2026-11-22-ai-dev-tools-design.md
- `ai-dev-tools Project` --semantically_similar_to--> `ai-dev-tools Project`  [INFERRED] [semantically similar]
  docs/superpowers/plans/2026-11-22-ai-dev-tools-plan.md → docs/superpowers/specs/2026-11-22-ai-dev-tools-design.md
- `AI-Powered Developer Tool Suite (CLI)` --semantically_similar_to--> `Goals: Unified CLI, Provider-Agnostic, Zero-Config, Distribution-Ready, Portfolio Showcase`  [INFERRED] [semantically similar]
  docs/superpowers/plans/2026-11-22-ai-dev-tools-plan.md → docs/superpowers/specs/2026-11-22-ai-dev-tools-design.md
- `Modular Go Monorepo Architecture` --semantically_similar_to--> `Monorepo Structure (cmd/, pkg/, .github/, docs/, scripts/, testdata/)`  [INFERRED] [semantically similar]
  docs/superpowers/plans/2026-11-22-ai-dev-tools-plan.md → docs/superpowers/specs/2026-11-22-ai-dev-tools-design.md
- `Task 6: Configuration (pkg/config)` --semantically_similar_to--> `Config Struct (Provider, Tools, GitHub, Output)`  [INFERRED] [semantically similar]
  docs/superpowers/plans/2026-11-22-ai-dev-tools-plan.md → docs/superpowers/specs/2026-11-22-ai-dev-tools-design.md

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Core Library Tasks (LLM, Prompt, Git, GitHub, Config)** — docs_superpowers_plans_2026_11_22_ai_dev_tools_plan_task_2_llm_provider, docs_superpowers_plans_2026_11_22_ai_dev_tools_plan_task_3_prompt_system, docs_superpowers_plans_2026_11_22_ai_dev_tools_plan_task_4_git_ops, docs_superpowers_plans_2026_11_22_ai_dev_tools_plan_task_5_github_integration, docs_superpowers_plans_2026_11_22_ai_dev_tools_plan_task_6_config [EXTRACTED 0.85]
- **CLI Command Tasks (Commit, Review, TestGen, Changelog, Config/Doctor)** — docs_superpowers_plans_2026_11_22_ai_dev_tools_plan_task_7_commit_cmd, docs_superpowers_plans_2026_11_22_ai_dev_tools_plan_task_8_review_cmd, docs_superpowers_plans_2026_11_22_ai_dev_tools_plan_task_9_testgen_cmd, docs_superpowers_plans_2026_11_22_ai_dev_tools_plan_task_10_changelog_cmd, docs_superpowers_plans_2026_11_22_ai_dev_tools_plan_task_11_config_doctor [EXTRACTED 0.85]
- **Distribution & Polish Tasks (GoReleaser, npm, Docs, CI, Install, E2E, Release)** — docs_superpowers_plans_2026_11_22_ai_dev_tools_plan_task_12_distribution, docs_superpowers_plans_2026_11_22_ai_dev_tools_plan_task_13_docs, docs_superpowers_plans_2026_11_22_ai_dev_tools_plan_task_14_ci_benchmarks, docs_superpowers_plans_2026_11_22_ai_dev_tools_plan_task_15_install_polish, docs_superpowers_plans_2026_11_22_ai_dev_tools_plan_task_16_e2e_tests, docs_superpowers_plans_2026_11_22_ai_dev_tools_plan_task_17_release [EXTRACTED 0.85]

## Communities (17 total, 7 thin omitted)

### Community 0 - "LLM Provider Abstraction"
Cohesion: 0.08
Nodes (21): Anthropic Provider (Claude 3.5 Sonnet, Haiku, Opus), Custom Provider (OpenAI-compatible endpoints), JSON Schema Validation for Prompts, pkg/llm.Provider Interface, Ollama Provider (llama3, codellama, qwen2.5-coder), OpenAI Provider (gpt-4o, gpt-4o-mini, o1, o3-mini), Prompt Template Registry (Built-in + User Override), Retry/Backoff with Circuit Breaker (+13 more)

### Community 1 - "CLI Commands (Cobra)"
Cohesion: 0.23
Nodes (9): NewCmd(), NewCmd(), NewCmd(), NewCmd(), NewCmd(), init(), initConfig(), NewCmd() (+1 more)

### Community 2 - "Implementation Tasks"
Cohesion: 0.12
Nodes (18): Benchmark Tracking in CI (Weekly), Golden File Tests (testdata/golden/), Task 11: Config CLI & Doctor (cmd/config, cmd/doctor), Task 12: GoReleaser, Homebrew, npm, GitHub Action, Task 13: Documentation Site (VitePress), Task 14: CI Pipeline & Benchmarks, Task 15: Install Script & Polish, Task 16: E2E Tests & Final Integration (+10 more)

### Community 3 - "Linting Config (golangci)"
Cohesion: 0.12
Nodes (16): threshold, min-len, min-occurrences, min-complexity, check-shadowing, linters-settings, dupl, goconst (+8 more)

### Community 4 - "Project Overview & Architecture"
Cohesion: 0.15
Nodes (14): ai-dev binary, ai-dev-tools Project, AI-Powered Developer Tool Suite (CLI), Cobra CLI Framework, Four AI Commands (commit, review, testgen, changelog), GitHub Action (Docker-based), GoReleaser for Multi-Platform Binaries, Modular Go Monorepo Architecture (+6 more)

### Community 5 - "Code Review & Changelog"
Cohesion: 0.16
Nodes (15): Test Framework Auto-Detection (gotest, jest, vitest, pytest), Keep a Changelog Format, Multi-Pass Reviewer (Security, Bugs, Style, Performance), SARIF Output Format for Code Review, Task 10: CLI Changelog/PR Summarizer (cmd/changelog), Task 6: Configuration (pkg/config), Task 7: CLI Commit Generator (cmd/commit), Task 8: CLI Code Reviewer (cmd/review) (+7 more)

### Community 6 - "Git & GitHub Integration"
Cohesion: 0.17
Nodes (12): Git CLI Fallback (shell-out to git), GitHub API Client (go-github wrapper), libgit2 Backend (git2go) for Git Operations, Task 4: Git Operations (pkg/git), Task 5: GitHub Integration (pkg/github), Commit Struct (Hash, ShortHash, Message, Author, Email, Date, Files), DiffOptions (IncludePatterns, ExcludePatterns, MaxSize), GitRepo Interface (Diff, Log, Status, Tags, Branches) (+4 more)

### Community 7 - "Linting Rules"
Cohesion: 0.17
Nodes (11): linters, path, issues, exclude-rules, linters, disable, enable-all, run (+3 more)

### Community 8 - "Main Entry & Testing"
Cohesion: 0.22
Nodes (4): main(), Execute(), NewRootCmd(), TestRootCommand_HasSubcommands()

### Community 9 - "Configuration System"
Cohesion: 0.22
Nodes (9): Config Priority: CLI > Env > Repo Config > User Config > Defaults, ChangelogConfig (Format, GroupBy, IncludePRs), CommitConfig (Style, MaxLength, IncludeBody), Config Resolution Priority: CLI > Env > Repo Config > User Config > Defaults, Config Struct (Provider, Tools, GitHub, Output), ProviderConfig (Name, Model, APIKeyEnv, BaseURL, Timeout, MaxRetries), ReviewConfig (SeverityThreshold, ExcludePatterns, IncludePatterns, MaxFilesPerReview), TestGenConfig (Framework, CoverageTarget, IncludeMocks, FollowPatterns) (+1 more)

## Knowledge Gaps
- **50 isolated node(s):** `github.com/ernedo/ai-dev-tools`, `ai-dev binary`, `Four AI Commands (commit, review, testgen, changelog)`, `Cobra CLI Framework`, `GoReleaser for Multi-Platform Binaries` (+45 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 73 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **7 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Task 6: Configuration (pkg/config)` connect `Code Review & Changelog` to `Configuration System`, `Git & GitHub Integration`?**
  _High betweenness centrality (0.210) - this node is a cross-community bridge._
- **What connects `github.com/ernedo/ai-dev-tools`, `ai-dev binary`, `Four AI Commands (commit, review, testgen, changelog)` to the rest of the system?**
  _50 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `LLM Provider Abstraction` be split into smaller, more focused modules?**
  _Cohesion score 0.08333333333333333 - nodes in this community are weakly interconnected._
- **Why does `Task 5: GitHub Integration (pkg/github)` connect `Git & GitHub Integration` to `Code Review & Changelog`?**
  _High betweenness centrality (0.189) - this node is a cross-community bridge._
- **Should `Implementation Tasks` be split into smaller, more focused modules?**
  _Cohesion score 0.12418300653594772 - nodes in this community are weakly interconnected._
- **Why does `Task 4: Git Operations (pkg/git)` connect `Git & GitHub Integration` to `LLM Provider Abstraction`?**
  _High betweenness centrality (0.188) - this node is a cross-community bridge._
- **Should `Linting Config (golangci)` be split into smaller, more focused modules?**
  _Cohesion score 0.125 - nodes in this community are weakly interconnected._