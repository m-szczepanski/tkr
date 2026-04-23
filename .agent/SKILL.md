---
name: tkr-index
description: >
  Master entry point for all tkr AI agent skills. Use this skill at the
  start of any work session on the tkr project to orient yourself and
  find the right skill for the task. Also use this when a task spans multiple
  subsystems and you need to understand how the skills relate to each other.
  This skill answers: "which skill should I read for this task?"
---

# tkr — Skill Index

This project has **6 skills**. This file is the map. It tells you which skill to read
for any given task and in what order.

---

## Skill Dependency Tree

```
go-conventions          ← Read FIRST for ANY task. All other skills build on this.
    ├── provider-integration    (adding/modifying API providers)
    ├── cli-command             (adding/modifying CLI commands)
    ├── database-layer          (SQL, migrations, Repository interface)
    ├── alert-engine            (condition parser, evaluator, MA logic)
    └── testing                 (test patterns for all of the above)
```

**`go-conventions` is always required.** Every other skill is additive on top of it.

---

## Task → Skill Lookup

| What you are doing | Skills to read |
|---|---|
| Starting fresh on any task | `go-conventions` |
| Adding a new stock exchange API | `go-conventions` → `provider-integration` |
| Fixing a provider's error handling | `go-conventions` → `provider-integration` |
| Adding a new CLI command or sub-command | `go-conventions` → `cli-command` |
| Changing a command's flags or output | `go-conventions` → `cli-command` |
| Adding a new DB table or column | `go-conventions` → `database-layer` |
| Implementing a Repository method | `go-conventions` → `database-layer` |
| Writing a migration file | `go-conventions` → `database-layer` |
| Changing alert condition syntax | `go-conventions` → `alert-engine` |
| Debugging why an alert didn't fire | `go-conventions` → `alert-engine` |
| Adding a new alert metric (e.g. RSI) | `go-conventions` → `alert-engine` |
| Writing tests for a provider | `go-conventions` → `provider-integration` → `testing` |
| Writing tests for DB methods | `go-conventions` → `database-layer` → `testing` |
| Writing tests for alert logic | `go-conventions` → `alert-engine` → `testing` |
| Writing tests for a CLI command | `go-conventions` → `cli-command` → `testing` |
| Any standalone test work | `go-conventions` → `testing` |

---

## Project Document Map

The skills reference these documents frequently:

| Document | What it contains |
|---|---|
| `README.md` | Public overview, quick start, project structure |
| `.agent/FUNCTIONAL_SPEC.md` | Authoritative behaviour spec — commands, data models, flows |
| `.agent/API_PROVIDERS.md` | API endpoints, response shapes, rate limits for each provider |
| `.agent/AI_AGENT_GUIDE.md` | Detailed rules for agents (package layout, interfaces, pitfalls) |
| `.agent/TASKS.md` | All deliverable tasks with effort estimates and milestone grouping |

When a skill says "see `.agent/FUNCTIONAL_SPEC.md` §3.3" — open that file and read that section.
The spec always wins over the skill if there is a conflict.

---

## How to Work a Task from TASKS.md

1. Find the task in `.agent/TASKS.md`. Read the task ID, description, and which milestone it belongs to.
2. Read `.agent/skills/go-conventions/SKILL.md`.
3. Look up the task in the table above and read the additional skill(s).
4. Check the **Definition of Done** checklist at the bottom of each relevant skill.
5. Implement.
6. Run `make lint test build` — all must pass clean.
7. Mark the task `[x]` in `.agent/TASKS.md`.

---

## Shared Patterns Quick Reference

### Sentinel errors → `internal/apperrors/errors.go`
### Domain types → `pkg/models/`
### Provider interface → `internal/provider/provider.go`
### Repository interface → `internal/db/repository.go`
### Notifier interface → `internal/notifier/notifier.go`
### Test in-memory DB → `db.Open(":memory:")`
### Test HTTP mocks → `httptest.NewServer(...)`
### Never `os.Exit` outside `main.go`
### Never `time.Now()` inside `internal/alert/`
### Never import sideways between sibling packages
