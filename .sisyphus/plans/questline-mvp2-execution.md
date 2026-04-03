# Questline MVP2 실행 계획

## TL;DR
> **Summary**: `specs/mvp2/`는 방향은 맞지만 현재 저장소 구조와 그대로 맞물리지는 않습니다. 이번 계획은 `specs/mvp2/*`와 `.opencode/document/requirement_2.md`를 기준 문서로 고정하고, MVP1 코드베이스 위에 **마이그레이션 가능한 코어 계약 → 공유 서비스 계층 → Bubble Tea TUI** 순서로 증분 구현하도록 재구성합니다.
> **Deliverables**:
> - MVP1 SQLite 데이터를 안전하게 올릴 수 있는 MVP2 스키마/마이그레이션 경로
> - Daily/Weekly/Epic/Guild/Sub(2-Depth), Lazy Evaluation, Flow XP 배율, 부모 PENDING 게이트를 포함한 공유 도메인/서비스 계층
> - `ql add|done|ls|me|check`가 동일한 startup sync 경로를 공유하는 CLI/TUI 하이브리드 진입 구조
> - Bubble Tea + Lip Gloss 기반의 좌우 분할 master-detail TUI와 자동 검증 가능한 테스트/스모크 체계
> - `develop`에서 분기한 전용 작업 브랜치 `feature/mvp2-master` 기반의 일관된 구현 흐름
> **Effort**: XL
> **Parallel**: YES - 4 waves
> **Critical Path**: 1 → 2 → 3 → 4 → 5 → 7 → 8 → 10 → F1-F4

## Context
### Original Request
- `specs/mvp2` 경로의 기본 설계 문서를 참고해 Questline MVP2용 실행 플랜을 구성한다.
- 플랜 문서는 한국어로 작성한다.
- MVP2 핵심 범위는 Bubble Tea 기반 `ql check`, 퀘스트 타입 확장, 04:00 Lazy Evaluation, Flow 배율, 부모 PENDING 상태 전환이다.

### Interview Summary
- 기준 문서는 `specs/mvp2/plan.md`, `specs/mvp2/data-model.md`, `specs/mvp2/contracts/services.go`, `specs/mvp2/contracts/tui.go`, `specs/mvp2/tasks.md`, `specs/mvp2/quickstart.md`, `.opencode/document/requirement_2.md`로 고정한다.
- 현재 저장소는 `cmd/ql/main.go`, `internal/cli/*.go`, `internal/domain/*.go`, `internal/repository/*.go`, `internal/engine/leveling.go` 중심의 MVP1 구조이며, CLI가 repository를 직접 호출한다.
- 테스트 인프라는 Go unit test 중심이며 `internal/cli/*_test.go`, `internal/engine/leveling_test.go`, `internal/repository/sqlite_test.go`가 기준 패턴이다. E2E/CI는 아직 없다.
- Bubble Tea 구현은 엄격한 Model/Update/View, `WindowSizeMsg` 대응, `tea.Cmd` 기반 비동기, 상태 전이 테스트를 따른다.
- SQLite 확장은 단순 `ALTER TABLE`이 아니라 schema version + 트랜잭션 + 필요 시 테이블 재생성 패턴으로 처리해야 한다.

### Metis Review (gaps addressed)
- TUI보다 코어 계약 안정화가 우선이라는 점을 반영해 **결정 로그/마이그레이션 맵/공유 bootstrap 경로**를 먼저 고정한다.
- 모호한 규칙은 다음 기본값으로 고정한다: Daily가 0개인 날은 `SMOOTH`, Daily가 존재하지만 0% 달성이면 `HAZY`, 내부 timestamp는 UTC 저장 + 로컬 timezone 04:00 컷오프 계산, MVP1 `TODO→pending`, `DONE→completed`, `DROPPED→archived` 마이그레이션.
- Daily/Weekly는 매일 새 레코드를 무한 복제하지 않고 **단일 활성 정의 + 이력 테이블 누적** 방식으로 고정한다.
- `player` 스키마는 MVP1의 `current_xp`, `total_xp_earned`, `quests_completed`를 유지한 채 `flow_status`, `last_synced_at`, `last_evaluated`, `streak_days`를 추가하는 방식으로 고정한다.
- `specs/mvp2/contracts/*.go`는 **reference-only 산출물**로 취급하고, 첫 전역 `go test ./...` 이전에 build tag 추가 또는 import path 정렬로 모듈 빌드를 깨지 않게 정리한다.
- `ql --migrate`, `ql sync --force`, `me --verbose` 같은 quickstart 상 문구는 공식 MVP2 CLI 범위가 아니므로 구현 계획에서 제외하고, 자동 bootstrap으로 대체한다.

## Work Objectives
### Core Objective
- 기존 MVP1 CLI 코드를 안전하게 확장하여, 데이터 유실 없이 MVP2의 시간 기반 규칙과 계층형 퀘스트 모델을 도입하고, 동일한 비즈니스 규칙을 CLI와 TUI 양쪽에서 공유하도록 만드는 decision-complete 실행 순서를 제공한다.

### Deliverables
- `internal/repository/sqlite.go` 기반 schema version/migration 러너와 MVP1→MVP2 업그레이드 경로
- `internal/domain`, `internal/engine`, `internal/service`에 걸친 MVP2 규칙 집합 (퀘스트 타입/상태, Flow, Lazy Evaluation, PENDING 게이트)
- `internal/cli/add.go`, `done.go`, `ls.go`, `me.go`, 신규 `check.go`가 공통 bootstrap/service를 호출하는 구조
- `internal/tui/` 이하 Bubble Tea TUI (model/update/view/theme/commands)
- repo-패턴을 따르는 migration, engine, service, CLI, TUI 테스트와 fixture 기반 스모크 검증
- `develop`에서 생성한 `feature/mvp2-master` 브랜치 위에서만 진행되는 MVP2 구현 이력

### Definition of Done (verifiable conditions with commands)
- `go test ./... -count=1`가 성공한다.
- `go vet ./...`가 성공한다.
- `mkdir -p ./.tmp && go build -o ./.tmp/ql ./cmd/ql`가 성공한다.
- `HOME="$(mktemp -d)" ./.tmp/ql add "아침 스트레칭" -t daily`가 성공한다.
- `HOME="$(mktemp -d)" ./.tmp/ql add "웹툰 연재" -t epic` 및 이어지는 `sub` 생성 테스트가 성공한다.
- `HOME="$(mktemp -d)" ./.tmp/ql me --flow`가 Flow 상태를 포함해 성공한다.
- `go test ./internal/tui -count=1`가 master-detail navigation, resize, toggle 상태 전이를 검증하며 성공한다.

### Must Have
- `specs/mvp2/plan.md:28-35`의 헌법 체크리스트와 `data-model.md:163-234`의 스키마 방향을 구현 경로에 반영
- 2-Depth 퀘스트 계층, 부모 `pending_completion` 게이트, 04:00 Lazy Evaluation, Flow 배율 규칙
- 모든 진입점(`add`, `done`, `ls`, `me`, `check`)이 **같은 startup sync/bootstrap 서비스**를 통과
- MVP1 데이터 업그레이드 fixture 테스트와 반복 실행 안전성(idempotency) 검증
- TUI는 키보드 전용 master-detail, 좌측 고정 정렬(Daily → Weekly → Epic → Guild), 우측 detail/subquest 패널 구현
- 모든 태스크에 자동 QA 시나리오와 증적 경로 포함
- 구현 시작 전에 `develop`에서 `feature/mvp2-master` 브랜치를 생성하고, 플랜의 모든 작업은 해당 브랜치에서만 수행

