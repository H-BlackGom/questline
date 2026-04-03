# Questline MVP 2 데이터 모델 (Data Model)

## 엔티티 다이어그램

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Quest (퀘스트)                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│ id (PK)          TEXT              고유 식별자 (UUID v4 앞 8자)              │
│ title            TEXT              퀘스트 제목                                │
│ status           TEXT              pending | in_progress | pending_completion │
│                  │                 | completed | archived                     │
│ type             TEXT              daily | weekly | epic | guild | sub         │
│ parent_id (FK)   TEXT ←────┐       부모 퀘스트 ID (2-Depth 중첩)            │
│ scheduled_date   TEXT      │       예정 실행일 (YYYY-MM-DD)                  │
│ due_date         TEXT      │       마감일 (YYYY-MM-DD, optional)             │
│ created_at       DATETIME  │       생성 시점                                │
│ completed_at     DATETIME  │       완료 시점                                │
│ deleted_at       DATETIME  │       삭제 시점 (Soft Delete)                   │
└────────────────────────────┼────────────────────────────────────────────────┘
                             │
                             │ 1:N
                             ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                             Player (플레이어)                                │
├─────────────────────────────────────────────────────────────────────────────┤
│ id (PK)          INTEGER CHECK(id=1)  단일 레코드 (싱글톤)                   │
│ level            INTEGER              현재 레벨 (1 시작)                      │
│ total_xp         INTEGER              누적 XP                               │
│ quests_completed INTEGER              완료한 퀘스트 수                       │
│ flow_status      TEXT                 burning | smooth | hazy               │
│ last_synced_at   DATETIME             마지막 동기화 시점                     │
│ last_evaluated   TEXT                 마지막 평가일 (YYYY-MM-DD, 4 AM 기준)  │
│ streak_days      INTEGER              연속 달성 일수                        │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                         QuestHistory (퀘스트 이력)                            │
├─────────────────────────────────────────────────────────────────────────────┤
│ id (PK)          INTEGER              자동 증가                            │
│ quest_id (FK)    TEXT                 퀘스트 ID                              │
│ date             TEXT                 해당 날짜 (YYYY-MM-DD)                  │
│ status           TEXT                 당일 상태                             │
│ completed        BOOLEAN              완료 여부                             │
│ xp_earned        INTEGER              획득 XP (Flow 배율 적용됨)          │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                      DailyEvaluation (일일 평가)                            │
├─────────────────────────────────────────────────────────────────────────────┤
│ id (PK)          INTEGER              자동 증가                            │
│ date             TEXT                 평가 대상일 (YYYY-MM-DD)                │
│ total_routines   INTEGER              전체 루틴 수                         │
│ completed        INTEGER              완료된 루틴 수                        │
│ completion_rate  REAL                 달성률 (0.0 ~ 1.0)                     │
│ flow_grade       TEXT                 BURNING | SMOOTH | HAZY               │
│ evaluated_at     DATETIME           평가 실행 시점                         │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 상태 정의 (State Definitions)

### QuestStatus (퀘스트 상태)

| 상태 | 설명 | 전이 가능한 상태 |
|------|------|------------------|
| `pending` | 대기 중 | in_progress, completed |
| `in_progress` | 진행 중 (하위 퀘스트 진행 시) | pending_completion, completed |
| `pending_completion` | 완료 대기 (하위 완료, 부모 수동 확인 대기) | completed |
| `completed` | 완료 | archived |
| `archived` | 보관됨 | - |

**상태 전이 규칙**:

```go
// Sub 퀘스트 완료 → 부모 in_progress 또는 pending_completion
if quest.Type == Sub && allSiblingsCompleted() {
    parent.Status = pending_completion
}

// 사용자 Space 입력 → 최종 완료
if quest.Status == pending_completion && userConfirms() {
    quest.Status = completed
    quest.CompletedAt = now
    grantXP(quest, player.FlowStatus) // Flow 배율 적용
}

// 새벽 4시 지연 평가 → Daily 루틴 archived 및 새 루틴 생성
if is4AM(evalTime) && quest.Type == Daily && quest.Status == pending {
    quest.Status = archived
    createNewDailyQuest(quest.Title)
}
```

