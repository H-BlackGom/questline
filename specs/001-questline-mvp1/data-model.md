# 데이터 모델: Questline MVP 1

**생성일**: 2025-03-20
**문서 경로**: `/specs/001-questline-mvp1/data-model.md`

---

## 1. 개요

Questline MVP 1의 핵심 데이터 모델은 두 개의 엔티티로 구성됩니다:
1. **Quest** (퀘스트): 사용자가 수행해야 할 작업
2. **Player** (플레이어): 사용자의 진행 상태와 레벨 정보

---

## 2. 엔티티 정의

### 2.1. Quest (퀘스트)

사용자가 생성하고 완료하는 작업 단위입니다.

#### 속성 (Fields)

| 필드명 | 타입 | 필수 | 기본값 | 설명 |
|--------|------|------|--------|------|
| id | string | Yes | UUID v4 | 고유 식별자 |
| title | string | Yes | - | 퀘스트 제목 |
| status | Status | Yes | TODO | TODO, DONE, DROPPED |
| due_date | *time.Time | No | nil | 마감일 (선택, YYYY-MM-DD) |
| due_date | *time.Time | No | nil | 마감일 (선택) |
| created_at | time.Time | Yes | now | 생성 시간 |
| completed_at | *time.Time | No | nil | 완료 시간 |

#### 상태 (Status)

```go
type Status string

const (
    StatusTODO    Status = "TODO"    // 진행 예정
    StatusDONE    Status = "DONE"    // 완료됨
    StatusDROPPED Status = "DROPPED" // 포기됨
)
```

#### 유효성 검증 (Validation)

- `title`: 비어있지 않아야 함, 최대 200자
- `status`: 유효한 Status 값이어야 함
- `due_date`: 있다면 미래 날짜여야 함

#### 상태 (Status)
- `TODO`: 진행 예정
- `DONE`: 완료됨
- `DROPPED`: 예약 상태 (MVP2에서 구현 예정)

---

### 2.2. Player (플레이어)

사용자의 게임 진행 상태를 저장하는 단일 레코드입니다.

#### 속성 (Fields)

| 필드명 | 타입 | 필수 | 기본값 | 설명 |
|--------|------|------|--------|------|
| id | int | Yes | 1 | 고정값 (단일 로우) |
| level | int | Yes | 1 | 현재 레벨 |
| current_xp | int | Yes | 0 | 현재 레벨에서의 XP |
| total_xp_earned | int | Yes | 0 | 누적 XP (감소 없음) |
| quests_completed | int | Yes | 0 | 완료한 퀘스트 수 |
| updated_at | time.Time | Yes | now | 마지막 업데이트 |

#### 칭호 (Title)

```go
// GetTitle 레벨별 칭호 반환
//
// Lv.1~9: Intern
// Lv.10~19: Junior
// Lv.20~29: Senior
// Lv.30~49: Lead
// Lv.50~98: Principal
// Lv.99: Guru

func (p *Player) GetTitle() string {
    switch {
    case p.level >= 99:
        return "Guru"
    case p.level >= 50:
        return "Principal"
    case p.level >= 30:
        return "Lead"
    case p.level >= 20:
        return "Senior"
    case p.level >= 10:
        return "Junior"
    default:
        return "Intern"
    }
}
```

#### 레벨업 계산

```go
// GetRequiredXPForNextLevel 다음 레벨 필요 XP
// 공식: 100 + (현재 레벨 * 50)
//
// Lv.1 → Lv.2: 150 XP 필요
// Lv.2 → Lv.3: 200 XP 필요
// Lv.3 → Lv.4: 250 XP 필요

func (p *Player) GetRequiredXPForNextLevel() int {
    return 100 + (p.level * 50)
}

// GetProgressPercent 다음 레벨 진행률 (0-100)
func (p *Player) GetProgressPercent() int {
    required := p.GetRequiredXPForNextLevel()
    if required == 0 {
        return 100
    }
    percent := (p.current_xp * 100) / required
    if percent > 100 {
        return 100
    }
    return percent
}
```

#### 유효성 검증 (Validation)

- `level`: 1 이상
- `current_xp`: 0 이상, 현재 레벨의 최대 XP 미만
- `total_xp_earned`: 0 이상, 누적값

---

## 3. 관계 (Relationships)

```
┌─────────┐       ┌─────────┐
│ Player  │   1:N │ Quest   │
│ (1 로우)│◄──────│ (N 로우)│
└─────────┘       └─────────┘
```

**비즈니스 규칙**:
- Player는 반드시 존재 (앱 초기화 시 자동 생성)
- Quest 완료 시 Player XP 증가 및 레벨업 가능
- Quest 삭제 시 Player 통계에는 영향 없음

---

## 4. 데이터베이스 스키마

### 4.1. 마이그레이션

```sql
-- 001_initial.sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO schema_migrations (version) VALUES (1);
```

### 4.2. Quests 테이블

