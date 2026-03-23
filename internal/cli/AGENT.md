# CLI Guide

## 책임 범위
- Cobra 명령 정의와 입력 검증
- 사용자 출력과 종료 코드
- 비즈니스 계산은 `engine`, 저장은 `repository`에 위임

## 출력 규칙
- 정상 메시지: `cmd.OutOrStdout()`
- 오류 메시지: `cmd.ErrOrStderr()`
- 성공: `✓` 접두어
- 오류: `✗ 오류:` 형식

## 종료 코드 규칙
| 코드 | 의미 |
|------|------|
| 0 | 성공 |
| 1 | 일반 오류 |
| 2 | 잘못된 입력 |
| 3 | 대상 없음 |
| 4 | 데이터베이스 오류 |

## 명령별 규칙
### add
- 제목: trim 후 비어 있지 않음, 최대 200자
- due date: `YYYY-MM-DD`만, 오늘 또는 미래

### done
- 이미 완료된 quest: XP 재지급 없음
- 성공 출력: `+50 XP`

### ls
- 기본: TODO만
- 정렬: `created_at DESC`, tie-break `id ASC`
- 충돌 플래그: 종료 코드 2

### me
- 첫 실행 자동 초기화
- 박스형 출력, 20칸 진행 바

## 테스트 규칙
- CLI 테스트: 항상 temp HOME 사용
- 실제 `~/.questline` 건드리지 않음
