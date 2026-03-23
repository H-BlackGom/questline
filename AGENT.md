# Questline MVP1 - Agent Guidelines

## Project Overview

**Questline**은 RPG 스타일의 진행 시스템을 갖춘 CLI 기반 퀘스트 관리 도구입니다.

- **Language**: Go 1.23+
- **Architecture**: Layered (cmd → internal/cli → domain/engine/repository)
- **Database**: SQLite (pure-go, `~/.questline/data.db`)
- **CLI Framework**: Cobra

## Core Domain Rules

### XP & Leveling System (CRITICAL)

- **Fixed XP**: 모든 퀘스트 완료 시 **50 XP** 고정 (난이도 시스템 없음)
- **Level Up Formula**: `next_level_xp = 100 + (current_level * 50)`
  - Lv.1→2: 150 XP
  - Lv.2→3: 200 XP
  - Lv.3→4: 250 XP
- **Titles by Level**:
  - Lv.1-9: Intern
  - Lv.10-19: Junior
  - Lv.20-29: Senior
  - Lv.30-49: Lead
  - Lv.50-98: Principal
  - Lv.99: Guru

### Data Model

```go
// Quest
ID          string     // UUID v4 first 8 chars lowercase
Title       string     // Max 200 chars, required
Status      Status     // TODO, DONE, DROPPED (reserved)
DueDate     *time.Time // Optional, YYYY-MM-DD format
CreatedAt   time.Time
CompletedAt *time.Time

// Player (Singleton - always 1 row with ID=1)
Level           int
CurrentXP       int
TotalXPEarned   int
QuestsCompleted int
```

## Architecture Guidelines

### Package Structure

```
cmd/ql/               # Entry point only
  └── main.go

internal/
  ├── cli/            # Cobra commands
  │   ├── root.go     # Root cmd + DB path helper
  │   ├── add.go      # ql add
  │   ├── done.go     # ql done
  │   ├── ls.go       # ql ls
  │   └── me.go       # ql me
  ├── domain/         # Domain models
  │   ├── quest.go
  │   └── player.go
  ├── engine/         # Business logic (no DB)
  │   ├── leveling.go
  │   └── leveling_test.go
  └── repository/     # Data access
      ├── sqlite.go
      ├── quest_repo.go
      └── player_repo.go
```

### Rules

1. **NO `service/` or `pkg/` in MVP1** - These are reserved for MVP2
2. **NO business logic in `cmd/`** - Only wiring
3. **NO DB logic in `engine/`** - Pure functions only
4. **NO external deps in `domain/`** - Plain structs only

## CLI Command Contracts

### Exit Codes

| Code | Meaning | When Used |
|------|---------|-----------|
| 0 | Success | Command completed successfully |
| 1 | General error | Unexpected errors |
| 2 | Invalid input | Empty title, invalid date format, conflicting flags |
| 3 | Not found | Quest ID doesn't exist |
| 4 | Database error | SQLite errors, file system errors |

### Command Specifications

**`ql add "<title>" [-d YYYY-MM-DD]`**
- Title: trimmed, non-empty, max 200 chars
- Due date: `YYYY-MM-DD` format only (no natural language like "tomorrow")
- Due date: today or future only
- Output: `✓ 퀘스트 #<id> 생성됨: "<title>"`

**`ql done <id>`**
- Transaction: quest status + player XP atomically
- No XP for already-done quests
- Output on level up: shows new level, title, and progress

**`ql ls [--done|--all]`**
- Default: shows TODO quests only
- Sort: `created_at DESC`, tie-breaker `id ASC`
- Empty list: shows message "퀘스트가 없습니다..."
- Conflicting flags (`--done` + `--all`): exit code 2

**`ql me`**
- Auto-creates player on first run
- Output: Box format with level, title, XP, progress bar (20 chars), completed count

## Testing Guidelines

### Test Strategy

1. **Table-driven tests** for engine logic
2. **Temp HOME for CLI tests** - Never touch real `~/.questline`
3. **Temp-file DB for repository tests** - Not in-memory

### Example Pattern

```go
func TestCommand(t *testing.T) {
    tmpDir := t.TempDir()
    origHome := os.Getenv("HOME")
    os.Setenv("HOME", tmpDir)
    defer os.Setenv("HOME", origHome)

    // ... test cases
}
```

