<!--
================================================================================
SYNC IMPACT REPORT
================================================================================
Version Change: 1.0.1 → 2.0.0 (MVP2 Major Architecture Update)
Modified Principles:
  - I (Layered Architecture): TUI 레이어 추가, Bubble Tea 아키텍처 통합
  - II (CLI Command Design): `ql check` 명령어 추가, 퀘스트 타입별 옵션 확장
  - III (Data Persistence): 퀘스트 타입, Flow 기록, 루틴 히스토리 스키마 추가
  - IV (RPG Gamification): Flow(몰입도) 시스템 추가
Added Sections:
  - VII. TUI 아키텍처 (Bubble Tea Elm 아키텍처)
  - VIII. 퀘스트 타입 시스템 (Daily/Weekly/Epic/Guild/Sub)
  - IX. 지연 평가와 시간 계산 (Lazy Evaluation)
  - X. Flow(몰입도) 시스템
  - XI. AI 페어 프로그래밍 상세 규칙 (Gemini CLI + OpenCode)
Removed Sections: N/A
Templates Requiring Updates:
  - ✅ plan-template.md: Constitution Check 섹션 MVP2 원칙 반영
  - ✅ spec-template.md: 퀘스트 타입, Flow 시스템 언급 추가
  - ✅ tasks-template.md: TUI, 퀘스트 타입 태그 추가
Follow-up TODOs: None
================================================================================
-->

# Questline Constitution (퀘스트라인 헌법)

> **버전**: 2.0.0 | **최초 비준**: 2025-03-20 | **최종 개정**: 2026-04-02

## 서문

본 헌법은 Questline MVP 2 프로젝트의 모든 개발 활동에 적용되는 불가침의 원칙을 정의한다.
Questline은 터미널에서 할 일을 RPG 퀘스트처럼 관리하고 완료 시 경험치(XP)를 획득하여
레벨업하는 TUI/CLI 하이브리드 애플리케이션이다. MVP 2에서는 기존 CLI를 넘어
Bubble Tea 기반의 마스터-디테일 TUI 대시보드를 도입하며, 지연 평가(Lazy Evaluation)를
통한 시간 흐름 계산과 Flow(몰입도) 기반 XP 배율 시스템을 구현한다.

본 프로젝트는 Go 언어, Cobra CLI 프레임워크, Bubble Tea TUI 프레임워크,
pure-go SQLite를 사용하여 개발된다.

---

## 핵심 원칙 (Core Principles)

### I. 계층형 아키텍처 (Layered Architecture)

