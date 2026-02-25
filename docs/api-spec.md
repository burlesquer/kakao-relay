# Relay Server API Specification

> ※ 이 서비스는 카카오에서 제공하는 공식 서비스가 아닙니다.

공유 카카오톡 채널을 통해 다수의 OpenClaw 인스턴스를 연결하는 중계 서버 API 명세.

## Base URL

```
https://{YOUR_RELAY_SERVER}
```

---

## Authentication

### Relay Token (OpenClaw → Relay)

OpenClaw 인스턴스가 Relay 서버에 요청할 때 사용:

```
Authorization: Bearer <relay_token>
```

- 계정 생성 시 발급
- `relayTokenHash`로 DB에 저장 (원본 저장 안 함)
- 토큰으로 `accountId` 식별

---

## Endpoints

### 1. 카카오톡 채널 Webhook (Public)

카카오톡 채널 플랫폼이 호출하는 웹훅 엔드포인트.

```
POST /kakao-talkchannel/webhook
```

**Request Headers:**
```
Content-Type: application/json
X-Kakao-Signature: <hmac_signature>  (optional)
```

**Request Body:** Kakao SkillPayload

**Response (Success):**
```json
{
  "version": "2.0",
  "useCallback": true
}
```

**Behavior:**
1. (선택) 카카오 서명 검증
2. `plusfriendUserKey`로 `conversationKey` 생성
3. mapping 테이블에서 `accountId` 조회
4. 매핑 존재 → 메시지 큐에 추가
5. 매핑 없음 → 페어링 안내 응답 또는 UNPAIRED 상태로 저장
6. 즉시 `useCallback: true` 반환

---

### 2. Poll Messages (OpenClaw)

OpenClaw 인스턴스가 새 메시지를 가져오는 엔드포인트.

```
GET /openclaw/messages
```

**Headers:**
```
Authorization: Bearer <relay_token>
```

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cursor` | string | No | 마지막 메시지 ID (페이지네이션) |
| `wait` | number | No | Long-polling 대기 시간 (ms, max 30000) |
| `limit` | number | No | 최대 메시지 수 (default: 10, max: 100) |

**Response:**
```json
{
  "messages": [
    {
      "id": "msg_abc123",
      "conversationKey": "channel_123:user_xyz",
      "timestamp": 1706700000000,
      "kakaoPayload": { /* Original SkillPayload */ },
      "normalized": {
        "userId": "user_xyz",
        "text": "안녕하세요",
        "channelId": "channel_123"
      },
      "callbackUrl": "https://bot-api.kakao.com/callback/xxx",
      "callbackExpiresAt": 1706700060000
    }
  ],
  "cursor": "msg_abc123",
  "hasMore": false
}
```

**Error Responses:**
| Status | Error | Description |
|--------|-------|-------------|
| 401 | `UNAUTHORIZED` | 유효하지 않은 토큰 |
| 429 | `RATE_LIMITED` | 요청 한도 초과 |

**Behavior:**
- `relayToken` → `accountId` 매핑
- 해당 계정의 `QUEUED` 상태 메시지만 반환
- 반환된 메시지는 `DELIVERED` 상태로 변경
- Long-polling: 메시지 없으면 `wait`ms 대기 후 반환

---

### 3. Send Reply (OpenClaw)

OpenClaw 인스턴스가 응답을 보내는 엔드포인트.

```
POST /openclaw/reply
```

**Headers:**
```
Authorization: Bearer <relay_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "messageId": "msg_abc123",
  "conversationKey": "channel_123:user_xyz",
  "response": {
    "version": "2.0",
    "template": {
      "outputs": [
        {
          "simpleText": {
            "text": "안녕하세요! 무엇을 도와드릴까요?"
          }
        }
      ]
    }
  }
}
```

**Response (Success):**
```json
{
  "success": true,
  "deliveredAt": 1706700005000
}
```

**Error Responses:**
| Status | Error | Description |
|--------|-------|-------------|
| 400 | `INVALID_RESPONSE` | 응답 형식 오류 |
| 401 | `UNAUTHORIZED` | 유효하지 않은 토큰 |
| 403 | `FORBIDDEN` | 다른 계정의 메시지 |
| 404 | `MESSAGE_NOT_FOUND` | 메시지 없음 |
| 410 | `CALLBACK_EXPIRED` | 콜백 URL 만료 |

**Behavior:**
1. `relayToken` → `accountId` 검증
2. `messageId`의 소유자가 요청 계정인지 확인
3. 저장된 `callbackUrl`로 카카오에 응답 전송
4. 메시지 상태를 `ACKED`로 변경

---

### 4. Acknowledge Messages (OpenClaw)

메시지 처리 완료 확인 (선택적, 재시도 방지용).

```
POST /openclaw/messages/ack
```

**Headers:**
```
Authorization: Bearer <relay_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "messageIds": ["msg_abc123", "msg_def456"]
}
```

**Response:**
```json
{
  "acknowledged": 2
}
```

---

### 5. Pairing - Generate Code (OpenClaw)

봇 오너가 페어링 코드를 생성.

```
POST /openclaw/pairing/generate
```

**Headers:**
```
Authorization: Bearer <relay_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "expiresInSeconds": 600,
  "metadata": {
    "label": "My Bot Instance"
  }
}
```

**Response:**
```json
{
  "code": "ABCD-1234",
  "expiresAt": 1706700600000
}
```

**Constraints:**
- 코드 형식: `[A-Z0-9]{4}-[A-Z0-9]{4}`
- 기본 유효기간: 10분
- 최대 유효기간: 30분
- 계정당 활성 코드: 최대 5개

---

### 6. Pairing - Verify Code (Internal/Kakao Webhook)

사용자가 입력한 페어링 코드 검증 (내부 호출).

```
POST /internal/pairing/verify
```

**Request Body:**
```json
{
  "code": "ABCD-1234",
  "plusfriendUserKey": "user_xyz",
  "kakaoChannelId": "channel_123"
}
```

**Response (Success):**
```json
{
  "success": true,
  "accountId": "acc_xxx",
  "conversationKey": "channel_123:user_xyz"
}
```

**Response (Failure):**
```json
{
  "success": false,
  "error": "INVALID_CODE" | "EXPIRED_CODE" | "ALREADY_PAIRED"
}
```

---

### 7. Unpair (OpenClaw)

매핑 해제.

```
POST /openclaw/pairing/unpair
```

**Headers:**
```
Authorization: Bearer <relay_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "conversationKey": "channel_123:user_xyz"
}
```

**Response:**
```json
{
  "success": true
}
```

---

### 8. List Paired Users (OpenClaw)

페어링된 사용자 목록 조회.

```
GET /openclaw/pairing/list
```

**Headers:**
```
Authorization: Bearer <relay_token>
```

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cursor` | string | No | 페이지네이션 커서 |
| `limit` | number | No | 최대 수 (default: 50) |