### Required Test Coverage

- **Engine**: Level up boundaries, multi-level, negative/edge cases
- **Repository**: Create, Get, List, Complete transaction
- **CLI**: Happy path + error cases (empty title, invalid date, not found)

## Code Style

### Go Conventions

- **NO `as any`** - Keep type safety
- **NO `@ts-ignore`** - Not applicable but reminder
- **NO unnecessary comments** - Code should be self-documenting
- **NO `fmt.Print` for user output** - Use `cmd.OutOrStdout()` and `cmd.ErrOrStderr()`
- **NO debug prints** - Remove before commit

### Error Handling

```go
// CLI commands: return error, let Cobra handle exit code
if err != nil {
    fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: %v\n", err)
    return ErrDatabase  // or ErrInvalidInput, ErrNotFound
}

// Repository: wrap errors with context
return fmt.Errorf("failed to create quest: %w", err)
```

### Output Formatting

- **Success**: `✓` prefix
- **Error**: `✗ 오류:` prefix
- **Level up**: `🎉 레벨업!` with details
- **Table**: Use `text/tabwriter` with consistent spacing

## Database Schema

### SQLite Types

```sql
CREATE TABLE quests (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    status TEXT DEFAULT 'TODO' CHECK (status IN ('TODO', 'DONE', 'DROPPED')),
    due_date TEXT,           -- YYYY-MM-DD
    created_at TEXT NOT NULL, -- RFC3339
    completed_at TEXT        -- RFC3339
);

CREATE TABLE player (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    level INTEGER DEFAULT 1 NOT NULL,
    current_xp INTEGER DEFAULT 0 NOT NULL,
    total_xp_earned INTEGER DEFAULT 0 NOT NULL,
    quests_completed INTEGER DEFAULT 0 NOT NULL,
    updated_at TEXT NOT NULL
);
```

### Initialization

- Auto-create `~/.questline/` directory
- Auto-create singleton player row (ID=1)
- Use `CREATE TABLE IF NOT EXISTS` (no migrations for MVP1)

## MVP1 Scope Boundaries

### ✅ In Scope

- 4 commands: `add`, `done`, `ls`, `me`
- Fixed 50 XP per quest
- `YYYY-MM-DD` dates only
- File-based SQLite
- Basic colored output
- Basic tests

### ❌ Out of Scope (MVP2)

- Difficulty levels (EASY/NORMAL/HARD)
- Variable XP (30/50/100)
- Natural language dates ("tomorrow", "tmr")
- `DROPPED` status CLI commands
- `service/` or `pkg/` packages
- CI/lint configuration
- Release binaries
- Homebrew distribution

## Git Workflow

### Commit Messages

Follow conventional commits:
- `feat: implement ql <command>`
- `fix: handle <specific error case>`
- `test: add tests for <feature>`
- `docs: update README/specs`
- `chore: update dependencies`

### Pre-commit Checks

```bash
go test ./...
go vet ./...
go build -o ./.tmp/ql ./cmd/ql
```

## Documentation

### README.md

Keep in sync with actual implementation:
- Installation instructions (source build)
- Usage examples
- Exit codes
- Troubleshooting

### Specs (`specs/001-questline-mvp1/`)

- **plan.md**: Project structure and tasks
- **data-model.md**: Schema and structures
- **contracts/cli-commands.md**: Command specs
- **quickstart.md**: User guide

Update specs when contracts change, but mark MVP2 features as "reserved".

## Agent Checklist

When working on this project:

- [ ] Read existing code in same package before writing
- [ ] Follow table-driven test patterns
- [ ] Use temp HOME for CLI tests
- [ ] Check exit codes match specification
- [ ] Verify no MVP2 features creeping in
- [ ] Run `go test ./...` and `go vet ./...` before commit
- [ ] Update specs if contracts changed
- [ ] Keep README in sync with code

## References

- **Source of Truth**: `/.opencode/document/requirement.md`
- **Data Model**: `/specs/001-questline-mvp1/data-model.md`
- **CLI Contracts**: `/specs/001-questline-mvp1/contracts/cli-commands.md`
- **Quickstart**: `/specs/001-questline-mvp1/quickstart.md`
