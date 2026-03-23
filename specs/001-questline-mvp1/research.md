# 연구 보고서: Questline MVP 1 기술 스택 조사

**생성일**: 2025-03-20
**브랜치**: develop
**문서 경로**: `/specs/001-questline-mvp1/research.md`

---

## 1. 요약 (Summary)

Questline MVP 1은 Go 언어로 개발되는 CLI 기반 RPG 퀘스트 관리 도구입니다.
핵심 기술 스택은 Cobra, SQLite, fatih/color를 사용하며, 순수 Go SQLite 구현체를
사용하여 CGO 의존성을 제거합니다.

---

## 2. 기술 스택 결정 (Technology Decisions)

### 2.1. CLI 프레임워크: Cobra

**결정**: Cobra 사용
**근거**:
- Go 생태계에서 가장 널리 사용되는 CLI 프레임워크
- Hugo, Kubernetes, Docker 등 대규모 프로젝트에서 검증됨
- 자동 플래그 파싱, 자동 생성 도움말, 쉘 자동완성 지원
- Subcommand 구조에 최적화 (ql add, ql done 등)

**대안 고려**:
- urfave/cli: 더 가벼움, Cobra보다 단순한 구조
- spf13/cobra: 선택 - 더 풍부한 생태계와 문서화

### 2.2. 데이터베이스: SQLite (pure-go)

**결정**: modernc.org/sqlite 사용
**근거**:
- CGO 의존성 없이 순수 Go로 구현된 SQLite
- Cross-compilation 가능 (단일 바이너리 배포)
- Windows, macOS, Linux 동일한 바이너리로 배포 가능
- SQLite의 내구성과 신뢰성 유지

**대안 고려**:
- mattn/go-sqlite3: CGO 필요, 크로스 컴파일 복잡함
- boltdb/bbolt: Key-value, 관계형 쿼리 필요 시 불편함
- modernc.org/sqlite: 선택 - CGO-free, 표준 database/sql 인터페이스

### 2.3. 색상 출력: fatih/color

**결정**: fatih/color 사용
**근거**:
- Go에서 가장 인기 있는 터미널 색상 라이브러리
- ANSI 색상 코드 자동 처리
- Windows 지원 자동
- 간단한 API: `color.Green("성공 메시지")`

**대안 고려**:
- chalk: Node.js 스타일 체이닝
- aurora: 메소드 체이닝 지원
- fatih/color: 선택 - 가장 간단하고 널리 사용됨

### 2.4. 테스트 라이브러리: stretchr/testify

**결정**: testify 사용
**근거**:
- Go 표준 테스트 라이브러리의 assert/require 확장
- Table-driven 테스트 작성에 최적화
- Mock 지원 (mock 패키지)
- Suite 지원 (suite 패키지)

### 2.5. 날짜 파싱: natalie-lang/naturaldate (참고)

**결정**: 요구사항에 `-d tmr` 형식 언급됨
**접근법**:
- 기본: Go 표준 `time.Parse` 사용
- 확장: "tmr", "내일", "오늘" 등 자연어 날짜 파싱은 MVP2로 연기
- MVP1에서는 ISO 8601 형식(YYYY-MM-DD) 우선 지원

---

## 3. 프로젝트 구조 (Project Structure)

```
questline/
├── cmd/ql/                      # CLI 진입점
│   └── main.go                  # Cobra root command 설정
├── internal/
│   ├── cli/                     # Cobra command 구현체
│   │   ├── add.go               # ql add
│   │   ├── done.go              # ql done
│   │   ├── ls.go                # ql ls
│   │   ├── me.go                # ql me
│   │   └── root.go              # root command, flag 바인딩
│   ├── domain/                  # 도메인 모델
│   │   ├── quest.go             # Quest struct, 상태 상수
│   │   └── player.go            # Player struct, 칭호
│   ├── engine/                  # 비즈니스 로직 (헌법 원칙 준수)
│   │   └── leveling.go          # XP/레벨업 계산
│   ├── repository/              # 데이터 접근 계층
│   │   ├── sqlite.go            # DB 연결, 마이그레이션
│   │   ├── quest_repo.go        # Quest CRUD
│   │   └── player_repo.go       # Player CRUD
│   └── service/                 # 유스케이스
│       ├── quest_service.go     # 퀘스트 비즈니스 로직
│       └── player_service.go    # 플레이어 비즈니스 로직
├── pkg/                         # 재사용 가능한 유틸리티
│   ├── color/                   # fatih/color 래퍼
│   └── formatter/               # 출력 포맷터 (tabwriter 등)
├── go.mod
├── go.sum
└── questline                    # 빌드된 바이너리
```

