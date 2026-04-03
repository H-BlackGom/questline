# Tasks: Questline MVP 2

**입력**: [MVP2 요구사항](/Users/hyoungrolee/Documents/1_Project/01_dev/questline/.opencode/document/requirement_2.md)
**사전조건**: [plan.md](./plan.md), [data-model.md](./data-model.md), [contracts/](./contracts/)

## Phase 1: 기반 (Foundation) - Task 6

**브랜치**: `feature/mvp2-db-migration`
**목적**: DB 마이그레이션과 모델 확장

### Phase 1.1: 데이터베이스 마이그레이션

- [ ] **T001** [P] 마이그레이션 파일 작성: `internal/repository/migrations/002_mvp2.up.sql`
  - quests 테이블: `type`, `parent_id`, `scheduled_date`, `deleted_at` 컬럼 추가
  - player 테이블: `flow_status`, `last_synced_at`, `last_evaluated`, `streak_days` 컬럼 추가
  - 새 테이블: `quest_history`, `daily_evaluation` 생성
  - 인덱스 생성: `idx_quests_type`, `idx_quests_parent`, `idx_quests_scheduled`
  - 태그: `[DB]`, `[TYPE]`

- [ ] **T002** 롤백 마이그레이션 작성: `internal/repository/migrations/002_mvp2.down.sql`
  - 추가된 컬럼 제거
  - 새 테이블 제거
  - 태그: `[DB]`

### Phase 1.2: 도메인 모델 확장

- [ ] **T003** [P] Quest 타입 정의: `internal/domain/quest_type.go`
  - `QuestType` enum: Daily, Weekly, Epic, Guild, Sub
  - `QuestStatus` 업데이트: Pending, InProgress, PendingCompletion, Completed, Archived
  - 유효성 검사 메서드
  - 태그: `[ARCH]`, `[TYPE]`

- [ ] **T004** [P] Flow 상태 정의: `internal/domain/flow.go`
  - `FlowStatus` enum: Burning, Smooth, Hazy
  - `FlowGrade` struct
  - `CalculateFlowGrade(rate float64)` 함수
  - 태그: `[ARCH]`, `[RPG]`, `[FLOW]`

- [ ] **T005** Quest 모델 확장: `internal/domain/quest.go` 수정
  - 새 필드: `Type`, `ParentID`, `ScheduledDate`, `DeletedAt`
  - `IsSubQuest()`, `CanHaveSubQuests()` 메서드
  - 태그: `[ARCH]`, `[TYPE]`

- [ ] **T006** Player 모델 확장: `internal/domain/player.go` 수정
  - 새 필드: `FlowStatus`, `LastSyncedAt`, `LastEvaluated`, `StreakDays`
  - `GetFlowMultiplier()` 메서드
  - 태그: `[ARCH]`, `[RPG]`, `[FLOW]`

### Phase 1.3: Repository 계층

- [ ] **T007** [P] Quest Repository 확장: `internal/repository/quest_repo.go` 수정
  - `GetByType()`, `GetByParent()`, `GetTree()` 메서드 추가
  - Soft delete 지원: `Delete()` → `deleted_at` 업데이트
  - 계층 조회 쿼리
  - 태그: `[DB]`, `[TYPE]`

- [ ] **T008** [P] Player Repository 확장: `internal/repository/player_repo.go` 수정
  - `UpdateFlowStatus()`, `UpdateSyncTime()` 메서드
  - 태그: `[DB]`, `[FLOW]`

- [ ] **T009** History Repository: `internal/repository/history_repo.go` (신규)
  - `SaveQuestHistory()`, `GetDailyStats()` 메서드
  - `SaveDailyEvaluation()`, `GetEvaluations()` 메서드
  - 태그: `[DB]`, `[FLOW]`

**체크포인트**: Task 6 완료 - DB 마이그레이션 및 모델 확장 완료

---

## Phase 2: 코어 엔진 (Core Engine) - Task 7, 8

**브랜치**: `feature/mvp2-lazy-eval`, `feature/mvp2-quest-hierarchy`
**목적**: 지연 평가, Flow 계산, 퀘스트 계층 구조 비즈니스 로직

### Phase 2.1: 지연 평가 엔진 (Lazy Evaluation)

- [ ] **T010** 시간 계산 유틸리티: `internal/engine/time_calc.go`
  - `GetLogicalDate(t time.Time) time.Time` - 새벽 4시 기준 논리적 일자
  - `DaysBetween(from, to time.Time) int`
  - `Is4AM(t time.Time) bool`
  - 태그: `[LAZY]`

