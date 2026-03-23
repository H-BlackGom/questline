# Questline MVP1 실행 계획

## TL;DR
> **Summary**: 기존 `specs/001-questline-mvp1/` 문서는 방향은 맞지만 범위 드리프트와 실행 정보 부족이 있습니다. 이번 계획은 `/.opencode/document/requirement.md`를 기준 문서로 고정하고, 그린필드 Go CLI를 수직 슬라이스 방식으로 구현하도록 재구성합니다.
> **Deliverables**:
> - 드리프트가 제거된 MVP1 기능 계약 (`add`, `done`, `ls`, `me`)
> - 자동 부트스트랩 SQLite 저장소와 싱글톤 player 초기화
> - 기본 검증 범위의 Go 테스트/검증 체계 (`go test ./...`, `go vet ./...`)
> - 구현 결과와 일치하는 quickstart/README/스펙 문서 정렬
> **Effort**: Medium
> **Parallel**: YES - 4 waves
> **Critical Path**: 1 → 2 → 3 → 5 → 7 → 8 → F1-F4

## Context
### Original Request
- `specs/001-questline-mvp1/` 폴더 내 기존 계획 문서를 검토하고, 보강 지점을 반영한 새 실행 계획을 세운다.
- 결과와 중간 정리는 한국어로 유지한다.
- 이전 세션을 참고하고, 이전에 막혀 있던 권한 이슈가 해소되었다는 전제를 반영한다.

### Interview Summary
- 저장소는 사실상 그린필드 상태이며, 구현 코드와 테스트 인프라가 거의 없다.
- `specs/001-questline-mvp1/plan.md`, `specs/001-questline-mvp1/research.md`, `specs/001-questline-mvp1/data-model.md`, `specs/001-questline-mvp1/quickstart.md`, `specs/001-questline-mvp1/contracts/cli-commands.md`, `/.opencode/document/requirement.md`를 함께 검토했다.
- 품질 범위는 사용자 선택에 따라 `go test ./...`와 `go vet ./...`까지만 MVP1에 포함하고, 린트/CI는 제외한다.
- 설계 기본값은 다음과 같이 고정한다: 고정 `+50 XP`, 날짜 입력은 `YYYY-MM-DD`만 허용, 자연어 날짜(`tmr`) 제외, `DROPPED`는 예약 상태로만 유지, `due_date`는 today-or-later 허용.
- 구현 순서는 계층 중심 브랜치 분할 대신 `계약 정렬 → 부트스트랩 → 엔진 → add → done → ls → me → 문서/검증` 수직 슬라이스로 재편한다.

### Metis Review (gaps addressed)
- `ql me` 출력 계약, 첫 실행 자동 초기화, ID 표현 방식, `ql ls` 정렬 기준을 명시적으로 고정해야 한다는 지적을 반영한다.
- SQLite shared in-memory 대신 테스트마다 temp-file DB를 사용하도록 고정한다.
- 날짜 비교는 로컬 날짜 의미를 유지할 수 있도록 date-only 규칙으로 정리하고, 테스트는 고정된 clock/location 전제를 사용한다.
- 범위 확장을 막기 위해 difficulty/가변 XP/자연어 날짜/추가 서비스 계층/CI/lint/`DROPPED` CLI 행동을 명시적으로 제외한다.

## Work Objectives
### Core Objective
- 비어 있는 Go 저장소에 대해 Questline MVP1의 핵심 루프를 구현할 수 있도록, 문서 드리프트를 제거하고 구현/검증/커밋 기준까지 포함한 decision-complete 실행 계획을 제공한다.

### Deliverables
- `/.opencode/document/requirement.md`와 일치하도록 정렬된 `specs/001-questline-mvp1/` 파생 문서
- `cmd/ql`, `internal/cli`, `internal/engine`, `internal/repository`, 최소한의 `internal/domain` 중심 구조를 사용하는 CLI MVP1
- `~/.questline/data.db` 자동 생성, 스키마 마이그레이션, player 싱글톤 자동 보장
- 명령어 계약 테스트와 패키지 테스트를 포함한 기본 검증 체계
- 구현 결과에 맞는 `README.md` 및 quickstart 정리

### Definition of Done (verifiable conditions with commands)
- `go test ./...`가 성공한다.
- `go vet ./...`가 성공한다.
- `go build -o ./.tmp/ql ./cmd/ql`가 성공한다.
- `HOME="$(mktemp -d)" ./.tmp/ql add "Buy milk"`가 성공하고 새 퀘스트 ID를 출력한다.
- `HOME="$(mktemp -d)" ./.tmp/ql ls`가 첫 실행에도 DB를 초기화하고 빈 목록 또는 생성된 TODO 목록을 안정적으로 출력한다.
- `HOME="$(mktemp -d)" ./.tmp/ql me`가 첫 실행에도 player 상태를 생성하고 출력한다.

### Must Have
- `/.opencode/document/requirement.md:9` 기준의 MVP1 범위 고수
- 고정 `50 XP` 보상과 `100 + (level * 50)` 레벨업 공식
- 사용자 홈 아래 `~/.questline` 자동 생성 및 SQLite 초기화
- `add`, `done`, `ls`, `me` 네 명령어만 구현
- 모든 구현 태스크에 happy/failure QA와 증적 경로 포함
- temp HOME 기반 에이전트 검증으로 실제 사용자 환경을 건드리지 않음

