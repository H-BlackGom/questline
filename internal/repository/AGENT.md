# Repository Guide

## 책임 범위
- SQLite 연결, 초기화, CRUD, 트랜잭션

## 저장 규칙
- DB 경로: `~/.questline/data.db`
- 테이블: `quests`, `player` 자동 생성
- player: 항상 단일 row (id=1)

## 형식 규칙
- due date: `YYYY-MM-DD` 문자열
- 시각: 일관된 텍스트 형식

## 트랜잭션 규칙
- quest 완료 + player XP 갱신: 원자적 처리
- 실패 시 롤백

## 테스트 규칙
- temp-file DB 사용
- in-memory shared DB 기본값 금지