**불가침 규칙**:
- **cmd/**: Cobra CLI 명령어 진입점 (`cmd/ql/main.go`).
  - TUI 모드 진입: `ql check`는 Bubble Tea 프로그램 실행
  - CLI 모드: 기존 명령어들은 직접 repository 호출
- **internal/**: 비즈니스 로직 및 도메인 모델
  - **domain/**: 순수 Go 구조체 (Quest, Player, Status, QuestType, Flow 등)
  - **cli/**: Cobra 명령어 구현 (add, done, ls, me, check)
  - **tui/**: Bubble Tea TUI 구현 (Model, Update, View 패턴)
    - **models/**: TUI 상태 관리 구조체
    - **views/**: 화면 렌더링 컴포넌트
    - **commands/**: TUI 내 사용자 액션 처리
  - **engine/**: 비즈니스 규칙 (XP/레벨링/Flow 계산, 시간 흐름 계산)
  - **repository/**: 데이터 접근 계층 (SQLite 구현체)
  - **service/**: 고수준 비즈니스 로직 (퀘스트 완료 처리, 상태 전환)

**의거**: 관심사 분리(Separation of Concerns)를 통해 테스트 용이성과 유지보수성을 확보한다.
CLI/TUI 프레임워크 변경 시에도 비즈니스 로직은 그대로 유지될 수 있어야 한다.
TUI는 오직 UI 상태 관리와 사용자 입력 처리에만 집중하고, 모든 비즈니스 결정은
engine과 service 계층에서 이루어진다.

---

### II. CLI 명령어 설계 (CLI Command Design)

**불가침 규칙**:
- 모든 명령어는 `ql <command>` 형태로 일관되게 설계
- **check**: `ql check`
  - TUI 대시보드 실행 (Bubble Tea 기반)
  - 좌우 패널 마스터-디테일 구조: 좌측 퀘스트 목록, 우측 상세 정보
  - Vim-like 단축키 지원 (j/k 이동, Enter 선택, q 종료)
  - 앱 실행 시 Lazy Evaluation 트리거
- **add**: `ql add "퀘스트 제목" [-t TYPE] [-d YYYY-MM-DD] [-p PARENT_ID]`
  - 퀘스트 생성 시 고유 ID 자동 생성 (UUID v4 앞 8자)
  - **TYPE**: `daily` | `weekly` | `epic` | `guild` | `sub`
  - **`-p PARENT_ID`**: 하위 퀘스트(Sub) 생성 시 부모 지정 (2-Depth 제한)
  - XP 보상: 기본 50 XP × Flow 배율
- **done**: `ql done <quest-id>`
  - 퀘스트 완료 시 XP 자동 지급 (Flow 배율 적용)
  - 하위 퀘스트(Sub) 완료 시: 부모 퀘스트 상태를 PENDING으로 전환
  - 에픽/길드 완료 시: 연결된 모든 하위 퀘스트 자동 완료 (설정 가능)
- **ls**: `ql ls [--done|--all|--type TYPE]`
  - 기본값: TODO 퀘스트만 표시
  - `--type`: 특정 타입 필터링
  - 색상으로 상태 및 타입 구분
- **me**: `ql me [--flow]`
  - 현재 레벨, 총 XP, Flow 상태, 연속 달성 일수 표시
  - `--flow`: 상세 Flow 히스토리 출력

**MVP2 범위 가드 (Scope Guard)**:
- 3-Depth 이상의 중첩 퀘스트 배제 (최대 2-Depth)
- 실시간 백그라운드 데몬 배제 (Lazy Evaluation만 사용)
- 웹 동기화/클라우드 저장 배제 (로컬 SQLite만 사용)
- 멀티플레이어/길드 간 상호작용 배제 (길드는 개인 카테고리로만 사용)

**의거**: 직관적인 명령어 구조는 사용자 경험의 핵심이다. Unix 철학을 따르는
단순하고 조합 가능한 명령어 설계를 지향한다. TUI는 CLI의 시각적 확장이며,
동일한 비즈니스 로직을 공유한다.

---

### III. 데이터 지속성 (Data Persistence)

**불가침 규칙**:
- 데이터 저장 위치: `~/.questline/data.db` (SQLite)
- **pure-go SQLite** 사용 (modernc.org/sqlite 또는 유사)
- 데이터베이스 스키마 버전 관리 필수 (마이그레이션 지원)
- 초기화 시 `~/.questline/` 디렉토리 자동 생성
- 백업: 주기적 자동 백업 또는 수동 백업 명령 제공

**스키마 설계 원칙**:
```sql
-- quests 테이블 (MVP2 확장)
CREATE TABLE quests (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('daily', 'weekly', 'epic', 'guild', 'sub')),
    status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'in_progress', 'pending_completion', 'completed', 'archived')),
    parent_id TEXT REFERENCES quests(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    due_date TEXT,
    scheduled_date TEXT,  -- 루틴/주간: 예정된 실행 날짜
    recurrence TEXT       -- 반복 패턴 (daily, weekly)
);

-- player_stats 테이블 (Flow 시스템 추가)
CREATE TABLE player_stats (
    id INTEGER PRIMARY KEY CHECK(id = 1),
    level INTEGER DEFAULT 1,
    total_xp INTEGER DEFAULT 0,
    quests_completed INTEGER DEFAULT 0,
    current_flow REAL DEFAULT 1.0,  -- 현재 Flow 배율 (0.5 ~ 2.0)
    streak_days INTEGER DEFAULT 0,  -- 연속 달성 일수
    last_activity_date TEXT         -- 마지막 활동일 (YYYY-MM-DD)
);

-- flow_history 테이블 (Flow 변동 기록)
CREATE TABLE flow_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    previous_flow REAL NOT NULL,
    new_flow REAL NOT NULL,
    routine_completion_rate REAL,  -- 루틴 달성률 (%)
    reason TEXT                      -- 'daily_evaluation', 'weekly_bonus', etc.
);

-- routine_completions 테이블 (루틴 완료 기록)
CREATE TABLE routine_completions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    total_routines INTEGER NOT NULL,
    completed_routines INTEGER NOT NULL,
    completion_rate REAL NOT NULL
);

-- quest_evaluations 테이블 (지연 평가 기록)
CREATE TABLE quest_evaluations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    evaluated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    reference_time TEXT NOT NULL,    -- 기준 시점 (새벽 4시 기준)
    days_processed INTEGER NOT NULL, -- 처리된 일수
    quests_archived INTEGER,         -- 보관된 퀘스트 수
    flow_adjustments TEXT            -- Flow 조정 JSON
);
```

**의거**: 로컬 파일 기반 저장은 오프라인 사용과 개인정보 보호를 보장한다.
pure-go 구현은 크로스 컴파일과 단일 바이너리 배포를 가능하게 한다.
퀘스트 타입별 이력과 Flow 변화 추적은 게이미피케이션의 지속 가능한 동기부여를 제공한다.

---

### IV. RPG 게이미피케이션 (RPG Gamification)

**불가침 규칙**:
- **XP 시스템**: 퀘스트 완료 시 XP 획득 (기본 50 XP × Flow 배율)
- **Flow 배율**: 0.5x ~ 2.0x (어제 루틴 달성률 기반)
  - 0~39%: 0.5x (낮은 집중)
  - 40~69%: 1.0x (기본)
  - 70~89%: 1.5x (좋은 흐름)
  - 90~100%: 2.0x (몰입 상태)
- **레벨업 공식**: `required_xp = 100 + (level * 50)`
  - 예: 레벨 1→2: 150 XP 필요, 레벨 2→3: 200 XP 필요
- **레벨업 시**: 콘솔에 축하 메시지 및 색상 출력
- **칭호 시스템**: 레벨 구간별 칭호 제공 (Intern, Junior, Senior, Lead, Principal, Guru)
- **시각적 피드백**:
  - `me` 명령어: 진행 바, Flow 상태, 연속 달성 일수 표시
  - TUI: 퀘스트 타입별 아이콘, Flow 배율 실시간 표시

**의거**: 게이미피케이션은 할 일 관리의 지속 가능한 동기부여를 제공한다.
Flow 시스템은 어제의 성과를 오늘의 동기로 전환하여 긍정적 피드백 루프를 만든다.
즉각적인 피드백과 시각적 보상은 사용자 참여도를 높인다.

---

### V. 코드 품질 및 테스트 (Code Quality & Testing)

**불가침 규칙**:
- **Go 표준 준수**: `gofmt`, `go vet`, `golint` 통과 필수
- **테스트 커버리지**: 핵심 비즈니스 로직 80% 이상
  - `*_test.go` 파일은 테스트 대상 파일과 동일 패키지에 위치
  - 테이블 기반 테스트(Table-driven tests) 사용
  - Bubble Tea TUI 테스트: `tea.Program` 모킹 또는 상태 기반 테스트
- **에러 처리**: 모든 에러는 명시적으로 처리, `log.Fatal`은 main에서만 사용
- **의존성 관리**: `go mod tidy`로 불필요한 의존성 제거
- **빌드**: `go build -o ql`로 단일 바이너리 생성
- **TUI 테스트**: Update 함수의 순수성 보장 (부작용 없는 상태 변환)

**의거**: 테스트는 회귀 방지와 리팩토링 안전성을 보장한다.
Go의 철학인 "명시적이고 간결한 코드"를 따른다.
TUI의 복잡한 상태 관리는 테스트를 통해 예측 가능하게 유지한다.

---

### VI. AI 페어 프로그래밍 (AI Pair Programming)

**불가침 규칙**:
- **Gemini CLI 사용 규칙**:
  - 복잡한 알고리즘 설계나 아키텍처 결정 시 Gemini CLI 활용
  - 코드 리뷰 및 리팩토링 제안 요청
  - Bubble Tea의 Msg/Update 패턴 설계 컨설팅
  - 명령: `gemini` 또는 `gcloud alpha code-tools gemini`
- **OpenCode 사용 규칙**:
  - 반복적이고 기계적인 작업 (보일러플레이트, 테스트 케이스 생성)
  - 문서화 및 주석 작성, 헌법/스펙 업데이트
  - SQLite 스키마 마이그레이션 코드 생성
  - 명령: `opencode` (현재 사용 중인 도구)
- **AI 협업 프로토콜**:
  1. AI에게 작업 설명 시 "왜(Why)"를 먼저 설명
  2. AI가 생성한 코드는 반드시 인간이 검토 후 커밋
  3. AI 제안에 대한 피드백 제공 (학습 개선)
  4. 민감한 정보(API 키, 개인정보)는 AI에 노출 금지
  5. Bubble Tea Msg 타입 설계는 인간이 결정, AI는 구현만 지원

**의거**: AI는 생산성 향상 도구이지 인간 개발자의 대체재가 아니다.
의사결정과 최종 책임은 인간 개발자에게 있다.
특히 TUI의 복잡한 상태 흐름과 메시지 타입 설계는 인간의 설계 의도가 필수적이다.

---

### VII. TUI 아키텍처 (TUI Architecture - Bubble Tea)

**불가침 규칙**:
- **Elm 아키텍처 준수**: Model → Update → View 순환
  - **Model**: 애플리케이션 상태를 담는 순수한 구조체
  - **Update**: Msg를 받아 새로운 Model을 반환하는 순수 함수
  - **View**: Model을 받아 문자열을 반환하는 순수 함수
- **Bubble Tea 구조**:
  ```go
  type model struct {
      quests      []domain.Quest      // 퀘스트 목록
      cursor      int                 // 선택된 인덱스
      selected    *domain.Quest       // 선택된 퀘스트
      panelFocus  PanelType           // 좌/우 패널 포커스
      width       int                 // 터미널 너비
      height      int                 // 터미널 높이
      err         error               // 에러 상태
  }
  ```
- **메시지 타입 설계**:
  - `Msg` 인터페이스 구현체로 모든 사용자 액션 표현
  - `tea.KeyMsg`: 키보드 입력 (j, k, Enter, q, 등)
  - `tea.WindowSizeMsg`: 터미널 크기 변경
  - 커스텀 Msg: `QuestSelectedMsg`, `QuestCompletedMsg`, `FlowUpdatedMsg`
- **커맨드(Command)**: 비동기 작업을 위한 `tea.Cmd`
  - DB 조회/저장은 Command로 처리
  - 시간 계산(Lazy Evaluation)은 Command로 트리거
- **lipgloss 스타일링**:
  - 테마 색상은 중앙집중식으로 관리 (`internal/tui/theme/`)
  - 퀘스트 타입별 색상 구분 (Daily: 파랑, Weekly: 보라, Epic: 주황, Guild: 초록, Sub: 회색)

**의거**: Bubble Tea의 Elm 아키텍처는 예측 가능한 상태 관리와 테스트 용이성을 제공한다.
순수 함수 기반 설계는 부작용을 제어하고, 메시지 중심 흐름은 복잡한 상호작용을
명시적으로 모델링할 수 있게 한다.

---

### VIII. 퀘스트 타입 시스템 (Quest Type System)

**불가침 규칙**:
- **타입 정의**:
  - **루틴(Daily)**: 매일 반복, 자정(새벽 4시 기준) 자동 보관
  - **주간(Weekly)**: 매주 반복, 월요일 자동 생성
  - **에픽(Epic)**: 큰 목표, 여러 하위 퀘스트 포함 가능
  - **길드(Guild)**: 카테고리/프로젝트 단위 그룹핑
  - **하위(Sub)**: 2-Depth 하위 작업, 부모 퀘스트에 종속
- **상태 전환**:
  - `pending` → `in_progress`: 첫 하위 퀘스트 시작
  - `in_progress` → `pending_completion`: 모든 하위 완료 (부모는 PENDING)
  - `pending_completion` → `completed`: 사용자 수동 확인
  - `completed` → `archived`: 평가 후 보관
- **하위 퀘스트 완료 로직**:
  - 모든 Sub 퀘스트가 completed 상태가 되면, 부모는 자동으로 pending_completion 상태로 전환
  - 부모는 수동으로 done 명령어 실행해야 completed로 최종 전환
  - 부모 완료 시: XP 계산에 Flow 배율 적용, 완료 시간 기록
- **루틴/주간 자동 생성**:
  - 앱 실행 시 Lazy Evaluation으로 지난 시간 계산
  - 새벽 4시 기준으로 과거 루틴 자동 보관 및 새 루틴 생성

**의거**: 타입별 다른 동작은 현실적인 작업 관리를 반영한다.
루틴은 습관 형성, 에픽은 큰 목표 달성, 하위는 세분화된 실행을 지원한다.
PENDING 상태는 "완료 검토"의 의식적인 순간을 제공하여 성취감을 극대화한다.

---

### IX. 지연 평가와 시간 계산 (Lazy Evaluation)

**불가침 규칙**:
- **백그라운드 데몬 금지**: 실시간 타이머나 데몬 프로세스 없음
- **실행 시점 계산**: `ql` 명령어 실행 시점에 과거 시간 흐름 계산
- **기준 시각**: 새벽 4:00 AM (04:00)를 하루의 시작으로 간주
- **계산 로직**:
  1. 마지막 실행 시점(`last_activity_date`) 조회
  2. 현재 시점과의 차이 계산 (새벽 4시 기준 일수)
  3. 누락된 일수만큼 루틴/주간 평가 및 Flow 계산
  4. 각 날짜별로:
     - 완료되지 않은 루틴 자동 보관
     - 새 루틴/주간 자동 생성 (설정된 경우)
     - Flow 배율 재계산 (루틴 달성률 기반)
- **거리 계산**:
  ```go
  func CalculateDaysSince(lastDate, currentDate time.Time) int {
      // 새벽 4시 기준으로 일수 계산
      lastCheckpoint := get4AMCheckpoint(lastDate)
      currentCheckpoint := get4AMCheckpoint(currentDate)
      return int(currentCheckpoint.Sub(lastCheckpoint).Hours() / 24)
  }
  ```
- **원자성**: 시간 계산과 상태 업데이트는 트랜잭션으로 처리

**의거**: 지연 평가는 백그라운드 프로세스 없이 시간 기반 로직을 구현하는
간결한 방법이다. 사용자가 앱을 실행하는 순간 모든 "누적된 시간"을 한 번에 계산하여
리소스를 절약하고 복잡성을 줄인다.

---

### X. Flow(몰입도) 시스템 (Flow System)

**불가침 규칙**:
- **Flow 정의**: 어제 루틴(Daily) 달성률에 기반한 오늘 XP 배율
- **일일 평가 시점**: 새벽 4시 기준으로 어제 루틱 평가
- **달성률 계산**: `completion_rate = completed_routines / total_routines × 100`
- **Flow 배율 테이블**:
  | 달성률 | Flow 배율 | 상태 |
  |--------|-----------|------|
  | 0-39%  | 0.5x      | 흐름 끊김 |
  | 40-69% | 1.0x      | 보통 |
  | 70-89% | 1.5x      | 흐름 진입 |
  | 90-100%| 2.0x      | 몰입 상태 |
- **연속 보너스**: 7일 연속 70%+ 달성 시 주간 보너스 XP
- **UI 표시**:
  - TUI 헤더에 현재 Flow 배율 표시
  - `ql me --flow`로 히스토리 그래프 출력
- **초기화**: 새로운 날짜 첫 실행 시 Flow 재계산 (Lazy Evaluation)

**의거**: Flow 시스템은 "어제의 나"가 "오늘의 나"에게 주는 선물이다.
높은 달성률로 얻은 배율은 긍정적 강화(positive reinforcement)가 되고,
낮은 달성률의 패널티는 다음 날의 동기부여가 된다. 이는 게이미피케이션의
핵심 메커니즘으로, 지속적인 참여를 유도한다.

---

### XI. AI 페어 프로그래밍 상세 규칙 (AI Pair Programming Guidelines)

**불가침 규칙**:
- **Gemini CLI 활용 시나리오**:
  - Bubble Tea 아키텍처 설계 및 Msg 타입 설계
  - 복잡한 상태 전환 로직 검토
  - TUI 성능 최적화 제안
  - SQL 쿼리 및 인덱스 설계
  - 명령: `gemini` - 상황 설명 후 "왜"에 집중한 질문
- **OpenCode 활용 시나리오**:
  - SQLite 스키마 및 마이그레이션 코드 생성
  - Bubble Tea View 함수 구현
  - 테이블 기반 테스트 케이스 작성
  - 문서화 및 README 업데이트
  - 명령: `opencode` - 구현 중심 작업 위임
- **작업 분배 원칙**:
  - 설계/아키텍처: 인간 주도, AI 자문 (Gemini)
  - 구현/코딩: 인간 검토, AI 생성 (OpenCode)
  - 테스트: 인간 명세, AI 생성 (OpenCode)
  - 문서화: AI 생성, 인간 승인 (OpenCode)
- **금지 사항**:
  - AI에게 보안 관련 결정 위임 금지
  - AI에게 데이터베이스 스키마 중요 변경 위임 금지 (검토 필수)
  - AI에게 사용자 경험(UX) 핵심 결정 위임 금지

**의거**: AI는 강력한 도구이지만, 설계 의도와 사용자 경험에 대한
책임은 인간 개발자에게 있다. 명확한 역할 분리는 생산성을 높이면서도
품질과 일관성을 보장한다.

---

## 거버넌스 (Governance)

### 헌법 개정 절차

1. **제안**: `.specify/memory/constitution.md` 수정 제안
2. **검토**: 모든 원칙 변경은 프로젝트 영향도 분석 필수
3. **비준**: 개발자 1인 프로젝트이므로 자체 판단 후 적용
4. **기록**: 개정 내용은 Sync Impact Report에 문서화

### 버전 관리 정책

- **MAJOR**: 아키텍처 원칙 변경, 기술 스택 교체 (예: TUI 프레임워크 변경)
- **MINOR**: 새로운 원칙 추가, 기존 원칙 확장 (예: 새 퀘스트 타입 추가)
- **PATCH**: 문구 명확화, 오타 수정, 예시 추가

### 준수 검증

- 모든 PR은 헌법 원칙 준수 여부 자체 검토
- 복잡도 증가 시 정당성 문서화
- `.specify/templates/` 내 템플릿은 헌법 원칙과 동기화 유지
- MVP2의 새로운 원칙(VII~X)은 `ql check` 구현 시 필수 검증

---

## 부록: 프로젝트 구조 예시 (MVP2)

```
questline/
├── cmd/
│   └── ql/
│       └── main.go              # CLI entry point
├── internal/
│   ├── cli/                     # Cobra commands
│   │   ├── root.go
│   │   ├── add.go
│   │   ├── done.go
│   │   ├── ls.go
│   │   ├── me.go
│   │   └── check.go             # TUI entry command
│   ├── tui/                     # Bubble Tea TUI
│   │   ├── model.go             # TUI state model
│   │   ├── update.go            # Update function (Msg handlers)
│   │   ├── view.go              # View function
│   │   ├── commands.go          # tea.Cmd factories
│   │   ├── theme/
│   │   │   └── styles.go        # lipgloss styles
│   │   └── views/
│   │       ├── quest_list.go    # 좌측 패널: 퀘스트 목록
│   │       ├── quest_detail.go  # 우측 패널: 상세 정보
│   │       └── status_bar.go    # 하단 상태 표시줄
│   ├── domain/                  # 순수 도메인 모델
│   │   ├── quest.go             # Quest struct with types
│   │   ├── player.go            # Player stats with Flow
│   │   ├── status.go            # Status enum
│   │   └── flow.go              # Flow calculation types
│   ├── engine/                  # Business rules
│   │   ├── leveling.go          # XP/level calculation
│   │   ├── flow_calc.go         # Flow evaluation logic
│   │   └── time_eval.go         # Lazy time evaluation
│   ├── service/                 # High-level business logic
│   │   ├── quest_service.go     # Quest completion flow
│   │   └── evaluation_service.go # Time/Flow evaluation orchestration
│   └── repository/              # Data access
│       ├── sqlite.go            # DB connection
│       ├── quest_repo.go
│       ├── player_repo.go
│       └── flow_repo.go         # Flow history storage
├── go.mod
├── go.sum
└── README.md
```

---

**Version**: 2.0.0 | **Ratified**: 2025-03-20 | **Last Amended**: 2026-04-02
