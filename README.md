# Questline

A gamified task management CLI tool with RPG-style progression.

## Features

- **Quest Management**: Create and complete tasks with due dates
- **XP System**: Earn 50 XP for each completed quest
- **Level Up**: Progress through levels with increasing difficulty
- **Titles**: Unlock titles from Intern to Guru as you level up

## Installation

### Build from Source

```bash
# Clone the repository
git clone https://github.com/H-BlackGom/questline.git
cd questline

# Build
go build -o ql ./cmd/ql

# Optional: Move to PATH
mv ql /usr/local/bin/
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

## Commands

| Command | Description |
|---------|-------------|
| `ql add "<title>" [-d YYYY-MM-DD]` | Add a new quest |
| `ql done <id>` | Complete a quest and earn XP |
| `ql ls [--done\|--all]` | List quests |
| `ql me` | Show player profile |

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

## Data Storage

Questline stores data in `~/.questline/data.db` (SQLite).

## License

MIT
