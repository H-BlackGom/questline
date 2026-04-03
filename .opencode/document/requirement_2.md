# 🚀 Questline MVP 2: 통합 테크 스펙 및 개발 계획서 (최종본)

## 1. MVP 2 목표 및 개발 범위
## 1.1. 목표 (Goal)
**"Gamification의 극대화: 시각적 몰입과 일일 습관 형성"**

텍스트 기반의 단순 입력을 넘어, `Bubble Tea`를 활용한 직관적인 **마스터-디테일(Master-Detail) TUI 대시보드**를 제공합니다. 또한 매일의 루틴을 관리하는 지연 평가(4 AM)와 Flow(몰입도) 시스템을 탑재하여 사용자의 일일 리텐션을 끌어올립니다.

## 1.2. 개발 범위 (Scope)
- **✅ 포함 (In-Scope):**
	- 마스터-디테일 구조의 TUI 대시보드 구현 (`ql check` 및 기본 `ql` 명령어).  
    - `Daily(루틴)`, `Weekly`, `Epic`, `Guild`, `Sub(2-Depth)` 퀘스트 타입 확장.
    - 좌측 리스트 고정 정렬 로직 (루틴 ➜ 위클리 ➜ 에픽 ➜ 길드).
    - 새벽 4시 기준 지연 평가(Lazy Eval) 및 일일 Flow 등급 정산.
    - 부모-자식 퀘스트 간의 상태 전이 로직 (`PENDING` 상태 도입).

---

## 2. TUI 대시보드 아키텍처 (Master-Detail UX)
## 2.1. 예상 화면 (The Guild Board)
Plaintext

```
╭──────────────────────────────────────────────────────────────────────────────╮
│ 🧙 [Lv.30 Lead Developer]              [▓▓▓▓▓▓░░░░] 60%           🔥 BURNING │
├──────────────────────────────────────┬───────────────────────────────────────┤
│ [ 📋 QUEST LIST ]                    │ [ 🔍 SUB QUESTS ]                     │
│                                      │                                       │
│  [ DAILY TRAINING ]                  │ 📜 가족 웹툰 '서빙하는 아빠곰' 연재   │
│   [x] Velog '얕은 지식' 시리즈 작성  │ ───────────────────────────────────── │
│   [ ] 코틀린(Kotlin) 1강 수강        │  진행: [▓▓▓▓░░░░░░] 40% (마감: D-30)  │
│                                      │                                       │
│  [ WEEKLY RAID ]                     │    [✔] 전체 시나리오 초안 작성        │
│   🐲 주간 백로그 레이드              │  ❯ [ ] 당나귀 아내 캐릭터 시트 채색   │
│                                      │    [ ] 3살 딸내미 에피소드 콘티 짜기  │
│  [ EPIC QUEST ]                      │                                       │
│ ❯ 📜 가족 웹툰 '서빙하는 아빠곰' 연재│                                       │
╰──────────────────────────────────────┴───────────────────────────────────────╯
❯ ql check (이동: ↑/↓, 펼치기: →/Enter, 뒤로: ←/Esc, 완료: Space)
```

## 2.2. 상태 관리 (Model)
- `CurrentFocus`: 현재 활성화된 패널 (`FocusMaster` 또는 `FocusDetail`)
- `MasterCursor` / `DetailCursor`: 각 패널의 선택된 항목 인덱스

## 2.3. 키보드 이벤트 분기 처리 (Update Loop)
화면의 포커스 상태에 따라 키보드 입력 동작이 동적으로 변환되는 `internal/tui/update.go` 뼈대입니다.

```Go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q": // [공통] 프로그램 종료
			return m, tea.Quit

		case "up", "k", "down", "j": // [공통] 상하 이동
			// 현재 Focus된 패널(Master or Detail)의 Cursor 증감 처리 (Max 길이 제한 적용)
			return m, nil

		case "right", "enter": // [네비게이션] 하위 패널 진입
			if m.CurrentFocus == FocusMaster && len(m.Quests[m.MasterCursor].SubQuests) > 0 {
				m.CurrentFocus = FocusDetail
				m.DetailCursor = 0
			}
			return m, nil

		case "left", "esc": // [네비게이션] 상위 패널 복귀
			if m.CurrentFocus == FocusDetail {
				m.CurrentFocus = FocusMaster
			}
			return m, nil

		case " ": // [액션] 완료 (Mark as Done) 트리거
			var targetID int
			if m.CurrentFocus == FocusMaster {
				targetID = m.Quests[m.MasterCursor].ID
			} else {
				targetID = m.Quests[m.MasterCursor].SubQuests[m.DetailCursor].ID
			}
			// DB Update 명령 트리거 (Sub Quest 완료 시 부모 PENDING 전환 포함)
			_ = targetID 
		}
	}
	return m, nil
}
```

