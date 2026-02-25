# Relay Server Architecture

> ※ 이 서비스는 카카오에서 제공하는 공식 서비스가 아닙니다.

## Overview

Relay Server는 **공유 카카오톡 채널**을 통해 다수의 OpenClaw 인스턴스를 연결하는 멀티테넌트 메시지 라우터입니다.

```
┌─────────────────────────────────────────────────────────────────────┐
│                    카카오톡 채널 플랫폼                               │
│                    (공유 카카오톡 채널 봇)                            │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  │ POST /kakao-talkchannel/webhook
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│                         Relay Server                                │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    Routing Layer                             │   │
│  │   plusfriendUserKey → conversationKey → accountId            │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│        ┌─────────────────────┼─────────────────────┐               │
│        ↓                     ↓                     ↓               │
│  ┌──────────┐         ┌──────────┐         ┌──────────┐           │
│  │ Queue A  │         │ Queue B  │         │ Queue C  │           │
│  │(owner_A) │         │(owner_B) │         │(owner_C) │           │
│  └──────────┘         └──────────┘         └──────────┘           │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
        │                     │                     │
        │ GET /messages       │ GET /messages       │ GET /messages
        ↓                     ↓                     ↓
   ┌─────────┐          ┌─────────┐          ┌─────────┐
   │OpenClaw │          │OpenClaw │          │OpenClaw │
   │   A     │          │   B     │          │   C     │
   └─────────┘          └─────────┘          └─────────┘
```

---

## Core Concepts

### 1. Conversation Key

사용자를 고유하게 식별하는 복합 키:

```
conversationKey = ${kakaoChannelId}:${plusfriendUserKey}
```

- `kakaoChannelId`: 공유 카카오 채널 ID
- `plusfriendUserKey`: 카카오 사용자 고유 ID (봇 간 안정적)

### 2. Account

OpenClaw 인스턴스(봇 오너)를 나타냄:

```typescript
interface Account {
  id: string;
  openclawUserId: string;     // OpenClaw 사용자 식별자
  relayTokenHash: string;     // 인증 토큰 해시
  rateLimitPerMinute: number; // 분당 요청 한도
  createdAt: Date;
  disabledAt?: Date;
}
```

### 3. Conversation Mapping

카카오 사용자 ↔ OpenClaw 인스턴스 매핑:

```typescript
interface ConversationMapping {
  conversationKey: string;    // PK
  accountId?: string;         // FK to Account (null if unpaired)
  state: 'UNPAIRED' | 'PENDING' | 'PAIRED' | 'BLOCKED';
  pairedAt?: Date;
}
```

---

## Message Flow

### Inbound (카카오톡 채널 → OpenClaw)

```
1. Kakao webhook 수신
   POST /kakao-talkchannel/webhook
   Body: { userRequest: { user: { properties: { plusfriendUserKey } }, utterance } }

2. conversationKey 생성
   conversationKey = `${channelId}:${plusfriendUserKey}`

3. 라우팅 결정
   mapping = SELECT * FROM mappings WHERE conversation_key = ?
   
   IF mapping.state == 'PAIRED':
     → 메시지 큐에 추가 (accountId = mapping.accountId)
   ELSE IF mapping.state == 'UNPAIRED':
     → 페어링 안내 응답
   ELSE IF mapping.state == 'PENDING':
     → 페어링 대기 중 응답

4. 즉시 ACK 반환
   { "version": "2.0", "useCallback": true }

5. OpenClaw 폴링
   GET /openclaw/messages (Authorization: Bearer <token>)
   → accountId에 해당하는 메시지만 반환
```

### Outbound (OpenClaw → 카카오톡 채널)

```
1. OpenClaw 응답 전송
   POST /openclaw/reply
   Body: { messageId, conversationKey, response }

2. 권한 검증
   - relayToken → accountId
   - message.accountId == accountId?

3. 카카오 콜백 호출
   POST callbackUrl
   Body: KakaoSkillResponse

4. 메시지 상태 업데이트
   UPDATE messages SET status = 'ACKED' WHERE id = ?
```

---

## Database Schema

### Tables

```sql
-- 계정 (OpenClaw 인스턴스)
CREATE TABLE accounts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  openclaw_user_id TEXT NOT NULL UNIQUE,
  relay_token_hash TEXT NOT NULL,
  rate_limit_per_minute INT DEFAULT 60,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  disabled_at TIMESTAMPTZ
);

-- 대화 매핑 (사용자 ↔ 계정)
CREATE TABLE conversation_mappings (
  conversation_key TEXT PRIMARY KEY,
  kakao_channel_id TEXT NOT NULL,
  plusfriend_user_key TEXT NOT NULL,
  account_id UUID REFERENCES accounts(id),
  state TEXT NOT NULL DEFAULT 'UNPAIRED',
  last_callback_url TEXT,
  last_callback_url_expires_at TIMESTAMPTZ,
  first_seen_at TIMESTAMPTZ DEFAULT NOW(),
  last_seen_at TIMESTAMPTZ DEFAULT NOW(),
  paired_at TIMESTAMPTZ,
  UNIQUE (kakao_channel_id, plusfriend_user_key)
);

-- 인바운드 메시지 큐
CREATE TABLE inbound_messages (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id UUID NOT NULL REFERENCES accounts(id),
  conversation_key TEXT NOT NULL,
  kakao_payload JSONB NOT NULL,
  normalized JSONB NOT NULL,
  callback_url TEXT NOT NULL,
  callback_expires_at TIMESTAMPTZ NOT NULL,
  status TEXT NOT NULL DEFAULT 'QUEUED',
  source_event_id TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  delivered_at TIMESTAMPTZ,
  acked_at TIMESTAMPTZ,
  UNIQUE (source_event_id)
);

-- 페어링 코드
CREATE TABLE pairing_codes (
  code TEXT PRIMARY KEY,
  account_id UUID NOT NULL REFERENCES accounts(id),
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ,
  used_by TEXT,
  metadata JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 인덱스
CREATE INDEX idx_messages_account_status ON inbound_messages(account_id, status);
CREATE INDEX idx_messages_expires ON inbound_messages(callback_expires_at);
CREATE INDEX idx_mappings_account ON conversation_mappings(account_id);
CREATE INDEX idx_pairing_expires ON pairing_codes(expires_at);
```