### Must NOT Have (guardrails, AI slop patterns, scope boundaries)
- difficulty 선택, 가변 `xp_reward`, 자연어 날짜 파싱, 추가 명령어
- `service/` 또는 `pkg/` 레이어를 “미래 대비” 명분으로 선제 도입
- `golangci-lint`, GitHub Actions, 릴리스 자동화, Homebrew 배포 작업
- `DROPPED` 상태에 대한 CLI 명령/출력/QA 추가
- 공유 in-memory SQLite 전략을 기본 테스트 경로로 채택
- 장식 중심의 ASCII 아트 작업을 기능/검증보다 우선함

## Verification Strategy
> ZERO HUMAN INTERVENTION — all verification is agent-executed.
- Test decision: tests-after + Go 기본 `testing` 패키지 중심, 필요 시 `testify` 사용
- QA policy: 모든 태스크는 temp HOME 또는 temp DB를 사용하는 구체적 happy/failure 시나리오를 가진다.
- Evidence: `.sisyphus/evidence/task-{N}-{slug}.{ext}`
- Static verification: 각 기능 슬라이스 완료 시 `go test ./...` 후 `go vet ./...`
- Smoke verification: `go build -o ./.tmp/ql ./cmd/ql` 후 temp HOME으로 CLI 명령 실행

## Execution Strategy
### Parallel Execution Waves
> Target: 5-8 tasks per wave. <3 per wave (except final) = under-splitting.
> Extract shared dependencies as Wave-1 tasks for max parallelism.

Wave 1: 범위/계약 고정, 저장소 부트스트랩 기초
Wave 2: 엔진/도메인 규칙, `ql add`
Wave 3: `ql done`, `ql ls`, `ql me`
Wave 4: 문서/quickstart/README 정렬

### Dependency Matrix (full, all tasks)
| Task | Depends On | Unlocks |
|------|------------|---------|
| 1. 계약 정렬 | - | 2, 3, 4, 5, 6, 7, 8 |
| 2. 저장소 부트스트랩 | 1 | 3, 4, 5, 6, 7 |
| 3. 엔진/도메인 규칙 | 1, 2 | 4, 5, 7 |
| 4. `ql add` | 1, 2, 3 | 5, 6, 8 |
| 5. `ql done` | 1, 2, 3, 4 | 7, 8 |
| 6. `ql ls` | 1, 2, 4 | 8 |
| 7. `ql me` | 1, 2, 3, 5 | 8 |
| 8. 문서 정렬/최종 스모크 | 1, 4, 5, 6, 7 | F1-F4 |

### Agent Dispatch Summary (wave → task count → categories)
| Wave | Task Count | Recommended Categories |
|------|------------|------------------------|
| 1 | 2 | `writing`, `unspecified-high` |
| 2 | 2 | `unspecified-high`, `quick` |
| 3 | 3 | `unspecified-high`, `quick` |
| 4 | 1 | `writing` |
| Final | 4 | `oracle`, `unspecified-high`, `deep` |

## TODOs
> Implementation + Test = ONE task. Never separate.
> EVERY task MUST have: Agent Profile + Parallelization + QA Scenarios.

- [x] 1. MVP1 계약 드리프트 정렬

  **What to do**: `/.opencode/document/requirement.md`를 source of truth로 선언하고, `specs/001-questline-mvp1/data-model.md`, `specs/001-questline-mvp1/contracts/cli-commands.md`, `specs/001-questline-mvp1/quickstart.md`, `specs/001-questline-mvp1/plan.md`를 다음 기준으로 정렬한다: 고정 `+50 XP`, `YYYY-MM-DD`만 허용, 자연어 날짜 제외, `DROPPED`는 예약 상태로만 유지, `due_date`는 date-only(`YYYY-MM-DD`) 의미, `ql ls` 정렬은 `created_at DESC` 후 `id ASC`, `ql me`는 박스형 출력과 20칸 진행 바를 유지하되 테스트 대상은 핵심 라벨/값/순서로 제한한다. 사용자 ID 계약은 “UUID v4 앞 8자 소문자 문자열을 저장/표시/입력에 동일 사용”으로 고정한다.
  **Must NOT do**: difficulty, 가변 XP, `tmr`, 추가 명령, 린트/CI, 릴리스 자동화, `service/`/`pkg/` 확장 구조를 문서에 다시 넣지 않는다.

  **Recommended Agent Profile**:
  - Category: `writing` — Reason: 스펙 문서 간 충돌을 해소하고 실행 기준 문구를 정리하는 작업이다.
  - Skills: `[]` — 별도 스킬 없이 문서 정렬과 계약 동결에 집중한다.
  - Omitted: [`git-master`] — 커밋 계획만 필요하며 실제 git 조작은 이 태스크의 핵심이 아니다.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 2, 3, 4, 5, 6, 7, 8 | Blocked By: -

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `/.opencode/document/requirement.md:9` — MVP1 포함/제외 범위의 최우선 기준이다.
  - Pattern: `/.opencode/document/requirement.md:13` — 고정 `50 XP`, 레벨업 공식, 칭호 범위를 정의한다.
  - Pattern: `/.opencode/document/requirement.md:18` — DB 핵심 필드 요구를 정의한다.
  - Pattern: `/.opencode/document/requirement.md:27` — `ql add` 계약의 기준이다.
  - Pattern: `/.opencode/document/requirement.md:30` — `ql done` 계약의 기준이다.
  - Pattern: `/.opencode/document/requirement.md:33` — `ql ls` 플래그 범위의 기준이다.
  - Pattern: `/.opencode/document/requirement.md:38` — `ql me` 필수 표시 필드의 기준이다.
  - Pattern: `specs/001-questline-mvp1/data-model.md:28` — 현재 drift 원인인 status/difficulty/xp_reward 확장을 보여준다.
  - Pattern: `specs/001-questline-mvp1/data-model.md:77` — MVP1 밖 상태 전이를 추가하고 있으므로 제거 또는 보류 표기 대상이다.
  - Pattern: `specs/001-questline-mvp1/research.md:69` — 자연어 날짜를 MVP2로 미루는 기존 결정이다.
  - Pattern: `specs/001-questline-mvp1/contracts/cli-commands.md:192` — 현재 계약서가 가변 XP를 도입해 source-of-truth와 충돌한다.
  - Pattern: `specs/001-questline-mvp1/quickstart.md:152` — same-day due date 예시가 있으므로 validation 규칙을 today-or-later로 맞춘다.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `rg -n "tmr|[Dd]ifficulty|xp_reward|golangci|GitHub Actions" specs/001-questline-mvp1 .opencode/document/requirement.md` 결과를 검토했을 때, 금지 기능은 source-of-truth 기준에서 “MVP1 제외” 또는 제거 상태로 정리되어 있다.
  - [ ] `rg -n "50 XP|YYYY-MM-DD|created_at DESC|20 문자|20칸|8자" specs/001-questline-mvp1` 결과를 검토했을 때, 핵심 계약이 파생 문서 전반에 일관되게 반영되어 있다.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```text
  Scenario: Drift removal verification
    Tool: Bash
    Steps: `rg -n "tmr|[Dd]ifficulty|xp_reward|service/|pkg/|CI|lint" specs/001-questline-mvp1`
    Expected: 금지 항목이 남아 있더라도 모두 "후속 범위" 또는 "MVP1 제외"로만 언급되고 기능 계약으로는 남아 있지 않다.
    Evidence: .sisyphus/evidence/task-1-contract-freeze.txt

  Scenario: Source-of-truth preservation
    Tool: Bash
    Steps: `rg -n "add|done|ls|me|50 XP|YYYY-MM-DD" .opencode/document/requirement.md specs/001-questline-mvp1`
    Expected: 네 명령어와 고정 XP, 날짜 규칙이 일관되며 상충 문구가 남아 있지 않다.
    Evidence: .sisyphus/evidence/task-1-contract-freeze-check.txt
  ```

  **Commit**: YES | Message: `docs(spec): reconcile questline mvp1 contracts` | Files: `.opencode/document/requirement.md`, `specs/001-questline-mvp1/plan.md`, `specs/001-questline-mvp1/data-model.md`, `specs/001-questline-mvp1/contracts/cli-commands.md`, `specs/001-questline-mvp1/quickstart.md`

