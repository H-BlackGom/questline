# 구현 계획: Questline MVP 1

**Branch**: `001-questline-mvp1` | **Date**: 2025-03-20 | **Spec**: [.opencode/document/requirement.md](/Users/hyoungrolee/Documents/1_Project/01_dev/questline/.opencode/document/requirement.md)
**Input**: Questline MVP 1 요구사항 문서

---

## 요약

Questline MVP 1은 CLI 기반 RPG 퀘스트 관리 도구입니다. 사용자가 터미널에서 할 일을
추가하고 완료할 때마다 XP를 획득하여 레벨업하는 게이미피케이션 요소를 포함합니다.
Go 언어, Cobra CLI 프레임워크, pure-go SQLite, fatih/color 라이브러리를 사용합니다.

핵심 목표: "The Core Engine: 핵심 루프 검증" - 복잡한 UI 없이 퀘스트 완료 시
레벨업하는 경험이 실제로 유효한지 빠르게 테스트합니다.

---

## 기술 맥락

**언어/버전**: Go 1.26.1

**주요 의존성**:
- [spf13/cobra](https://github.com/spf13/cobra) - CLI 프레임워크
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) - pure-go SQLite (CGO-free)
- [fatih/color](https://github.com/fatih/color) - 터미널 색상 출력
- [stretchr/testify](https://github.com/stretchr/testify) - 테스트 어서션
- [google/uuid](https://github.com/google/uuid) - UUID 생성

**저장소**: SQLite (`~/.questline/data.db`)

**테스트**: `go test` + testify, SQLite In-Memory 모드

**타겟 플랫폼**: macOS, Linux, Windows (단일 바이너리 배포)

**프로젝트 유형**: CLI 애플리케이션

**성능 목표**: 명령어 실행 < 100ms, 메모리 < 50MB

**제약사항**:
- 단일 사용자 (로컬 파일 기반)
- 오프라인 사용 필수
- 순수 Go로 작성 (CGO 사용 금지)

**규모/범위**:
- 4개 CLI 명령어 (add, done, ls, me)
- 2개 엔티티 (Quest, Player)
- MVP 1: 핵심 루프만 구현 (TUI, Sub Quest 등 연기)

---

## 헌법 검증

*GATE: Phase 0 연구 전에 통과해야 함. Phase 1 설계 후 재검증.*

**Questline 원칙 검증 체크리스트**:

- [x] **I. 계층형 아키텍처**: cmd/에는 CLI 명령어만, 비즈니스 로직은 internal/로 분리
  - 설계 확인: `internal/cli/` (Cobra 명령어), `internal/engine/` (레벨업 로직), `internal/repository/` (데이터 접근)
- [x] **II. CLI 명령어 설계**: `ql <command>` 형태, 직관적인 명령어 구조
  - 설계 확인: `ql add`, `ql done`, `ql ls`, `ql me`
- [x] **III. 데이터 지속성**: pure-go SQLite 사용, `~/.questline/data.db` 경로
  - 설계 확인: modernc.org/sqlite 사용, ~/.questline/ 경로
- [x] **IV. RPG 게이미피케이션**: XP/레벨업 시스템, 색상 피드백 고려
  - 설계 확인: XP=50/퀘스트, 레벨업 공식: 100 + (level * 50), fatih/color 사용
- [x] **V. AI 페어 프로그래밍**: 복잡한 설계는 Gemini CLI, 반복 작업은 OpenCode
  - 설계 확인: 복잡한 알고리즘(Gemini), 보일러플레이트(OpenCode)
- [x] **VI. 코드 품질**: gofmt, 테스트 커버리지 80%+, 명시적 에러 처리
  - 설계 확인: Table-driven tests, 테스트 커버리지 목표 80%

**검증 결과**: ✅ 모든 원칙 준수 확인됨

---

## 프로젝트 구조

### 문서 (이 피처)

```text
specs/001-questline-mvp1/
├── plan.md              # 이 파일 (구현 계획)
├── research.md          # Phase 0 출력 (기술 스택 조사)
├── data-model.md        # Phase 1 출력 (데이터 모델)
├── quickstart.md        # Phase 1 출력 (퀵스타트 가이드)
├── contracts/           # Phase 1 출력 (CLI 계약서)
│   └── cli-commands.md
└── tasks.md             # Phase 2 출력 (작업 목록)
```

### 소스 코드 (저장소 루트)

```text
questline/
├── cmd/
│   └── ql/
│       └── main.go          # CLI 진입점, Cobra root 설정
├── internal/
│   ├── cli/                 # Cobra 명령어 구현 (cmd/로 위임)
│   │   ├── root.go          # Root command, 플래그 바인딩
│   │   ├── add.go           # ql add
│   │   ├── done.go          # ql done
│   │   ├── ls.go            # ql ls
│   │   └── me.go            # ql me
│   ├── domain/              # 도메인 모델
│   │   ├── quest.go         # Quest 구조체, 상태/난이도 상수
│   │   └── player.go        # Player 구조체, 칭호 메서드
│   ├── engine/              # 비즈니스 로직
│   │   ├── leveling.go      # XP/레벨업 계산
│   │   └── leveling_test.go # 테이블 기반 테스트
│   ├── repository/          # 데이터 접근 계층
│   │   ├── sqlite.go        # DB 연결, 마이그레이션
│   │   ├── quest_repo.go    # Quest CRUD
│   │   ├── quest_repo_test.go
│   │   ├── player_repo.go   # Player CRUD
│   │   └── player_repo_test.go
│   └── service/             # 유스케이스 (MVP2에서 구현 예정)
│       ├── quest_service.go
│       └── player_service.go
├── pkg/                     # 재사용 가능한 유틸리티 (MVP2에서 구현 예정)
│   ├── color/               # fatih/color 래퍼
│   │   └── color.go
│   └── formatter/           # 출력 포맷터
│       └── table.go         # tabwriter 래퍼
├── go.mod
├── go.sum
└── questline                # 빌드된 바이너리
```

**구조 결정**: Go 표준 프로젝트 레이아웃 + 헌법 I.계층형 아키텍처 준수
- `cmd/`: 진입점만
- `internal/`: 비즈니스 로직 (domain, engine, repository, service)
- `pkg/`: 재사용 가능한 유틸리티

---

## 태스크별 브랜치 전략

요구사항에 정의된 4개 태스크를 순차적으로 구현합니다.

### Task 1: 코어 비즈니스 로직 (레벨링 엔진)
**Branch**: `feature/core-engine`
**목표**: DB 의존성 없이 순수하게 동작하는 경험치 및 레벨업 계산기 구현

구현 내용:
- `internal/engine/leveling.go`: CalculateLevelUp 함수
- 초과 XP 누적 및 연속 레벨업 재귀 처리
- 레벨 구간별 칭호 반환 함수
- 테이블 기반 단위 테스트

### Task 2: 데이터 접근 계층 (Repository)
**Branch**: `feature/data-repository`
**목표**: 데이터베이스 CRUD 로직 및 트랜잭션 구현

구현 내용:
- `internal/repository/sqlite.go`: DB 초기화 및 마이그레이션
- `internal/repository/quest_repo.go`: Quest CRUD
- `internal/repository/player_repo.go`: Player CRUD
- 동적 쿼리 (상태별 목록 필터링)
- `MarkAsDone` + XP 증가 트랜잭션
- In-Memory DB 테스트

### Task 3: CLI 뼈대 구축 및 라우팅
**Branch**: `feature/cli-scaffold`
**목표**: Cobra 연동 및 명령어 껍데기, 플래그 파싱 구현

구현 내용:
- `cmd/ql/main.go`: Root command 설정
- `internal/cli/*.go`: add, done, ls, me 명령어 등록
- `-d` 마감일 문자열 파싱 (ISO 8601)
- `ls` 플래그 바인딩 (-a, -d)

### Task 4: 로직 통합 및 출력 포맷팅
**Branch**: `feature/integration`
**목표**: CLI와 내부 로직 결합 및 터미널 UX 개선

구현 내용:
- `fatih/color` 적용 (레벨업, 성공/에러 메시지)
- `text/tabwriter` 적용 (ql ls 테이블 출력)
- ASCII 아트 또는 색상으로 시각적 피드백
- 전체 흐름 통합 테스트

---

## 생성된 산출물

이 계획 명령어는 다음 문서들을 생성했습니다:

1. ✅ `specs/001-questline-mvp1/research.md` - Phase 0: 기술 스택 조사
2. ✅ `specs/001-questline-mvp1/data-model.md` - Phase 1: 데이터 모델
3. ✅ `specs/001-questline-mvp1/quickstart.md` - Phase 1: 퀙스타트 가이드
4. ✅ `specs/001-questline-mvp1/contracts/cli-commands.md` - Phase 1: CLI 계약서

---

## 다음 단계

다음 명령어로 작업 목록을 생성하세요:

```bash
/speckit.tasks @specs/001-questline-mvp1
```

또는 직접 `specs/001-questline-mvp1/tasks.md`를 작성하여 태스크별 구현을 시작하세요.

---

**Version**: 1.0.0 | **Generated**: 2025-03-20
