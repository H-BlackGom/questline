# AGENT 계층 구조 정리 실행 계획

## TL;DR
> **Summary**: 현재 하나의 대형 `AGENT.md`에 섞여 있는 전역 규칙과 디렉터리별 규칙을 분리해, 짧은 루트 가이드와 5개의 로컬 가이드로 재구성한다. 문서 전용 작업이므로 이번 실행에서는 "코드+테스트" 완료 규칙에 문서 예외를 명시하고, 완료 기준은 문서 변경 + 기계 검증 + 원치 않는 중복 제거로 고정한다.
> **Deliverables**:
> - 전역 불변 규칙만 남긴 루트 `AGENT.md`
> - `internal/cli`, `internal/domain`, `internal/engine`, `internal/repository`, `specs` 하위의 로컬 `AGENT.md`
> - 루트/로컬 우선순위 및 중복 제거 규칙
> - `.specify/memory/constitution.md` 동기화
> **Effort**: Short
> **Parallel**: NO
> **Critical Path**: 1 → 2 → 3 → 4 → 5 → 6 → 7 → F1-F4

## Context
### Original Request
- 현재 `AGENT.md`를 토큰 낭비가 적은 계층 구조로 재구성하고 싶다.
- 공통 지침은 상위에, 세부 지침은 하위 디렉터리별로 나누고 싶다.
- 완료된 task는 코드/테스트/검증 기준으로 정의하고, task마다 commit 되길 원한다.

### Interview Summary
- 루트 `AGENT.md`는 짧게 유지하고, 하위 디렉터리별 `AGENT.md`로 세부 규칙을 분산한다.
- `cmd/ql`은 현재 wiring-only 경로이므로 별도 `AGENT.md`를 두지 않는다.
- 이번 작업에서는 문서 전용 task에 대한 예외를 허용한다.
- `.specify/memory/constitution.md`도 함께 동기화한다.
- `specs/AGENT.md`는 MVP1 세부가 아니라 공통 spec 유지 규칙만 담는다.

### Metis Review (gaps addressed)
- 문서 전용 작업의 완료 규칙 예외를 명시하지 않으면 실행자가 임의 해석할 수 있으므로 루트 `AGENT.md`에 명문화한다.
- 루트와 로컬 간 중복/권한 충돌을 막기 위해 "전역 규칙은 루트, 경로 전용 규칙은 로컬" 우선순위를 명시한다.
- `cmd/ql`에 로컬 가이드를 만들지 않는 대신, wiring-only 규칙은 루트에 남겨 discoverability를 보장한다.
- constitution은 별도 trailing task로 동기화해 AGENT 계층 작업과 정책 동기화를 분리한다.

## Work Objectives
### Core Objective
- 기존 대형 `AGENT.md`를 전역/로컬 책임 기준으로 분해해, 에이전트가 항상 가장 가까운 지침만 읽어도 올바르게 작업할 수 있는 계층형 가이드 구조를 만든다.

### Deliverables
- `AGENT.md` 재작성
- `internal/cli/AGENT.md`
- `internal/domain/AGENT.md`
- `internal/engine/AGENT.md`
- `internal/repository/AGENT.md`
- `specs/AGENT.md`
- `.specify/memory/constitution.md` 동기화

### Definition of Done (verifiable conditions with commands)
- `test -f AGENT.md` 가 성공한다.
- `test -f internal/cli/AGENT.md && test -f internal/domain/AGENT.md && test -f internal/engine/AGENT.md && test -f internal/repository/AGENT.md && test -f specs/AGENT.md` 가 성공한다.
- `test ! -f cmd/ql/AGENT.md` 가 성공한다.
- `grep -n "Task 완료 계약" AGENT.md` 와 `grep -n "Git 계약" AGENT.md` 가 성공한다.
- `grep -n "문서 전용 작업 예외" AGENT.md` 가 성공한다.
- `grep -n "exit code\|종료 코드" internal/cli/AGENT.md` 가 성공한다.
- `grep -n "50 XP\|레벨" internal/engine/AGENT.md` 가 성공한다.
- `grep -n "singleton\|단일 row\|id=1" internal/repository/AGENT.md` 가 성공한다.
- `grep -n "specs/AGENT.md\|하위 가이드 위치" AGENT.md` 가 성공한다.
- `go test ./...` 와 `go vet ./...` 가 계속 성공한다.