- [x] 2. 저장소 부트스트랩과 SQLite 초기화 뼈대 구축

  **What to do**: 최소 구조를 생성한다: `cmd/ql/main.go`, `internal/cli/root.go`, `internal/repository/sqlite.go`, `internal/repository/sqlite_test.go`, 필요 최소한의 `internal/domain/player.go`와 `internal/domain/quest.go`. `go.mod`에 `cobra`, `modernc.org/sqlite`, `testify`를 추가한다. `internal/repository/sqlite.go`는 외부 migration 디렉터리 없이 `CREATE TABLE IF NOT EXISTS` 기반으로 `~/.questline` 생성, SQLite open, schema ensure, player singleton ensure를 수행한다. 스키마는 `quests(id TEXT PRIMARY KEY, title TEXT NOT NULL, status TEXT NOT NULL, due_date TEXT NULL, created_at TEXT NOT NULL, completed_at TEXT NULL)`와 `player(id INTEGER PRIMARY KEY CHECK (id = 1), level INTEGER NOT NULL, current_xp INTEGER NOT NULL, total_xp_earned INTEGER NOT NULL, quests_completed INTEGER NOT NULL, updated_at TEXT NOT NULL)`를 기준으로 한다. `due_date`는 timezone 없는 canonical `YYYY-MM-DD` 문자열로 저장하고, 첫 실행 시 `ls`, `me`, `add`, `done`가 모두 자동 초기화 경로를 공유하도록 공용 bootstrap 함수를 제공한다.
  **Must NOT do**: `service/`/`pkg/`/CI 디렉터리를 미리 만들지 않는다. shared in-memory SQLite를 테스트 기본값으로 채택하지 않는다.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` — Reason: Go CLI 엔트리포인트, 저장소 초기화, temp DB 테스트 전략을 함께 잡아야 한다.
  - Skills: `[]` — 표준 Go와 SQLite 초기화 구현만으로 충분하다.
  - Omitted: [`frontend-ui-ux`] — CLI 프로젝트이므로 무관하다.

  **Parallelization**: Can Parallel: YES | Wave 1 | Blocks: 3, 4, 5, 6, 7 | Blocked By: 1

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `specs/001-questline-mvp1/plan.md:93` — 의도된 최상위 디렉터리 구조다.
  - Pattern: `/.opencode/document/requirement.md:18` — DB 경로와 핵심 테이블 필드 요구다.
  - Pattern: `specs/001-questline-mvp1/research.md:32` — pure-go SQLite 선택 근거다.
  - Pattern: `specs/001-questline-mvp1/research.md:343` — 기본 빌드 명령 예시다.
  - External: `https://pkg.go.dev/modernc.org/sqlite` — driver import 및 `database/sql` 연동 기준이다.
  - External: `https://pkg.go.dev/github.com/spf13/cobra` — root command 구성 기준이다.
  - External: `https://pkg.go.dev/os#UserHomeDir` — 홈 디렉터리 해석을 구현할 때 참고한다.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `go build -o ./.tmp/ql ./cmd/ql`가 성공한다.
  - [ ] `go test ./internal/repository/...`가 성공하고 temp-file SQLite 기반 bootstrap 테스트가 포함된다.
  - [ ] `HOME="$(mktemp -d)" ./.tmp/ql --help`가 성공하며 root command가 네 하위 명령을 노출한다.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```text
  Scenario: Root command exposes CLI surface without side effects
    Tool: Bash
    Steps: `TMP_HOME=$(mktemp -d) && go build -o ./.tmp/ql ./cmd/ql && HOME="$TMP_HOME" ./.tmp/ql --help | rg "add|done|ls|me"`
    Expected: root help가 네 하위 명령을 모두 노출하고 temp HOME 사용 시에도 실패하지 않는다.
    Evidence: .sisyphus/evidence/task-2-bootstrap.txt

  Scenario: Bootstrap test strategy creates schema and singleton player
    Tool: Bash
    Steps: `go test ./internal/repository/... -run TestBootstrap`
    Expected: 테스트가 temp-file DB 생성, `quests`/`player` 테이블 보장, player singleton 보장을 검증하며 성공한다.
    Evidence: .sisyphus/evidence/task-2-bootstrap-test.txt
  ```

  **Commit**: YES | Message: `chore: bootstrap questline persistence` | Files: `go.mod`, `go.sum`, `cmd/ql/main.go`, `internal/cli/root.go`, `internal/repository/sqlite.go`, `internal/repository/sqlite_test.go`, `internal/domain/player.go`, `internal/domain/quest.go`