**Response:**
```json
{
  "users": [
    {
      "conversationKey": "channel_123:user_xyz",
      "plusfriendUserKey": "user_xyz",
      "state": "PAIRED",
      "pairedAt": 1706700000000,
      "lastSeenAt": 1706750000000
    }
  ],
  "cursor": "next_page_token",
  "hasMore": true
}
```

---

### 9. Health Check (Public)

```
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "timestamp": 1706700000000,
  "version": "1.0.0"
}
```

---

### 10. SSE Events Stream (OpenClaw)

실시간 이벤트 수신을 위한 Server-Sent Events 스트림.

```
GET /v1/events
```

**Headers:**
```
Authorization: Bearer <relay_token>
Accept: text/event-stream
```

**Event Types:**

#### `connected`
연결 성공 시 전송.

```json
{
  "accountId": "acc_xxx",
  "sessionId": "sess_yyy",
  "status": "paired" | "pending_pairing"
}
```

#### `message`
새 메시지 수신 시 전송.

```typescript
interface SSEMessageEvent {
  id: string;                        // 메시지 ID
  conversationKey: string;           // "${channelId}:${userKey}"
  kakaoPayload: KakaoSkillPayload;   // 카카오 원본 페이로드
  normalized: {
    userId: string;                  // plusfriendUserKey
    text: string;                    // 사용자 발화
    channelId: string;               // 카카오 채널 ID
  };
  createdAt: string;                 // ISO 8601 (예: "2025-01-31T21:00:00Z")
}
```

#### `pairing_complete`
페어링 완료 시 전송.

```json
{
  "conversationKey": "channel_123:user_xyz",
  "pairedAt": "2025-01-31T21:00:00Z"
}
```

#### `heartbeat`
연결 유지용 (30초 간격).

```json
{}
```

**Connection Notes:**
- 연결 끊김 시 자동 재연결 권장
- `Last-Event-ID` 헤더로 이벤트 재수신 불가 (stateless)
- 메시지 유실 방지를 위해 `GET /openclaw/messages`와 병행 사용 권장

---

## Data Models

### ConversationMapping

```typescript
type PairingState = 'UNPAIRED' | 'PENDING' | 'PAIRED' | 'BLOCKED';

interface ConversationMapping {
  id: string;
  kakaoChannelId: string;
  plusfriendUserKey: string;
  conversationKey: string;           // `${kakaoChannelId}:${plusfriendUserKey}`
  
  accountId?: string;                // null if UNPAIRED
  state: PairingState;
  
  lastCallbackUrl?: string;
  lastCallbackUrlExpiresAt?: Date;
  
  firstSeenAt: Date;
  lastSeenAt: Date;
  pairedAt?: Date;
}
```

### InboundMessage

```typescript
type DeliveryStatus = 'QUEUED' | 'DELIVERED' | 'ACKED' | 'EXPIRED' | 'FAILED';

interface InboundMessage {
  id: string;
  accountId: string;
  conversationKey: string;
  
  kakaoPayload: KakaoSkillPayload;
  normalized: {
    userId: string;
    text: string;
    channelId: string;
  };
  
  callbackUrl: string;
  callbackExpiresAt: Date;
  
  status: DeliveryStatus;
  sourceEventId?: string;            // Idempotency key
  
  createdAt: Date;
  deliveredAt?: Date;
  ackedAt?: Date;
}
```

### PairingCode

```typescript
interface PairingCode {
  code: string;                      // "ABCD-1234"
  accountId: string;
  expiresAt: Date;
  usedAt?: Date;
  usedBy?: string;                   // plusfriendUserKey
  metadata?: Record<string, unknown>;
  createdAt: Date;
}
```

---

## Rate Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| `POST /kakao-talkchannel/webhook` | 1000 req | per minute per channel |
| `GET /openclaw/messages` | 60 req | per minute per account |
| `POST /openclaw/reply` | 120 req | per minute per account |
| `POST /openclaw/pairing/generate` | 10 req | per minute per account |
| `POST /internal/pairing/verify` | 30 req | per minute per user |
| `GET /v1/events` | 5 connections | per account (concurrent) |

---

## Error Response Format

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable message",
    "details": {}
  }
}
```

---

## Webhook Signature Verification (Optional)

카카오가 서명을 제공하는 경우:

```
X-Kakao-Signature: sha256=<hmac_hex>
```

검증:
```typescript
const expected = crypto
  .createHmac('sha256', KAKAO_SIGNATURE_SECRET)
  .update(rawBody)
  .digest('hex');

if (signature !== `sha256=${expected}`) {
  return c.json({ error: 'Invalid signature' }, 401);
}
```
