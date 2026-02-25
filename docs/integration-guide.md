# 카카오톡 채널 릴레이 서버 연동 가이드

> ※ 이 서비스는 카카오에서 제공하는 공식 서비스가 아닙니다.

## 개요

이 문서는 카카오톡 채널을 OpenClaw 인스턴스와 연동하기 위한 릴레이 서버의 연동 절차를 설명합니다.

## 아키텍처

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   카카오      │     │  릴레이 서버  │     │   OpenClaw   │
│  톡채널 사용자 │◀───▶│              │◀───▶│   인스턴스   │
└──────────────┘     └──────────────┘     └──────────────┘
       │                    │                    │
       │  Webhook POST      │  Long-poll GET     │
       │  Callback POST     │  Reply POST        │
       └────────────────────┴────────────────────┘
```

**핵심 개념:**
- 하나의 카카오톡 채널(봇)을 여러 OpenClaw 인스턴스가 공유
- 사용자는 페어링 코드로 특정 OpenClaw에 연결
- 릴레이 서버가 메시지 라우팅 담당

---

## 연동 절차

### STEP 1: 관리자 설정 (1회)

#### 1-1. Account 생성

**Admin UI 사용:**
```
URL: https://{YOUR_RELAY_SERVER}/admin/
메뉴: Accounts → Create Account
```

**API 사용:**
```bash
curl -X POST https://kakao-talkchannel-relay-....run.app/admin/api/accounts \
  -H "Content-Type: application/json" \
  -d '{"openclawUserId": "my-openclaw-instance"}'
```

**응답:**
```json
{
  "id": "uuid",
  "openclawUserId": "my-openclaw-instance",
  "relayToken": "64자_hex_토큰_1회만_표시됨",
  "mode": "relay",
  "rateLimitPerMinute": 60
}
```

> ⚠️ `relayToken`은 이 응답에서만 확인 가능. 반드시 안전하게 저장할 것.

#### 1-2. 카카오 채널 웹훅 등록

**카카오 비즈니스 채널 관리센터:**
1. 채널 설정 → 챗봇 → 스킬 설정
2. 웹훅 URL 등록:
   ```
   https://{YOUR_RELAY_SERVER}/kakao-talkchannel/webhook
   ```
3. (선택) 서명 검증 활성화 시 `KAKAO_SIGNATURE_SECRET` 환경변수 설정

---

### STEP 2: OpenClaw 연동 설정

#### 2-1. OpenClaw 플러그인 사용 (권장)

[openclaw-kakao-talkchannel-plugin](https://github.com/kakao-bart-lee/openclaw-kakao-talkchannel-plugin)을 설치하면 SSE 연결, 메시지 폴링, 응답 전송, 페어링 등이 자동 처리됩니다.

```bash
openclaw plugins install @openclaw/kakao-talkchannel
```

플러그인 설정에서 `relayUrl`과 `relayToken`을 지정합니다:

```json
{
  "channels": {
    "kakao-talkchannel": {
      "accounts": {
        "default": {
          "relayUrl": "https://{YOUR_RELAY_SERVER}",
          "relayToken": "<발급받은_토큰>"
        }
      }
    }
  }
}
```

#### 2-2. 직접 연동 (플러그인 미사용)

플러그인 없이 직접 릴레이 서버 API를 호출할 수도 있습니다.

**환경변수 설정:**

```env
KAKAO_RELAY_URL=https://{YOUR_RELAY_SERVER}
KAKAO_RELAY_TOKEN=<발급받은_64자_토큰>
```

**구현 요구사항:**

**필수 구현:**

| 기능 | 엔드포인트 | 설명 |
|------|-----------|------|
| 메시지 폴링 | `GET /openclaw/messages` | Long-poll로 메시지 수신 |
| 응답 전송 | `POST /openclaw/reply` | 카카오로 응답 전송 |
| 메시지 ACK | `POST /openclaw/messages/ack` | 수신 확인 |
| 페어링 코드 생성 | `POST /openclaw/pairing/generate` | 사용자 연결용 코드 |

**인증 방식:**
```
Authorization: Bearer <relay_token>
# 또는
?token=<relay_token>
```

---

### STEP 3: 사용자 페어링

#### 3-1. 페어링 코드 생성

**API:**
```bash
curl -X POST https://...run.app/openclaw/pairing/generate \
  -H "Authorization: Bearer <relay_token>" \
  -H "Content-Type: application/json" \
  -d '{"expirySeconds": 600}'
```

**응답:**
```json
{
  "code": "ABCD-1234",
  "expiresAt": "2025-01-31T21:10:00Z"
}
```

#### 3-2. 사용자에게 안내

```
카카오톡에서 [채널명]을 추가한 후,
채팅창에 다음을 입력하세요:

/pair ABCD-1234
```

#### 3-3. 페어링 완료

사용자가 `/pair ABCD-1234` 입력 시:
- 성공: "✅ OpenClaw에 연결되었습니다!"
- 실패: "❌ 유효하지 않은 코드입니다."

---

### STEP 4: 메시지 흐름

```
사용자 → 카카오: "안녕하세요"
    ↓