### Must NOT Have (guardrails, AI slop patterns, scope boundaries)
- 3-Depth 이상 중첩, 클라우드 동기화, 백그라운드 데몬, 마우스 지원, 검색/필터 UI, 테마 커스터마이징
- Bubble Tea 도입 전 `service/` 외의 과도한 아키텍처 리라이트 (`pkg/`, generalized platform abstraction 등)
- 외부 migration framework, GitHub Actions, release automation, Homebrew 배포 작업
- quickstart에만 있고 기준 요구사항에 없는 `sync --force`, `--migrate`, `me --verbose`를 MVP2 공식 CLI로 추가
- legacy `DROPPED` 상태를 신규 UX에 노출하거나, 자식 상태 해석을 추가 범위로 확대
- `develop` 또는 다른 기능 브랜치에서 직접 구현을 진행하지 않는다.

## Verification Strategy
> ZERO HUMAN INTERVENTION — all verification is agent-executed.
- Test decision: **TDD**. repository/engine/service/TUI 상태 전이는 failing test → 구현 → green test 순서로 진행한다. 단, 커밋은 green 상태에서만 한다.
- QA policy: 각 태스크는 temp DB/temp HOME/모킹된 Bubble Tea Msg 기반 happy/failure 시나리오를 가진다.
- Evidence: `.sisyphus/evidence/task-{N}-{slug}.{ext}`
- Static verification: 각 wave 종료 시 `go test ./... -count=1`, `go vet ./...`, `go build -o ./.tmp/ql ./cmd/ql`
- Static verification: 각 wave 종료 시 `go test ./... -count=1`, `go vet ./...`, `mkdir -p ./.tmp && go build -o ./.tmp/ql ./cmd/ql`
- Migration verification: fixture로 만든 MVP1 DB 파일을 업그레이드 후 row count, status mapping, schema version, singleton player 상태를 검증
- TUI verification: 실제 수동 관찰 대신 `go test ./internal/tui -count=1`로 navigation, resize, toggle, focus switching, small-terminal fallback을 검증

## Execution Strategy
### Parallel Execution Waves
> Target: 5-8 tasks per wave. <3 per wave (except final) = under-splitting.
> Extract shared dependencies as Wave-1 tasks for max parallelism.

Wave 1: 마이그레이션 계약/도메인 규칙/공유 bootstrap 기반
Wave 2: repository 확장, lazy evaluation + flow 엔진, hierarchy completion 유스케이스
Wave 3: CLI 재배선, TUI model/layout/theme
Wave 4: TUI update/commands/`ql check`, 문서/acceptance 고정

### Branch Policy
- 시작 브랜치: `develop`
- 작업 브랜치: `feature/mvp2-master`
- 시작 절차:
  1. `git checkout develop`
  2. `git pull --ff-only`
  3. `git checkout -b feature/mvp2-master`
- 본 플랜의 모든 구현/테스트/커밋은 `feature/mvp2-master`에서 수행한다.
- `specs/mvp2/plan.md`에 적힌 `feature/mvp2-db-migration`, `feature/mvp2-lazy-eval` 등은 **논리적 작업 슬라이스 참고값**으로만 사용하고, 별도 브랜치로 실제 분기하지 않는다.
- 사용자가 별도로 요청하지 않는 한 MVP2 작업 중간에 추가 feature branch를 만들지 않는다.

### Dependency Matrix (full, all tasks)
| Task | Depends On | Unlocks |
|------|------------|---------|
| 1. 마이그레이션 하니스와 상태 매핑 | - | 2, 3, 4, 5, 6, 7, 8, 9, 10 |
| 2. 도메인 enum/시계/Flow 계약 | 1 | 3, 4, 5, 6, 7, 8, 9 |
| 3. 공유 bootstrap/service seam | 1, 2 | 4, 5, 6, 7, 8, 9 |
| 4. v2 repository 재구성 | 1, 2, 3 | 5, 6, 7, 8, 9, 10 |
| 5. Lazy Evaluation + Flow 엔진 | 1, 2, 3, 4 | 6, 7, 8, 9, 10 |
| 6. 계층 구조/PENDING 완료 유스케이스 | 1, 2, 3, 4, 5 | 7, 8, 9, 10 |
| 7. CLI 명령 재배선과 플래그 확장 | 3, 4, 5, 6 | 8, 9, 10 |
| 8. TUI model/layout/theme | 2, 3, 4, 5, 6, 7 | 9, 10 |
| 9. TUI update/commands/`ql check` | 5, 6, 7, 8 | 10 |
| 10. 문서/fixture acceptance 정렬 | 4, 5, 6, 7, 8, 9 | F1-F4 |

### Agent Dispatch Summary (wave → task count → categories)
| Wave | Task Count | Recommended Categories |
|------|------------|------------------------|
| 1 | 3 | `unspecified-high`, `deep` |
| 2 | 3 | `unspecified-high`, `quick` |
| 3 | 2 | `unspecified-high`, `visual-engineering` |
| 4 | 2 | `visual-engineering`, `writing` |
| Final | 4 | `oracle`, `unspecified-high`, `deep` |

## TODOs
> Implementation + Test = ONE task. Never separate.
> EVERY task MUST have: Agent Profile + Parallelization + QA Scenarios.

