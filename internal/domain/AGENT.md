# Domain Guide

## 책임 범위
- Quest, Player 구조체와 최소 표현
- 인프라 의존성 없음

## Quest 규칙
- ID: UUID v4 앞 8자 소문자
- 상태: `TODO`, `DONE`, `DROPPED` (MVP1에서 DROPPED는 예약)
- due_date: 선택값

## Player 규칙
- 단일 사용자 기준 (싱글톤)
- XP, 레벨, 완료 수 최소 필드

## 금지 사항
- 외부 라이브러리 의존성 금지
- 저장소/CLI/비즈니스 계산 로직 금지