### QuestType (퀘스트 타입)

| 타입 | 설명 | 특징 | 정렬 순서 |
|------|------|------|-----------|
| `daily` | 일일 루틴 | 매일 반복, 새벽 4시 자동 보관/생성 | 1 |
| `weekly` | 주간 퀘스트 | 매주 반복, 일요일 기준 | 2 |
| `epic` | 에픽 퀘스트 | 큰 목표, 여러 Sub 포함 가능 | 3 |
| `guild` | 길드 퀘스트 | 프로젝트/카테고리 그룹핑 | 4 |
| `sub` | 하위 퀘스트 | Epic/Guild의 2-Depth 하위 | - |

**2-Depth 제한**:

```go
// 부모가 Sub인 경우 Sub를 가질 수 없음 (2-Depth 제한)
func CanHaveSubQuests(parent *Quest) bool {
    return parent.Type == Epic || parent.Type == Guild
}

func CanBeSubQuest(child *Quest) bool {
    if child.ParentID == nil {
        return false
    }
    parent := GetQuest(*child.ParentID)
    return parent.Type != Sub // 부모가 Sub면 안 됨
}
```

### FlowStatus (몰입도 상태)

| 상태 | 달성률 | XP 배율 | 설명 |
|------|--------|---------|------|
| `BURNING` | ≥ 80% | 1.5x | 몰입 상태 |
| `SMOOTH` | 50~79% | 1.0x | 기본 상태 |
| `HAZY` | < 50% | 0.5x | 흐름 끊김 |

**Flow 정산 규칙**:

```go
func CalculateFlowGrade(completionRate float64) FlowStatus {
    switch {
    case completionRate >= 0.8:
        return BURNING
    case completionRate >= 0.5:
        return SMOOTH
    default:
        return HAZY
    }
}

// 어제 Daily 루틴 평가 → 오늘 Flow 상태 결정
func EvaluateYesterdayFlow(player *Player, date time.Time) FlowStatus {
    yesterday := date.Add(-24 * time.Hour)
    totalRoutines := CountDailyRoutines(player, yesterday)
    completedRoutines := CountCompletedDailyRoutines(player, yesterday)
    
    if totalRoutines == 0 {
        return SMOOTH // Divide by Zero 방지
    }
    
    rate := float64(completedRoutines) / float64(totalRoutines)
    return CalculateFlowGrade(rate)
}
```

## 데이터베이스 스키마

### 테이블 생성 SQL