- [ ] 1. SQLite schema versioning과 MVP1→MVP2 마이그레이션 하니스 구축

  **What to do**: 먼저 `develop`에서 `feature/mvp2-master` 브랜치를 생성하고 해당 브랜치로 전환한 뒤, `internal/repository/sqlite.go`의 `initSchema()` 부트스트랩을 schema version 기반 마이그레이션 러너로 바꾼다. `PRAGMA user_version`을 사용해 v1과 v2를 구분하고, v1 `quests/player` 테이블을 v2 스키마로 올리는 트랜잭션 경로를 만든다. 단순 컬럼 추가로 끝나지 않는 상태 enum/제약 변경은 SQLite 테이블 재생성 패턴(신규 테이블 생성 → 데이터 복사 → 기존 테이블 교체)으로 처리한다. MVP1 fixture DB를 만들어 `TODO→pending`, `DONE→completed`, `DROPPED→archived`, `current_xp`/`total_xp_earned`/`quests_completed` 보존, player singleton 유지, `flow_status`/`last_synced_at`/`last_evaluated`/`streak_days` 확장, `quest_history`/`daily_evaluation` 신규 생성까지 검증한다.
  또한 `specs/mvp2/contracts/services.go`, `specs/mvp2/contracts/tui.go`가 전역 빌드를 깨지 않도록 `//go:build ignore` 같은 reference-only build tag를 추가하거나, 실제 모듈 경로(`github.com/H-BlackGom/questline/...`)와 의존성 정합성을 맞추는 방식 중 하나를 선택해 **첫 전역 `go test ./...` 전에** 정리한다.
  **Must NOT do**: 외부 migration framework를 도입하지 않는다. 프로덕션 DB를 파괴하는 hard reset이나 `DROP TABLE quests` 직행 경로를 허용하지 않는다. 마이그레이션 실패 시 부분 커밋을 남기지 않는다.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` — Reason: SQLite 제약과 legacy 데이터 보존을 동시에 다뤄야 하는 고위험 저장소 작업이다.
  - Skills: `[]` — repo 내 패턴과 표준 `database/sql`만으로 처리 가능하다.
  - Omitted: [`writing`] — 핵심은 문서 정리가 아니라 트랜잭션 마이그레이션과 fixture 검증이다.

  **Parallelization**: Can Parallel: YES | Wave 1 | Blocks: 2, 3, 4, 5, 6, 7, 8, 9, 10 | Blocked By: -

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `internal/repository/sqlite.go:17-35` — 현재 저장소 생성과 schema bootstrap 진입점이다.
  - Pattern: `internal/repository/sqlite.go:43-85` — 현재 `CREATE TABLE IF NOT EXISTS` 기반의 단순 schema 초기화다.
  - Pattern: `internal/repository/sqlite_test.go:9-61` — temp-file DB 기반 bootstrap 테스트 패턴이다.
  - Pattern: `internal/cli/root.go:21-39` — 실제 DB 경로와 data dir 생성 규약이다.
  - Pattern: `specs/mvp2/data-model.md:163-234` — 목표 v2 스키마와 인덱스 정의다.
  - Pattern: `.opencode/document/requirement_2.md:117-119` — 최소 마이그레이션 요구 컬럼이다.
  - External: `https://sqlite.org/pragma.html#pragma_user_version` — schema version 추적 기준이다.
  - External: `https://sqlite.org/lang_altertable.html#otheralter` — SQLite 비단순 schema 변경의 안전한 재생성 절차다.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `git branch --show-current` 결과가 `feature/mvp2-master`다.
  - [ ] `go test ./internal/repository -count=1 -run 'TestBootstrap|TestMigrateV1ToV2|TestMigrationRollback'`가 성공한다.
  - [ ] `go test ./... -run '^$'`가 spec contract 파일 때문에 실패하지 않는다.
  - [ ] MVP1 fixture DB를 업그레이드한 뒤 `quests` row count와 `player(id=1)`가 보존되고, `user_version`이 v2로 갱신된다.
  - [ ] 업그레이드 후 `quests` 테이블은 `type`, `parent_id`, `scheduled_date`, `deleted_at`을 가지고, `player` 테이블은 기존 `current_xp`, `total_xp_earned`, `quests_completed`를 유지한 채 `flow_status`, `last_synced_at`, `last_evaluated`, `streak_days`를 가진다.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```text
  Scenario: MVP1 fixture DB upgrades cleanly
    Tool: Bash
    Steps: `git branch --show-current && go test ./internal/repository -count=1 -run TestMigrateV1ToV2`
    Expected: fixture DB의 legacy status가 `pending/completed/archived`로 정확히 매핑되고, schema version이 증가하며 row count가 유지된다.
    Evidence: .sisyphus/evidence/task-1-migrate-v1-v2.txt

  Scenario: Failed migration rolls back atomically
    Tool: Bash
    Steps: `go test ./internal/repository -count=1 -run TestMigrationRollback`
    Expected: 오류 주입 시 마이그레이션이 부분 반영되지 않고, 기존 v1 schema와 데이터가 그대로 남는다.
    Evidence: .sisyphus/evidence/task-1-migrate-rollback.txt

  Scenario: reference-only spec contracts do not break global build
    Tool: Bash
    Steps: `go test ./... -run '^$'`
    Expected: `specs/mvp2/contracts/*.go`가 빌드 대상에 남아 있더라도 build tag 또는 import 정합성 조치로 인해 전역 테스트 탐색이 실패하지 않는다.
    Evidence: .sisyphus/evidence/task-1-spec-contract-build.txt
  ```

  **Commit**: YES | Message: `feat(repository): add mvp2 schema migration runner` | Files: `internal/repository/sqlite.go`, `internal/repository/sqlite_test.go`, `internal/repository/migrations_test.go` 또는 동등 fixture 테스트 파일

- [ ] 2. MVP2 도메인 enum, 로컬 04:00 clock, Flow 규칙을 canonical contract로 고정

  **What to do**: `internal/domain/quest.go`와 `internal/domain/player.go`를 MVP2 enum 중심으로 재구성하고, 필요한 경우 `internal/domain/status.go`, `quest_type.go`, `flow.go`로 분리한다. 퀘스트 상태는 `pending`, `in_progress`, `pending_completion`, `completed`, `archived`로 고정하고, 퀘스트 타입은 `daily`, `weekly`, `epic`, `guild`, `sub`로 고정한다. Player는 MVP1의 `CurrentXP`, `TotalXPEarned`, `QuestsCompleted`, `Level`을 유지한 채 `FlowStatus`, `LastSyncedAt`, `LastEvaluated`, `StreakDays`를 확장한다. `internal/engine/clock.go` 또는 동등 파일에서 로컬 timezone 04:00 컷오프 계산을 캡슐화하고, `internal/engine/flow.go`에서 `BURNING(>=80%)`, `SMOOTH(50~79%)`, `HAZY(<50%)` 및 "Daily가 0개면 SMOOTH" 기본값을 구현한다. DST/경계 시각 테스트가 가능하도록 clock/location 주입 구조를 만든다.
  **Must NOT do**: `time.Now().Add(-4 * time.Hour)`를 코드 전역에 복붙하지 않는다. 기존 `TODO/DONE/DROPPED`와 신규 status enum을 혼용하지 않는다. 0개 Daily와 0% 달성을 같은 케이스로 취급하지 않는다.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` — Reason: 이후 repository/service/TUI가 모두 의존하는 canonical business contract를 고정해야 한다.
  - Skills: `[]` — 순수 도메인/엔진 코드와 table-driven test가 중심이다.
  - Omitted: [`quick`] — 규칙 결정이 많아 단순 수정 범위가 아니다.

  **Parallelization**: Can Parallel: YES | Wave 1 | Blocks: 3, 4, 5, 6, 7, 8, 9 | Blocked By: 1

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `internal/domain/quest.go:5-30` — 현재 MVP1 status/quest 구조이며, 여기서 enum 체계가 확장된다.
  - Pattern: `internal/engine/leveling.go:14-72` — pure engine style과 XP/title 규칙 기준이다.
  - Pattern: `internal/engine/leveling_test.go:8-142` — table-driven engine test 패턴이다.
  - Pattern: `specs/mvp2/data-model.md:63-156` — QuestStatus, QuestType, FlowStatus canonical 정의다.
  - Pattern: `specs/mvp2/data-model.md:237-299` — TUI가 기대하는 Focus/QuestNode 정렬 규칙이다.
  - Pattern: `.opencode/document/requirement_2.md:92-113` — 04:00 Lazy Evaluation, 5대 방어 로직, Flow 배율, 부모 PENDING 요구다.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `go test ./internal/domain ./internal/engine -count=1`가 성공한다.
  - [ ] `go test ./internal/engine -count=1 -run 'TestLogicalDate|TestFlowGrade|TestFourAMBoundary|TestNoDailyDefaultsToSmooth'`가 성공한다.
  - [ ] `go test ./internal/domain -count=1 -run 'TestQuestTypeValidation|TestStatusTransitions'`가 성공한다.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```text
  Scenario: 04:00 컷오프와 Flow 기본값이 정확히 계산된다
    Tool: Bash
    Steps: `go test ./internal/engine -count=1 -run 'TestFourAMBoundary|TestNoDailyDefaultsToSmooth|TestZeroPercentDailyBecomesHazy'`
    Expected: Daily가 0개인 날은 `SMOOTH`, Daily가 있지만 완료율 0%인 날은 `HAZY`, 04:00:00 정각은 당일로 계산된다.
    Evidence: .sisyphus/evidence/task-2-clock-flow.txt

  Scenario: 잘못된 타입/상태 조합은 생성 시 거부된다
    Tool: Bash
    Steps: `go test ./internal/domain -count=1 -run 'TestQuestTypeValidation|TestInvalidParentRules'`
    Expected: `sub` without parent, `daily` with parent, 3-depth 유도 조합이 모두 명시적 에러로 실패한다.
    Evidence: .sisyphus/evidence/task-2-domain-validation.txt
  ```

  **Commit**: YES | Message: `feat(domain): define mvp2 quest and flow contracts` | Files: `internal/domain/*.go`, `internal/engine/clock.go`, `internal/engine/flow.go`, 관련 테스트 파일

- [ ] 3. 모든 진입점이 공유하는 bootstrap/service seam 도입

  **What to do**: `internal/service/`를 MVP2 범위에 한해 도입하고, 최소한 `bootstrap.go`, `quest_service.go`, `sync_service.go`, `player_service.go`를 만든다. `bootstrap`은 DB open → migration ensure → lazy sync 실행 → service 묶음 반환의 단일 경로가 되어야 하며, `ql add`, `done`, `ls`, `me`, `check` 모두 이 경로를 사용하도록 강제한다. 기존 repository가 갖고 있던 퀘스트 완료+XP 지급+status 변경 책임은 service 계층으로 이동시키고, repository는 읽기/쓰기/트랜잭션 primitive만 남긴다.
  **Must NOT do**: CLI와 TUI가 각각 별도 sync 로직을 갖게 하지 않는다. repository에서 business rule 분기를 계속 유지하지 않는다. MVP2 밖 명령어/모듈로 service 레이어 확장을 일반화하지 않는다.

  **Recommended Agent Profile**:
  - Category: `deep` — Reason: 현재 direct CLI→repo 구조를 최소 침습으로 shared service seam으로 바꿔야 한다.
  - Skills: `[]` — 기존 패턴 이해와 구조 재편이 핵심이다.
  - Omitted: [`visual-engineering`] — 이 태스크는 UI가 아닌 orchestration 구조 작업이다.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 4, 5, 6, 7, 8, 9 | Blocked By: 1, 2

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `internal/cli/root.go:10-39` — 현재 CLI 진입과 DB path 규약이다.
  - Pattern: `internal/cli/done.go:30-83` — 현재 command가 repository를 직접 열고 player/XP를 직접 엮는 구조다.
  - Pattern: `internal/cli/me.go:22-57` — read-only 명령도 bootstrap 없이 repository를 직접 호출하는 현재 패턴이다.
  - Pattern: `internal/repository/player_repo.go:58-120` — 현재 repository 안에 quest completion + XP 지급이 결합되어 있는 지점이다.
  - Pattern: `specs/mvp2/contracts/services.go:13-191` — QuestService, SyncService, PlayerService 목표 계약이다.
  - Pattern: `specs/mvp2/plan.md:84-90` — MVP2가 의도한 `internal/service/` 구조다.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `go test ./internal/service -count=1`가 성공한다.
  - [ ] `go test ./internal/cli -count=1 -run 'TestCommandsUseSharedBootstrap'`가 성공한다.
  - [ ] `rg -n "repository\.New\(|NewPlayerRepository\(|NewQuestRepository\(" internal/cli` 결과에서 허용된 bootstrap 내부 외에는 direct repository construction이 남아 있지 않다.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```text
  Scenario: 모든 CLI 진입점이 동일한 bootstrap을 사용한다
    Tool: Bash
    Steps: `go test ./internal/cli -count=1 -run TestCommandsUseSharedBootstrap`
    Expected: `add`, `done`, `ls`, `me`, `check`가 동일한 service/bootstrap seam을 호출하고 lazy sync가 중복 실행되지 않는다.
    Evidence: .sisyphus/evidence/task-3-shared-bootstrap.txt

  Scenario: repository direct access가 CLI에서 제거된다
    Tool: Bash
    Steps: `rg -n "repository\.New\(|NewPlayerRepository\(|NewQuestRepository\(" internal/cli`
    Expected: 결과가 bootstrap/service 초기화 파일 외에는 비어 있거나 허용된 위치로만 제한된다.
    Evidence: .sisyphus/evidence/task-3-no-direct-repo.txt
  ```

  **Commit**: YES | Message: `feat(service): add shared bootstrap and orchestration seam` | Files: `internal/service/*.go`, `internal/cli/*.go` 중 bootstrap 배선 파일, 관련 테스트

- [ ] 4. v2 repository 재구성과 계층/이력 조회 API 추가

  **What to do**: `internal/repository/quest_repo.go`, `player_repo.go`를 v2 스키마에 맞게 재작성하고, 필요 시 `history_repo.go`를 추가한다. quest repository는 root quest 조회, parent/child 조회, type/status filter, soft delete 제외 기본 조회, progress 계산에 필요한 집계 query를 제공한다. player repository는 flow/status/sync timestamp/streak를 다루고, quest history/daily evaluation 저장소는 일일 이력과 정산 결과를 저장한다. 모든 write는 service가 트랜잭션을 관리할 수 있도록 low-level 메서드와 tx helper를 노출한다.
  **Must NOT do**: repository 내부에서 Flow grade 계산, parent status 전환, XP 지급 공식, UI 정렬 로직을 구현하지 않는다. soft delete restore를 MVP2 필수 기능처럼 확장하지 않는다.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` — Reason: 스키마 변경과 query 확장이 크며 이후 모든 유스케이스의 기반이 된다.
  - Skills: `[]` — `database/sql`과 현재 repo 패턴 확장으로 충분하다.
  - Omitted: [`quick`] — 파일 수와 제약이 많아 단순 패치가 아니다.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 5, 6, 7, 8, 9, 10 | Blocked By: 1, 2, 3

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `internal/repository/quest_repo.go:11-136` — 현재 quest CRUD/list repository shape다.
  - Pattern: `internal/repository/player_repo.go:10-120` — 현재 player singleton repository와 transaction example이다.
  - Pattern: `internal/repository/sqlite.go:43-85` — schema와 singleton player 보장 위치다.
  - Pattern: `specs/mvp2/data-model.md:163-234` — repository가 만족해야 하는 v2 table/index 계약이다.
  - Pattern: `specs/mvp2/contracts/services.go:48-143` — QuestNode, QuestFilter, SyncResult, FlowAdjustment 구조 요구다.
  - Pattern: `specs/mvp2/plan.md:87-90` — repository 하위에서 기대하는 파일 구조다.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `go test ./internal/repository -count=1 -run 'TestQuestTreeQueries|TestPlayerFlowPersistence|TestDailyEvaluationPersistence'`가 성공한다.
  - [ ] root quest 조회 시 soft-deleted row가 기본 결과에서 제외된다.
  - [ ] `GetQuestTree()` 또는 동등 repository API가 Daily → Weekly → Epic → Guild root ordering에 필요한 데이터를 안정적으로 반환한다.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```text
  Scenario: 계층/이력 조회가 기대한 shape를 반환한다
    Tool: Bash
    Steps: `go test ./internal/repository -count=1 -run 'TestQuestTreeQueries|TestDailyEvaluationPersistence'`
    Expected: root quest, sub quest, flow history, daily evaluation row가 fixture 기준으로 정확히 조회된다.
    Evidence: .sisyphus/evidence/task-4-repository-tree.txt

  Scenario: soft delete와 legacy null 값이 조회를 오염시키지 않는다
    Tool: Bash
    Steps: `go test ./internal/repository -count=1 -run 'TestSoftDeletedQuestExcluded|TestLegacyNullMigrationValues'`
    Expected: soft-deleted row는 기본 list/tree에서 보이지 않고, legacy null 컬럼은 안전한 기본값으로 읽힌다.
    Evidence: .sisyphus/evidence/task-4-repository-softdelete.txt
  ```

  **Commit**: YES | Message: `feat(repository): add mvp2 quest tree and history adapters` | Files: `internal/repository/quest_repo.go`, `internal/repository/player_repo.go`, `internal/repository/history_repo.go`, 관련 테스트

- [ ] 5. Lazy Evaluation + Flow 엔진을 순수 규칙과 idempotent service로 구현

  **What to do**: `internal/engine/`에 `sync.go`, `flow.go`, 필요 시 `weekly.go`를 추가하고, `internal/service/sync_service.go`가 이를 orchestration하도록 구현한다. 규칙은 다음으로 고정한다: (1) 논리적 일자는 로컬 timezone 04:00 컷오프, (2) Daily가 0개면 `SMOOTH`, (3) 이미 정산된 날짜는 재정산하지 않음, (4) 장기 미접속 시 패널티는 마지막 미정산 1일치만 적용, (5) Weekly 이월 마감은 다음 주 일요일로 재계산, (6) `pending_completion` 부모는 expiry/archive 대상에서 제외. 모든 sync는 반복 실행해도 같은 결과를 내는 idempotent 흐름으로 만들고, bootstrap/service에서 앱 시작 시 1회만 호출되게 한다.
  **Must NOT do**: sync 로직에서 직접 CLI/TUI 출력 포맷을 알지 않는다. `time.Now()`를 테스트 불가능한 방식으로 하드코딩하지 않는다. 미접속 일수만큼 penalty를 누적 적용하지 않는다.

  **Recommended Agent Profile**:
  - Category: `deep` — Reason: 시간 경계, idempotency, weekly rollover, flow multiplier가 얽힌 핵심 비즈니스 규칙이다.
  - Skills: `[]` — 순수 로직 + service orchestration이 중심이다.
  - Omitted: [`visual-engineering`] — UI보다 규칙 엔진 정확도가 우선이다.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 6, 7, 8, 9, 10 | Blocked By: 1, 2, 3, 4

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `internal/engine/leveling.go:14-72` — pure engine 함수 스타일 기준이다.
  - Pattern: `internal/engine/leveling_test.go:8-142` — table-driven edge-case 테스트 패턴이다.
  - Pattern: `specs/mvp2/data-model.md:130-156` — Flow 정산 규칙과 Daily 0개 기본값이다.
  - Pattern: `specs/mvp2/data-model.md:334-360` — 5대 방어 로직의 검증 방향이다.
  - Pattern: `.opencode/document/requirement_2.md:92-113` — 04:00 Lazy Evaluation, 5대 방어 로직, Flow 배율, PENDING 보존 요구다.
  - Pattern: `specs/mvp2/contracts/services.go:86-143` — SyncService/SyncResult/FlowGrade 구조다.
  - External: `https://sqlite.org/lang_transaction.html` — sync write를 atomic transaction으로 묶기 위한 기준이다.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `go test ./internal/engine ./internal/service -count=1 -run 'TestEvaluateLazySync|TestWeeklyRollover|TestFlowGrade|TestPendingPreserved'`가 성공한다.
  - [ ] 동일 fixture DB에 대해 sync를 2회 연속 호출해도 두 번째 호출에서 추가 penalty, 추가 daily evaluation row, 추가 archive가 발생하지 않는다.
  - [ ] Weekly due date 재계산 테스트에서 결과가 "다음 주 일요일"로 고정된다.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```text
  Scenario: 장기 미접속과 04:00 경계가 안전하게 처리된다
    Tool: Bash
    Steps: `go test ./internal/service -count=1 -run 'TestEvaluateLazySync_LongInactivitySinglePenalty|TestEvaluateLazySync_BoundaryAtFourAM'`
    Expected: 미접속 일수와 무관하게 penalty는 1회만 반영되고, 04:00:00 정각 실행은 당일 기준으로 처리된다.
    Evidence: .sisyphus/evidence/task-5-lazy-sync.txt

  Scenario: sync는 재실행해도 동일 결과를 유지한다
    Tool: Bash
    Steps: `go test ./internal/service -count=1 -run TestEvaluateLazySync_Idempotent`
    Expected: 같은 DB에 대해 연속 실행 시 두 번째 결과의 추가 변경 수가 0이며 상태가 변하지 않는다.
    Evidence: .sisyphus/evidence/task-5-lazy-idempotent.txt
  ```

  **Commit**: YES | Message: `feat(engine): implement lazy evaluation and flow rules` | Files: `internal/engine/*.go`, `internal/service/sync_service.go`, 관련 테스트

- [ ] 6. Quest hierarchy 완료 유스케이스와 부모 PENDING 게이트 구현

  **What to do**: `internal/service/quest_service.go`와 필요한 engine helper에서 Sub 생성 제한, parent-child 조회, progress 계산, Sub 완료 시 부모 `in_progress`/`pending_completion` 전이, 부모 수동 완료 시에만 XP 지급되는 유스케이스를 구현한다. `add -p`와 `done`/TUI Space 모두 동일한 `CompleteQuest()` 유스케이스를 사용하도록 만들고, 부모가 `pending_completion`일 때 날짜가 지나도 `archived`/`expired`로 이동하지 않는 규칙을 보존한다. `DROPPED`는 legacy migration 용도만 유지하고 MVP2 UX 규칙에서는 blocking child로 취급하지 않는다.
  **Must NOT do**: parent quest를 모든 Sub 완료 즉시 `completed`로 만들지 않는다. hierarchy 로직을 CLI update handler에 중복 구현하지 않는다. 3-depth 허용, sibling auto-reorder, progress bar 비즈니스 규칙을 저장소에 숨기지 않는다.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` — Reason: 상태 전이, XP 지급 시점, 계층 제한이 동시에 걸린 핵심 유스케이스다.
  - Skills: `[]` — 현재 repo/service 패턴 위에서 구현 가능하다.
  - Omitted: [`quick`] — 규칙 충돌 가능성이 높아 세밀한 테스트가 필요하다.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 7, 8, 9, 10 | Blocked By: 1, 2, 3, 4, 5

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `internal/cli/add.go:28-86` — 현재 quest 생성 입력/검증/저장 흐름이다.
  - Pattern: `internal/cli/done.go:22-85` — 현재 완료 UX와 출력 흐름이다.
  - Pattern: `internal/repository/player_repo.go:58-120` — 현재 완료+XP 트랜잭션이 묶여 있는 위치이며 service로 이동해야 한다.
  - Pattern: `specs/mvp2/data-model.md:73-120` — parent/sub 상태 전이와 2-Depth 규칙이다.
  - Pattern: `.opencode/document/requirement_2.md:111-113` — Sub 완료 후 부모가 PENDING이 되는 핵심 요구다.
  - Pattern: `specs/mvp2/contracts/services.go:13-84` — QuestService와 CompletionResult/ParentTransition 계약이다.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `go test ./internal/service -count=1 -run 'TestCreateSubQuestValidation|TestCompleteSubQuestTransitionsParent|TestCompleteParentPendingAwardsXP'`가 성공한다.
  - [ ] 모든 Sub가 완료된 뒤 부모 상태는 즉시 `pending_completion`이 되며, XP는 부모 완료 전까지 증가하지 않는다.
  - [ ] `pending_completion` 상태의 부모는 lazy sync 후에도 그대로 유지된다.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```text
  Scenario: 모든 Sub 완료 후 부모가 XP 없이 pending_completion으로 전이된다
    Tool: Bash
    Steps: `go test ./internal/service -count=1 -run TestCompleteSubQuestTransitionsParent`
    Expected: 마지막 Sub 완료 시 부모 status는 `pending_completion`, player XP는 아직 증가하지 않는다.
    Evidence: .sisyphus/evidence/task-6-parent-pending.txt

  Scenario: 잘못된 parent/sub 조합이 명시적으로 거부된다
    Tool: Bash
    Steps: `go test ./internal/service -count=1 -run 'TestCreateSubQuestRejectsDailyParent|TestCreateSubQuestRejectsThirdDepth'`
    Expected: `daily`/`weekly`/`sub` 밑으로 Sub를 추가하려는 시도는 오류로 실패한다.
    Evidence: .sisyphus/evidence/task-6-sub-validation.txt
  ```

  **Commit**: YES | Message: `feat(service): add hierarchy completion gate` | Files: `internal/service/quest_service.go`, 관련 engine/repository 보조 파일, 테스트

- [ ] 7. CLI 명령 재배선과 MVP2 플래그/출력 계약 확장

  **What to do**: `internal/cli/add.go`, `done.go`, `ls.go`, `me.go`, `root.go`를 서비스 계층 기반으로 재배선한다. `add`는 `-t`/`--type`, `-p`/`--parent`, `-d`/`--due`를 지원하고, `sub`는 parent 필수, `daily/weekly`는 parent 금지 규칙을 따른다. `done`은 서비스 결과에 따라 `+XP`, Flow 배율, 부모 PENDING 전이 메시지를 다르게 출력한다. `ls`는 `--type` 필터를 지원하고 기본 root ordering이 Daily → Weekly → Epic → Guild가 되도록 맞춘다. `me`는 `--flow`를 지원해 현재 Flow 상태, 배율, 연속 일수, 다음 평가 시점을 출력한다.
  **Must NOT do**: quickstart에만 있던 `--migrate`, `sync --force`, `--verbose`를 이 단계에서 슬쩍 추가하지 않는다. CLI별로 독자적인 XP 계산/상태 전환 로직을 넣지 않는다. 기존 한국어 오류 메시지 톤을 깨지 않는다.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` — Reason: 여러 Cobra 명령을 한 번에 재배선하되 shared service contract를 유지해야 한다.
  - Skills: `[]` — Cobra와 현재 CLI 테스트 패턴을 그대로 확장하면 된다.
  - Omitted: [`writing`] — 핵심은 문서보다 CLI 행동 계약이다.

  **Parallelization**: Can Parallel: YES | Wave 3 | Blocks: 8, 9, 10 | Blocked By: 3, 4, 5, 6

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `internal/cli/add.go:14-87` — 현재 `ql add` Cobra 등록과 입력 검증 패턴이다.
  - Pattern: `internal/cli/done.go:11-85` — 현재 `done` 명령의 출력/에러 처리 기준이다.
  - Pattern: `internal/cli/ls.go:45-120` — flag 처리와 표 출력 패턴이다.
  - Pattern: `internal/cli/me.go:12-55` — profile box 출력 패턴과 helper 재사용 기준이다.
  - Pattern: `internal/cli/add_test.go:11-68` — temp HOME 기반 CLI black-box 테스트 패턴이다.
  - Pattern: `specs/mvp2/plan.md:28-35` — CLI surface와 MVP2 헌법 체크리스트다.
  - Pattern: `specs/mvp2/quickstart.md:44-146` — 사용자가 기대하는 CLI/TUI 명령 흐름이다.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `go test ./internal/cli -count=1`가 성공한다.
  - [ ] `mkdir -p ./.tmp && HOME="$(mktemp -d)" ./.tmp/ql add "아침 스트레칭" -t daily`가 성공한다.
  - [ ] `mkdir -p ./.tmp && HOME="$(mktemp -d)" ./.tmp/ql me --flow`가 Flow 상태/배율/연속일수 라벨을 포함해 성공한다.
  - [ ] `mkdir -p ./.tmp && HOME="$(mktemp -d)" ./.tmp/ql ls --type epic`가 성공하고 타입 필터가 적용된다.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```text
  Scenario: MVP2 CLI 플래그가 end-to-end로 동작한다
    Tool: Bash
    Steps: `TMP_HOME=$(mktemp -d) && mkdir -p ./.tmp && go build -o ./.tmp/ql ./cmd/ql && HOME="$TMP_HOME" ./.tmp/ql add "웹툰 연재" -t epic && EPIC_ID=$(HOME="$TMP_HOME" ./.tmp/ql ls --type epic | awk 'NR==3 {print $1}') && HOME="$TMP_HOME" ./.tmp/ql add "캐릭터 디자인" -t sub -p "$EPIC_ID"`
    Expected: epic 생성 후 sub 생성이 성공하고, invalid parent 없이 정상 저장된다.
    Evidence: .sisyphus/evidence/task-7-cli-flags.txt

  Scenario: 잘못된 타입/부모/플래그 조합은 명시적으로 실패한다
    Tool: Bash
    Steps: `TMP_HOME=$(mktemp -d) && mkdir -p ./.tmp && go build -o ./.tmp/ql ./cmd/ql && ! HOME="$TMP_HOME" ./.tmp/ql add "잘못된 서브" -t sub`
    Expected: parent 없이 `sub`를 만들면 비정상 종료와 명시적 오류 메시지가 출력된다.
    Evidence: .sisyphus/evidence/task-7-cli-invalid-sub.txt
  ```

  **Commit**: YES | Message: `feat(cli): wire mvp2 quest commands through services` | Files: `internal/cli/add.go`, `internal/cli/done.go`, `internal/cli/ls.go`, `internal/cli/me.go`, `internal/cli/root.go`, 관련 테스트

- [ ] 8. Bubble Tea model/layout/theme 뼈대와 two-pane render contract 구현

  **What to do**: `internal/tui/`에 `model.go`, `view.go`, `theme/styles.go`, `views/master.go`, `views/detail.go`, `views/status.go`를 만들고, TUI의 state model을 단일 `Model`로 정의한다. 좌측 패널은 root quest list, 우측 패널은 선택된 quest의 detail/sub list, 상단은 player/flow 헤더, 하단은 key hint 상태바로 고정한다. `FocusMaster`/`FocusDetail`, `MasterCursor`/`DetailCursor`, `QuestNode.Progress`, `MinWidth=80`, `MinHeight=24`, Daily → Weekly → Epic → Guild 정렬, Lip Gloss 색상 테마를 명시적으로 구현한다. `specs/mvp2/contracts/tui.go`의 `FocusFocusDetail` 오타는 구현 시 `FocusDetail`로 교정하고 문서 주석도 정렬한다.
  **Must NOT do**: 이 단계에서 key handling, DB write, async command orchestration을 넣지 않는다. mouse support, search box, modal, alternate screen 최적화 같은 범위 외 UX를 추가하지 않는다.

  **Recommended Agent Profile**:
  - Category: `visual-engineering` — Reason: Bubble Tea 레이아웃과 Lip Gloss 스타일의 UI 구조 설계가 핵심이다.
  - Skills: `[]` — 외부 UI skill 없이도 Bubble Tea/Lip Gloss 패턴을 따르는 구현이 가능하다.
  - Omitted: [`quick`] — state model과 panel contract를 세밀하게 고정해야 한다.

  **Parallelization**: Can Parallel: YES | Wave 3 | Blocks: 9, 10 | Blocked By: 2, 3, 4, 5, 6, 7

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `specs/mvp2/data-model.md:237-299` — FocusType, Model, QuestNode, root ordering 계약이다.
  - Pattern: `specs/mvp2/contracts/tui.go:11-60` — TUIModel, FocusType, NavigateMsg 계약이며 `FocusFocusDetail` 오타를 수정해야 한다.
  - Pattern: `specs/mvp2/contracts/tui.go:145-190` — TUIConfiguration, ThemeColors 기본값이다.
  - Pattern: `.opencode/document/requirement_2.md:23-39` — 목표 master-detail 화면과 key hint 표현이다.
  - External: `https://github.com/charmbracelet/bubbletea/blob/main/tea.go#L366-L368` — Model/Init/Update/View 계약이다.
  - External: `https://github.com/charmbracelet/bubbletea/blob/main/screen.go#L1-L5` — `WindowSizeMsg` 기반 resize 처리 기준이다.
  - External: `https://github.com/charmbracelet/lipgloss/blob/main/examples/color/standalone/main.go#L23-L26` — Lip Gloss 스타일 선언 패턴이다.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `go test ./internal/tui -count=1 -run 'TestInitialModel|TestRenderMasterDetail|TestSmallTerminalFallback'`가 성공한다.
  - [ ] TUI 초기 render는 80x24 이상에서 좌우 2패널을, 80x24 미만에서는 명시적 fallback 상태를 반환한다.
  - [ ] root list 정렬이 Daily → Weekly → Epic → Guild로 고정된다.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```text
  Scenario: 초기 렌더가 2패널 구조를 안정적으로 만든다
    Tool: Bash
    Steps: `go test ./internal/tui -count=1 -run 'TestInitialModel|TestRenderMasterDetail'`
    Expected: 헤더, 좌측 quest list, 우측 detail pane, 상태바가 모두 렌더 문자열에 포함된다.
    Evidence: .sisyphus/evidence/task-8-tui-layout.txt

  Scenario: 작은 터미널에서 안전하게 fallback 된다
    Tool: Bash
    Steps: `go test ./internal/tui -count=1 -run TestSmallTerminalFallback`
    Expected: width/height가 최소치 미만일 때 panic 없이 축소 레이아웃 또는 명시적 안내 문구가 렌더된다.
    Evidence: .sisyphus/evidence/task-8-tui-small-terminal.txt
  ```

  **Commit**: YES | Message: `feat(tui): add master-detail layout skeleton` | Files: `internal/tui/model.go`, `internal/tui/view.go`, `internal/tui/theme/styles.go`, `internal/tui/views/*.go`, 관련 테스트

- [ ] 9. Bubble Tea update/commands와 `ql check` 통합

  **What to do**: `internal/tui/update.go`, `commands.go`, 신규 `internal/cli/check.go`를 구현해 `ql check`가 Bubble Tea 프로그램을 실행하도록 연결한다. `Update`는 `tea.KeyMsg`와 `tea.WindowSizeMsg`만으로 `j/k`, `↑/↓`, `enter/→`, `esc/←`, `space`, `q/ctrl+c`를 처리하고, side effect는 모두 `tea.Cmd`를 통해 서비스 호출로 넘긴다. `LoadQuestsCmd`, `LoadPlayerCmd`, `PerformSyncCmd`, `ToggleQuestCmd`를 구현하고, `space`는 root quest와 sub quest 모두 동일한 `CompleteQuest()`를 호출하게 한다. 앱 시작 시 sync는 bootstrap에서 이미 1회 수행되므로 TUI 내부에서 재실행하지 않으며, resize 시 panel width와 cursor bounds를 안전하게 재계산한다.
  **Must NOT do**: `Update` 안에서 직접 DB를 열거나 repository를 호출하지 않는다. sync를 bootstrap과 TUI command 양쪽에서 중복 실행하지 않는다. Bubble Tea 내부에서 business rule 분기(Flow 계산, hierarchy 결정)를 구현하지 않는다.

  **Recommended Agent Profile**:
  - Category: `visual-engineering` — Reason: Bubble Tea event loop, navigation, async command wiring이 핵심이다.
  - Skills: `[]` — 공식 Bubble Tea 패턴만 따르면 된다.
  - Omitted: [`deep`] — 규칙 자체는 이미 서비스 계층에서 고정되므로 UI orchestration에 집중한다.

  **Parallelization**: Can Parallel: YES | Wave 4 | Blocks: 10 | Blocked By: 5, 6, 7, 8

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `specs/mvp2/contracts/tui.go:37-143` — Navigate/Toggle/Load/Sync Msg 및 Cmd shape다.
  - Pattern: `specs/mvp2/plan.md:67-75` — `internal/tui/update.go`, `view.go`, views 구조 의도다.
  - Pattern: `.opencode/document/requirement_2.md:46-86` — key 분기 update loop skeleton과 focus 전환 규칙이다.
  - External: `https://github.com/charmbracelet/bubbletea/blob/main/tea.go#L366-L368` — Update/View/Cmd 계약이다.
  - External: `https://github.com/charmbracelet/bubbletea/blob/main/examples/sequence/main.go#L12-L16` — ordered vs concurrent cmd orchestration 패턴이다.
  - External: `https://github.com/charmbracelet/bubbletea/blob/main/tea.go#L825-L831` — `WindowSizeMsg` 처리 예시다.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `go test ./internal/tui -count=1 -run 'TestUpdateNavigation|TestToggleQuestCmd|TestWindowResize|TestCheckCommandStartsProgram'`가 성공한다.
  - [ ] `mkdir -p ./.tmp && go build -o ./.tmp/ql ./cmd/ql` 후 `HOME="$(mktemp -d)" ./.tmp/ql check`가 즉시 panic 없이 시작되고 종료 가능하다.
  - [ ] `space` 동작 테스트에서 root/sub 모두 service `CompleteQuest()` 호출 경로를 사용한다.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```text
  Scenario: key navigation과 focus 전환이 deterministic 하다
    Tool: Bash
    Steps: `go test ./internal/tui -count=1 -run 'TestUpdateNavigation|TestWindowResize'`
    Expected: `j/k`는 현재 포커스된 패널의 cursor만 움직이고, `enter/→`는 subquest가 있을 때만 detail pane으로 이동하며, resize 시 cursor 범위가 보정된다.
    Evidence: .sisyphus/evidence/task-9-tui-update.txt

  Scenario: `ql check` 시작과 종료가 안전하다
    Tool: Bash
    Steps: `TMP_HOME=$(mktemp -d) && mkdir -p ./.tmp && go build -o ./.tmp/ql ./cmd/ql && script -q /dev/null env HOME="$TMP_HOME" ./.tmp/ql check <<'EOF'
q
EOF`
    Expected: 프로그램이 panic 없이 시작되고 `q` 입력 후 정상 종료한다.
    Evidence: .sisyphus/evidence/task-9-ql-check-smoke.txt
  ```

  **Commit**: YES | Message: `feat(cli): add ql check bubble tea entrypoint` | Files: `internal/tui/update.go`, `internal/tui/commands.go`, `internal/cli/check.go`, 관련 테스트

- [x] 10. 스펙/quickstart 정렬과 fixture 기반 acceptance 회귀 묶음 완성

  **What to do**: 구현이 끝나면 `specs/mvp2/quickstart.md`, 필요 시 `specs/mvp2/plan.md`, `README.md`를 실제 동작과 맞게 정렬한다. 동시에 temp HOME + fixture DB 기반 acceptance 테스트 묶음을 추가해 `add → ls → done → me --flow → check` 핵심 경로와 migration smoke를 자동 검증한다. quickstart에만 남아 있는 비범위 명령(`--migrate`, `sync --force`, `--verbose`)은 제거하거나 "미포함"으로 명시하고, 실제 지원하는 플래그/출력으로 치환한다.
  **Must NOT do**: 구현과 맞지 않는 quickstart 예시를 그대로 남기지 않는다. manual visual check만으로 완료 처리하지 않는다. README에 아직 구현되지 않은 기능을 예고 문구로 추가하지 않는다.

  **Recommended Agent Profile**:
  - Category: `writing` — Reason: 구현 결과와 문서/스모크 검증을 정렬하는 마감 태스크다.
  - Skills: `[]` — 문서와 acceptance 스크립트 정리가 중심이다.
  - Omitted: [`quick`] — 문서+검증 정렬 범위가 넓다.

  **Parallelization**: Can Parallel: YES | Wave 4 | Blocks: F1-F4 | Blocked By: 4, 5, 6, 7, 8, 9

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `specs/mvp2/quickstart.md:1-284` — 현재 예시 명령과 문서 drift 원천이다.
  - Pattern: `specs/mvp2/plan.md:236-257` — quickstart 상 노출되어야 하는 핵심 사용자 흐름이다.
  - Pattern: `README.md:63-150` — 기존 CLI 문서 톤과 표 형식이다.
  - Pattern: `internal/cli/add_test.go:11-68` — temp HOME black-box CLI 테스트 패턴이다.
  - Pattern: `internal/repository/sqlite_test.go:36-61` — temp DB bootstrap 검증 패턴이다.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `go test ./... -count=1`가 성공한다.
  - [ ] `go vet ./...`가 성공한다.
  - [ ] `mkdir -p ./.tmp && go build -o ./.tmp/ql ./cmd/ql`가 성공한다.
  - [ ] `TMP_HOME=$(mktemp -d)` 기반 스모크 스크립트로 `add`, `ls`, `done`, `me --flow`, `check` 흐름이 모두 성공한다.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```text
  Scenario: temp HOME 기반 MVP2 전체 핵심 루프가 통과한다
    Tool: Bash
    Steps: `TMP_HOME=$(mktemp -d) && mkdir -p ./.tmp && go build -o ./.tmp/ql ./cmd/ql && HOME="$TMP_HOME" ./.tmp/ql add "아침 스트레칭" -t daily && HOME="$TMP_HOME" ./.tmp/ql ls && QID=$(HOME="$TMP_HOME" ./.tmp/ql ls | awk 'NR==3 {print $1}') && HOME="$TMP_HOME" ./.tmp/ql done "$QID" && HOME="$TMP_HOME" ./.tmp/ql me --flow`
    Expected: add/list/done/me 흐름이 모두 성공하고 `me --flow` 출력에 Flow 상태 라벨이 포함된다.
    Evidence: .sisyphus/evidence/task-10-mvp2-smoke.txt

  Scenario: 문서와 실제 지원 명령이 일치한다
    Tool: Bash
    Steps: `if command -v rg >/dev/null 2>&1; then rg -n -- "--migrate|sync --force|--verbose" specs/mvp2/quickstart.md README.md; else grep -R -n -E -- "--migrate|sync --force|--verbose" specs/mvp2/quickstart.md README.md; fi && ./.tmp/ql --help && ./.tmp/ql me --help`
    Expected: 비범위 명령은 제거 또는 미포함으로 정리되고, 문서/도움말/실제 CLI가 상충하지 않는다.
    Evidence: .sisyphus/evidence/task-10-doc-sync.txt
  ```

  **Commit**: YES | Message: `docs(mvp2): align quickstart and acceptance flows` | Files: `specs/mvp2/quickstart.md`, `specs/mvp2/plan.md`, `README.md`, acceptance 테스트/스크립트 파일

## Final Verification Wave (MANDATORY — after ALL implementation tasks)
> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.
> **Do NOT auto-proceed after verification. Wait for user's explicit approval before marking work complete.**
> **Never mark F1-F4 as checked before getting user's okay.** Rejection or user feedback -> fix -> re-run -> present again -> wait for okay.
- [ ] F1. Plan Compliance Audit — oracle
- [ ] F2. Code Quality Review — unspecified-high
- [ ] F3. Real Manual QA — unspecified-high (+ playwright if UI)
- [ ] F4. Scope Fidelity Check — deep

## Commit Strategy
- Green commits only. failing-test-only 상태는 로컬 작업 중간 단계로만 허용하고 커밋하지 않는다.
- 권장 커밋 순서:
  1. `test(repository): add mvp2 migration fixtures`
  2. `feat(repository): migrate sqlite schema to v2`
  3. `test(engine): cover lazy eval flow rules`
  4. `feat(engine): implement lazy evaluation and flow`
  5. `test(service): cover hierarchy and bootstrap orchestration`
  6. `feat(service): add shared quest and sync services`
  7. `feat(cli): rewire add done ls me for mvp2`
  8. `test(tui): cover model update and resize`
  9. `feat(tui): add master-detail ql check dashboard`
  10. `docs(mvp2): align quickstart and acceptance docs`

## Success Criteria
- MVP1 fixture DB를 그대로 업그레이드해도 데이터 손실 없이 MVP2 동작이 가능하다.
- 하루가 바뀐 뒤 어떤 진입점으로 앱을 시작하든 Lazy Evaluation이 정확히 1회 적용된다.
- Daily/Weekly/Epic/Guild/Sub 규칙이 CLI와 TUI에서 동일하게 동작한다.
- 부모 퀘스트는 모든 Sub 완료 후에도 즉시 XP를 지급하지 않고 `pending_completion`에서 수동 확인을 요구한다.
- TUI는 80x24 이상에서 일관된 2패널 레이아웃을 제공하고, 작은 터미널에서는 명시적 fallback 메시지 또는 축소 레이아웃으로 실패 없이 동작한다.