카카오 → 릴레이: POST /kakao-talkchannel/webhook
    ↓
릴레이: inbound_messages 테이블에 저장 (status: queued)
    ↓
릴레이 → 카카오: { useCallback: true }
    ↓
OpenClaw → 릴레이: GET /openclaw/messages?wait=10
    ↓
릴레이 → OpenClaw: [{ id, payload, callbackUrl }]
    ↓
릴레이: status 변경 (queued → delivered)
    ↓
OpenClaw: AI 처리
    ↓
OpenClaw → 릴레이: POST /openclaw/reply { messageId, response }
    ↓
릴레이 → 카카오: POST callbackUrl { response }
    ↓
카카오 → 사용자: "안녕하세요! 무엇을 도와드릴까요?"
```

---

## API 레퍼런스

### 인증

모든 `/openclaw/*` 엔드포인트는 인증 필요:

```
Authorization: Bearer <relay_token>
```

### 엔드포인트

#### GET /openclaw/messages

메시지 폴링 (Long-poll 지원)

**Query Parameters:**
| 파라미터 | 타입 | 기본값 | 설명 |
|---------|------|-------|------|
| limit | number | 20 | 최대 메시지 수 (1-100) |
| wait | number | 0 | 대기 시간 초 (0-30) |

**Response:**
```json
{
  "messages": [
    {
      "id": "uuid",
      "payload": { /* 카카오 웹훅 전체 */ },
      "callbackUrl": "https://...",
      "callbackExpiresAt": "2025-01-31T21:05:00Z",
      "createdAt": "2025-01-31T21:00:00Z"
    }
  ],
  "hasMore": false
}
```

#### POST /openclaw/reply

메시지 응답

**Request:**
```json
{
  "messageId": "uuid",
  "response": {
    "version": "2.0",
    "template": {
      "outputs": [
        { "simpleText": { "text": "응답 메시지" } }
      ]
    }
  }
}
```

**Response:**
```json
{
  "success": true,
  "outboundMessageId": "uuid"
}
```

#### POST /openclaw/messages/ack

메시지 수신 확인

**Request:**
```json
{
  "messageIds": ["uuid1", "uuid2"]
}
```

**Response:**
```json
{
  "acknowledged": 2,
  "requested": 2
}
```

#### POST /openclaw/pairing/generate

페어링 코드 생성

**Request:**
```json
{
  "expirySeconds": 600,
  "metadata": {}
}
```

**Response:**
```json
{
  "code": "ABCD-1234",
  "expiresAt": "2025-01-31T21:10:00Z"
}
```

#### GET /openclaw/pairing/list

페어링된 대화 목록

**Query Parameters:**
| 파라미터 | 타입 | 기본값 |
|---------|------|-------|
| limit | number | 50 |
| offset | number | 0 |

**Response:**
```json
{
  "conversations": [
    {
      "conversationKey": "channel:user123",
      "state": "paired",
      "pairedAt": "2025-01-31T20:00:00Z",
      "lastSeenAt": "2025-01-31T21:00:00Z"
    }
  ],
  "total": 1,
  "hasMore": false
}
```

#### POST /openclaw/pairing/unpair

페어링 해제

**Request:**
```json
{
  "conversationKey": "channel:user123"
}
```

---

## 사용자 명령어

카카오 채팅에서 사용 가능한 명령어:

| 명령어 | 설명 |
|--------|------|
| `/pair <코드>` | OpenClaw에 연결 |
| `/unpair` | 연결 해제 |
| `/status` | 연결 상태 확인 |
| `/help` | 도움말 |

---

## 설정값

| 환경변수 | 기본값 | 설명 |
|---------|-------|------|
| `DATABASE_URL` | - | PostgreSQL 연결 문자열 |
| `KAKAO_SIGNATURE_SECRET` | - | (선택) 웹훅 서명 검증 키 |
| `CALLBACK_TTL_SECONDS` | 55 | 카카오 콜백 만료 시간 |
| `QUEUE_TTL_SECONDS` | 900 | 메시지 큐 만료 시간 |
| `MAX_POLL_WAIT_SECONDS` | 30 | 최대 Long-poll 대기 |

---

## 문제 해결

### 페어링 코드가 작동하지 않음
- 코드 만료 확인 (기본 10분)
- 대소문자 구분 없음
- 이미 사용된 코드인지 확인

### 메시지가 전달되지 않음
- 페어링 상태 확인 (`/status`)
- OpenClaw 폴링 로그 확인
- 릴레이 서버 Health 체크: `GET /health`

### 응답이 카카오로 전달되지 않음
- `callbackExpiresAt` 만료 여부 확인
- 응답 형식이 카카오 스펙에 맞는지 확인

---

## 서비스 URL

배포 후 생성되는 URL을 사용합니다:

- **릴레이 서버**: `https://{YOUR_RELAY_SERVER}`
- **Admin UI**: `https://{YOUR_RELAY_SERVER}/admin/`
- **Health Check**: `https://{YOUR_RELAY_SERVER}/health`
