# Pairing Flow

> ※ 이 서비스는 카카오에서 제공하는 공식 서비스가 아닙니다.

카카오톡 채널 사용자를 특정 OpenClaw 인스턴스에 연결하는 페어링 프로세스.

---

## Why Pairing?

공유 카카오톡 채널에서는 **하나의 봇**이 **다수의 OpenClaw 인스턴스**를 대신합니다:

```
                    ┌─────────────────┐
                    │   공유 봇        │
                    │(카카오톡 채널)  │
                    └─────────────────┘
                            │
         ┌──────────────────┼──────────────────┐
         ↓                  ↓                  ↓
    ┌─────────┐        ┌─────────┐        ┌─────────┐
    │ User A  │        │ User B  │        │ User C  │
    └─────────┘        └─────────┘        └─────────┘
         │                  │                  │
         │                  │                  │
         ?                  ?                  ?
         │                  │                  │
         ↓                  ↓                  ↓
    ┌─────────┐        ┌─────────┐        ┌─────────┐
    │OpenClaw │        │OpenClaw │        │OpenClaw │
    │    1    │        │    2    │        │    3    │
    └─────────┘        └─────────┘        └─────────┘
```

**문제**: User A의 메시지가 어느 OpenClaw로 가야 하는지 어떻게 알 수 있는가?

**해결**: 명시적 페어링 - 사용자가 코드를 입력하여 특정 인스턴스에 연결

---

## Pairing States

```
┌──────────┐    첫 메시지    ┌──────────┐
│          │ ─────────────→ │          │
│  (없음)   │                │ UNPAIRED │
│          │                │          │
└──────────┘                └──────────┘
                                  │
                                  │ /pair 명령어 입력
                                  ↓
                            ┌──────────┐
                            │          │
                            │ PENDING  │ (코드 검증 중)
                            │          │
                            └──────────┘
                                  │
                    ┌─────────────┴─────────────┐
                    │ 성공                      │ 실패
                    ↓                          ↓
              ┌──────────┐              ┌──────────┐
              │          │              │          │
              │  PAIRED  │              │ UNPAIRED │
              │          │              │          │
              └──────────┘              └──────────┘
                    │
                    │ /unpair 또는 관리자 해제
                    ↓
              ┌──────────┐
              │          │
              │ UNPAIRED │
              │          │
              └──────────┘
```

---

## Flow 1: Bot Owner Generates Code

봇 오너가 페어링 코드를 생성합니다.

### CLI 방식

```bash
$ openclaw pairing generate kakao
Pairing code: ABCD-1234
Expires in: 10 minutes

Share this code with your user.
They should send "/pair ABCD-1234" to the Kakao bot.
```

### API 방식

```
POST /openclaw/pairing/generate
Authorization: Bearer <relay_token>
Content-Type: application/json

{
  "expiresInSeconds": 600,
  "metadata": {
    "label": "Customer Support Bot"
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

### 내부 동작

```sql
INSERT INTO pairing_codes (code, account_id, expires_at, metadata)
VALUES ('ABCD-1234', 'acc_xxx', NOW() + INTERVAL '10 minutes', '{"label": "..."}');
```

---

## Flow 2: New User First Message

페어링되지 않은 사용자가 처음 메시지를 보냅니다.

### Sequence

```
User                    Relay Server                    Database
  │                          │                              │
  │  "안녕하세요"             │                              │
  │ ────────────────────────→│                              │
  │                          │  SELECT FROM mappings        │
  │                          │  WHERE conversation_key = ?  │
  │                          │─────────────────────────────→│
  │                          │                              │
  │                          │  (결과 없음)                  │
  │                          │←─────────────────────────────│
  │                          │                              │
  │                          │  INSERT INTO mappings        │
  │                          │  (state = 'UNPAIRED')        │
  │                          │─────────────────────────────→│
  │                          │                              │
  │  "연결이 필요합니다.      │                              │
  │   /pair <코드>를 입력하세요"                              │
  │ ←────────────────────────│                              │
  │                          │                              │
