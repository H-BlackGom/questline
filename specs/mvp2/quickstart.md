# Questline MVP 2 Quickstart Guide (퀵스타트 가이드)

## 개발 환경 설정

### 1. 저장소 클론 및 의존성 설치

```bash
# 저장소 클론
git clone https://github.com/H-BlackGom/questline.git
cd questline

# 브랜치 전환
git checkout feature/mvp2-master

# 의존성 다운로드
go mod download

# Bubble Tea 및 관련 패키지 확인
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
```

### 2. 데이터베이스 마이그레이션

```bash
# 기존 데이터베이스 백업 (권장)
cp ~/.questline/data.db ~/.questline/data.db.backup.$(date +%Y%m%d)

# 마이그레이션은 앱 시작 시 자동으로 실행됩니다
ql me  # 정상 동작 확인
```

## 빌드 및 실행

### 빌드

```bash
# 개발 빌드
go build -o ql ./cmd/ql

# 프로덕션 빌드 (최적화)
go build -ldflags="-s -w" -o ql ./cmd/ql

# 교차 컴파일 예시
# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o ql-darwin-arm64 ./cmd/ql

# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o ql-linux-amd64 ./cmd/ql
```

### 실행

```bash
# TUI 대시보드 실행
./ql check

# 또는 기본 명령어
./ql add "새 퀘스트"
./ql me
```

## MVP2 기능 테스트

### 1. 퀘스트 타입별 생성

```bash
# Daily 루틴
./ql add "아침 스트레칭 10분" -t daily
./ql add "코딩 테스트 1문제" -t daily

# Weekly 퀘스트
./ql add "주간 회고 작성" -t weekly -d 2026-04-07

# Epic 퀘스트 (큰 목표)
./ql add "웹툰 연재" -t epic
EPIC_ID=$(./ql ls --type epic | grep "웹툰" | awk '{print $1}')

# Sub 퀘스트 (Epic 하위)
./ql add "시나리오 초안" -t sub -p $EPIC_ID
./ql add "캐릭터 디자인" -t sub -p $EPIC_ID

# Guild 퀘스트 (프로젝트 카테고리)
./ql add "Side Project" -t guild
```

### 2. TUI 대시보드 조작

```bash
# TUI 실행
./ql check
```

**키보드 조작법**:

| 키 | 동작 |
|----|------|
| `↑` / `k` | 위로 이동 |
| `↓` / `j` | 아래로 이동 |
| `→` / `Enter` | 하위 패널 진입 (Sub 보기) |
| `←` / `Esc` | 상위 패널 복귀 |
| `Space` | 퀘스트 완료/해제 |
| `q` / `Ctrl+C` | 종료 |

**화면 구조**:

```
╔═══════════════════════════════════════════════════════════════════════════╗
║ 🧙 [Lv.5 Junior]              XP: [██████░░░░] 60%           🔥 BURNING     ║
╠════════════════════════════════════════╦════════════════════════════════════╣
║ [ 📋 QUEST LIST ]                      ║ [ 🔍 SUB QUESTS ]                  ║
║                                        ║                                    ║
║  [ DAILY TRAINING ]                    ║ 📜 웹툰 연재                       ║
║   [x] 아침 스트레칭 10분               ║ ─────────────────────────────────  ║
║ ❯ [ ] 코딩 테스트 1문제                ║    [x] 시나리오 초안              ║
║                                        ║ ❯ [ ] 캐릭터 디자인               ║
║  [ WEEKLY RAID ]                       ║                                    ║
║   [ ] 주간 회고 작성                   ║                                    ║
║                                        ║                                    ║
║  [ EPIC QUEST ]                        ║                                    ║
║   📜 웹툰 연재                         ║                                    ║
╚════════════════════════════════════════╩════════════════════════════════════╝
❯ ql check (이동: ↑/↓, 펼치기: →/Enter, 뒤로: ←/Esc, 완료: Space)
```

### 3. Flow 시스템 테스트

```bash
# 어제 날짜로 Daily 루틴 생성 및 완료 (테스트용)
# (실제로는 시간을 조작하거나 테스트 DB 사용)

# Flow 상태 확인
./ql me --flow

# 출력 예시:
# ╔══════════════════════════════════╗
# ║        퀘스트라인 캐릭터         ║
# ╠══════════════════════════════════╣
# ║  레벨: Lv.5                       ║
# ║  칭호: Junior                      ║
# ║  누적 XP: 450                      ║
# ║                                  ║
# ║  Flow 상태: 🔥 BURNING            ║
# ║  XP 배율: 1.5x                    ║
# ║                                  ║
# ║  어제 달성률: 100%                 ║
# ║  연속 달성: 3일                    ║
# ╚══════════════════════════════════╝

# 오늘 퀘스트 완료 시 Flow 배율 적용 확인
./ql done <quest-id>
# "+75 XP 획득! (기본 50 XP × BURNING 1.5x)"
```

### 4. 지연 평가 (Lazy Evaluation) 테스트

```bash
# 1. 일부 Daily 루틴을 어제 완료하지 않은 상태로 두기

# 2. 24시간 후 (또는 시간 조작으로) 앱 재실행
./ql check

# 3. 자동으로 수행되는 작업 확인:
#    - 어제 미완료 Daily 자동 보관 (archived)
#    - 새 Daily 자동 생성
#    - Flow 등급 재계산
```

