# Questline MVP 2 Implementation Plan (구현 계획)

**브랜치**: `feature/mvp2-master` | **날짜**: 2026-04-02 | **스펙**: [requirement_2.md](/Users/hyoungrolee/Documents/1_Project/01_dev/questline/.opencode/document/requirement_2.md)
**입력**: Questline MVP 2 통합 테크 스펙

## 개요

MVP 2에서는 텍스트 기반 CLI를 넘어 **Bubble Tea 기반의 TUI 대시보드**를 구현하고, 지연 평가(Lazy Evaluation)를 통한 시간 동기화와 Flow(몰입도) 시스템을 도입하여 Gamification을 극대화합니다. 마스터-디테일 구조의 인터페이스와 새벽 4시 기준의 일일 습관 형성 메커니즘이 핵심입니다.

## 기술 컨텍스트

| 항목 | 값 |
|------|-----|
| **언어/버전** | Go 1.23+ |
| **주요 의존성** | Cobra (CLI), Bubble Tea (TUI), Lip Gloss (스타일), pure-go SQLite |
| **저장소** | SQLite (`~/.questline/data.db`) |
| **테스팅** | Go 테스트 + Table-driven tests |
| **타겟 플랫폼** | Linux, macOS, Windows (터미널) |
| **프로젝트 유형** | CLI/TUI 하이브리드 도구 |
| **성능 목표** | TUI 60fps, Lazy Eval < 100ms |
| **제약사항** | 백그라운드 데몬 없음, 단일 바이너리 배포 |
| **스코프/규모** | 5개 Task로 분할, 2-Depth 중첩 퀘스트 지원 |

## 헌법 체크리스트

**Questline MVP2 헌법 원칙 검증**:

- [x] **I. 계층형 아키텍처**: cmd/ (CLI 명령어), internal/tui/ (TUI), internal/service/ (비즈니스 로직) 분리
- [x] **II. CLI 명령어 설계**: `ql check`로 TUI 진입, Vim-like 단축키 (j/k, Enter, Esc)
- [x] **III. 데이터 지속성**: pure-go SQLite 사용, `quests` 테이블에 `parent_id`, `type`, `deleted_at` 추가
- [x] **IV. RPG 게이미피케이션**: Flow 시스템 (BURNING/SMOOTH/HAZY), XP 배율 (1.5x/1.0x/0.5x)
- [x] **VII. TUI 아키텍처**: Bubble Tea Elm 아키텍처 (Model → Update → View)
- [x] **VIII. 퀘스트 타입 시스템**: Daily, Weekly, Epic, Guild, Sub (2-Depth)
- [x] **IX. 지연 평가**: 새벽 4시 기준 `time.Now().Add(-4 * time.Hour)`, 5대 방어 로직
- [x] **X. Flow 시스템**: Daily 달성률 기반 (80%+/50-79%/50%-)

**검증 실패 시**: 해당 원칙을 준수하는 방향으로 설계 재검토

## 프로젝트 구조

### 문서 (이 기능)

```text
specs/mvp2/
├── plan.md              # 이 파일 (구현 계획)
├── research.md          # 연구 문서 (필요시)
├── data-model.md        # 데이터 모델 정의
├── quickstart.md        # 퀵스타트 가이드
├── contracts/           # 인터페이스 계약
└── tasks.md             # 태스크 목록
```

### 소스 코드 (레포지토리 루트)

```text
questline/
├── cmd/ql/
│   └── main.go              # CLI 진입점
├── internal/
│   ├── cli/
│   │   ├── root.go
│   │   ├── add.go           # -t TYPE, -p PARENT_ID 옵션 추가
│   │   ├── done.go
│   │   ├── ls.go            # --type 필터 추가
│   │   ├── me.go            # Flow 상태 표시
│   │   └── check.go         # TUI 진입점 (NEW)
│   ├── tui/                 # NEW: Bubble Tea TUI
│   │   ├── model.go         # FocusMaster/FocusDetail 상태
│   │   ├── update.go        # KeyMsg 핸들러 (j/k, Enter, Esc, Space)
│   │   ├── view.go          # 마스터-디테일 레이아웃
│   │   ├── theme/
│   │   │   └── styles.go    # Lip Gloss 스타일 정의
│   │   └── views/
│   │       ├── master.go    # 좌측 퀘스트 목록
│   │       └── detail.go    # 우측 상세 정보
│   ├── domain/
│   │   ├── quest.go         # QuestType enum 추가
│   │   ├── player.go        # FlowStatus 추가
│   │   └── evaluation.go    # SyncResult, FlowGrade
│   ├── engine/
│   │   ├── leveling.go
│   │   ├── flow.go          # Flow 계산 로직
│   │   └── sync.go          # Lazy Evaluation (4 AM 기준)
│   ├── service/
│   │   ├── quest_service.go # PENDING 상태 전환 로직
│   │   └── sync_service.go  # 지연 평가 오케스트레이션
│   └── repository/
│       ├── sqlite.go        # 마이그레이션
│       ├── quest_repo.go    # 계층 조회, soft delete
│       └── player_repo.go   # flow_status 업데이트
└── go.mod
```