```

### Response to Unpaired User

```json
{
  "version": "2.0",
  "template": {
    "outputs": [
      {
        "simpleText": {
          "text": "OpenClaw에 연결되지 않았습니다.\n\n연결하려면 봇 관리자에게 페어링 코드를 요청한 후:\n/pair <코드>\n\n를 입력해주세요."
        }
      }
    ]
  }
}
```

---

## Flow 3: User Enters Pairing Code

사용자가 페어링 코드를 입력합니다.

### Sequence

```
User                    Relay Server                    Database
  │                          │                              │
  │  "/pair ABCD-1234"       │                              │
  │ ────────────────────────→│                              │
  │                          │                              │
  │                          │  SELECT FROM pairing_codes   │
  │                          │  WHERE code = 'ABCD-1234'    │
  │                          │    AND expires_at > NOW()    │
  │                          │    AND used_at IS NULL       │
  │                          │─────────────────────────────→│
  │                          │                              │
  │                          │  { account_id: 'acc_xxx' }   │
  │                          │←─────────────────────────────│
  │                          │                              │
  │                          │  UPDATE mappings             │
  │                          │  SET account_id = 'acc_xxx', │
  │                          │      state = 'PAIRED',       │
  │                          │      paired_at = NOW()       │
  │                          │─────────────────────────────→│
  │                          │                              │
  │                          │  UPDATE pairing_codes        │
  │                          │  SET used_at = NOW(),        │
  │                          │      used_by = 'user_xyz'    │
  │                          │─────────────────────────────→│
  │                          │                              │
  │  "연결되었습니다!         │                              │
  │   이제 대화를 시작하세요." │                              │
  │ ←────────────────────────│                              │
  │                          │                              │
```

### Success Response

```json
{
  "version": "2.0",
  "template": {
    "outputs": [
      {
        "simpleText": {
          "text": "✅ OpenClaw에 연결되었습니다!\n\n이제 자유롭게 대화를 시작하세요."
        }
      }
    ]
  }
}
```

### Error Responses

**잘못된 코드:**
```json
{
  "version": "2.0",
  "template": {
    "outputs": [
      {
        "simpleText": {
          "text": "❌ 유효하지 않은 코드입니다.\n\n코드를 다시 확인하거나 관리자에게 새 코드를 요청하세요."
        }
      }
    ]
  }
}
```

**만료된 코드:**
```json
{
  "version": "2.0",
  "template": {
    "outputs": [
      {
        "simpleText": {
          "text": "⏰ 코드가 만료되었습니다.\n\n관리자에게 새 코드를 요청하세요."
        }
      }
    ]
  }
}
```

---

## Flow 4: Paired User Sends Message

페어링된 사용자의 메시지는 자동으로 라우팅됩니다.

### Sequence

```
User                    Relay Server                    OpenClaw A
  │                          │                              │
  │  "날씨 알려줘"            │                              │
  │ ────────────────────────→│                              │
  │                          │                              │
  │                          │  mapping.state == 'PAIRED'   │
  │                          │  mapping.accountId == 'A'    │
  │                          │                              │
  │                          │  INSERT INTO messages        │
  │                          │  (account_id = 'A', ...)     │
  │                          │                              │
  │  { useCallback: true }   │                              │
  │ ←────────────────────────│                              │
  │                          │                              │
  │                          │  GET /messages (polling)     │
  │                          │←─────────────────────────────│
  │                          │                              │
  │                          │  [{ text: "날씨 알려줘" }]   │
  │                          │─────────────────────────────→│
  │                          │                              │
  │                          │     (AI 처리)                │
  │                          │                              │
  │                          │  POST /reply                 │
  │                          │←─────────────────────────────│
  │                          │                              │
  │  "서울 현재 기온 5도..."  │                              │
  │ ←────────────────────────│                              │
  │                          │                              │