- [ ] **T011** [P] 지연 평가 엔진: `internal/engine/sync.go`
  - `EvaluateLazySync(lastSync time.Time) (*SyncResult, error)`
  - 5대 방어 로직 구현:
    1. Divide by Zero 방지 (루틴 0개)
    2. 중복 패널티 방지 (이미 평가된 날짜 skip)
    3. 무한 과거 방지 (최대 처리 일수 제한)
    4. 04:00:00 경계선 처리
    5. PENDING 상태 보존
  - 태그: `[LAZY]`, `[QUALITY]`

- [ ] **T012** 지연 평가 테스트: `internal/engine/sync_test.go`
  - 5대 Edge Case 단위 테스트
  - `TestLazyEval_DivideByZero`
  - `TestLazyEval_DuplicatePenalty`
  - `TestLazyEval_InfinitePast`
  - `TestLazyEval_Boundary4AM`
  - `TestLazyEval_PendingPreserve`
  - 태그: `[LAZY]`, `[QUALITY]`

### Phase 2.2: Flow 계산 엔진

- [ ] **T013** Flow 계산기: `internal/engine/flow.go`
  - `CalculateDailyFlow(date time.Time) (*FlowGrade, error)`
  - `ApplyFlowMultiplier(baseXP int, grade FlowStatus) int`
  - 루틴 달성률 계산
  - 태그: `[FLOW]`, `[RPG]`

- [ ] **T014** Flow 테스트: `internal/engine/flow_test.go`
  - 80% → BURNING (1.5x)
  - 60% → SMOOTH (1.0x)
  - 30% → HAZY (0.5x)
  - 0% → SMOOTH (Divide by Zero 방지)
  - 태그: `[FLOW]`, `[QUALITY]`

### Phase 2.3: 퀘스트 계층 로직

- [ ] **T015** Quest Service 계층: `internal/service/quest_service.go` 수정
  - `CreateQuest()` 타입 검증 및 2-Depth 제한
  - `CompleteQuest()` Sub 완료 시 부모 PENDING 전환
  - `GetQuestTree()` 정렬 (Daily → Weekly → Epic → Guild)
  - 태그: `[TYPE]`, `[ARCH]`

- [ ] **T016** PENDING 상태 전이 로직: `internal/service/quest_service.go`
  - `checkParentTransition(questID string) error`
  - 모든 Sub 완료 체크
  - 부모 상태 PendingCompletion으로 변경
  - 태그: `[TYPE]`

- [ ] **T017** 계층 구조 테스트: `internal/service/quest_service_test.go`
  - Sub 생성 시 부모 검증
  - Sub 완료 시 부모 PENDING 전환 검증
  - 부모 완료 시 XP 지급 검증
  - 태그: `[TYPE]`, `[QUALITY]`

### Phase 2.4: CLI 명령어 확장

- [ ] **T018** `ql add` 확장: `internal/cli/add.go` 수정
  - `-t TYPE` 옵션: daily, weekly, epic, guild, sub
  - `-p PARENT_ID` 옵션: 부모 퀘스트 지정
  - 타입별 유효성 검사
  - 태그: `[CLI]`, `[TYPE]`

- [ ] **T019** `ql done` 확장: `internal/cli/done.go` 수정
  - Flow 배율 적용된 XP 표시
  - 레벨업 메시지 개선
  - 태그: `[CLI]`, `[RPG]`, `[FLOW]`

- [ ] **T020** `ql ls` 확장: `internal/cli/ls.go` 수정
  - `--type TYPE` 필터 옵션
  - 타입별 색상 표시
  - 태그: `[CLI]`, `[TYPE]`

- [ ] **T021** `ql me` 확장: `internal/cli/me.go` 수정
  - Flow 상태 표시 (BURNING/SMOOTH/HAZY)
  - `--flow` 옵션: 상세 Flow 히스토리
  - 연속 달성 일수 표시
  - 태그: `[CLI]`, `[RPG]`, `[FLOW]`

**체크포인트**: Task 7, 8 완료 - 지연 평가 및 계층 로직 완료

---

## Phase 3: TUI 구현 (TUI Implementation) - Task 9, 10

**브랜치**: `feature/mvp2-tui-layout`, `feature/mvp2-tui-interaction`
**목적**: Bubble Tea 기반 마스터-디테일 TUI 대시보드

