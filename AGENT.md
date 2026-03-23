# Agent Guide

## 프로젝트 범위
- Questline은 RPG 진행 시스템이 있는 Go 기반 CLI 프로젝트다.
- MVP1 범위는 `add`, `done`, `ls`, `me` 네 명령만 포함한다.
- 퀘스트 완료 보상은 항상 고정 `50 XP`다.
- 날짜 입력은 `YYYY-MM-DD` 형식만 허용한다.

## Task 완료 계약
다음 조건을 모두 만족해야 task가 완료된 것으로 본다.
1. 구현 코드가 작성되거나 수정되었다.
2. 관련 테스트 코드가 작성되거나 수정되었다.
3. 대상 테스트가 통과했다.
4. `go test ./...` 가 통과했다.
5. `go vet ./...` 가 통과했다.

## 문서 전용 작업 예외
문서 정리/리팩터링 작업에 한해 예외를 허용한다:
- 완료 기준: 문서 변경 + 기계 검증 + 의도하지 않은 중복 제거
- 코드 변경 없이 문서만 바뀌어도 검증 명령은 계속 수행한다.

## Git 계약
- Commit after every completed task.
- One completed task = one atomic commit.
- Do not batch unrelated finished tasks.
- Never commit before tests are green.

## Global Verification
```bash
go test ./...
go vet ./...
go build -o ./.tmp/ql ./cmd/ql
```

## 범위 가드
Do not add:
- difficulty / variable XP
- natural-language dates (`tmr`, `tomorrow`)
- extra commands
- `service/` or `pkg/` layers for MVP1
- CI/lint/release automation

## Directory Guide
- `internal/cli/AGENT.md` — CLI commands and output rules
- `internal/domain/AGENT.md` — Domain model rules
- `internal/engine/AGENT.md` — XP/leveling business logic
- `internal/repository/AGENT.md` — SQLite persistence rules
- `specs/AGENT.md` — Spec maintenance rules

## References
- Source of Truth: `/.opencode/document/requirement.md`
- Quickstart: `/specs/001-questline-mvp1/quickstart.md`
- CLI Contracts: `/specs/001-questline-mvp1/contracts/cli-commands.md`