### 브랜치 전략

| Task ID | 목표 | 브랜치 |
|---------|------|--------|
| Task 6 | DB 마이그레이션 & 모델 확장 | `feature/mvp2-db-migration` |
| Task 7 | 코어 엔진: 지연 평가(Sync) | `feature/mvp2-lazy-eval` |
| Task 8 | 계층 구조 비즈니스 로직 | `feature/mvp2-quest-hierarchy` |
| Task 9 | TUI 아키텍처: 모델 및 레이아웃 | `feature/mvp2-tui-layout` |
| Task 10 | TUI 인터랙션: Update 루프 | `feature/mvp2-tui-interaction` |

## 복잡도 추적

| 위반 사유 | 필요 이유 | 더 간단한 대안이 거부된 이유 |
|-----------|-----------|------------------------------|
| 5개 Task 분할 | 독립적인 테스트와 배포를 위해 | 단일 커밋으로는 리뷰와 롤백이 어려움 |
| TUI 계층 추가 | CLI와 TUI의 관심사 분리 | 한 파일에 합치면 테스트와 재사용이 어려움 |
| 5대 방어 로직 | Edge case 안정성 보장 | 생략 시 데이터 무결성 훼손 위험 |

## Phase 0: 연구 (Research)

**NEEDS CLARIFICATION 해소 완료**:

| 항목 | 결정 | 근거 | 대안 |
|------|------|------|------|
| TUI 프레임워크 | Bubble Tea v1.x | Go에서 가장 성숙, Lip Gloss와 통합 | Termui (덜 성숙), Tview (복잡함) |
| 마이그레이션 도구 | 수동 SQL | MVP2 범위, 단순 ALTER TABLE | golang-migrate (오버엔지니어링) |
| 시간 계산 | `time.Now().Add(-4 * time.Hour)` | 새벽 4시 기준 논리적 일자 | Cron (데몬 필요, 제외) |
| Flow 등급 | 3단계 (BURNING/SMOOTH/HAZY) | 요구사항에 명시됨 | 5단계 (복잡도 증가) |

**연구 산출물**: `research.md` (필요시 생성)

## Phase 1: 설계 및 계약 (Design & Contracts)

### 데이터 모델 (data-model.md)

**핵심 엔티티**:

1. **Quest (확장)**:
   - `id`, `title`, `status` (existing)
   - `type`: enum (daily, weekly, epic, guild, sub)
   - `parent_id`: FK to Quest (nullable, 2-Depth 제한)
   - `scheduled_date`: YYYY-MM-DD (루틴/주간용)
   - `deleted_at`: soft delete

2. **Player (확장)**:
   - `level`, `total_xp`, `quests_completed` (existing)
   - `flow_status`: enum (burning, smooth, hazy)
   - `last_synced_at`: 마지막 동기화 시점
   - `last_evaluated_date`: 마지막 평가일 (4 AM 기준)

3. **SyncResult (NEW)**:
   - `evaluated_date`: 평가 대상일
   - `daily_total`: 총 데일리 수
   - `daily_completed`: 완료된 데일리 수
   - `completion_rate`: 달성률 (%)
   - `new_flow_grade`: 새로운 Flow 등급

4. **TUIState (NEW)**:
   - `current_focus`: master | detail
   - `master_cursor`: 좌측 패널 선택 인덱스
   - `detail_cursor`: 우측 패널 선택 인덱스
   - `quests`: 현재 표시 중인 퀘스트 목록

**상태 전이**:

```
Sub Quest 완료:
  pending → completed
  → 모든 Sub completed → 부모: pending_completion (PENDING)

부모 Quest 완료 (사용자 Space 입력):
  pending_completion → completed
  → XP 지급 (Flow 배율 적용)

Daily 루틴 새벽 4시:
  completed → archived
  → 새 Daily 생성 (scheduled_date = 오늘)

Flow 정산 (새벽 4시):
  completion_rate >= 80% → BURNING (1.5x)
  completion_rate >= 50% → SMOOTH (1.0x)
  completion_rate < 50% → HAZY (0.5x)
```

### 인터페이스 계약 (contracts/)

**QuestService**:

```go
type QuestService interface {
    // Quest 생성 (타입 지정, 부모 지정 가능)
    CreateQuest(title string, questType QuestType, parentID *string, dueDate *time.Time) (*Quest, error)
    
    // Quest 완료 (Sub 완료 시 부모 PENDING 전환)
    CompleteQuest(questID string) (*CompletionResult, error)
    
    // 퀘스트 목록 조회 (타입별 필터, 계층 포함)
    ListQuests(filter QuestFilter) ([]*QuestNode, error)
    
    // 마스터-디테일용 계층 조회
    GetQuestTree() ([]*QuestNode, error)
}

type QuestNode struct {
    Quest       *Quest
    SubQuests   []*QuestNode
    Progress    float64  // 에픽/길드 진행률
}
```