```sql
-- quests 테이블 (MVP2 확장)
CREATE TABLE quests (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('pending', 'in_progress', 'pending_completion', 'completed', 'archived'))
        DEFAULT 'pending',
    type TEXT NOT NULL CHECK(type IN ('daily', 'weekly', 'epic', 'guild', 'sub'))
        DEFAULT 'daily',
    parent_id TEXT,
    scheduled_date TEXT,           -- YYYY-MM-DD
    due_date TEXT,                 -- YYYY-MM-DD
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    deleted_at DATETIME,           -- Soft Delete용
    
    FOREIGN KEY (parent_id) REFERENCES quests(id) ON DELETE CASCADE,
    -- 2-Depth 제한: parent의 parent는 NULL이어야 함
    CHECK (
        parent_id IS NULL OR 
        (SELECT type FROM quests AS parent WHERE parent.id = parent_id) IN ('epic', 'guild')
    )
);

-- 인덱스
CREATE INDEX idx_quests_type ON quests(type);
CREATE INDEX idx_quests_status ON quests(status);
CREATE INDEX idx_quests_parent ON quests(parent_id);
CREATE INDEX idx_quests_scheduled ON quests(scheduled_date);
CREATE INDEX idx_quests_deleted ON quests(deleted_at) WHERE deleted_at IS NULL;

-- player 테이블 (MVP2 확장)
CREATE TABLE player (
    id INTEGER PRIMARY KEY CHECK(id = 1),
    level INTEGER DEFAULT 1,
    total_xp INTEGER DEFAULT 0,
    quests_completed INTEGER DEFAULT 0,
    flow_status TEXT DEFAULT 'smooth' CHECK(flow_status IN ('burning', 'smooth', 'hazy')),
    last_synced_at DATETIME,
    last_evaluated TEXT,           -- YYYY-MM-DD
    streak_days INTEGER DEFAULT 0
);

-- quest_history 테이블 (Daily 완료 이력)
CREATE TABLE quest_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    quest_id TEXT NOT NULL,
    date TEXT NOT NULL,            -- YYYY-MM-DD
    status TEXT NOT NULL,
    completed BOOLEAN DEFAULT FALSE,
    xp_earned INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (quest_id) REFERENCES quests(id) ON DELETE CASCADE,
    UNIQUE(quest_id, date)
);

CREATE INDEX idx_history_quest ON quest_history(quest_id);
CREATE INDEX idx_history_date ON quest_history(date);

-- daily_evaluation 테이블 (일일 평가 기록)
CREATE TABLE daily_evaluation (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL UNIQUE,     -- YYYY-MM-DD (평가 대상일)
    total_routines INTEGER NOT NULL,
    completed_routines INTEGER NOT NULL,
    completion_rate REAL NOT NULL,
    flow_grade TEXT NOT NULL CHECK(flow_grade IN ('burning', 'smooth', 'hazy')),
    evaluated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_evaluation_date ON daily_evaluation(date);
```

## TUI 상태 모델 (TUI State Model)

### FocusType (패널 포커스)

```go
type FocusType int

const (
    FocusMaster FocusType = iota  // 좌측: 퀘스트 목록
    FocusDetail                   // 우측: 상세 정보/Sub 퀘스트
)
```

### Model (Bubble Tea)

```go
type Model struct {
    // 패널 상태
    CurrentFocus  FocusType        // 현재 활성화된 패널
    MasterCursor  int              // 좌측 패널 선택 인덱스
    DetailCursor  int              // 우측 패널 선택 인덱스
    
    // 데이터
    Quests        []*QuestNode     // 계층 구조의 퀘스트 목록
    Player        *Player           // 현재 플레이어 상태
    
    // TUI 상태
    Width         int               // 터미널 너비
    Height        int               // 터미널 높이
    Loading       bool              // 로딩 상태
    Error         error             // 에러 상태
    
    // 동기화
    LastSync      time.Time         // 마지막 동기화 시점
    SyncStatus    string            // 동기화 상태 메시지
}

type QuestNode struct {
    Quest       *domain.Quest
    SubQuests   []*QuestNode
    IsExpanded  bool              // 에픽/길드 펼침 여부
    Progress    float64           // 진행률 (0.0 ~ 1.0)
}

// 정렬 순서: Daily → Weekly → Epic → Guild
func (m *Model) GetSortedQuests() []*QuestNode {
    typeOrder := map[domain.QuestType]int{
        domain.Daily:  1,
        domain.Weekly: 2,
        domain.Epic:   3,
        domain.Guild:  4,
    }
    
    sorted := make([]*QuestNode, len(m.Quests))
    copy(sorted, m.Quests)
    
    sort.Slice(sorted, func(i, j int) bool {
        return typeOrder[sorted[i].Quest.Type] < typeOrder[sorted[j].Quest.Type]
    })
    
    return sorted
}
```

## 검증 규칙 (Validation Rules)

### Quest 생성 검증