---

## 4. 데이터베이스 스키마 (Database Schema)

### 4.1. 마이그레이션 버전 관리

```sql
-- migrations/001_initial.sql
CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### 4.2. Quests 테이블

```sql
CREATE TABLE quests (
    id TEXT PRIMARY KEY,                    -- UUID v4
    title TEXT NOT NULL,
    status TEXT DEFAULT 'TODO',             -- TODO, DONE, DROPPED
    difficulty TEXT DEFAULT 'NORMAL',       -- EASY, NORMAL, HARD
    xp_reward INTEGER DEFAULT 50,
    due_date DATETIME,                      -- ISO 8601
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

CREATE INDEX idx_quests_status ON quests(status);
CREATE INDEX idx_quests_created ON quests(created_at DESC);
```

### 4.3. Player 테이블

```sql
CREATE TABLE player (
    id INTEGER PRIMARY KEY CHECK (id = 1),  -- 단일 로우
    level INTEGER DEFAULT 1,
    current_xp INTEGER DEFAULT 0,
    total_xp_earned INTEGER DEFAULT 0,      -- 누적 XP (레벨업 시 차감X)
    quests_completed INTEGER DEFAULT 0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 초기 데이터 삽입
INSERT INTO player (id) VALUES (1);
```

---

## 5. 핵심 알고리즘 (Core Algorithms)

### 5.1. 레벨업 계산

```go
// CalculateLevelUp 계산 로직
// 다음 레벨 필요 XP = 100 + (현재 레벨 * 50)
// 
// 예시:
// - Lv.1 → Lv.2: 150 XP 필요 (100 + 1*50)
// - Lv.2 → Lv.3: 200 XP 필요 (100 + 2*50)
// - 연속 레벨업: 400 XP 획득 (Lv.1, 50 XP) → Lv.3 직행 가능

type LevelingResult struct {
    OldLevel       int
    NewLevel       int
    LevelsGained   int
    RemainingXP    int
    TotalXPToNext  int
}

func CalculateLevelUp(currentLevel, currentXP, xpGained int) LevelingResult {
    // 구현: 초과 XP 누적 및 연속 레벨업 재귀 처리
}
```

### 5.2. 칭호 계산

```go
// GetTitleForLevel 레벨별 칭호 반환
//
// Lv.1~9: Intern
// Lv.10~19: Junior
// Lv.20~29: Senior
// Lv.30~49: Lead
// Lv.50~98: Principal
// Lv.99: Guru

func GetTitleForLevel(level int) string {
    // 구현: 구간별 칭호 매핑
}
```

---

## 6. CLI 인터페이스 명세 (CLI Contract)

### 6.1. ql add

```bash
ql add "퀘스트 제목" [-d "2025-03-25"]
```

**입력**:
- `<title>` (필수): 퀘스트 제목
- `-d, --due` (선택): 마감일 (ISO 8601)

**출력**:
```
✓ 퀘스트 #a1b2c3d4 생성됨: "퀘스트 제목"
```

**에러**:
```
✗ 오류: 퀘스트 제목이 필요합니다.
```

### 6.2. ql done

```bash
ql done <quest-id>
```

**입력**:
- `<quest-id>` (필수): 퀘스트 ID

**출력 (레벨업 시)**:
```
✓ 퀘스트 완료! +50 XP

🎉 레벨업! Lv.1 → Lv.2
   칭호: Junior
   다음 레벨까지: 150 XP
```

**출력 (레벨업 없음)**:
```
✓ 퀘스트 완료! +50 XP
   다음 레벨까지: 30 XP
```

### 6.3. ql ls

```bash
ql ls [-a|--all] [-d|--done]
```

**입력**:
- `-a, --all` (선택): 전체 퀘스트
- `-d, --done` (선택): 완료된 퀘스트만
- 기본: 진행 중인 퀘스트만

**출력**:
```
ID          제목                    상태    마감일
----        ------                  ----    ------
a1b2c3d4    리팩토링하기            TODO    03-25
b2c3d4e5    문서 작성               TODO    03-26
```

### 6.4. ql me

```bash
ql me
```

**출력**:
```
╔══════════════════════════════════╗
║        퀘스트라인 캐릭터         ║
╠══════════════════════════════════╣
║  레벨: Lv.2                      ║
║  칭호: Junior                    ║
║  누적 XP: 150                    ║
║                                  ║
║  다음 레벨까지: 150/200 XP       ║
║  [████████░░░░░░░░░░░░] 75%      ║
║                                  ║
║  완료한 퀘스트: 3개              ║
╚══════════════════════════════════╝
```

---

## 7. 테스트 전략 (Testing Strategy)

### 7.1. Engine Layer (internal/engine)

**Table-Driven Tests**:
```go
func TestCalculateLevelUp(t *testing.T) {
    tests := []struct {
        name         string
        level        int
        currentXP    int
        xpGained     int
        wantLevel    int
        wantLevelsUp int
        wantRemain   int
    }{
        {"단일 레벨업", 1, 100, 50, 2, 1, 0},
        {"연속 레벨업", 1, 0, 400, 3, 2, 50},
        {"정확히 경계", 1, 100, 50, 2, 1, 0},
        {"레벨업 없음", 1, 0, 30, 1, 0, 30},
    }
    // ...
}
```

### 7.2. Repository Layer (internal/repository)

**In-Memory SQLite Tests**:
```go
func TestQuestRepository(t *testing.T) {
    // file::memory:?cache=shared 사용
    // 각 테스트마다 새로운 DB 인스턴스
}
```

### 7.3. CLI Layer (internal/cli)

**Output Capture Tests**:
```go
func TestAddCommand(t *testing.T) {
    // stdout/stderr 캡처
    // Cobra 에러 출력 검증
}
```

---

## 8. 빌드 및 배포 (Build & Distribution)

### 8.1. 로컬 빌드

```bash
# 개발 빌드
go build -o questline ./cmd/ql

# 릴리스 빌드 (최적화)
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o questline-darwin-amd64 ./cmd/ql
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o questline-linux-amd64 ./cmd/ql
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o questline-windows-amd64.exe ./cmd/ql
```

### 8.2. 데이터 위치

```
~/.questline/
├── data.db          # SQLite 데이터베이스
└── config.json      # 향후 설정 파일
```

### 8.3. 설치 방법

```bash
# 홈브루 (향후)
brew install questline

# 직접 다운로드
curl -L https://github.com/H-BlackGom/questline/releases/latest/download/questline-$(uname -s)-$(uname -m) -o questline
chmod +x questline
mv questline /usr/local/bin/
```

---

## 9. 위험 요소 및 완화책 (Risks & Mitigations)

| 위험 | 가능성 | 영향 | 완화책 |
|------|--------|------|--------|
| SQLite 동시성 | 낮음 | 중간 | 단일 사용자 CLI이므로 제한적. WAL 모드 사용 |
| XP 오버플로우 | 낮음 | 높음 | uint64 사용, 경계값 테스트 강화 |
| 날짜 파싱 복잡성 | 중간 | 낮음 | MVP1에서는 ISO 8601만 지원 |
| 크로스 플랫폼 경로 | 중간 | 중간 | filepath.Join 사용, 테스트 필요 |

---

## 10. 참고 자료 (References)

- [Cobra Documentation](https://github.com/spf13/cobra)
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite)
- [fatih/color](https://github.com/fatih/color)
- [stretchr/testify](https://github.com/stretchr/testify)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

---

**결론**: 모든 기술 스택이 검증되었으며, 헌법 원칙(I-VI)을 준수하는 구조로 설계되었습니다.