```

---

## Flow 5: Unpair

사용자 또는 관리자가 연결을 해제합니다.

### User-initiated

```
User → "/unpair"
Relay → UPDATE mappings SET state = 'UNPAIRED', account_id = NULL
Relay → "연결이 해제되었습니다."
```

### Admin-initiated (API)

```
POST /openclaw/pairing/unpair
Authorization: Bearer <relay_token>
Content-Type: application/json

{
  "conversationKey": "channel_123:user_xyz"
}
```

---

## Special Commands

Relay가 인식하는 특수 명령어:

| Command | Description |
|---------|-------------|
| `/pair <code>` | 페어링 코드 입력 |
| `/unpair` | 연결 해제 |
| `/status` | 현재 연결 상태 확인 |
| `/help` | 도움말 |

### Implementation

```typescript
function parseCommand(utterance: string): Command | null {
  const trimmed = utterance.trim();
  
  if (trimmed.startsWith('/pair ')) {
    const code = trimmed.slice(6).trim().toUpperCase();
    return { type: 'PAIR', code };
  }
  
  if (trimmed === '/unpair') {
    return { type: 'UNPAIR' };
  }
  
  if (trimmed === '/status') {
    return { type: 'STATUS' };
  }
  
  if (trimmed === '/help') {
    return { type: 'HELP' };
  }
  
  return null;
}
```

---

## Security Considerations

### 1. Code Brute-Force Protection

```typescript
const pairingAttempts = new RateLimiter({
  keyPrefix: 'pairing:attempts:',
  points: 5,           // 5번 시도
  duration: 300,       // 5분 내
  blockDuration: 900,  // 초과 시 15분 차단
});

async function verifyPairingCode(userKey: string, code: string) {
  try {
    await pairingAttempts.consume(userKey);
  } catch (e) {
    throw new Error('Too many attempts. Please try again later.');
  }
  
  // ... verify code
}
```

### 2. Code Entropy

```typescript
function generatePairingCode(): string {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789'; // 혼동 문자 제외
  const part1 = Array.from({ length: 4 }, () => 
    chars[crypto.randomInt(chars.length)]
  ).join('');
  const part2 = Array.from({ length: 4 }, () => 
    chars[crypto.randomInt(chars.length)]
  ).join('');
  return `${part1}-${part2}`;
}

// 엔트로피: 32^8 = 1,099,511,627,776 조합
```

### 3. Single Use

```sql
-- 코드 사용 시 atomic update
UPDATE pairing_codes 
SET used_at = NOW(), used_by = $1
WHERE code = $2 
  AND used_at IS NULL 
  AND expires_at > NOW()
RETURNING account_id;
```

### 4. Account Code Limit

```typescript
async function generateCode(accountId: string) {
  const activeCount = await db.query(`
    SELECT COUNT(*) FROM pairing_codes 
    WHERE account_id = $1 
      AND expires_at > NOW() 
      AND used_at IS NULL
  `, [accountId]);
  
  if (activeCount >= 5) {
    throw new Error('Maximum active codes reached. Wait for expiry or delete existing codes.');
  }
  
  // ... generate new code
}
```

---

## Edge Cases

### Already Paired

사용자가 이미 페어링된 상태에서 다른 코드 입력:

```
Option 1: 거부
"이미 연결되어 있습니다. 먼저 /unpair로 연결을 해제하세요."

Option 2: 교체 (권장)
"기존 연결이 해제되고 새로운 봇에 연결되었습니다."
```

### Code Owner Mismatch

사용자 A가 자신의 코드를 사용자 B에게 공유:

- **문제 없음**: 코드는 계정에 연결, 사용자는 코드 사용
- **결과**: B가 A의 OpenClaw에 연결됨
- **주의**: 의도적인 공유인지 확인 (코드 유출 주의)

### Multiple Devices

같은 카카오 계정으로 여러 기기에서 접속:

- **해결**: `plusfriendUserKey`로 식별 (기기 무관)
- **결과**: 모든 기기에서 같은 OpenClaw로 연결