- [x] 3. 엔진 규칙과 최소 도메인 계약 구현

  **What to do**: `internal/engine/leveling.go`와 `internal/engine/leveling_test.go`를 만들고, 고정 `50 XP` 지급을 전제로 레벨업 계산과 칭호 계산을 구현한다. `internal/domain/player.go`와 `internal/domain/quest.go`에는 MVP1에 필요한 최소 필드와 helper만 둔다: quest ID는 UUID v4 앞 8자 소문자, status는 `TODO`/`DONE`/예약용 `DROPPED`, due date는 date-only 문자열 또는 이를 감싼 최소 표현으로 유지한다. `Player`는 `Level`, `CurrentXP`, `TotalXPEarned`, `QuestsCompleted`와 다음 레벨 필요 XP 계산을 제공한다.
  **Must NOT do**: difficulty enum, 가변 XP 계산, drop/undo 전이, 복잡한 서비스 추상화 추가.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` — Reason: 순수 비즈니스 규칙과 경계 테스트가 핵심이다.
  - Skills: `[]` — 표준 Go 테스트만 사용한다.
  - Omitted: [`writing`] — 문서보다 코드와 테스트가 중심이다.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 4, 5, 7 | Blocked By: 1, 2

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `/.opencode/document/requirement.md:13` — XP, 레벨업 공식, 칭호 범위를 확정한다.
  - Pattern: `specs/001-questline-mvp1/plan.md:140` — 엔진 태스크의 원래 의도다.
  - Pattern: `specs/001-questline-mvp1/research.md:164` — `CalculateLevelUp` 결과 구조 예시다.
  - Pattern: `specs/001-questline-mvp1/research.md:188` — 칭호 매핑 기준이다.
  - Pattern: `specs/001-questline-mvp1/data-model.md:321` — player XP 증가에 필요한 최소 필드와 규칙이 드러난다.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `go test ./internal/engine/...`가 성공한다.
  - [ ] `go test ./... -run TestCalculateLevelUp`가 단일 레벨업, 연속 레벨업, 경계값, 레벨업 없음 케이스를 모두 통과한다.
  - [ ] `go vet ./...`가 성공한다.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```text
  Scenario: Exact boundary and multi-level progression
    Tool: Bash
    Steps: `go test ./internal/engine/... -run 'TestCalculateLevelUp|TestGetTitle'`
    Expected: Lv.1에서 150 XP는 정확히 Lv.2가 되고, 400 XP 입력은 2단계 이상 연속 레벨업 케이스를 통과한다.
    Evidence: .sisyphus/evidence/task-3-leveling.txt

  Scenario: Invalid/non-positive XP handling
    Tool: Bash
    Steps: `go test ./internal/engine/... -run TestRejectNonPositiveXP`
    Expected: 0 또는 음수 XP 입력 시 상태가 비정상적으로 변하지 않음을 검증하고 테스트가 성공한다.
    Evidence: .sisyphus/evidence/task-3-leveling-error.txt
  ```

  **Commit**: YES | Message: `feat: add quest engine core` | Files: `internal/engine/leveling.go`, `internal/engine/leveling_test.go`, `internal/domain/player.go`, `internal/domain/quest.go`

- [ ] 4. `ql add` 구현과 날짜 검증 계약 고정

  **What to do**: `internal/cli/add.go`를 구현하고 `root.go`에 등록한다. 입력 계약은 `ql add "<title>" [-d YYYY-MM-DD]`만 허용한다. 제목은 trim 후 비어 있지 않고 200자 이하, due date는 today-or-later 규칙을 적용한다. 성공 시 `✓ 퀘스트 #<id> 생성됨: "<title>"`를 출력하고, ID는 저장/표시 모두 8자 문자열을 사용한다. 잘못된 인자는 종료 코드 `2`, DB/bootstrap 오류는 종료 코드 `4`로 고정한다. `internal/repository`에는 quest insert와 조회에 필요한 최소 API를 추가하고 `internal/cli/add_test.go` 또는 이에 준하는 black-box 테스트를 만든다.
  **Must NOT do**: `-d tmr`, difficulty 플래그, 제목 자동 잘림, ANSI 색상 의존 테스트.

  **Recommended Agent Profile**:
  - Category: `quick` — Reason: bootstrap과 engine이 준비된 뒤에는 단일 명령 계약 구현으로 좁아진다.
  - Skills: `[]` — Cobra 인자 처리와 저장소 호출만 있으면 된다.
  - Omitted: [`unspecified-high`] — 명령 자체는 범위가 좁고 명확하다.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 5, 6, 8 | Blocked By: 1, 2, 3

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `/.opencode/document/requirement.md:27` — `ql add` 기능 범위의 원본 계약이다.
  - Pattern: `specs/001-questline-mvp1/contracts/cli-commands.md:8` — 인자/에러 메시지 예시가 있다.
  - Pattern: `specs/001-questline-mvp1/quickstart.md:65` — 성공/실패 사용 예시가 있다.
  - Pattern: `specs/001-questline-mvp1/data-model.md:71` — 제목 길이와 due date validation 방향을 보여준다.
  - External: `https://pkg.go.dev/time#Parse` — `YYYY-MM-DD` 파싱 기준이다.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `go test ./... -run TestAddCommand`가 성공한다.
  - [ ] `go build -o ./.tmp/ql ./cmd/ql && TMP_HOME=$(mktemp -d) && HOME="$TMP_HOME" ./.tmp/ql add "Buy milk"`가 성공하고 `퀘스트 #`를 포함한 생성 메시지를 출력한다.
  - [ ] `go build -o ./.tmp/ql ./cmd/ql && TMP_HOME=$(mktemp -d) && HOME="$TMP_HOME" ./.tmp/ql add "Buy milk" -d 2026-03-24`가 성공한다.
  - [ ] `go build -o ./.tmp/ql ./cmd/ql && TMP_HOME=$(mktemp -d) && HOME="$TMP_HOME" ./.tmp/ql add "Buy milk" -d 2026/03/24`는 종료 코드 `2`로 실패한다.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```text
  Scenario: Add quest with and without due date
    Tool: Bash
    Steps: `go build -o ./.tmp/ql ./cmd/ql && TMP_HOME=$(mktemp -d) && HOME="$TMP_HOME" ./.tmp/ql add "Buy milk" && HOME="$TMP_HOME" ./.tmp/ql add "Write docs" -d 2026-03-24`
    Expected: 두 명령 모두 성공하고 각각 8자 ID가 포함된 생성 메시지를 출력한다.
    Evidence: .sisyphus/evidence/task-4-add.txt

  Scenario: Reject invalid title/date input
    Tool: Bash
    Steps: `go build -o ./.tmp/ql ./cmd/ql && TMP_HOME=$(mktemp -d) && HOME="$TMP_HOME" ./.tmp/ql add "" ; test $? -eq 2` 와 `HOME="$TMP_HOME" ./.tmp/ql add "Buy milk" -d 2026/03/24 ; test $? -eq 2`
    Expected: 빈 제목과 잘못된 날짜 형식이 모두 종료 코드 2로 실패하고 에러 메시지를 출력한다.
    Evidence: .sisyphus/evidence/task-4-add-error.txt
  ```

  **Commit**: YES | Message: `feat: implement ql add` | Files: `internal/cli/add.go`, `internal/cli/add_test.go`, `internal/cli/root.go`, `internal/repository/sqlite.go`, `internal/repository/sqlite_test.go`

