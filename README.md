# kakao-relay

카카오톡 채널 웹훅을 수신하여 OpenClaw 인스턴스에 실시간으로 중계하는 Go 기반 릴레이 서버.

> ※ 이 서비스는 카카오에서 제공하는 공식 서비스가 아닙니다.

## 아키텍처

```
카카오톡 사용자
    │
    │ 메시지 전송
    ▼
카카오톡 채널 (오픈빌더 스킬)
    │
    │ POST /kakao/webhook
    ▼
┌──────────────────────────────┐
│        kakao-relay           │
│                              │
│  웹훅 수신 → 메시지 큐잉     │
│  SSE 브로커 → 실시간 전달    │
│  콜백 프록시 → 카카오 응답   │
│                              │
│  PostgreSQL │ Redis          │
└──────────────────────────────┘
    │                ▲
    │ SSE 스트림     │ POST /openclaw/reply
    ▼                │
OpenClaw 인스턴스
```

**핵심 흐름:**
1. 카카오 웹훅 → 릴레이 서버가 메시지를 DB에 저장하고 SSE로 실시간 전달
2. OpenClaw가 AI 처리 후 `/openclaw/reply`로 응답 전송
3. 릴레이 서버가 카카오 콜백 URL로 응답을 프록시

## 빠른 시작

```bash
git clone https://github.com/burlesquer/kakao-relay.git
cd kakao-relay
docker compose up -d
```

서버가 시작되면:
- 대시보드: http://localhost:8080/dashboard/
- 헬스체크: http://localhost:8080/health

DB 스키마는 서버 시작 시 자동으로 생성됩니다 (`go:embed` 기반 마이그레이션).

## 환경 변수

| 변수 | 필수 | 기본값 | 설명 |
|------|:----:|-------|------|
| `DATABASE_URL` | O | - | PostgreSQL 연결 문자열 |
| `REDIS_URL` | O | - | Redis 연결 문자열 |
| `PORT` | | `8080` | 서버 포트 |
| `LOG_LEVEL` | | `info` | 로그 레벨 (debug, info, warn, error) |
| `KAKAO_SIGNATURE_SECRET` | | - | 카카오 웹훅 HMAC 서명 검증 키 |
| `CALLBACK_TTL_SECONDS` | | `55` | 카카오 콜백 URL 유효시간 (카카오 제한: 60초) |
| `QUEUE_TTL_SECONDS` | | `900` | 메시지 큐 TTL (15분) |

## 프로젝트 구조

```
cmd/server/main.go          서버 엔트리포인트, 라우팅 설정
internal/
  config/                    환경 변수 파싱, 상수 정의
  database/                  DB 연결, 자동 마이그레이션 (schema.sql embed)
  handler/                   HTTP 핸들러 (웹훅, SSE, 대시보드, 세션)
  middleware/                인증, Rate Limit, 서명 검증, 로깅
  model/                     데이터 모델 (Account, Message, Session 등)
  repository/                PostgreSQL 데이터 접근 계층
  service/                   비즈니스 로직 (메시지, 세션, 카카오 콜백)
  sse/                       SSE 브로커 (Redis Pub/Sub 기반)
  jobs/                      백그라운드 작업 (만료 메시지/세션 정리)
  util/                      토큰 생성, 해싱, 암호화 유틸리티
web/
  dashboard.html             임베디드 대시보드 UI
  embed.go                   go:embed 디렉티브
drizzle/migrations/          SQL 마이그레이션 히스토리 (참조용)
docker-compose.yml           PostgreSQL + Redis + 앱 통합 실행
Dockerfile                   멀티스테이지 빌드 (golang:1.25-alpine → alpine:3.19)
```

## 주요 기능

- **카카오 웹훅 수신**: HMAC-SHA256 서명 검증 (선택), `/pair`, `/unpair`, `/status`, `/help` 명령어 처리
- **SSE 실시간 스트리밍**: Redis Pub/Sub 기반, 30초 하트비트, 연결 시 대기 메시지 즉시 전달
- **세션 기반 페어링**: 대시보드에서 세션 생성 → 페어링 코드 발급 → 카카오에서 `/pair <코드>` 입력
- **콜백 프록시**: 카카오 허용 도메인만 허용 (*.kakao.com 등), HTTPS 필수, 5초 타임아웃
- **임베디드 대시보드**: 계정/세션/대화/메시지 관리, 토큰 재발급, 통계 조회
- **자동 정리**: 5분마다 만료 메시지(7일 보관) 및 세션 정리
- **보안**: 토큰 SHA256 해싱 (평문 미저장), 테넌트 격리, IP 기반 Rate Limiting

## ngrok 로컬 개발

카카오 오픈빌더 스킬은 **공인 HTTPS URL**을 요구합니다. 로컬 개발 시 [ngrok](https://ngrok.com/)을 사용하면 카카오 웹훅을 로컬 서버로 터널링할 수 있습니다.

### 1. ngrok 설치

```bash
# macOS
brew install ngrok

# Linux (snap)
snap install ngrok

# Windows (scoop)
scoop install ngrok

# 또는 https://ngrok.com/download 에서 직접 다운로드
```

### 2. ngrok 인증 (최초 1회)

[ngrok 대시보드](https://dashboard.ngrok.com/get-started/your-authtoken)에서 인증 토큰을 복사한 후:

```bash
ngrok config add-authtoken <YOUR_AUTHTOKEN>
```

### 3. 터널 시작

```bash
# 릴레이 서버(8080 포트)에 대한 HTTPS 터널
ngrok http 8080
```

출력 예시:
```
Forwarding  https://a1b2c3d4.ngrok-free.app → http://localhost:8080
```

### 4. 카카오 오픈빌더에 URL 등록

1. [카카오 i 오픈빌더](https://i.kakao.com/) → 스킬 관리
2. 스킬 URL을 ngrok URL로 설정:
   ```
   https://a1b2c3d4.ngrok-free.app/kakao/webhook
   ```
3. 콜백 URL 사용 활성화

### 5. 전체 로컬 개발 흐름

```bash
# 터미널 1: 릴레이 서버 + DB
cd kakao-relay
docker compose up -d

# 터미널 2: ngrok 터널
ngrok http 8080

# 터미널 3: OpenClaw + 플러그인
cd ../openclaw-kakao
pnpm install
openclaw plugins install -l .
openclaw gateway restart
```

### 주의사항

- ngrok 무료 플랜은 재시작마다 URL이 변경됩니다. 변경 시 오픈빌더 스킬 URL도 갱신 필요
- ngrok 유료 플랜에서는 고정 도메인(`ngrok http --url=your-domain.ngrok-free.app 8080`) 사용 가능
- `ngrok http 8080 --log=stdout` 으로 요청 로그를 터미널에서 확인 가능
- ngrok Inspector(`http://localhost:4040`)에서 웹훅 요청/응답을 상세히 확인 가능

---

## 테스트

```bash
# Docker 환경 내 빌드 검증
docker compose up -d --build

# Go 테스트 (로컬에 Go 설치 시)
go test ./...
```

## 문서

- [카카오 채널 설정 가이드](docs/setup-guide.md) — 카카오톡 채널 + 오픈빌더 + 릴레이 서버 연동
- [시스템 아키텍처](docs/architecture.md) — 메시지 흐름, 페어링, DB 스키마, 보안
- [API 명세](docs/api-spec.md) — 전체 엔드포인트 레퍼런스
