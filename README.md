# Questline

A gamified task management CLI tool with RPG-style progression.

## Features

### Core Features
- **Quest Management**: Create and complete tasks with due dates
- **XP System**: Earn 50 XP for each completed quest
- **Level Up**: Progress through levels with increasing difficulty
- **Titles**: Unlock titles from Intern to Guru as you level up

### MVP2 Features
- **Quest Types**: Classify quests as `daily`, `weekly`, `epic`, `guild`, or `sub`
  ```bash
  ql add "Morning routine" -t daily
  ql add "Big project" -t epic
  ```
- **Quest Hierarchy**: Create sub-quests under epic/guild quests
  ```bash
  ql add "Sub task" -t sub -p <parent_id>
  ```
- **Flow System**: Track daily completion streaks with XP multipliers
  ```bash
  ql me --flow  # Show Flow status and multiplier
  ```
- **TUI Dashboard**: Interactive terminal UI for quest management
  ```bash
  ql check  # Launch Bubble Tea dashboard
  ```
- **Lazy Evaluation**: Automatic daily quest archival at 04:00

## Installation

### Prerequisites

- Go 1.23 or later

### Build from Source

```bash
# Clone the repository
git clone https://github.com/H-BlackGom/questline.git
cd questline

# Download dependencies
go mod download

# Build the binary
go build -o ql ./cmd/ql
```

#### Option 1: Install to PATH (requires sudo)

```bash
# Move to system PATH (requires administrator privileges)
sudo mv ql /usr/local/bin/

# Or on macOS with Homebrew:
# sudo mv ql /opt/homebrew/bin/
```

#### Option 2: Use without installation

```bash
# Run directly from current directory
./ql add "My first quest"
./ql ls
./ql me

# Or move to a user directory
mkdir -p ~/bin
mv ql ~/bin/
export PATH="$HOME/bin:$PATH"
```

### Verify Installation

```bash
ql --version
ql --help
```

## Quick Start

```bash
# Add a quest
ql add "Buy milk"
ql add "Write documentation" -d 2026-03-24

# List quests
ql ls              # Show TODO quests
ql ls --done       # Show completed quests
ql ls --all        # Show all quests

# Complete a quest
ql done abc12345

# View your profile
ql me
```

## Detailed Usage

### Adding Quests

```bash
# Add a simple quest
ql add "Complete the report"

# Add a quest with a due date (YYYY-MM-DD format)
ql add "Submit project" -d 2026-03-25

# Using long flag
ql add "Review code" --due=2026-03-26

# Add typed quests (MVP2)
ql add "Morning stretch" -t daily
ql add "Big project" -t epic
ql add "Weekly review" -t weekly

# Add sub-quest under a parent (MVP2)
ql add "Sub task" -t sub -p <parent_id>
```

### Listing Quests

```bash
# Default: Show only TODO quests
ql ls

# Show only completed quests
ql ls --done
ql ls -d

# Show all quests (both TODO and DONE)
ql ls --all
ql ls -a
```

### Completing Quests

```bash
# Complete a quest (use the ID shown in ql ls)
ql done a1b2c3d4

# You'll earn 50 XP and see your progress
# If you level up, you'll see a celebration message!
```

### Viewing Profile

```bash
# Show your current level, XP, and progress
ql me

# Show Flow status with XP multiplier (MVP2)
ql me --flow

# Example output:
# ╔══════════════════════════════════╗
# ║        퀘스트라인 캐릭터         ║
# ╠══════════════════════════════════╣
# ║  레벨: Lv.2                       ║
# ║  칭호: Junior                      ║
# ║  누적 XP: 150                      ║
# ║                                  ║
# ║  다음 레벨까지: 50/200 XP          ║
# ║  [█████░░░░░░░░░░░░░] 25%         ║
# ║                                  ║
# ║  완료한 퀘스트: 3개                ║
# ╚══════════════════════════════════╝
```

### TUI Dashboard (MVP2)

```bash
# Launch interactive Bubble Tea dashboard
ql check

# Keyboard controls:
#   ↑/k, ↓/j    - Navigate
#   →/Enter     - View sub-quests
#   ←/Esc       - Back
#   Space       - Complete/undo quest
#   q/Ctrl+C    - Quit
```

## Commands

| Command | Description |
|---------|-------------|
| `ql add "<title>" [-d YYYY-MM-DD] [-t <type>] [-p <parent>]` | Add a new quest |
| `ql done <id>` | Complete a quest and earn XP |
| `ql ls [--done\|--all] [--type <type>]` | List quests |
| `ql me [--flow]` | Show player profile |
| `ql check` | Launch TUI dashboard |

## Error Codes

| Exit Code | Meaning |
|-----------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid input (e.g., empty title, invalid date format) |
| 3 | Quest not found |
| 4 | Database error |

## Level System

- **XP per Quest**: 50 XP
- **Level Up Formula**: `next_level_xp = 100 + (current_level * 50)`
- **Titles**:
  - Lv.1-9: Intern
  - Lv.10-19: Junior
  - Lv.20-29: Senior
  - Lv.30-49: Lead
  - Lv.50-98: Principal
  - Lv.99: Guru

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run specific package tests
go test ./internal/engine/...
go test ./internal/cli/...
go test ./internal/repository/...
```

### Project Structure

```
questline/
├── cmd/ql/           # CLI entry point
├── internal/
│   ├── cli/          # CLI commands (add, done, ls, me)
│   ├── domain/       # Domain models (Quest, Player)
│   ├── engine/       # XP/leveling logic
│   └── repository/   # SQLite persistence
├── go.mod
├── go.sum
└── README.md
```

### Building for Different Platforms

```bash
# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o ql-darwin-amd64 ./cmd/ql

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o ql-darwin-arm64 ./cmd/ql

# Linux
GOOS=linux GOARCH=amd64 go build -o ql-linux-amd64 ./cmd/ql

# Windows
GOOS=windows GOARCH=amd64 go build -o ql-windows-amd64.exe ./cmd/ql
```

## Troubleshooting

### Permission Denied

If you get "permission denied" when moving the binary to `/usr/local/bin/`, use `sudo`:

```bash
sudo mv ql /usr/local/bin/
```

### Database Locked

If the database is locked, make sure you're not running multiple instances of `ql` simultaneously.

### Reset Data

To reset all your data and start fresh:

```bash
rm ~/.questline/data.db
```

The database will be automatically recreated on the next run.

## Data Storage

Questline stores data in `~/.questline/data.db` (SQLite).

## License

MIT