- [ ] 5. `ql done` 트랜잭션과 XP 지급 구현

  **What to do**: `internal/cli/done.go`와 관련 테스트를 구현한다. `done`는 quest 존재 확인, 이미 `DONE` 여부 확인, quest 상태 변경, `completed_at` 기록, player XP/레벨/완료 수 갱신을 하나의 DB 트랜잭션으로 처리해야 한다. 성공 출력은 기본 `✓ 퀘스트 완료! +50 XP`이며, 레벨업이 발생하면 이어서 `🎉 레벨업! Lv.X → Lv.Y`, `칭호: ...`, `다음 레벨까지: CUR/REQ XP`를 출력한다. 잘못된 인자 없음은 종료 코드 `2`, 미존재 quest는 `3`, 이미 완료된 quest는 `1`, DB 오류는 `4`로 고정한다. success/error 출력에 색상을 적용하되 테스트는 ANSI 제거 후 핵심 텍스트를 검증한다.
  **Must NOT do**: 재완료 시 XP 재지급, `DROPPED` 전이 처리, 난이도별 XP, 사람이 수동으로 DB를 손봐야 하는 보정 로직.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` — Reason: 트랜잭션, 비즈니스 규칙, CLI 계약을 동시에 묶는 핵심 슬라이스다.
  - Skills: `[]` — Go 표준 DB 트랜잭션과 테스트만으로 충분하다.
  - Omitted: [`writing`] — 설명보다 상태 전이와 검증 구현이 핵심이다.

  **Parallelization**: Can Parallel: YES | Wave 3 | Blocks: 7, 8 | Blocked By: 1, 2, 3, 4

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `/.opencode/document/requirement.md:30` — `ql done`의 원래 기능 계약이다.
  - Pattern: `specs/001-questline-mvp1/contracts/cli-commands.md:49` — 성공/실패 출력 예시가 있다.
  - Pattern: `specs/001-questline-mvp1/plan.md:150` — repository 트랜잭션 필요성을 명시한다.
  - Pattern: `specs/001-questline-mvp1/research.md:239` — 레벨업 시 출력 형식 예시다.
  - Pattern: `specs/001-questline-mvp1/data-model.md:182` — quest 완료와 player 진행의 관계를 설명한다.
  - External: `https://pkg.go.dev/database/sql#Tx` — SQLite 트랜잭션 처리 기준이다.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `go test ./... -run 'TestDoneCommand|TestDoneTransaction|TestDoneRollback'`가 성공한다.
  - [ ] `go build -o ./.tmp/ql ./cmd/ql && TMP_HOME=$(mktemp -d) && HOME="$TMP_HOME" ./.tmp/ql add "Quest A"` 후 생성된 ID로 `done`를 실행하면 `+50 XP`가 출력된다.
  - [ ] 같은 ID에 대해 두 번째 `done` 호출은 종료 코드 `1`로 실패하고 player XP를 다시 올리지 않는다.
  - [ ] 존재하지 않는 ID에 대한 `done` 호출은 종료 코드 `3`으로 실패한다.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```text
  Scenario: Complete quest and gain XP exactly once
    Tool: Bash
    Steps: `go build -o ./.tmp/ql ./cmd/ql && TMP_HOME=$(mktemp -d) && ID=$(HOME="$TMP_HOME" ./.tmp/ql add "Quest A" | sed -E 's/.*#([a-z0-9]{8}).*/\1/') && HOME="$TMP_HOME" ./.tmp/ql done "$ID" && HOME="$TMP_HOME" ./.tmp/ql me`
    Expected: `done`는 `+50 XP`를 출력하고, 이어진 `me` 출력은 `누적 XP: 50`과 `완료한 퀘스트: 1개`를 보여준다.
    Evidence: .sisyphus/evidence/task-5-done.txt

  Scenario: Reject repeated or unknown quest completion
    Tool: Bash
    Steps: `go build -o ./.tmp/ql ./cmd/ql && TMP_HOME=$(mktemp -d) && ID=$(HOME="$TMP_HOME" ./.tmp/ql add "Quest A" | sed -E 's/.*#([a-z0-9]{8}).*/\1/') && HOME="$TMP_HOME" ./.tmp/ql done "$ID" >/dev/null && HOME="$TMP_HOME" ./.tmp/ql done "$ID" ; test $? -eq 1 && HOME="$TMP_HOME" ./.tmp/ql done deadbeef ; test $? -eq 3`
    Expected: 이미 완료된 퀘스트는 종료 코드 1, 미존재 ID는 종료 코드 3으로 실패하고 XP는 추가 지급되지 않는다.
    Evidence: .sisyphus/evidence/task-5-done-error.txt
  ```

  **Commit**: YES | Message: `feat: implement ql done` | Files: `internal/cli/done.go`, `internal/cli/done_test.go`, `internal/repository/sqlite.go`, `internal/repository/sqlite_test.go`, `internal/engine/leveling.go`