### Must Have
- 루트 `AGENT.md`는 전역 불변 규칙만 담는다.
- 문서 전용 작업 예외를 루트에 명시한다.
- task 완료 후 1 task = 1 commit 규칙을 루트에 명시한다.
- `cmd/ql/AGENT.md`는 생성하지 않는다.
- 각 로컬 `AGENT.md`는 자기 디렉터리 하위에만 적용되는 규칙만 담는다.
- `.specify/memory/constitution.md`는 현재 구현/운영 원칙과 충돌하지 않도록 정리한다.

### Must NOT Have (guardrails, AI slop patterns, scope boundaries)
- 루트와 로컬에 같은 규칙을 canonical하게 중복 기재하지 않는다.
- AGENT 구조 정리 중 실제 제품 기능/CLI 동작/DB 스키마 구현을 바꾸지 않는다.
- `cmd/ql/AGENT.md`를 새로 만들지 않는다.
- `service/`, `pkg/`, difficulty, 가변 XP, 자연어 날짜를 새 정책처럼 부활시키지 않는다.
- unrelated README/spec 정리를 끼워 넣지 않는다.

## Verification Strategy
> ZERO HUMAN INTERVENTION — all verification is agent-executed.
- Test decision: docs-only exception + repo safety checks 유지
- QA policy: 각 task는 파일 존재/부재, 헤더 존재, 금지 중복 제거를 쉘 명령으로 검증한다.
- Evidence: `.sisyphus/evidence/task-{N}-{slug}.txt`
- Repo safety: 각 완료 task 뒤 `go test ./...` 와 `go vet ./...` 실행
- Scope safety: 각 task 뒤 `git diff --name-only`로 허용 파일만 바뀌었는지 확인

## Execution Strategy
### Parallel Execution Waves
> 문서 구조 마이그레이션이므로 순차 실행이 더 안전하다. 각 단계가 직전 단계의 중복 제거를 전제한다.

Wave 1: 루트 기초 재작성 + 로컬 파일 뼈대
Wave 2: CLI/Domain 추출
Wave 3: Engine/Repository 추출
Wave 4: Specs 추출 + constitution 동기화

### Dependency Matrix (full, all tasks)
| Task | Depends On | Unlocks |
|------|------------|---------|
| 1. 루트 재작성 + skeleton | - | 2, 3, 4, 5, 6, 7 |
| 2. CLI 가이드 추출 | 1 | 6 |
| 3. Domain 가이드 추출 | 1 | 6 |
| 4. Engine 가이드 추출 | 1 | 6, 7 |
| 5. Repository 가이드 추출 | 1 | 6, 7 |
| 6. Specs 가이드 추출 + 루트 정리 | 2, 3, 4, 5 | 7 |
| 7. Constitution 동기화 | 4, 5, 6 | F1-F4 |

### Agent Dispatch Summary (wave → task count → categories)
| Wave | Task Count | Recommended Categories |
|------|------------|------------------------|
| 1 | 1 | `writing` |
| 2 | 2 | `writing` |
| 3 | 2 | `writing` |
| 4 | 2 | `writing` |
| Final | 4 | `oracle`, `unspecified-high`, `deep` |

## TODOs
> 문서 수정 + 검증 + 커밋 = ONE task.
> 모든 task는 허용 파일 범위를 벗어나면 실패로 간주한다.

