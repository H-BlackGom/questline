# 퀵스타트 가이드: Questline MVP 1

**문서 경로**: `/specs/001-questline-mvp1/quickstart.md`

---

## 설치

### 방법 1: 소스에서 빌드

```bash
# 저장소 클론
git clone https://github.com/H-BlackGom/questline.git
cd questline

# 의존성 설치
go mod download

# 빌드
go build -o ql ./cmd/ql

# PATH에 추가 (선택)
chmod +x ql
sudo mv ql /usr/local/bin/
```

### 방법 2: 릴리스 바이너리 (MVP2 - 예정)

**참고**: MVP1에서는 소스에서 빌드해주세요.

```bash
# macOS (Intel)
curl -L https://github.com/H-BlackGom/questline/releases/latest/download/questline-darwin-amd64 -o ql
chmod +x ql
mv ql /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/H-BlackGom/questline/releases/latest/download/questline-darwin-arm64 -o ql
chmod +x ql
mv ql /usr/local/bin/

# Linux
curl -L https://github.com/H-BlackGom/questline/releases/latest/download/questline-linux-amd64 -o ql
chmod +x ql
sudo mv ql /usr/local/bin/
```

---

## 첫 실행

```bash
# 버전 확인
ql --version

# 도움말
ql --help
```

**초기 데이터 디렉토리 생성**:
첫 실행 시 `~/.questline/data.db`가 자동으로 생성됩니다.

---

## 기본 사용법

### 1. 퀘스트 추가하기

```bash
# 기본 퀘스트 추가
ql add "코드 리뷰하기"
# 출력: ✓ 퀘스트 #a1b2c3d4 생성됨: "코드 리뷰하기"

# 마감일 지정
ql add "문서 작성" -d "2025-03-25"
ql add "버그 수정" --due="2025-12-31"
```

### 2. 퀘스트 목록 보기

```bash
# 진행 중인 퀘스트만 (기본)
ql ls

# 모든 퀘스트
ql ls -a
ql ls --all

# 완료된 퀘스트만
ql ls -d
ql ls --done
```

**출력 예시**:
```
ID          제목                    상태    마감일
----        ------                  ----    ------
a1b2c3d4    코드 리뷰하기           TODO    03-25
b2c3d4e5    문서 작성               TODO    03-26
```

### 3. 퀘스트 완료하기

```bash
# 퀘스트 완료
ql done a1b2c3d4
```

**출력 예시** (레벨업 없음):
```
✓ 퀘스트 완료! +50 XP
   다음 레벨까지: 30/150 XP
```

**출력 예시** (레벨업 있음):
```
✓ 퀘스트 완료! +50 XP

🎉 레벨업! Lv.1 → Lv.2
   칭호: Junior
   다음 레벨까지: 0/200 XP
```

### 4. 플레이어 상태 확인

```bash
ql me
```

**출력 예시**:
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

## 사용 시나리오

### 시나리오 1: 하루의 업무 관리

```bash
# 아침: 오늘 할 일 등록
ql add "회의 준비" -d "2025-03-20"
ql add "코드 리뷰" -d "2025-03-20"
ql add "버그 수정" -d "2025-03-20"

# 오전: 할 일 확인
ql ls

# 업무 완료 시
ql done a1b2c3d4
ql done b2c3d4e5

# 진행 상황 확인
ql me

# 하루 마무리
ql ls -d  # 오늘 완료한 것들
```

### 시나리오 2: 프로젝트 마일스톤

```bash
# 큰 작업 등록
ql add "MVP 1 릴리스 준비" -d "2025-03-31"
ql add "테스트 커버리지 80% 달성" -d "2025-03-25"

# 주기적으로 진행
ql ls  # 남은 작업 확인
ql done c3d4e5f6  # 완료 시 XP 획득

# 레벨업 동기부여
ql me  # 성장 확인
```

---

## 레벨업 시스템

**XP 획득**:
- 퀘스트 1개 완료 = +50 XP

**레벨업 필요 XP**:
```
다음 레벨 필요 XP = 100 + (현재 레벨 * 50)
```

**예시**:
- Lv.1 → Lv.2: 150 XP 필요
- Lv.2 → Lv.3: 200 XP 필요
- Lv.3 → Lv.4: 250 XP 필요

**칭호**:
- Lv.1~9: Intern
- Lv.10~19: Junior
- Lv.20~29: Senior
- Lv.30~49: Lead
- Lv.50~98: Principal
- Lv.99: Guru

---

## 문제 해결

### 데이터베이스 초기화

```bash
# 데이터 삭제 (주의!)
rm ~/.questline/data.db

# 다음 실행 시 자동 재생성
ql me
```

### 권한 문제

```bash
# 권한 확인
ls -la ~/.questline/

# 필요시 권한 수정
chmod 755 ~/.questline/
chmod 644 ~/.questline/data.db
```

### 빌드 문제

```bash
# Go 버전 확인 (1.21+ 권장)
go version

# 의존성 정리
go mod tidy

# 재빌드
go build -o ql ./cmd/ql
```

---

## 고급 사용법

### 별칭 설정 (Bash/Zsh)

```bash
# .bashrc 또는 .zshrc에 추가
alias qa='ql add'
alias qd='ql done'
alias ql='ql ls'
alias qm='ql me'

# 사용
qa "새로운 퀘스트"
qd a1b2c3d4
```

### 함수로 자동화

```bash
# 완료하고 바로 상태 확인
function qdone() {
    ql done "$1" && ql me
}

# 사용
qdone a1b2c3d4
```

---

## 개발자 정보

**GitHub**: https://github.com/H-BlackGom/questline

**버그 리포트**: GitHub Issues

**기여**: Pull Requests 환영

---

**즐겁게 퀘스트를 완료하세요! 🎮**
