# 📘 Questline MVP 1: 통합 테크 스펙 및 개발 계획서

## 1. MVP 1 기획 및 요구사항 (Product Requirements)

### 1.1. 목표 (Goal)
* **"The Core Engine: 핵심 루프 검증"**
* 복잡한 UI나 부가 기능을 배제하고, 터미널에서 명령어로 할 일을 추가하고 완료했을 때 레벨업을 하는 경험(도파민)이 실제로 유효한가를 가장 빠르게 테스트합니다.

### 1.2. 개발 범위 (Scope)
* **✅ 포함 (In-Scope):** 순수 CLI 인터페이스 (표준 출력), 일반 퀘스트(Guild Request) 단일 지원, 로컬 SQLite 연동, 고정 XP 획득 및 선형 레벨업 시스템.
* **❌ 제외 (Out-of-Scope -> MVP 2/3 연기):** TUI 대시보드, Sub Quest(2-Depth), Daily/Weekly 타입, Flow(몰입도) 시스템, 새벽 4시 지연 평가 로직.

### 1.3. 핵심 게임 메카닉
* **XP 보상:** 퀘스트 1개 완료 시 일괄 **50 XP** 지급
* **레벨업 공식:** `다음 레벨 필요 XP = 100 + (현재 레벨 * 50)`
* **칭호 (Title):** `Lv.1~9`: Intern / `Lv.10~19`: Junior / `Lv.20~29`: Senior / `Lv.30~49`: Lead / `Lv.50~98`: Principal / `Lv.99`: Guru

### 1.4. 데이터베이스 스키마 (SQLite)
* **경로:** `~/.questline/data.db`
* **`quests` 테이블:** `id` (PK), `title` (Text), `status` (TODO, DONE, DROPPED), `due_date`, `created_at`, `completed_at`
* **`player` 테이블:** `level` (Default 1), `current_xp` (Default 0) *※ 단일 로우 유지*

---

## 2. CLI 명령어 명세 (Command Interface)

### ① `ql add "<내용>" [-d <마감일>]`
새로운 퀘스트를 `TODO` 상태로 추가. 마감일은 선택 사항.

### ② `ql done <ID>`
퀘스트를 `DONE` 처리하고 50 XP 지급. 레벨업 시 축하 메시지 출력.

### ③ `ql ls [flags]`
진행 중인 퀘스트 목록 조회.
* `-d`, `--done`: 완료된 퀘스트만 출력
* `-a`, `--all`: 상태 무관 전체 퀘스트 출력

### ④ `ql me`
현재 플레이어의 레벨, 칭호, 다음 레벨까지의 누적/필요 XP 출력.

---

## 3. 태스크 분할 및 브랜치 전략 (Tasks & Branching)

팀원들은 아래의 태스크 단위로 브랜치를 생성하고 PR을 진행합니다.

### 📝 Task 1: 코어 비즈니스 로직 (레벨링 엔진)
* **Branch:** `feature/core-engine`
* **목표:** DB 의존성 없이 순수하게 동작하는 경험치 및 레벨업 계산기 구현.
* **구현 내용:** `internal/engine/leveling.go` 작성. 초과 XP 누적 및 연속 레벨업 재귀 처리가 포함된 `CalculateLevelUp` 함수와 레벨 구간별 칭호 반환 함수 구현.

### 📝 Task 2: 데이터 접근 계층 (Repository)
* **Branch:** `feature/data-repository`
* **목표:** 데이터베이스 CRUD 로직 및 트랜잭션 구현.
* **구현 내용:** `internal/repository/quest_repo.go`, `player_repo.go` 작성. `ls -a`, `ls -d` 상태별 조회를 위한 동적 쿼리 구현. 퀘스트 완료 처리(`MarkAsDone`)와 플레이어 XP 증가를 묶는 DB 트랜잭션 로직 포함.

### 📝 Task 3: CLI 뼈대 구축 및 라우팅
* **Branch:** `feature/cli-scaffold`
* **목표:** Cobra 연동 및 명령어 껍데기, 플래그 파싱 구현.
* **구현 내용:** `cmd/ql/main.go`에 root command 세팅. `internal/cli/` 하위에 `add`, `done`, `ls`, `me` 명령어 등록. `-d` 마감일 문자열 파싱 및 `ls` 플래그 바인딩.

### 📝 Task 4: 로직 통합 및 출력 포맷팅
* **Branch:** `feature/integration`
* **목표:** CLI와 내부 로직 결합 및 터미널 UX 개선.
* **구현 내용:** `fatih/color`를 적용하여 레벨업 및 성공/에러 메시지 색상 처리. `text/tabwriter`를 사용하여 `ql ls` 명령어의 테이블 간격 정렬 출력.

---

## 4. 유닛 테스트 전략 및 계획 (Unit Testing)

### 4.1. 테스트 환경
* **Assertion:** `stretchr/testify`
* **DB Mocking:** SQLite In-Memory 모드 (`file::memory:?cache=shared`) 활용

### 4.2. 계층별 테스트 대상
1. **Engine Layer (`internal/engine`) ➜ Table-Driven Test**
    * 한 번에 대량의 XP 획득 시 2업 이상 연속 레벨업을 하는 엣지 케이스 검증.
    * 정확히 요구 경험치를 채웠을 때의 레벨업 경계선 검증.
2. **Repository Layer (`internal/repository`) ➜ In-Memory DB Test**
    * 동적 쿼리에 따른 상태별 목록(`TODO`, `DONE`)이 정확히 필터링되는지 검증.
    * `done` 트랜잭션 수행 중 에러 발생 시 `player` 테이블의 XP가 롤백되는지 검증.
3. **CLI Layer (`internal/cli`) ➜ Output Capture Test**
    * 필수 인자(ID) 누락 시 Cobra 에러 메시지가 정상 출력되는지 검증.