- [ ] 1. 루트 `AGENT.md` 재작성과 로컬 AGENT skeleton 생성

  **What to do**: 현재 `AGENT.md`를 전면 재작성해 전역 규칙만 남긴다. 동시에 빈 뼈대가 아니라 최소 유효 섹션을 가진 로컬 파일 `internal/cli/AGENT.md`, `internal/domain/AGENT.md`, `internal/engine/AGENT.md`, `internal/repository/AGENT.md`, `specs/AGENT.md`를 생성한다. 루트에는 반드시 `프로젝트 범위`, `Task 완료 계약`, `문서 전용 작업 예외`, `Git 계약`, `공통 검증 명령`, `범위 가드`, `하위 가이드 위치` 섹션을 둔다.
  **Must NOT do**: 이 단계에서 세부 규칙을 전부 루트에서 제거하지 않은 채 로컬 파일 없이 참조만 남기지 않는다. `cmd/ql/AGENT.md`를 만들지 않는다.

  **Recommended Agent Profile**:
  - Category: `writing` — Reason: 구조 재정의와 문서 skeleton 생성이 중심이다.
  - Skills: `[]` — 별도 스킬 없이 문서 리팩터링으로 충분하다.
  - Omitted: [`git-master`] — 커밋은 task 종료 시 수행하면 된다.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 2, 3, 4, 5, 6, 7 | Blocked By: -

  **References**:
  - Pattern: `AGENT.md:1-281` — 현재 전역/로컬 규칙이 모두 섞여 있는 원본이다.
  - Pattern: `.sisyphus/drafts/agent-guidelines-hierarchy.md:92-161` — 루트와 로컬 분리 초안 및 한국어 문안이 정리되어 있다.
  - Pattern: `README.md:1-233` — 현재 사용자 문서의 톤과 설치/실행 설명 기준이다.

  **Acceptance Criteria**:
  - [ ] `test -f AGENT.md && test -f internal/cli/AGENT.md && test -f internal/domain/AGENT.md && test -f internal/engine/AGENT.md && test -f internal/repository/AGENT.md && test -f specs/AGENT.md` 가 성공한다.
  - [ ] `test ! -f cmd/ql/AGENT.md` 가 성공한다.
  - [ ] `grep -n "Task 완료 계약" AGENT.md && grep -n "문서 전용 작업 예외" AGENT.md && grep -n "Git 계약" AGENT.md` 가 성공한다.
  - [ ] `go test ./... && go vet ./...` 가 성공한다.

  **QA Scenarios**:
  ```text
  Scenario: Root and local AGENT skeletons exist
    Tool: Bash
    Steps: `test -f AGENT.md && test -f internal/cli/AGENT.md && test -f internal/domain/AGENT.md && test -f internal/engine/AGENT.md && test -f internal/repository/AGENT.md && test -f specs/AGENT.md && test ! -f cmd/ql/AGENT.md`
    Expected: 필요한 6개 파일만 존재하고 `cmd/ql/AGENT.md`는 없다.
    Evidence: .sisyphus/evidence/task-1-agent-foundation.txt

  Scenario: Root contains mandatory global sections
    Tool: Bash
    Steps: `grep -n "프로젝트 범위\|Task 완료 계약\|문서 전용 작업 예외\|Git 계약\|범위 가드\|하위 가이드 위치" AGENT.md`
    Expected: 전역 섹션들이 모두 존재한다.
    Evidence: .sisyphus/evidence/task-1-agent-foundation-sections.txt
  ```

  **Commit**: YES | Message: `docs(agent): establish hierarchical guide skeleton` | Files: `AGENT.md`, `internal/cli/AGENT.md`, `internal/domain/AGENT.md`, `internal/engine/AGENT.md`, `internal/repository/AGENT.md`, `specs/AGENT.md`

- [ ] 2. `internal/cli/AGENT.md` 채우기와 루트의 CLI 세부 규칙 제거

  **What to do**: 현재 루트에 있는 종료 코드, 명령별 입력 검증, stdout/stderr, CLI 테스트 규칙을 `internal/cli/AGENT.md`로 옮긴다. 루트에는 CLI-specific 세부를 남기지 않고, 필요하면 "CLI 세부는 `internal/cli/AGENT.md` 참조" 정도의 짧은 라우팅만 남긴다.
  **Must NOT do**: 전역 exit code 불변 규칙 자체를 완전히 삭제하지 않는다. 단, 명령별 상세 계약은 루트에서 제거한다.

  **Recommended Agent Profile**:
  - Category: `writing` — Reason: 기존 문안을 CLI 책임 기준으로 재배치하는 작업이다.
  - Skills: `[]` — 문서 편집만 필요하다.
  - Omitted: [`git-master`] — task 단위 커밋만 수행하면 된다.

  **Parallelization**: Can Parallel: NO | Wave 2 | Blocks: 6 | Blocked By: 1

  **References**:
  - Pattern: `AGENT.md:81-115` — 현재 CLI 계약과 출력 규칙의 원본이다.
  - Pattern: `.sisyphus/drafts/agent-guidelines-hierarchy.md:183-216` — CLI 로컬 가이드 한국어 초안이다.
  - Pattern: `specs/001-questline-mvp1/contracts/cli-commands.md:8-222` — CLI 계약 용어와 표현 기준이다.

  **Acceptance Criteria**:
  - [ ] `grep -n "종료 코드 규칙\|출력 규칙\|명령별 규칙" internal/cli/AGENT.md` 가 성공한다.
  - [ ] `! grep -n "ql add\|ql done\|ql ls\|ql me" AGENT.md` 가 성공한다.
  - [ ] `go test ./... && go vet ./...` 가 성공한다.

  **QA Scenarios**:
  ```text
  Scenario: CLI rules are canonical only in local guide
    Tool: Bash
    Steps: `grep -n "종료 코드 규칙\|명령별 규칙" internal/cli/AGENT.md && ! grep -n "ql add\|ql done\|ql ls\|ql me" AGENT.md`
    Expected: CLI-specific canonical 규칙은 `internal/cli/AGENT.md`에서만 찾을 수 있다.
    Evidence: .sisyphus/evidence/task-2-cli-guide.txt

  Scenario: Root still preserves only global routing
    Tool: Bash
    Steps: `grep -n "internal/cli/AGENT.md" AGENT.md`
    Expected: 루트에는 CLI 세부 대신 라우팅만 남는다.
    Evidence: .sisyphus/evidence/task-2-cli-guide-root.txt
  ```

  **Commit**: YES | Message: `docs(agent): extract cli-specific guidance` | Files: `AGENT.md`, `internal/cli/AGENT.md`