```sql
CREATE TABLE quests (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    status TEXT DEFAULT 'TODO' CHECK (status IN ('TODO', 'DONE', 'DROPPED')),
    due_date TEXT,  -- YYYY-MM-DD format
    created_at TEXT NOT NULL,  -- ISO 8601
    completed_at TEXT,  -- ISO 8601
    
    -- 제약조건
    CHECK (length(title) > 0 AND length(title) <= 200)
);

-- 인덱스
CREATE INDEX idx_quests_status ON quests(status);
CREATE INDEX idx_quests_created ON quests(created_at DESC);
```

### 4.3. Player 테이블

```sql
CREATE TABLE player (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    level INTEGER DEFAULT 1 CHECK (level >= 1),
    current_xp INTEGER DEFAULT 0 CHECK (current_xp >= 0),
    total_xp_earned INTEGER DEFAULT 0 CHECK (total_xp_earned >= 0),
    quests_completed INTEGER DEFAULT 0 CHECK (quests_completed >= 0),
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 초기 데이터
INSERT INTO player (id) VALUES (1);
```

---

## 5. Go 구조체

### 5.1. Quest 구조체

```go
package domain

import (
    "time"
    "github.com/google/uuid"
)

type Quest struct {
    ID           string     `json:"id"`
    Title        string     `json:"title"`
    Status       Status     `json:"status"`
    DueDate      *time.Time `json:"due_date,omitempty"`
    CreatedAt    time.Time  `json:"created_at"`
    CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// NewQuest 새로운 퀘스트 생성 (MVP1: 고정 50 XP)
func NewQuest(title string, dueDate *time.Time) *Quest {
    return &Quest{
        ID:         uuid.New().String()[:8], // 첫 8자만 사용 (간결성)
        Title:      title,
        Status:     StatusTODO,
        DueDate:    dueDate,
        CreatedAt:  time.Now(),
    }
}

// Complete 퀘스트 완료 처리
func (q *Quest) Complete() {
    now := time.Now()
    q.Status = StatusDONE
    q.CompletedAt = &now
}

// IsOverdue 마감일 지남 여부
func (q *Quest) IsOverdue() bool {
    if q.DueDate == nil || q.Status == StatusDONE {
        return false
    }
    return time.Now().After(*q.DueDate)
}
```

### 5.2. Player 구조체

```go
package domain

import "time"

type Player struct {
    ID              int       `json:"id"`
    Level           int       `json:"level"`
    CurrentXP       int       `json:"current_xp"`
    TotalXPEarned   int       `json:"total_xp_earned"`
    QuestsCompleted int       `json:"quests_completed"`
    UpdatedAt       time.Time `json:"updated_at"`
}

// NewPlayer 새로운 플레이어 생성 (초기값)
func NewPlayer() *Player {
    return &Player{
        ID:        1,
        Level:     1,
        CurrentXP: 0,
        UpdatedAt: time.Now(),
    }
}

// AddXP XP 획득 및 레벨업 처리
func (p *Player) AddXP(amount int) (leveledUp bool, levelsGained int) {
    if amount <= 0 {
        return false, 0
    }
    
    p.CurrentXP += amount
    p.TotalXPEarned += amount
    levelsGained = 0
    
    for p.CurrentXP >= p.GetRequiredXPForNextLevel() {
        p.CurrentXP -= p.GetRequiredXPForNextLevel()
        p.Level++
        levelsGained++
    }
    
    p.QuestsCompleted++
    p.UpdatedAt = time.Now()
    
    return levelsGained > 0, levelsGained
}
```

---

## 6. Repository 인터페이스

### 6.1. QuestRepository

```go
package repository

import "questline/internal/domain"

type QuestRepository interface {
    Create(quest *domain.Quest) error
    GetByID(id string) (*domain.Quest, error)
    Update(quest *domain.Quest) error
    Delete(id string) error
    ListByStatus(status domain.Status) ([]*domain.Quest, error)
    ListAll() ([]*domain.Quest, error)
}
```

### 6.2. PlayerRepository

```go
package repository

type PlayerRepository interface {
    Get() (*domain.Player, error)
    Update(player *domain.Player) error
    AddXP(amount int) (leveledUp bool, newLevel int, err error)
}
```

---

## 7. 검증 규칙 요약

### Quest
- [ ] ID는 고유해야 함 (UUID)
- [ ] Title은 비어있지 않고 200자 이하
- [ ] Status는 TODO/DONE/DROPPED 중 하나
- [ ] DueDate가 있다면 미래 날짜

### Player
- [ ] ID는 항상 1 (단일 로우)
- [ ] Level은 1 이상
- [ ] CurrentXP는 현재 레벨의 최대 XP 미만
- [ ] TotalXPEarned는 항상 CurrentXP 이상

---

## 8. 샘플 데이터

### 초기 플레이어
```json
{
  "id": 1,
  "level": 1,
  "current_xp": 0,
  "total_xp_earned": 0,
  "quests_completed": 0
}
```

### 샘플 퀘스트
```json
{
  "id": "a1b2c3d4",
  "title": "코드 리뷰하기",
  "status": "TODO",
  "difficulty": "NORMAL",
  "xp_reward": 50,
  "due_date": "2025-03-25T00:00:00Z",
  "created_at": "2025-03-20T10:30:00Z"
}
```

---

**참고**: 모든 데이터 모델은 헌법 원칙(I-VI)을 준수하여 설계되었습니다.