- [ ] 6. `ql ls` 필터/정렬/빈 상태 출력 구현

  **What to do**: `internal/cli/ls.go`와 목록 조회용 repository API를 구현한다. 기본 동작은 `TODO`만 보여주고, `--done`은 `DONE`만, `--all`은 전체를 보여준다. `--all`과 `--done` 동시 사용은 종료 코드 `2`의 잘못된 인자로 고정한다. 출력은 헤더가 있는 표 형식이며 정렬 기준은 `created_at DESC`, tie-break는 `id ASC`다. title은 자르지 않고 그대로 출력하며 due date는 있으면 `MM-DD`, 없으면 `-`로 렌더링한다. 첫 실행/빈 데이터 시에는 `퀘스트가 없습니다. 'ql add'로 새 퀘스트를 만들어보세요!`를 출력한다.
  **Must NOT do**: 완료/전체 플래그 동시 허용, 임의 정렬, 제목 자동 말줄임표 처리, 빈 목록에서 table 헤더만 덩그러니 출력.

  **Recommended Agent Profile**:
  - Category: `quick` — Reason: 조회용 필터와 출력 계약이 중심이며 의존성이 제한적이다.
  - Skills: `[]` — CLI와 쿼리 정렬 구현만 있으면 된다.
  - Omitted: [`unspecified-high`] — 복잡한 트랜잭션보다는 조회 계약이 핵심이다.

  **Parallelization**: Can Parallel: YES | Wave 3 | Blocks: 8 | Blocked By: 1, 2, 4

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `/.opencode/document/requirement.md:33` — `ql ls` 범위와 플래그 계약의 원본이다.
  - Pattern: `specs/001-questline-mvp1/contracts/cli-commands.md:93` — 테이블 출력과 빈 상태 예시가 있다.
  - Pattern: `specs/001-questline-mvp1/quickstart.md:77` — CLI 사용 흐름과 샘플 출력이 있다.
  - Pattern: `specs/001-questline-mvp1/plan.md:158` — 상태별 목록 필터링 요구를 확인할 수 있다.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `go test ./... -run TestListCommand`가 성공한다.
  - [ ] `go build -o ./.tmp/ql ./cmd/ql && TMP_HOME=$(mktemp -d) && HOME="$TMP_HOME" ./.tmp/ql ls`가 첫 실행 시 빈 상태 메시지를 출력한다.
  - [ ] `go build -o ./.tmp/ql ./cmd/ql && TMP_HOME=$(mktemp -d)` 환경에서 여러 quest를 추가/완료한 뒤 `ls`, `ls --done`, `ls --all`이 기대된 행만 출력한다.
  - [ ] `go build -o ./.tmp/ql ./cmd/ql && TMP_HOME=$(mktemp -d) && HOME="$TMP_HOME" ./.tmp/ql ls --all --done`는 종료 코드 `2`로 실패한다.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```text
  Scenario: Filter TODO, DONE, and ALL deterministically
    Tool: Bash
    Steps: `go build -o ./.tmp/ql ./cmd/ql && TMP_HOME=$(mktemp -d) && ID1=$(HOME="$TMP_HOME" ./.tmp/ql add "Alpha" | sed -E 's/.*#([a-z0-9]{8}).*/\1/') && sleep 1 && ID2=$(HOME="$TMP_HOME" ./.tmp/ql add "Beta" | sed -E 's/.*#([a-z0-9]{8}).*/\1/') && HOME="$TMP_HOME" ./.tmp/ql done "$ID1" >/dev/null && HOME="$TMP_HOME" ./.tmp/ql ls && HOME="$TMP_HOME" ./.tmp/ql ls --done && HOME="$TMP_HOME" ./.tmp/ql ls --all`
    Expected: 기본 `ls`는 `Beta`만, `--done`은 `Alpha`만, `--all`은 `Beta` 후 `Alpha` 순으로 출력한다.
    Evidence: .sisyphus/evidence/task-6-ls.txt

  Scenario: Reject conflicting filter flags
    Tool: Bash
    Steps: `go build -o ./.tmp/ql ./cmd/ql && TMP_HOME=$(mktemp -d) && HOME="$TMP_HOME" ./.tmp/ql ls --all --done ; test $? -eq 2`
    Expected: 충돌하는 플래그 조합이 종료 코드 2와 에러 메시지로 실패한다.
    Evidence: .sisyphus/evidence/task-6-ls-error.txt
  ```

  **Commit**: YES | Message: `feat: implement ql ls` | Files: `internal/cli/ls.go`, `internal/cli/ls_test.go`, `internal/repository/sqlite.go`, `internal/repository/sqlite_test.go`