- [ ] 3. `internal/domain/AGENT.md` 채우기와 루트의 도메인 세부 규칙 제거

  **What to do**: Quest/Player 필드 기대값, `DROPPED` 예약 상태, plain struct 원칙, 외부 의존성 금지 등 도메인 규칙을 `internal/domain/AGENT.md`에 정리하고 루트에서 제거한다.
  **Must NOT do**: 전역 MVP1 범위 가드에 필요한 핵심 금지사항까지 루트에서 없애지 않는다.

  **Recommended Agent Profile**:
  - Category: `writing` — Reason: 모델 수준 규칙을 한 파일로 고정하는 문서 작업이다.
  - Skills: `[]`
  - Omitted: [`git-master`]

  **Parallelization**: Can Parallel: NO | Wave 2 | Blocks: 6 | Blocked By: 1

  **References**:
  - Pattern: `AGENT.md:29-45` — 현재 도메인 모델 설명 원본이다.
  - Pattern: `AGENT.md:76-79` — domain에 외부 의존성을 두지 않는 전역 구조 규칙이다.
  - Pattern: `.sisyphus/drafts/agent-guidelines-hierarchy.md:218-238` — domain 가이드 한국어 초안이다.

  **Acceptance Criteria**:
  - [ ] `grep -n "Quest 규칙\|Player 규칙\|DROPPED" internal/domain/AGENT.md` 가 성공한다.
  - [ ] `! grep -n "UUID v4 first 8 chars\|Player \(Singleton" AGENT.md` 가 성공한다.
  - [ ] `go test ./... && go vet ./...` 가 성공한다.

  **QA Scenarios**:
  ```text
  Scenario: Domain rules moved to local guide
    Tool: Bash
    Steps: `grep -n "Quest 규칙\|Player 규칙\|DROPPED" internal/domain/AGENT.md && ! grep -n "UUID v4 first 8 chars\|Player \(Singleton" AGENT.md`
    Expected: domain canonical 규칙은 `internal/domain/AGENT.md`로 이동한다.
    Evidence: .sisyphus/evidence/task-3-domain-guide.txt

  Scenario: Root still keeps MVP1 scope guard
    Tool: Bash
    Steps: `grep -n "범위 가드\|difficulty\|service/\|pkg/" AGENT.md`
    Expected: 전역 범위 가드는 루트에 남아 있다.
    Evidence: .sisyphus/evidence/task-3-domain-root-guard.txt
  ```

  **Commit**: YES | Message: `docs(agent): extract domain guidance` | Files: `AGENT.md`, `internal/domain/AGENT.md`