### Phase 3.1: TUI 기반 구조

- [ ] **T022** [P] TUI 모델 정의: `internal/tui/model.go`
  - `Model` struct: CurrentFocus, MasterCursor, DetailCursor
  - `QuestNode` 트리 구조
  - `FocusType` enum: FocusMaster, FocusDetail
  - 태그: `[TUI]`, `[ARCH]`

- [ ] **T023** [P] 테마 스타일링: `internal/tui/theme/styles.go`
  - Lip Gloss 스타일 정의
  - 퀘스트 타입별 색상 (Daily: Blue, Weekly: Purple, Epic: Orange, Guild: Green, Sub: Gray)
  - Flow 상태별 스타일 (BURNING: Red, SMOOTH: Blue, HAZY: Yellow)
  - 태그: `[TUI]`

- [ ] **T024** TUI 초기화: `internal/tui/init.go`
  - `InitialModel()` 팩토리 함수
  - DB 연결 및 데이터 로드
  - 태그: `[TUI]`, `[DB]`

### Phase 3.2: Update 루프 (Key Handling)

- [ ] **T025** Update 핸들러: `internal/tui/update.go`
  - `tea.KeyMsg` 처리:
    - `j/k`, `↑/↓`: 커서 이동 (현재 포커스된 패널)
    - `→`, `Enter`: 하위 패널 진입 (Sub 보기)
    - `←`, `Esc`: 상위 패널 복귀
    - `Space`: 퀘스트 완료 토글
    - `q`, `Ctrl+C`: 종료
  - `tea.WindowSizeMsg`: 터미널 크기 변경 처리
  - 태그: `[TUI]`

- [ ] **T026** [P] 네비게이션 로직: `internal/tui/navigation.go`
  - `moveCursor(m *Model, dir Direction)`
  - `enterDetail(m *Model)` - Master → Detail 전환
  - `exitDetail(m *Model)` - Detail → Master 복귀
  - 커서 경계 검사
  - 태그: `[TUI]`

- [ ] **T027** 액션 핸들러: `internal/tui/actions.go`
  - `toggleQuest(m *Model)` - Space 키 처리
  - DB 업데이트 Command 생성
  - Sub 완료 시 부모 PENDING 전환
  - 태그: `[TUI]`, `[TYPE]`

- [ ] **T028** Commands: `internal/tui/commands.go`
  - `LoadQuestsCmd()` - 비동기 퀘스트 로드
  - `ToggleQuestCmd(questID)` - 완료 토글
  - `SyncCmd()` - 지연 평가 실행
  - 태그: `[TUI]`, `[DB]`

### Phase 3.3: View 렌더링

- [ ] **T029** [P] 마스터 패널: `internal/tui/views/master.go`
  - 좌측 퀘스트 목록 렌더링
  - 타입별 정렬 (Daily → Weekly → Epic → Guild)
  - 선택된 항목 하이라이트
  - 퀘스트 타입 아이콘 표시
  - 태그: `[TUI]`

- [ ] **T030** [P] 디테일 패널: `internal/tui/views/detail.go`
  - 우측 상세 정보 렌더링
  - Sub 퀘스트 목록 표시
  - 진행률 바 (Progress Bar)
  - 마감일 표시
  - 태그: `[TUI]`

- [ ] **T031** 헤더/푸터: `internal/tui/views/status.go`
  - 상단: 플레이어 정보 (레벨, XP, Flow 상태)
  - 하단: 키보드 단축키 안내
  - 태그: `[TUI]`, `[RPG]`

- [ ] **T032** 메인 View: `internal/tui/view.go`
  - 마스터-디테일 레이아웃 조합
  - `lipgloss.JoinHorizontal()` 사용
  - 터미널 크기에 따른 조정
  - 태그: `[TUI]`

### Phase 3.4: TUI 통합

- [ ] **T033** `ql check` 명령어: `internal/cli/check.go` (신규)
  - Cobra 명령어 정의
  - Bubble Tea 프로그램 실행
  - 초기 Lazy Evaluation 트리거
  - 태그: `[CLI]`, `[TUI]`, `[LAZY]`

- [ ] **T034** TUI 테스트: `internal/tui/*_test.go`
  - Update 함수 순수성 테스트
  - 상태 전이 테스트
  - 태그: `[TUI]`, `[QUALITY]`

