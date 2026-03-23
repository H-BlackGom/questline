<!--
================================================================================
3: SYNC IMPACT REPORT
4: ================================================================================
5: Version Change: 1.0.0 → 1.0.1 (MVP1 Policy Alignment)
6: Modified Principles: I (Architecture), II (CLI), III (Data), IV (Gamification)
7: Added Sections:
8:   - MVP1 Scope Guard (Section II)
9: Removed Sections: N/A
10: Templates Requiring Updates: N/A
11: Follow-up TODOs: None
12: ================================================================================
13: -->
14: 
15: # Questline Constitution (퀘스트라인 헌법)
16: 
17: > **버전**: 1.0.1 | **최초 비준**: 2025-03-20 | **최종 개정**: 2026-03-23
18: 
19: ## 서문
20: 
21: 본 헌법은 Questline MVP 1 프로젝트의 모든 개발 활동에 적용되는 불가침의 원칙을 정의한다.
22: Questline은 터미널에서 할 일을 RPG 퀘스트처럼 관리하고 완료 시 경험치(XP)를 획득하여
23: 레벨업하는 CLI 애플리케이션이다. 본 프로젝트는 Go 언어, Cobra CLI 프레임워크,
24: pure-go SQLite, fatih/color 라이브러리를 사용하여 개발된다.
25: 
26: ---
27: 
28: ## 핵심 원칙 (Core Principles)
29: 
30: ### I. 계층형 아키텍처 (Layered Architecture)
31: 
32: **불가침 규칙**:
33: - **cmd/**: Cobra CLI 명령어 진입점 (`cmd/ql/main.go`).
34: - **internal/**: 비즈니스 로직 및 도메인 모델
35:   - **domain/**: 순수 Go 구조체 (Quest, Player, Status 등)
36:   - **cli/**: Cobra 명령어 구현 (add, done, ls, me)
37:   - **engine/**: 비즈니스 규칙 (XP/레벨링 계산)
38:   - **repository/**: 데이터 접근 계층 (SQLite 구현체)
39: 
40: **의거**: 관심사 분리(Separation of Concerns)를 통해 테스트 용이성과 유지보수성을 확보한다.
41: CLI 프레임워크 변경 시에도 비즈니스 로직은 그대로 유지될 수 있어야 한다.
42: 
43: ---
44: 
45: ### II. CLI 명령어 설계 (CLI Command Design)
46: 
47: **불가침 규칙**:
48: - 모든 명령어는 `ql <command>` 형태로 일관되게 설계
49: - **add**: `ql add "퀘스트 제목" [-d YYYY-MM-DD]`
50:   - 퀘스트 생성 시 고유 ID 자동 생성 (UUID v4 앞 8자)
51:   - 고정 XP 보상: 퀘스트 완료 시 항상 50 XP
52:   - 날짜 형식: `YYYY-MM-DD`만 허용
53: - **done**: `ql done <quest-id>`
54:   - 퀘스트 완료 시 XP 자동 지급 (고정 50 XP)
55:   - 완료된 퀘스트는 상태 변경 및 완료 시간 기록
56: - **ls**: `ql ls [--done|--all]`
57:   - 기본값: TODO 퀘스트만 표시
58:   - 색상으로 상태 구분
59: - **me**: `ql me`
60:   - 현재 레벨, 총 XP, 다음 레벨까지 필요 XP, 완료한 퀘스트 수 표시
61:   - ASCII 아트 또는 색상으로 시각적 피드백 제공
62: 
63: **MVP1 범위 가드 (Scope Guard)**:
64: - 난이도 시스템 (Easy/Normal/Hard) 배제
65: - 가변 XP 보상 배제 (항상 50 XP)
66: - 자연어 날짜 파싱 배제 (`YYYY-MM-DD` 고정)
67: 
68: **의거**: 직관적인 명령어 구조는 사용자 경험의 핵심이다. Unix 철학을 따르는
69: 단순하고 조합 가능한 명령어 설계를 지향한다.
70: 
71: ---
72: 
73: ### III. 데이터 지속성 (Data Persistence)
74: 
75: **불가침 규칙**:
76: - 데이터 저장 위치: `~/.questline/data.db` (SQLite)
77: - **pure-go SQLite** 사용 (mattn/go-sqlite3 의 CGO-free 대안 또는 유사)
78: - 데이터베이스 스키마 버전 관리 필수 (마이그레이션 지원)
79: - 초기화 시 `~/.questline/` 디렉토리 자동 생성
80: - 백업 및 복구 메커니즘 고려 (향후 확장)
81: 
82: **스키마 설계 원칙**:
83: ```sql
84: -- quests 테이블
85: CREATE TABLE quests (
86:     id TEXT PRIMARY KEY,
87:     title TEXT NOT NULL,
88:     status TEXT DEFAULT 'pending',
89:     created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
90:     completed_at DATETIME,
91:     due_date TEXT
92: );
93: 
94: -- player_stats 테이블
95: CREATE TABLE player_stats (
96:     id INTEGER PRIMARY KEY CHECK (id = 1),
97:     level INTEGER DEFAULT 1,
98:     total_xp INTEGER DEFAULT 0,
99:     quests_completed INTEGER DEFAULT 0
100: );
101: ```
102: 
103: **의거**: 로컬 파일 기반 저장은 오프라인 사용과 개인정보 보호를 보장한다.
104: pure-go 구현은 크로스 컴파일과 단일 바이너리 배포를 가능하게 한다.
105: 
106: ---
107: 
108: ### IV. RPG 게이미피케이션 (RPG Gamification)
109: 
110: **불가침 규칙**:
111: - **XP 시스템**: 퀘스트 완료 시 XP 획득 (기본 50 XP)
112: - **레벨업 공식**: `required_xp = 100 + (level * 50)`
113:   - 예: 레벨 1→2: 150 XP 필요, 레벨 2→3: 200 XP 필요
114: - **레벨업 시**: 콘솔에 축하 메시지 및 색상 출력 (fatih/color)
115: - **시각적 피드백**:
116:   - `me` 명령어: 진행 바(progress bar)로 다음 레벨까지 진행도 표시
117:   - `ls` 명령어: 완료된 퀘스트는 초록색, 미완료는 노란색으로 표시
118: 
119: **의거**: 게이미피케이션은 할 일 관리의 지속 가능한 동기부여를 제공한다.
120: 즉각적인 피드백 루프는 사용자 참여도를 높인다.

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
│   └── ql/
│       └── main.go      # CLI entry point
├── internal/
│   ├── cli/             # CLI commands (add, done, ls, me)
│   ├── domain/
│   │   ├── quest.go     # Quest struct
│   │   └── player.go    # Player struct
│   ├── engine/          # XP/leveling logic
│   └── repository/
│       ├── sqlite.go    # DB connection
│       ├── quest_repo.go
│       └── player_repo.go
├── go.mod
├── go.sum
└── README.md
```

---

**Version**: 1.0.1 | **Ratified**: 2025-03-20 | **Last Amended**: 2026-03-23