- [ ] 4. `internal/engine/AGENT.md` 채우기와 루트의 XP/레벨링 규칙 제거

  **What to do**: 고정 `50 XP`, 레벨업 공식, 칭호 구간, pure logic 원칙, table-driven test 기대치를 `internal/engine/AGENT.md`로 옮긴다. 루트에는 이 수학적 세부를 남기지 않는다.
  **Must NOT do**: 루트의 MVP1 범위 가드에서 "difficulty/가변 XP 금지" 문구를 제거하지 않는다.

  **Recommended Agent Profile**:
  - Category: `writing` — Reason: 비즈니스 규칙을 엔진 소유 문서로 정리하는 작업이다.
  - Skills: `[]`
  - Omitted: [`git-master`]

  **Parallelization**: Can Parallel: NO | Wave 3 | Blocks: 7 | Blocked By: 1

  **References**:
  - Pattern: `AGENT.md:14-27` — XP/레벨링/칭호 원본 규칙이다.
  - Pattern: `AGENT.md:120-141` — 엔진 테스트 기대치 원본이다.
  - Pattern: `.sisyphus/drafts/agent-guidelines-hierarchy.md:240-260` — engine 가이드 한국어 초안이다.

  **Acceptance Criteria**:
  - [ ] `grep -n "50 XP\|레벨\|칭호\|table-driven" internal/engine/AGENT.md` 가 성공한다.
  - [ ] `! grep -n "Lv.1-9\|Level Up Formula\|Fixed XP" AGENT.md` 가 성공한다.
  - [ ] `go test ./... && go vet ./...` 가 성공한다.

  **QA Scenarios**:
  ```text
  Scenario: Engine formulas are local-only
    Tool: Bash
    Steps: `grep -n "50 XP\|레벨\|칭호" internal/engine/AGENT.md && ! grep -n "Level Up Formula\|Lv.1-9" AGENT.md`
    Expected: 엔진 수식/칭호는 local guide에만 canonical하게 존재한다.
    Evidence: .sisyphus/evidence/task-4-engine-guide.txt

  Scenario: Root still blocks variable XP
    Tool: Bash
    Steps: `grep -n "variable XP\|가변 XP\|difficulty" AGENT.md`
    Expected: 수학적 세부는 빠졌지만 전역 금지사항은 남아 있다.
    Evidence: .sisyphus/evidence/task-4-engine-root-guard.txt
  ```

  **Commit**: YES | Message: `docs(agent): extract engine guidance` | Files: `AGENT.md`, `internal/engine/AGENT.md`

- [ ] 5. `internal/repository/AGENT.md` 채우기와 루트의 저장소 세부 규칙 제거

  **What to do**: SQLite 스키마, 자동 초기화, singleton player, transaction, temp-file DB 테스트 규칙을 `internal/repository/AGENT.md`에 정리하고 루트의 저장소 세부를 제거한다.
  **Must NOT do**: 루트에서 `go test ./...`, `go vet ./...` 같은 전역 검증 규칙까지 제거하지 않는다.

  **Recommended Agent Profile**:
  - Category: `writing` — Reason: persistence 소유 규칙을 한곳으로 정리하는 작업이다.
  - Skills: `[]`
  - Omitted: [`git-master`]

  **Parallelization**: Can Parallel: NO | Wave 3 | Blocks: 7 | Blocked By: 1

  **References**:
  - Pattern: `AGENT.md:173-201` — SQLite 스키마/초기화 원본이다.
  - Pattern: `AGENT.md:121-123,139-141` — repository 테스트 전략 원본이다.
  - Pattern: `.sisyphus/drafts/agent-guidelines-hierarchy.md:262-284` — repository 가이드 한국어 초안이다.

  **Acceptance Criteria**:
  - [ ] `grep -n "SQLite\|singleton\|트랜잭션\|temp-file" internal/repository/AGENT.md` 가 성공한다.
  - [ ] `! grep -n "CREATE TABLE quests\|CREATE TABLE player" AGENT.md` 가 성공한다.
  - [ ] `go test ./... && go vet ./...` 가 성공한다.

  **QA Scenarios**:
  ```text
  Scenario: Repository schema and transaction guidance moved locally
    Tool: Bash
    Steps: `grep -n "SQLite\|singleton\|트랜잭션\|temp-file" internal/repository/AGENT.md && ! grep -n "CREATE TABLE quests\|CREATE TABLE player" AGENT.md`
    Expected: 저장소 세부 규칙은 local guide에만 남는다.
    Evidence: .sisyphus/evidence/task-5-repository-guide.txt

  Scenario: Root still keeps global verification commands
    Tool: Bash
    Steps: `grep -n "go test ./...\|go vet ./...\|go build -o ./.tmp/ql ./cmd/ql" AGENT.md`
    Expected: 전역 검증 명령은 루트에 남는다.
    Evidence: .sisyphus/evidence/task-5-repository-root-checks.txt
  ```

  **Commit**: YES | Message: `docs(agent): extract repository guidance` | Files: `AGENT.md`, `internal/repository/AGENT.md`