---

## Long Polling

효율적인 메시지 폴링을 위한 Long Polling 구현:

```typescript
async function pollMessages(accountId: string, wait: number, limit: number) {
  const deadline = Date.now() + wait;
  
  while (Date.now() < deadline) {
    const messages = await db.query(`
      SELECT * FROM inbound_messages 
      WHERE account_id = $1 AND status = 'QUEUED'
      ORDER BY created_at ASC
      LIMIT $2
      FOR UPDATE SKIP LOCKED
    `, [accountId, limit]);
    
    if (messages.length > 0) {
      // Mark as delivered
      await db.query(`
        UPDATE inbound_messages 
        SET status = 'DELIVERED', delivered_at = NOW()
        WHERE id = ANY($1)
      `, [messages.map(m => m.id)]);
      
      return messages;
    }
    
    // Wait before retry
    await sleep(Math.min(1000, deadline - Date.now()));
  }
  
  return [];
}
```

---

## Idempotency

웹훅 재시도와 폴링 경쟁으로 인한 중복 처리 방지:

### Source Event ID

```typescript
// 카카오 페이로드에서 고유 ID 추출 또는 생성
function getSourceEventId(payload: KakaoSkillPayload): string {
  // 카카오가 제공하는 경우
  if (payload.userRequest.eventId) {
    return payload.userRequest.eventId;
  }
  
  // 없으면 해시 생성
  const content = JSON.stringify({
    channelId: payload.bot.id,
    userId: payload.userRequest.user.id,
    utterance: payload.userRequest.utterance,
    timestamp: payload.userRequest.timestamp,
  });
  
  return crypto.createHash('sha256').update(content).digest('hex');
}
```

### Upsert with Conflict

```sql
INSERT INTO inbound_messages (source_event_id, ...)
VALUES ($1, ...)
ON CONFLICT (source_event_id) DO NOTHING
RETURNING id;
```

---

## TTL & Cleanup

### Message TTL

- **콜백 URL 만료**: 1분 (카카오 제한)
- **메시지 큐 TTL**: 15분 (설정 가능)
- **페어링 코드 TTL**: 10분 (기본)

### Background Job

```typescript
// 매 분 실행
async function cleanupExpiredMessages() {
  // 만료된 메시지 상태 변경
  await db.query(`
    UPDATE inbound_messages 
    SET status = 'EXPIRED'
    WHERE status IN ('QUEUED', 'DELIVERED')
      AND callback_expires_at < NOW()
  `);
  
  // 오래된 메시지 삭제 (7일)
  await db.query(`
    DELETE FROM inbound_messages 
    WHERE created_at < NOW() - INTERVAL '7 days'
  `);
  
  // 만료된 페어링 코드 삭제
  await db.query(`
    DELETE FROM pairing_codes 
    WHERE expires_at < NOW() AND used_at IS NULL
  `);
}
```

---

## Security Considerations

### 1. Tenant Isolation

- 모든 쿼리에 `account_id` 필터 필수
- 토큰 검증 실패 시 즉시 거부
- 다른 계정 데이터 접근 불가

### 2. Rate Limiting

계정별 분당 요청 한도:

```typescript
const rateLimiter = new RateLimiter({
  keyPrefix: 'ratelimit:',
  points: account.rateLimitPerMinute,
  duration: 60,
});

await rateLimiter.consume(account.id);
```

### 3. Token Security

- 원본 토큰 저장 안 함 (해시만)
- 토큰 생성 시 즉시 표시 후 폐기
- 토큰 재발급 가능 (기존 토큰 무효화)

### 4. Pairing Security

- 코드 단기 유효 (10분)
- 일회용 (사용 후 무효화)
- 계정당 활성 코드 제한 (5개)
- 실패 시도 Rate Limiting

---

## Monitoring

### Key Metrics

| Metric | Description |
|--------|-------------|
| `relay.webhook.received` | 수신 웹훅 수 |
| `relay.message.queued` | 큐에 추가된 메시지 |
| `relay.message.delivered` | 전달된 메시지 |
| `relay.message.expired` | 만료된 메시지 |
| `relay.callback.sent` | 카카오 콜백 전송 |
| `relay.callback.failed` | 콜백 실패 |
| `relay.pairing.created` | 생성된 페어링 |
| `relay.pairing.verified` | 완료된 페어링 |

### Health Indicators

```typescript
async function healthCheck() {
  return {
    database: await checkDatabase(),
    queueDepth: await getQueueDepth(),
    oldestUndelivered: await getOldestUndeliveredAge(),
    activeAccounts: await countActiveAccounts(),
  };
}
```