### 5. PENDING 상태 전이 테스트

```bash
# Epic 생성 및 Sub 추가
./ql add "웹툰 연재" -t epic
EPIC_ID=$(./ql ls --type epic | grep "웹툰" | head -1 | awk '{print $1}')

./ql add "시나리오" -t sub -p $EPIC_ID
./ql add "캐릭터" -t sub -p $EPIC_ID

# TUI에서 Sub 완료
./ql check
# 1. Epic 선택 (→ 눌러 Sub 보기)
# 2. 각 Sub 선택 후 Space로 완료
# 3. 모든 Sub 완료 후 Epic이 PENDING으로 변경 확인
# 4. Epic 선택 후 Space로 최종 완료
# 5. XP 지급 확인
```

## 테스트

### 단위 테스트 실행

```bash
# 전체 테스트
go test ./...

# 특정 패키지
go test ./internal/engine/... -v
go test ./internal/tui/... -v

# 커버리지 확인
go test ./... -cover
```

### 5대 방어 로직 테스트

```bash
# 지연 평가 엣지 케이스 테스트
go test ./internal/engine/... -run TestLazyEval -v

# 테스트 케이스:
# - TestLazyEval_DivideByZero: 루틴 0개
# - TestLazyEval_DuplicatePenalty: 중복 평가 방지
# - TestLazyEval_FutureDate: 미래 날짜 처리
# - TestLazyEval_Boundary4AM: 04:00:00 정각
# - TestLazyEval_PendingPreserve: PENDING 상태 보존
```

### TUI 통합 테스트

```bash
# Bubble Tea 프로그램 테스트
go test ./internal/tui/... -v

# 테스트는 tea.Program 모킹 또는 상태 기반 검증
```

## 디버깅

### 로그 출력

```bash
# 디버그 로그 활성화
DEBUG=1 ./ql check

# 또는
go run ./cmd/ql check --debug
```

### 데이터베이스 직접 조회

```bash
# SQLite CLI로 직접 확인
sqlite3 ~/.questline/data.db

# 유용한 쿼리
SELECT id, title, type, status, parent_id FROM quests WHERE deleted_at IS NULL;
SELECT * FROM player;
SELECT * FROM daily_evaluation ORDER BY date DESC LIMIT 7;
```

### TUI 레이아웃 디버깅

```bash
# 터미널 크기 확인
stty size  # 행 열

# 작은 터미널 테스트
# TUI는 최소 80x24 권장
```

## 문제 해결

### TUI가 실행되지 않음

```bash
# 오류 메시지 확인
./ql check 2>&1

# 터미널 지원 확인
echo $TERM  # should be xterm-256color or similar

# Bubble Tea 종속성 재설치
go mod tidy
go mod download
```

### 마이그레이션 실패

```bash
# 백업에서 복원
cp ~/.questline/data.db.backup.20260402 ~/.questline/data.db

# 스키마 수동 확인
sqlite3 ~/.questline/data.db ".schema"

# 앱 재실행 시 자동 마이그레이션 실행
./ql me
```

### Flow 계산 확인

```bash
# daily_evaluation 테이블 확인
sqlite3 ~/.questline/data.db "SELECT * FROM daily_evaluation ORDER BY date DESC;"

# player 테이블 flow_status 확인
sqlite3 ~/.questline/data.db "SELECT flow_status, last_evaluated FROM player;"

# Flow 상태는 앱 시작 시 자동으로 계산됩니다
./ql me --flow
```

## 개발 워크플로우

### 1. 기능 브랜치 개발

```bash
# Task별 브랜치
# Task 6: DB 마이그레이션
git checkout -b feature/mvp2-db-migration

# 작업 후 커밋
git add .
git commit -m "feat(db): add quest type and parent_id columns"

# PR 생성 후 develop에 머지
```

### 2. AI 페어 프로그래밍

```bash
# 복잡한 설계 (Gemini CLI)
gemini "Bubble Tea 마스터-디테일 패턴 설계"

# 보일러플레이트 생성 (OpenCode)
opencode "Create TUI model structs for FocusMaster/FocusDetail"
```

### 3. 코드 품질 검증

```bash
# 린트
go vet ./...
golint ./...

# 포맷팅
gofmt -w .

# 테스트 커버리지
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 배포

### 바이너리 빌드

```bash
# 버전 정보 포함
VERSION=$(git describe --tags --always)
go build -ldflags="-X main.Version=$VERSION -s -w" -o ql ./cmd/ql
```

### 설치

```bash
# 로컬 설치
go install ./cmd/ql

# 또는 수동 설치
cp ql /usr/local/bin/
```

## 참고 자료

- [MVP2 구현 계획](./plan.md)
- [데이터 모델](./data-model.md)
- [요구사항](../../.opencode/document/requirement_2.md)
- [헌법](../../.specify/memory/constitution.md)
- [Bubble Tea 문서](https://github.com/charmbracelet/bubbletea)
- [Lip Gloss 문서](https://github.com/charmbracelet/lipgloss)