- [ ] 6. `specs/AGENT.md` 채우기와 루트의 문서 동기화 규칙 정리

  **What to do**: `specs/AGENT.md`에는 공통 spec 유지 규칙만 넣는다. README/quickstart/contracts/data-model 동기화 규칙, MVP2 reserved 표기 규칙, 구현되지 않은 기능 문서화 금지 등을 옮긴다. 루트에는 문서 동기화의 존재만 남기고, 구체 대상 파일 목록은 `specs/AGENT.md`에 canonical하게 둔다.
  **Must NOT do**: MVP1 세부 구현 내용을 `specs/AGENT.md`로 새로 승격하지 않는다.

  **Recommended Agent Profile**:
  - Category: `writing` — Reason: spec process 규칙을 독립 문서로 분리하는 작업이다.
  - Skills: `[]`
  - Omitted: [`git-master`]

  **Parallelization**: Can Parallel: NO | Wave 4 | Blocks: 7 | Blocked By: 2, 3, 4, 5

  **References**:
  - Pattern: `AGENT.md:244-261` — README/spec 동기화 원본 규칙이다.
  - Pattern: `.sisyphus/drafts/agent-guidelines-hierarchy.md:286-307` — specs 가이드 한국어 초안이다.
  - Pattern: `specs/001-questline-mvp1/quickstart.md:7-92` — 설치/사용 예시가 실제 문서 대상임을 보여준다.
  - Pattern: `specs/001-questline-mvp1/contracts/cli-commands.md:8-222` — 계약 문서가 별도 canonical source임을 보여준다.

  **Acceptance Criteria**:
  - [ ] `grep -n "같이 봐야 할 문서\|동기화 규칙\|금지 사항" specs/AGENT.md` 가 성공한다.
  - [ ] `! grep -n "README.md\|quickstart.md\|contracts/cli-commands.md\|data-model.md" AGENT.md` 가 성공한다.
  - [ ] `go test ./... && go vet ./...` 가 성공한다.

  **QA Scenarios**:
  ```text
  Scenario: Docs sync rules are canonical in specs guide
    Tool: Bash
    Steps: `grep -n "같이 봐야 할 문서\|동기화 규칙\|금지 사항" specs/AGENT.md && ! grep -n "quickstart.md\|contracts/cli-commands.md\|data-model.md" AGENT.md`
    Expected: 문서 동기화 세부 규칙은 `specs/AGENT.md`로 이동한다.
    Evidence: .sisyphus/evidence/task-6-specs-guide.txt

  Scenario: specs guide remains cross-spec only
    Tool: Bash
    Steps: `! grep -n "ql add\|ql done\|50 XP\|Lv.1" specs/AGENT.md`
    Expected: `specs/AGENT.md`는 MVP1 구현 세부를 직접 소유하지 않는다.
    Evidence: .sisyphus/evidence/task-6-specs-scope.txt
  ```

  **Commit**: YES | Message: `docs(agent): extract specs maintenance guidance` | Files: `AGENT.md`, `specs/AGENT.md`

