<!--
================================================================================
SYNC IMPACT REPORT
================================================================================
Version Change: 0.0.0 → 1.0.0 (Initial ratification)
Modified Principles: N/A (New constitution)
Added Sections:
  - I. 계층형 아키텍처 (Layered Architecture)
  - II. CLI 명령어 설계 (CLI Command Design)
  - III. 데이터 지속성 (Data Persistence)
  - IV. RPG 게이미피케이션 (RPG Gamification)
  - V. AI 페어 프로그래밍 (AI Pair Programming)
  - VI. 코드 품질 및 테스트 (Code Quality & Testing)
Removed Sections: N/A
Templates Requiring Updates:
  - ✅ .specify/templates/plan-template.md (Constitution Check 섹션 참조)
  - ✅ .specify/templates/spec-template.md (요구사항 정의 시 원칙 준수)
  - ✅ .specify/templates/tasks-template.md (작업 분리 원칙 적용)
Follow-up TODOs: None
================================================================================
-->

# Questline Constitution (퀘스트라인 헌법)

> **버전**: 1.0.0 | **최초 비준**: 2025-03-20 | **최종 개정**: 2025-03-20

## 서문

본 헌법은 Questline MVP 1 프로젝트의 모든 개발 활동에 적용되는 불가침의 원칙을 정의한다.
Questline은 터미널에서 할 일을 RPG 퀘스트처럼 관리하고 완료 시 경험치(XP)를 획득하여
레벨업하는 CLI 애플리케이션이다. 본 프로젝트는 Go 언어, Cobra CLI 프레임워크,
pure-go SQLite, fatih/color 라이브러리를 사용하여 개발된다.

---

## 핵심 원칙 (Core Principles)

### I. 계층형 아키텍처 (Layered Architecture)

