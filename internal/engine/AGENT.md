# Engine Guide

## 책임 범위
- 순수 비즈니스 규칙만
- DB, 파일, CLI 출력 의존 금지

## XP/레벨 규칙
- 퀘스트 완료: 항상 `50 XP`
- 다음 레벨 필요 XP: `100 + (현재 레벨 * 50)`

## 칭호 규칙
- Lv.1~9: Intern
- Lv.10~19: Junior
- Lv.20~29: Senior
- Lv.30~49: Lead
- Lv.50~98: Principal
- Lv.99: Guru

## 테스트 규칙
- table-driven test 우선
- 경계값, 연속 레벨업, 레벨업 없음, 0/음수 입력 필수