- [ ] 7. `ql me` 상태 패널과 진행 바 구현

  **What to do**: `internal/cli/me.go`를 구현하고 첫 실행에도 자동 bootstrap 경로를 사용한다. 출력은 계약서 예시처럼 박스형 패널을 사용하고, 내부 논리 기준 필드는 순서대로 `레벨`, `칭호`, `누적 XP`, `다음 레벨까지`, 20칸 진행 바, `완료한 퀘스트`를 포함한다. 첫 실행 출력은 `Lv.1`, `Intern`, `누적 XP: 0`, `다음 레벨까지: 0/150 XP`, `완료한 퀘스트: 0개`를 보여줘야 한다. 진행 바는 `current_xp / required_xp` 비율을 기반으로 채우며, 색상은 있어도 테스트는 ANSI 제거 후 논리 라인만 검증한다.
  **Must NOT do**: `me` 실행을 위해 사전 데이터 초기화를 수동으로 요구, 임의 필드 추가, progress bar 길이 변경, exact spacing만을 테스트 핵심으로 삼기.

  **Recommended Agent Profile**:
  - Category: `quick` — Reason: 읽기 전용 명령이지만 출력 계약이 비교적 명확하다.
  - Skills: `[]` — CLI 출력 포맷과 player 조회만 있으면 된다.
  - Omitted: [`writing`] — 문서보다 실행 결과 계약이 중요하다.

  **Parallelization**: Can Parallel: YES | Wave 3 | Blocks: 8 | Blocked By: 1, 2, 3, 5

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `/.opencode/document/requirement.md:38` — `ql me` 필수 정보의 원본이다.
  - Pattern: `specs/001-questline-mvp1/contracts/cli-commands.md:143` — 박스형 출력과 20칸 진행 바 기준이다.
  - Pattern: `specs/001-questline-mvp1/quickstart.md:122` — 첫 사용자 관점에서 기대되는 출력 형태다.
  - Pattern: `specs/001-questline-mvp1/data-model.md:137` — 다음 레벨 XP 계산 기준이다.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `go test ./... -run TestMeCommand`가 성공한다.
  - [ ] `go build -o ./.tmp/ql ./cmd/ql && TMP_HOME=$(mktemp -d) && HOME="$TMP_HOME" ./.tmp/ql me`가 첫 실행 상태를 출력한다.
  - [ ] 세 개의 quest를 완료한 뒤 `HOME="$TMP_HOME" ./.tmp/ql me`가 `Lv.2`, `누적 XP: 150`, `완료한 퀘스트: 3개`를 출력한다.
  - [ ] 진행 바는 20칸으로 렌더링되고 `current/required` 비율과 일치한다.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```text
  Scenario: First-run profile output
    Tool: Bash
    Steps: `go build -o ./.tmp/ql ./cmd/ql && TMP_HOME=$(mktemp -d) && HOME="$TMP_HOME" ./.tmp/ql me`
    Expected: 출력에 `Lv.1`, `Intern`, `누적 XP: 0`, `0/150 XP`, `완료한 퀘스트: 0개`가 순서대로 포함된다.
    Evidence: .sisyphus/evidence/task-7-me.txt

  Scenario: Progress updates after leveling
    Tool: Bash
    Steps: `go build -o ./.tmp/ql ./cmd/ql && TMP_HOME=$(mktemp -d) && for TITLE in A B C; do ID=$(HOME="$TMP_HOME" ./.tmp/ql add "$TITLE" | sed -E 's/.*#([a-z0-9]{8}).*/\1/'); HOME="$TMP_HOME" ./.tmp/ql done "$ID" >/dev/null; done && HOME="$TMP_HOME" ./.tmp/ql me`
    Expected: 세 번째 완료 후 `Lv.2`와 `누적 XP: 150`이 보이며 진행 바 비율도 새 레벨 기준으로 갱신된다.
    Evidence: .sisyphus/evidence/task-7-me-levelup.txt
  ```

  **Commit**: YES | Message: `feat: implement ql me` | Files: `internal/cli/me.go`, `internal/cli/me_test.go`, `internal/cli/root.go`, `internal/repository/sqlite.go`