**불가침 규칙**:
- **cmd/**: Cobra CLI 명령어만 정의. 비즈니스 로직 포함 금지.
  - 플래그 파싱, 입력 검증, 출력 포맷팅만 담당
  - 모든 비즈니스 로직은 internal/ 패키지로 위임
- **internal/**: 비즈니스 로직 및 도메인 모델
  - **domain/**: 순수 Go 구조체 (Quest, Player, Status 등)
  - **service/**: 유스케이스 및 비즈니스 규칙 (QuestService, PlayerService)
  - **repository/**: 데이터 접근 계층 (SQLite 구현체)
- **pkg/**: 재사용 가능한 유틸리티 (color 출력, XP 계산 등)

**의거**: 관심사 분리(Separation of Concerns)를 통해 테스트 용이성과 유지보수성을 확보한다.
CLI 프레임워크 변경 시에도 비즈니스 로직은 그대로 유지될 수 있어야 한다.

---

### II. CLI 명령어 설계 (CLI Command Design)

**불가침 규칙**:
- 모든 명령어는 `questline <command>` 형태로 일관되게 설계
- **add**: `questline add "퀘스트 제목" [--difficulty easy|normal|hard]`
  - 퀘스트 생성 시 고유 ID 자동 생성 (UUID)
  - 난이도별 XP 보상 다르게 설정 (easy: 30, normal: 50, hard: 100)
- **done**: `questline done <quest-id>`
  - 퀘스트 완료 시 XP 자동 지급
  - 완료된 퀘스트는 상태 변경 및 완료 시간 기록
- **ls**: `questline ls [--status pending|completed|all]`
  - 기본값: pending 퀘스트만 표시
  - 색상으로 상태 구분 (fatih/color 사용)
- **me**: `questline me`
  - 현재 레벨, 총 XP, 다음 레벨까지 필요 XP, 완료한 퀘스트 수 표시
  - ASCII 아트 또는 색상으로 시각적 피드백 제공

**의거**: 직관적인 명령어 구조는 사용자 경험의 핵심이다. Unix 철학을 따르는
단순하고 조합 가능한 명령어 설계를 지향한다.

---

### III. 데이터 지속성 (Data Persistence)

**불가침 규칙**:
- 데이터 저장 위치: `~/.questline/data.db` (SQLite)
- **pure-go SQLite** 사용 (mattn/go-sqlite3 의 CGO-free 대안 또는 유사)
- 데이터베이스 스키마 버전 관리 필수 (마이그레이션 지원)
- 초기화 시 `~/.questline/` 디렉토리 자동 생성
- 백업 및 복구 메커니즘 고려 (향후 확장)

**스키마 설계 원칙**:
```sql
-- quests 테이블
CREATE TABLE quests (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    difficulty TEXT DEFAULT 'normal',
    status TEXT DEFAULT 'pending',
    xp_reward INTEGER DEFAULT 50,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

-- player_stats 테이블
CREATE TABLE player_stats (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    level INTEGER DEFAULT 1,
    total_xp INTEGER DEFAULT 0,
    quests_completed INTEGER DEFAULT 0
);
```

**의거**: 로컬 파일 기반 저장은 오프라인 사용과 개인정보 보호를 보장한다.
pure-go 구현은 크로스 컴파일과 단일 바이너리 배포를 가능하게 한다.

---

### IV. RPG 게이미피케이션 (RPG Gamification)

**불가침 규칙**:
- **XP 시스템**: 퀘스트 완료 시 XP 획득 (기본 50 XP)
- **레벨업 공식**: `required_xp = level * 100`
  - 예: 레벨 1→2: 100 XP 필요, 레벨 2→3: 200 XP 필요
- **레벨업 시**: 콘솔에 축하 메시지 및 색상 출력 (fatih/color)
- **시각적 피드백**:
  - `me` 명령어: 진행 바(progress bar)로 다음 레벨까지 진행도 표시
  - `ls` 명령어: 완료된 퀘스트는 초록색, 미완료는 노란색으로 표시

**의거**: 게이미피케이션은 할 일 관리의 지속 가능한 동기부여를 제공한다.
즉각적인 피드백 루프는 사용자 참여도를 높인다.

---

### V. AI 페어 프로그래밍 (AI Pair Programming)

**불가침 규칙**:
- **Gemini CLI 사용 규칙**:
  - 복잡한 알고리즘 설계나 아키텍처 결정 시 Gemini CLI 활용
  - 코드 리뷰 및 리팩토링 제안 요청
  - 명령: `gemini` 또는 `gcloud alpha code-tools gemini`
- **OpenCode 사용 규칙**:
  - 반복적이고 기계적인 작업 (보일러플레이트, 테스트 케이스 생성)
  - 문서화 및 주석 작성
  - 명령: `opencode` (현재 사용 중인 도구)
- **AI 협업 프로토콜**:
  1. AI에게 작업 설명 시 "왜(Why)"를 먼저 설명
  2. AI가 생성한 코드는 반드시 인간이 검토 후 커밋
  3. AI 제안에 대한 피드백 제공 (학습 개선)
  4. 민감한 정보(API 키, 개인정보)는 AI에 노출 금지

**의거**: AI는 생산성 향상 도구이지 인간 개발자의 대체재가 아니다.
의사결정과 최종 책임은 인간 개발자에게 있다.

---

### VI. 코드 품질 및 테스트 (Code Quality & Testing)

**불가침 규칙**:
- **Go 표준 준수**: `gofmt`, `go vet`, `golint` 통과 필수
- **테스트 커버리지**: 핵심 비즈니스 로직 80% 이상
  - `*_test.go` 파일은 테스트 대상 파일과 동일 패키지에 위치
  - 테이블 기반 테스트(Table-driven tests) 사용 권장
- **에러 처리**: 모든 에러는 명시적으로 처리, `log.Fatal`은 main에서만 사용
- **의존성 관리**: `go mod tidy`로 불필요한 의존성 제거
- **빌드**: `go build -o questline`로 단일 바이너리 생성

**의거**: 테스트는 회귀 방지와 리팩토링 안전성을 보장한다.
Go의 철학인 "명시적이고 간결한 코드"를 따른다.

---

## 거버넌스 (Governance)

### 헌법 개정 절차

1. **제안**: `.specify/memory/constitution.md` 수정 제안
2. **검토**: 모든 원칙 변경은 프로젝트 영향도 분석 필수
3. **비준**: 개발자 1인 프로젝트이므로 자체 판단 후 적용
4. **기록**: 개정 내용은 Sync Impact Report에 문서화

### 버전 관리 정책

- **MAJOR**: 아키텍처 원칙 변경, 기술 스택 교체
- **MINOR**: 새로운 원칙 추가, 기존 원칙 확장
- **PATCH**: 문구 명확화, 오타 수정, 예시 추가

### 준수 검증

- 모든 PR은 헌법 원칙 준수 여부 자체 검토
- 복잡도 증가 시 정당성 문서화
- `.specify/templates/` 내 템플릿은 헌법 원칙과 동기화 유지

---

## 부록: 프로젝트 구조 예시

```
questline/
├── cmd/
│   ├── root.go          # Cobra root command
│   ├── add.go           # questline add
│   ├── done.go          # questline done
│   ├── ls.go            # questline ls
│   └── me.go            # questline me
├── internal/
│   ├── domain/
│   │   ├── quest.go     # Quest struct
│   │   └── player.go    # Player struct
│   ├── service/
│   │   ├── quest_service.go
│   │   └── player_service.go
│   └── repository/
│       ├── sqlite.go    # DB connection
│       ├── quest_repo.go
│       └── player_repo.go
├── pkg/
│   ├── color/           # fatih/color wrapper
│   └── xp/              # XP calculation utilities
├── main.go
├── go.mod
└── go.sum
```

---

**Version**: 1.0.0 | **Ratified**: 2025-03-20 | **Last Amended**: 2025-03-20