**SyncService**:

```go
type SyncService interface {
    // 지연 평가 실행 (앱 시작 시 호출)
    // last_synced_at ~ 현재까지의 시간 동기화
    EvaluateLazySync() (*SyncResult, error)
    
    // 특정 날짜의 Flow 등급 계산
    CalculateFlowGrade(date time.Time) (*FlowGrade, error)
}

type SyncResult struct {
    ProcessedDays     int
    EvaluatedDates    []string
    FlowAdjustments   []*FlowAdjustment
    ArchivedQuests    int
    NewQuests         int
}
```

**TUI Commands** (Bubble Tea Msg):

```go
// KeyMsg 기반 명령
type ToggleQuestMsg struct { QuestID string }
type NavigateMsg struct { Direction Direction }
type FocusChangeMsg struct { Panel FocusType }
```

### 퀵스타트 (quickstart.md)

```bash
# 1. DB 마이그레이션 (기존 데이터 보존)
ql --migrate

# 2. 새로운 퀘스트 타입으로 생성
ql add "매일 운동 30분" -t daily
ql add "주간 회의 준비" -t weekly -d 2026-04-07
ql add "웹툰 연재" -t epic
ql add "캐릭터 디자인" -t sub -p <epic_id>

# 3. TUI 대시보드 실행
ql check

# TUI 조작법:
#   j/k 또는 ↑/↓: 목록 이동
#   Enter/→: 하위 패널 진입 (Sub 보기)
#   Esc/←: 상위 패널 복귀
#   Space: 퀘스트 완료
#   q: 종료

# 4. Flow 상태 확인
ql me --flow
```

## Phase 2: 태스크 목록 (Tasks)

**단계별 실행 순서**:

### Phase 2.1: 기반 (Foundation) - Task 6
- [ ] DB 마이그레이션 스크립트 작성
- [ ] Quest 모델 확장 (type, parent_id, deleted_at)
- [ ] Player 모델 확장 (flow_status, last_synced_at)
- [ ] Repository 계층 조회 메서드 추가

### Phase 2.2: 핵심 로직 (Core Logic) - Task 7, 8
- [ ] 지연 평가 엔진 구현 (4 AM 기준)
- [ ] 5대 방어 로직 단위 테스트
- [ ] Flow 계산 로직 (BURNING/SMOOTH/HAZY)
- [ ] PENDING 상태 전환 로직
- [ ] QuestService 계층 구조 메서드

### Phase 2.3: TUI 구현 (TUI Implementation) - Task 9, 10
- [ ] Bubble Tea Model 정의 (FocusMaster/FocusDetail)
- [ ] Update 루프 구현 (KeyMsg 핸들러)
- [ ] 마스터-디테일 View 레이아웃
- [ ] Lip Gloss 테마 스타일링
- [ ] `ql check` 명령어 연결

### Phase 2.4: 통합 및 테스트 (Integration)
- [ ] End-to-end 테스트 (TUI 흐름)
- [ ] 5대 Edge Case 검증
- [ ] Flow 배율 적용 확인
- [ ] 기존 데이터와의 하위 호환성

## 에이전트 컨텍스트 업데이트

**실행 명령**:
```bash
.specify/scripts/bash/update-agent-context.sh opencode
```

**추가될 컨텍스트**:
- Bubble Tea TUI 아키텍처 (Model/Update/View)
- 지연 평가 (Lazy Evaluation) 패턴
- Flow 시스템 및 XP 배율 로직
- 마스터-디테일 UI 패턴

## 생성된 산출물

| 파일 | 경로 | 상태 |
|------|------|------|
| 구현 계획 | `specs/mvp2/plan.md` | ✅ 완료 |
| 데이터 모델 | `specs/mvp2/data-model.md` | 📝 작성 필요 |
| 퀵스타트 | `specs/mvp2/quickstart.md` | 📝 작성 필요 |
| 계약 | `specs/mvp2/contracts/` | 📝 작성 필요 |
| 태스크 목록 | `specs/mvp2/tasks.md` | 📝 작성 필요 |

## 다음 단계

1. `data-model.md` 작성: 엔티티, 상태 전이, 관계 정의
2. `quickstart.md` 작성: 개발자용 테스트 가이드
3. `contracts/` 작성: 인터페이스 정의
4. `tasks.md` 작성: 세부 태스크 분할
5. 각 Task 브랜치에서 구현 시작

## 참고 자료

- [헌법](/Users/hyoungrolee/Documents/1_Project/01_dev/questline/.specify/memory/constitution.md) v2.0.0
- [요구사항](/Users/hyoungrolee/Documents/1_Project/01_dev/questline/.opencode/document/requirement_2.md)
- [Bubble Tea Docs](https://github.com/charmbracelet/bubbletea)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss)