**체크포인트**: Task 9, 10 완료 - TUI 구현 완료

---

## Phase 4: 통합 및 검증 (Integration) - Cross-Task

**목적**: 전체 시스템 통합 및 End-to-End 테스트

### Phase 4.1: 통합 테스트

- [ ] **T035** End-to-End 테스트: `tests/e2e/tui_test.go`
  - `ql check` 실행 및 조작 시뮬레이션
  - 퀘스트 생성 → 완료 → Flow 정산 흐름
  - 태그: `[TUI]`, `[QUALITY]`

- [ ] **T036** 마이그레이션 테스트
  - v1 데이터로부터 마이그레이션 검증
  - 하위 호환성 확인
  - 태그: `[DB]`, `[QUALITY]`

### Phase 4.2: 수동 검증

- [ ] **T037** 5대 방어 로직 검증
  - Divide by Zero 시 SMOOTH 유지 확인
  - 중복 평가 방지 확인
  - 무한 과거 방지 확인 (최대 일수 제한)
  - 04:00:00 정각 처리 확인
  - PENDING 상태 보존 확인
  - 태그: `[LAZY]`, `[QUALITY]`

- [ ] **T038** Flow 배율 적용 검증
  - BURNING (1.5x): 50 XP → 75 XP
  - SMOOTH (1.0x): 50 XP → 50 XP
  - HAZY (0.5x): 50 XP → 25 XP
  - 태그: `[FLOW]`, `[RPG]`, `[QUALITY]`

- [ ] **T039** 퀘스트 계층 검증
  - Sub 생성 시 2-Depth 제한 확인
  - Sub 완료 시 부모 PENDING 전환 확인
  - 부모 완료 시 XP 지급 확인
  - 태그: `[TYPE]`, `[QUALITY]`

### Phase 4.3: 문서화 및 완료

- [ ] **T040** README 업데이트
  - MVP2 기능 설명 추가
  - TUI 사용법 문서화
  - Flow 시스템 설명
  - 태그: `[AI]`

- [ ] **T041** CHANGELOG 업데이트
  - v2.0.0 변경사항 정리
  - Breaking changes 명시
  - 태그: `[AI]`

**최종 체크포인트**: MVP 2 완료

---

## 의존성 및 실행 순서

### Phase 의존성

```
Phase 1 (DB 마이그레이션)
    │
    ▼
Phase 2 (코어 엔진) ───────┐
    │                      │
    ▼                      ▼
Phase 3 (TUI) ◄────────── (Flow/계층 로직)
    │
    ▼
Phase 4 (통합)
```

### 브랜치 병합 순서

1. `feature/mvp2-db-migration` → develop
2. `feature/mvp2-lazy-eval` → develop (DB 의존)
3. `feature/mvp2-quest-hierarchy` → develop (DB 의존)
4. `feature/mvp2-tui-layout` → develop (Flow/계층 의존)
5. `feature/mvp2-tui-interaction` → develop (Layout 의존)

### 병렬 실행 가능 작업

- T001 ~ T004: 병렬 (서로 다른 파일)
- T010 ~ T014: 병렬 (엔진 로직)
- T022 ~ T024: 병렬 (TUI 초기 구조)
- T029 ~ T031: 병렬 (View 컴포넌트)

---

## 원칙 준수 태그 레퍼런스

| 태그 | 원칙 | 설명 |
|------|------|------|
| `[ARCH]` | I | 계층형 아키텍처 준수 |
| `[CLI]` | II | CLI 명령어 설계 |
| `[DB]` | III | 데이터 지속성 |
| `[RPG]` | IV | RPG 게이미피케이션 |
| `[QUALITY]` | V | 코드 품질 및 테스트 |
| `[AI]` | VI | AI 페어 프로그래밍 |
| `[TUI]` | VII | TUI 아키텍처 |
| `[TYPE]` | VIII | 퀘스트 타입 시스템 |
| `[LAZY]` | IX | 지연 평가 |
| `[FLOW]` | X | Flow 시스템 |

---

## 메모

- **MVP2 범위 가드**: 3-Depth 이상 중첩 금지, 백그라운드 데몬 없음, 웹 동기화 없음
- **TUI 최소 크기**: 80x24 권장
- **Flow 정산 시점**: 새벽 4시 기준, 앱 실행 시 Lazy Evaluation
- **XP 기본값**: 50 XP, Flow 배율 적용 (0.5x ~ 1.5x)
