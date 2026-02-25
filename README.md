# 카카오톡 채널 릴레이 서버

> ※ 이 서비스는 카카오에서 제공하는 공식 서비스가 아닙니다.

공유 카카오톡 채널을 여러 OpenClaw 인스턴스에 연결하는 Go 기반 릴레이 서버입니다. 카카오톡 채널 웹훅을 수신해 계정/대화 매핑을 기준으로 메시지를 라우팅하고, OpenClaw 측과의 long-poll 및 콜백 흐름을 처리합니다.

## 주요 기능
- 카카오톡 채널 웹훅 수신 및 서명 검증 지원
- OpenClaw 메시지 폴링/응답/ACK 처리
- 페어링 코드 기반 사용자-계정 매핑
- Admin/Portal UI 제공

## 프로젝트 구조
- `cmd/server/main.go`: 서버 엔트리포인트
- `internal/`: 핸들러/서비스/레포지토리/미들웨어 등 핵심 로직
- `admin/`, `portal/`: 프론트엔드 소스
- `public/`, `static/`: 정적 자산(서빙 대상)
- `drizzle/migrations/`: SQL 마이그레이션
- `docs/`: 아키텍처/연동 가이드/페어링 플로우

## 요구 사항
- Go 1.25+
- PostgreSQL (로컬 기본 포트: 5433)
- Redis
- Bun (프론트 빌드/테스트용)

## 로컬 실행
1) 환경 변수 설정
```
cp .env.example .env
```
필수: `DATABASE_URL`, `REDIS_URL`, `ADMIN_PASSWORD`, `ADMIN_SESSION_SECRET`, `PORTAL_SESSION_SECRET`

2) PostgreSQL 실행
```
make docker-up
```

3) Redis 실행 (로컬/도커 등 임의 방식)

4) 마이그레이션 적용
- `drizzle/migrations/`의 SQL 파일을 순서대로 적용하세요.

5) 서버 실행
```
go run ./cmd/server
```

## 프론트엔드 빌드
- Admin UI 빌드: `bun run build:admin`
- Portal UI 빌드: `bun run build:portal`
- 전체 빌드: `bun run build:all`

빌드 산출물은 기본적으로 `public/`에 생성되며, Docker 이미지에서는 `public/`이 `static/`으로 복사되어 서빙됩니다.

## 테스트
- Go: `go test ./...`
- 프론트: `bun test`, `bun test admin/`, `bun test portal/`

## 환경 변수
예시는 `.env.example`를 참고하세요.
- `DATABASE_URL`, `REDIS_URL`: 필수 연결 정보
- `KAKAO_SIGNATURE_SECRET`: 카카오 서명 검증 (선택)
- `ADMIN_PASSWORD`, `ADMIN_SESSION_SECRET`, `PORTAL_SESSION_SECRET`: 관리자/포털 세션
- `QUEUE_TTL_SECONDS`, `CALLBACK_TTL_SECONDS`: 큐/콜백 TTL 조정
- `LOG_LEVEL`, `PORT`

## 배포
- `Dockerfile`: 런타임 이미지 빌드
- `deploy.sh`: Cloud Run 배포 스크립트 (프로젝트/리전 값 확인 필요)

## 관련 프로젝트
- [openclaw-kakao-talkchannel-plugin](https://github.com/kakao-bart-lee/openclaw-kakao-talkchannel-plugin): OpenClaw 카카오톡 채널 플러그인 — SSE 연결, 메시지 폴링/응답, 페어링을 자동 처리. OpenClaw에 플러그인 설치 후 `relayUrl`과 `relayToken`을 설정하면 바로 연동됩니다.

## 문서
- `docs/setup-guide.md`: **설정 가이드** (카카오 채널 + 오픈빌더 + 릴레이 서버 전체 설정)
- `docs/architecture.md`: 시스템 아키텍처
- `docs/integration-guide.md`: OpenClaw 연동 가이드
- `docs/pairing-flow.md`: 페어링 플로우
- `docs/api-spec.md`: API 스펙