- [ ] 8. README/quickstart 정렬 및 end-to-end 스모크 검증

  **What to do**: 구현 결과에 맞춰 `README.md`와 `specs/001-questline-mvp1/quickstart.md`를 정리한다. quickstart에는 `tmr`, difficulty, 릴리스 자동화, Homebrew 설치 같은 MVP1 외 항목을 제거하거나 후속 범위로 명시한다. temp HOME을 사용하는 end-to-end 스모크 스크립트 또는 테스트를 추가하여 `add → ls → done → me → ls --done` 전체 플로우를 검증한다. 최종 사용자 문서는 실제 구현된 종료 코드와 입력 형식(`YYYY-MM-DD`)을 반영해야 한다.
  **Must NOT do**: 구현되지 않은 배포/설치 경로를 확정된 기능처럼 문서화, 수동 검증 지시만 남기기, source-of-truth와 다른 CLI 예시 유지.

  **Recommended Agent Profile**:
  - Category: `writing` — Reason: 구현 완료 후 사용자 문서와 스모크 검증 경로를 정렬하는 작업이다.
  - Skills: `[]` — 문서 정리와 간단한 쉘 검증이면 충분하다.
  - Omitted: [`git-master`] — 문서 커밋 메시지는 필요하지만 git 조작 자체는 핵심이 아니다.

  **Parallelization**: Can Parallel: YES | Wave 4 | Blocks: F1-F4 | Blocked By: 1, 4, 5, 6, 7

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `README.md:1` — 현재 README가 비어 있으므로 사용자 진입 문서를 보강해야 한다.
  - Pattern: `specs/001-questline-mvp1/quickstart.md:7` — 현재 quickstart는 MVP1 외 설치/권한/고급 사용법까지 포함한다.
  - Pattern: `specs/001-questline-mvp1/contracts/cli-commands.md:216` — 종료 코드 표를 실제 구현과 맞춰야 한다.
  - Pattern: `/.opencode/document/requirement.md:25` — 문서가 반드시 반영해야 하는 네 명령 계약의 시작점이다.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `go test ./...`가 성공한다.
  - [ ] `go vet ./...`가 성공한다.
  - [ ] `go build -o ./.tmp/ql ./cmd/ql && TMP_HOME=$(mktemp -d)` 환경에서 `add → ls → done → me → ls --done` 전체 스모크 플로우가 성공한다.
  - [ ] `rg -n "tmr|Difficulty|xp_reward|brew install|GitHub Actions" README.md specs/001-questline-mvp1/quickstart.md` 결과가 MVP1 범위와 충돌하지 않는다.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```text
  Scenario: End-to-end happy path
    Tool: Bash
    Steps: `go build -o ./.tmp/ql ./cmd/ql && TMP_HOME=$(mktemp -d) && ID=$(HOME="$TMP_HOME" ./.tmp/ql add "Ship MVP" -d 2026-03-24 | sed -E 's/.*#([a-z0-9]{8}).*/\1/') && HOME="$TMP_HOME" ./.tmp/ql ls && HOME="$TMP_HOME" ./.tmp/ql done "$ID" && HOME="$TMP_HOME" ./.tmp/ql me && HOME="$TMP_HOME" ./.tmp/ql ls --done`
    Expected: add/ls/done/me/ls --done가 모두 성공하고 동일한 ID와 누적 XP 50을 일관되게 보여준다.
    Evidence: .sisyphus/evidence/task-8-e2e.txt

  Scenario: Documentation no longer advertises deferred scope
    Tool: Bash
    Steps: `rg -n "tmr|difficulty|xp_reward|Homebrew|GitHub Actions" README.md specs/001-questline-mvp1/quickstart.md`
    Expected: 구현 범위를 넘는 항목이 제거되었거나 명확히 "후속 범위"로만 남아 있다.
    Evidence: .sisyphus/evidence/task-8-docs.txt
  ```

  **Commit**: YES | Message: `docs: align quickstart and readme` | Files: `README.md`, `specs/001-questline-mvp1/quickstart.md`, `specs/001-questline-mvp1/contracts/cli-commands.md`

## Final Verification Wave (MANDATORY — after ALL implementation tasks)
> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.
> **Do NOT auto-proceed after verification. Wait for user's explicit approval before marking work complete.**
> **Never mark F1-F4 as checked before getting user's okay.** Rejection or user feedback -> fix -> re-run -> present again -> wait for okay.
- [ ] F1. Plan Compliance Audit — oracle
- [ ] F2. Code Quality Review — unspecified-high
- [ ] F3. Real Manual QA — unspecified-high
- [ ] F4. Scope Fidelity Check — deep

## Commit Strategy
- 커밋은 수직 슬라이스 단위로 끊고, 각 커밋마다 `go test ./...`와 `go vet ./...`가 통과해야 한다.
- 권장 순서: `docs: reconcile questline mvp1 contracts`, `chore: bootstrap questline persistence`, `feat: add leveling engine core`, `feat: implement ql add`, `feat: implement ql done`, `feat: implement ql ls`, `feat: implement ql me`, `docs: align quickstart and readme`
- 하나의 커밋에 여러 명령어를 섞지 않는다.

## Success Criteria
- 구현자는 추가 판단 없이 각 태스크의 입력/출력/검증/커밋 경계를 따라 작업할 수 있다.
- 최종 산출물은 고정 50 XP CLI MVP1이며, 파생 문서와 실제 동작이 일치한다.
- 기본 검증은 전부 에이전트 실행만으로 완료되고, 사용자 로컬 홈 디렉터리에 부작용을 남기지 않는다.