---

## 3. 핵심 시스템 업데이트 명세 (Core Mechanics)
## 3.1. 새벽 4시 지연 평가 (Lazy Evaluation)
크론잡 없이 앱 최초 실행 시점에 과거의 시간 흐름을 동기화합니다. (`논리적 일자 = time.Now().Add(-4 * time.Hour)`)

## 3.2. 🚨 지연 평가 시 5대 방어 로직 (Edge Cases)
단위 테스트 작성 시 반드시 검증해야 하는 핵심 예외 상황입니다.

1. **Divide by Zero 방지:** 어제 등록된 Daily가 0개일 경우, 달성률 정산 시 0으로 나누는 에러를 방지하고 기본 Flow 상태(`SMOOTH`)를 유지합니다.
2. **패널티 중복 적용 방지 (장기 미접속):** 며칠간 접속하지 않았더라도, `last_synced_at`이 가리키는 가장 마지막 접속일 단 하루 치의 달성률만 정산하여 1회의 패널티만 부여합니다.
3. **무한 과거 날짜 배정 방지:** 주간 레이드(Weekly) 이월 시 단순히 7일을 더하는 것이 아니라, 앱을 실행한 오늘 시점 기준의 "다가오는 다음 주 일요일"로 마감일을 덮어씁니다.
4. **04:00:00 경계선 처리:** `04:00:00` 정각에 실행된 로직은 '오늘'로 간주하여 처리합니다.
5. **PENDING 상태 보존:** 부모 퀘스트가 완료 보고를 대기 중인 `PENDING` 상태일 때는 날짜가 경과하더라도 `EXPIRED`로 변경하지 않고 사용자의 수동 완료 보상을 위해 영구 보존합니다.

## 3.3. Flow(몰입도) 및 XP 배율

Daily 달성률에 따라 당일 획득하는 모든 XP 효율이 변동됩니다.
- 🔥 **BURNING** (80% 이상): `획득 XP x 1.5`
- 💧 **SMOOTH** (50~79%): `획득 XP x 1.0`
- ☁️ **HAZY** (50% 미만): `획득 XP x 0.5`

## 3.4. Sub Quest 상태 전이 (PENDING)

- 모든 하위 퀘스트(`Sub`)를 완료하더라도 부모 퀘스트는 즉시 완료되지 않고 `PENDING` 상태가 됩니다. 사용자가 직접 `Space`(완료)를 입력해야 최종 XP가 지급됩니다.

---

## 4. 데이터베이스 마이그레이션 (Schema Updates)
- **`quests` 테이블 확장:** `parent_id` (Integer, FK), `type` (Text), `deleted_at` (DateTime - Soft Delete용) 추가.
- **`player` 테이블 확장:** `flow_status` (Text), `last_synced_at` (DateTime) 추가.

---

## 5. 태스크 분할 및 브랜치 전략 (Action Items)

| **Task ID** | **목표 및 작업 내용**                                                               | **Branch 명**                   |
| ----------- | ---------------------------------------------------------------------------- | ------------------------------ |
| **Task 6**  | **DB 마이그레이션 & 모델 확장**<br>• `quests` 계층/타입 및 `player` 동기화 컬럼 추가 DDL 실행        | `feature/mvp2-db-migration`    |
| **Task 7**  | **코어 엔진: 지연 평가(Sync)**<br>• 4 AM 논리적 일자 변환 및 5대 엣지 케이스 방어 로직 구현              | `feature/mvp2-lazy-eval`       |
| **Task 8**  | **계층 구조 비즈니스 로직**<br>• `add -p` 매핑 및 하위 퀘스트 All-Done 시 부모 `PENDING` 전환       | `feature/mvp2-quest-hierarchy` |
| **Task 9**  | **TUI 아키텍처: 모델 및 레이아웃**<br>• `internal/tui/model.go` 패널 포커스 모델 정의 및 뷰 렌더링    | `feature/mvp2-tui-layout`      |
| **Task 10** | **TUI 인터랙션: Update 루프 및 정렬**<br>• 리스트 정렬 및 `tea.KeyMsg`에 따른 패널/커서 이동, 완료 트리거 | `feature/mvp2-tui-interaction` |