- [ ] 7. `.specify/memory/constitution.md`를 새 AGENT 구조와 현재 MVP1 정책에 맞게 동기화

  **What to do**: constitution에서 현재 프로젝트와 충돌하는 항목을 수정한다. 구체적으로 `service/`, `pkg/`, difficulty, 가변 XP, `questline <command>` 명령 형식, in-constitution schema 예시의 outdated 필드를 현재 MVP1 정책에 맞게 정리한다. AGENT 계층 구조를 직접 헌법에 넣을 필요는 없지만, AI 협업/코드 품질 원칙이 새 AGENT 체계와 충돌하지 않도록 업데이트한다. Sync Impact Report와 버전/최종 개정일도 함께 갱신한다.
  **Must NOT do**: constitution을 AGENT.md의 상세 문구 복제본으로 만들지 않는다. 템플릿 파일까지 건드리는 것은 실제 충돌이 확인된 경우에만 포함하고, 없으면 제외한다.

  **Recommended Agent Profile**:
  - Category: `writing` — Reason: 정책 문서 정합성 정리가 핵심이다.
  - Skills: `[]`
  - Omitted: [`git-master`]

  **Parallelization**: Can Parallel: NO | Wave 4 | Blocks: F1-F4 | Blocked By: 4, 5, 6

  **References**:
  - Pattern: `.specify/memory/constitution.md:38-49` — 현재 service/pkg를 요구해 충돌한다.
  - Pattern: `.specify/memory/constitution.md:57-71` — difficulty/questline command 규칙이 현재 구현과 충돌한다.
  - Pattern: `.specify/memory/constitution.md:79-106` — outdated schema 예시가 존재한다.
  - Pattern: `.specify/memory/constitution.md:115-123` — 레벨업 공식과 시각 피드백이 현재 정책과 다르다.
  - Pattern: `AGENT.md:47-79` — 현재 아키텍처/레이어 원칙의 최신 기준이다.
  - Pattern: `AGENT.md:203-223` — 현재 MVP1 scope boundary 기준이다.

  **Acceptance Criteria**:
  - [ ] `! grep -n "service/\|pkg/\|difficulty\|easy: 30\|normal: 50\|hard: 100\|questline add" .specify/memory/constitution.md` 가 성공한다.
  - [ ] `grep -n "ql add\|50 XP\|YYYY-MM-DD\|cmd/\|internal/" .specify/memory/constitution.md` 가 성공한다.
  - [ ] `go test ./... && go vet ./...` 가 성공한다.

  **QA Scenarios**:
  ```text
  Scenario: Constitution no longer conflicts with current MVP1 policy
    Tool: Bash
    Steps: `! grep -n "service/\|pkg/\|difficulty\|easy: 30\|normal: 50\|hard: 100\|questline add" .specify/memory/constitution.md && grep -n "ql add\|50 XP\|YYYY-MM-DD" .specify/memory/constitution.md`
    Expected: 오래된 정책이 제거되고 현재 MVP1 기준이 반영된다.
    Evidence: .sisyphus/evidence/task-7-constitution-sync.txt

  Scenario: Constitution remains principle-level, not AGENT duplicate
    Tool: Bash
    Steps: `! grep -n "Task 완료 계약\|문서 전용 작업 예외\|하위 가이드 위치" .specify/memory/constitution.md`
    Expected: constitution은 원칙 문서로 남고 AGENT 상세 운영 규칙 복제본이 되지 않는다.
    Evidence: .sisyphus/evidence/task-7-constitution-scope.txt
  ```

  **Commit**: YES | Message: `docs(constitution): align governance with current MVP1 guidance` | Files: `.specify/memory/constitution.md`

## Final Verification Wave (MANDATORY — after ALL implementation tasks)
> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.
> **Do NOT auto-proceed after verification. Wait for user's explicit approval before marking work complete.**
> **Never mark F1-F4 as checked before getting user's okay.** Rejection or user feedback -> fix -> re-run -> present again -> wait for okay.
- [ ] F1. Plan Compliance Audit — oracle
- [ ] F2. Code Quality Review — unspecified-high
- [ ] F3. Real Manual QA — unspecified-high
- [ ] F4. Scope Fidelity Check — deep

## Commit Strategy
- 커밋은 반드시 task 단위로 끊는다.
- 각 커밋 전 검증은 해당 task acceptance criteria + `go test ./...` + `go vet ./...`를 따른다.
- 권장 순서:
  1. `docs(agent): establish hierarchical guide skeleton`
  2. `docs(agent): extract cli-specific guidance`
  3. `docs(agent): extract domain guidance`
  4. `docs(agent): extract engine guidance`
  5. `docs(agent): extract repository guidance`
  6. `docs(agent): extract specs maintenance guidance`
  7. `docs(constitution): align governance with current MVP1 guidance`

## Success Criteria
- 루트 `AGENT.md`는 전역 규칙만 남고 5개의 로컬 AGENT 파일이 책임별 세부 규칙을 가진다.
- 실행자는 더 이상 루트 대형 문서 전체를 읽지 않고도 현재 작업 디렉터리 기준 가이드를 찾을 수 있다.
- 문서 전용 작업 예외, task 완료 규칙, task당 commit 규칙이 루트에 명시된다.
- constitution이 현재 MVP1 정책과 충돌하지 않는다.