```go
func ValidateQuestCreate(q *Quest) error {
    // 1. Sub 타입은 반드시 부모가 필요
    if q.Type == Sub && q.ParentID == nil {
        return errors.New("sub quest must have a parent")
    }
    
    // 2. 2-Depth 제한: Sub의 부모는 Epic 또는 Guild여야 함
    if q.Type == Sub && q.ParentID != nil {
        parent := GetQuest(*q.ParentID)
        if parent.Type != Epic && parent.Type != Guild {
            return errors.New("sub quest parent must be epic or guild")
        }
    }
    
    // 3. Daily/Weekly는 parent_id를 가질 수 없음
    if (q.Type == Daily || q.Type == Weekly) && q.ParentID != nil {
        return errors.New("daily and weekly quests cannot have parent")
    }
    
    // 4. scheduled_date는 Daily/Weekly에만 유효
    if q.ScheduledDate != nil && (q.Type != Daily && q.Type != Weekly) {
        return errors.New("scheduled_date is only valid for daily/weekly")
    }
    
    return nil
}
```

### 지연 평가 검증 (5대 방어 로직)

```go
func ValidateLazyEvaluation(result *SyncResult) error {
    // 1. Divide by Zero 방지: 루틴 0개면 rate = 1.0 (SMOOTH)
    if result.TotalRoutines == 0 {
        result.CompletionRate = 1.0
        result.FlowGrade = SMOOTH
    }
    
    // 2. 중복 패널티 방지: 이미 평가된 날짜는 skip
    if IsAlreadyEvaluated(result.Date) {
        return errors.New("already evaluated")
    }
    
    // 3. 무한 과거 방지: 최대 30일까지만 처리
    days := DaysSinceLastSync()
    if days > 30 {
        days = 30 // 또는 에러 처리
    }
    
    // 4. 04:00:00 경계선: 정각은 오늘로 처리
    if IsExactly4AM(time.Now()) {
        result.TargetDate = Today()
    }
    
    // 5. PENDING 보존: 수동 완료 대기 중인 퀘스트는 archive 안 함
    if quest.Status == PendingCompletion {
        quest.ShouldArchive = false
    }
    
    return nil
}
```

## 관계 다이어그램 (Relationship Diagram)

```
Player (1)
  │
  │ 1:1
  ▼
FlowStatus (current)
  │
  │ 1:N (히스토리)
  ▼
DailyEvaluation (N)

Quest (N)
  │
  │ 1:N (self-referential, 2-Depth)
  ├─────► Sub Quest (N)
  │
  │ 1:N (히스토리)
  ▼
QuestHistory (N)

Epic/Guild (1)
  │
  │ 1:N
  ▼
Sub Quest (N, max 2-Depth)
```

## 마이그레이션 전략

### v1 → v2 마이그레이션

```sql
-- 1. 기존 quests 테이블에 컬럼 추가
ALTER TABLE quests ADD COLUMN type TEXT DEFAULT 'daily';
ALTER TABLE quests ADD COLUMN parent_id TEXT;
ALTER TABLE quests ADD COLUMN scheduled_date TEXT;
ALTER TABLE quests ADD COLUMN deleted_at DATETIME;

-- 2. 기존 데이터 마이그레이션
UPDATE quests SET type = 'daily' WHERE type IS NULL;
UPDATE quests SET status = 'pending' WHERE status IS NULL;

-- 3. player 테이블 확장
ALTER TABLE player ADD COLUMN flow_status TEXT DEFAULT 'smooth';
ALTER TABLE player ADD COLUMN last_synced_at DATETIME;
ALTER TABLE player ADD COLUMN last_evaluated TEXT;
ALTER TABLE player ADD COLUMN streak_days INTEGER DEFAULT 0;

-- 4. 신규 테이블 생성
-- quest_history, daily_evaluation (위 SQL 참조)

-- 5. 인덱스 생성
CREATE INDEX idx_quests_type ON quests(type);
CREATE INDEX idx_quests_parent ON quests(parent_id);
```

**하위 호환성**: 기존 `quests` 데이터는 `type = 'daily'`로 자동 변환됨
