# API 명세

## 인증

OpenClaw → Relay 요청 시 Bearer 토큰 인증:

```
Authorization: Bearer <relay_token>
```

또는 쿼리 파라미터:

```
?token=<relay_token>
```

토큰은 세션 페어링 완료 시 발급되며, SHA256 해시로만 DB에 저장됩니다.

---

## 공개 엔드포인트

### GET /

루트 경로. 대시보드로 리다이렉트.

**응답:** `302 Found` → `/dashboard/`

### GET /health

헬스체크.

**응답:**
```json
{
  "status": "ok",
  "timestamp": 1706700000000
}
```

### POST /kakao/webhook

카카오톡 채널 오픈빌더 스킬이 호출하는 웹훅.

**헤더:**
```
Content-Type: application/json
X-Kakao-Signature: <hmac_hex>  (KAKAO_SIGNATURE_SECRET 설정 시 필수)
```

**요청:** 카카오 SkillPayload (오픈빌더 스펙)

**응답:**
```json
{
  "version": "2.0",
  "useCallback": true
}
```

**동작:**
1. (선택) HMAC-SHA256 서명 검증
2. `plusfriendUserKey` + `channelId`로 `conversationKey` 생성
3. `conversation_mappings` 조회/생성
4. 명령어 파싱: `/pair <코드>`, `/unpair`, `/status`, `/help`
5. 페어링된 사용자 → `inbound_messages`에 저장 + SSE 발행
6. 미페어링 → 안내 응답 반환

---

## 인증 필요 엔드포인트

### GET /v1/events

SSE(Server-Sent Events) 실시간 이벤트 스트림.

**인증:** Bearer 토큰 (계정 또는 세션 토큰)

**이벤트 타입:**

| 이벤트 | 설명 |
|--------|------|
| `connected` | 연결 성공. `{ accountId, sessionId, status }` |
| `message` | 새 인바운드 메시지. `{ id, conversationKey, kakaoPayload, normalized, createdAt }` |
| `pairing_complete` | 페어링 완료. `{ conversationKey, pairedAt }` |
| `: ping` | 30초 간격 하트비트 (SSE 코멘트) |

**동작:**
- 연결 시 대기 중인 `queued` 메시지를 즉시 전달 후 `delivered`로 변경
- Redis Pub/Sub 기반으로 새 이벤트 실시간 수신

### POST /openclaw/reply

카카오 사용자에게 응답 전송.

**인증:** Bearer 토큰

**요청:**
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

**응답 (성공):**
```json
{
  "success": true,
  "deliveredAt": 1706700005000
}
```

**에러:**

| 상태 | 설명 |
|------|------|
| 401 | 유효하지 않은 토큰 |
| 403 | 다른 계정의 메시지 |
| 404 | 메시지 없음 |
| 410 | 콜백 URL 만료 (카카오 1분 제한) |

---

## 세션 엔드포인트

### POST /v1/sessions/create

새 페어링 세션 생성. 인증 불필요, IP Rate Limit 적용 (10회/5분).

**응답:**
```json
{
  "sessionToken": "64자_hex_토큰",
  "pairingCode": "ABCD-1234",
  "expiresIn": 300,
  "status": "pending_pairing"
}
```

- `sessionToken`: SSE 연결 및 상태 조회에 사용 (1회만 노출)
- `pairingCode`: 카카오에서 `/pair ABCD-1234` 입력용 (형식: XXXX-XXXX)
- 유효기간: 5분

### GET /v1/sessions/{sessionToken}/status

세션 상태 조회. IP Rate Limit 적용 (30회/분).

**쿼리:** `?token=<sessionToken>`

**응답:**
```json
{
  "status": "paired",
  "accountId": "uuid",
  "relayToken": "64자_hex_토큰",
  "pairedAt": "2025-01-31T21:00:00Z"
}
```

**status 값:** `pending_pairing`, `paired`, `expired`, `disconnected`

---

## 대시보드 엔드포인트

### GET /dashboard/

임베디드 대시보드 UI.

### GET /dashboard/api/overview

전체 통계.

**응답:**
```json
{
  "accounts": 1,
  "sessionTotal": 5,
  "sessionPending": 0,
  "sessionPaired": 1,
  "conversationPaired": 3,
  "conversationUnpaired": 1,
  "inboundTotal": 150,
  "outboundTotal": 145,
  "outboundFailed": 2,
  "sseClients": 1,
  "timestamp": 1706700000000
}
```

### GET /dashboard/api/accounts

계정 목록.

### GET /dashboard/api/accounts/{id}/conversations

계정의 대화 매핑 목록.

### GET /dashboard/api/accounts/{id}/messages

계정의 메시지 목록. `?type=outbound`로 아웃바운드 조회.

### GET /dashboard/api/accounts/{id}/stats

계정 통계 (인바운드/아웃바운드 수, 실패 수, 대화 수, SSE 클라이언트 수).

### GET /dashboard/api/accounts/{id}/failed-messages

최근 실패한 아웃바운드 메시지 (최대 50건).

### POST /dashboard/api/accounts/{id}/regenerate-token

릴레이 토큰 재발급. 기존 토큰은 즉시 무효화.

### DELETE /dashboard/api/accounts/{id}

계정 삭제.

### DELETE /dashboard/api/accounts/{id}/conversations/{convId}

대화 매핑 삭제.

### GET /dashboard/api/sessions

최근 세션 목록 (기본 50건, `?limit=N`).

### POST /dashboard/api/sessions/create

대시보드에서 세션 생성 (페어링 코드 발급).

### POST /dashboard/api/sessions/{id}/disconnect

세션 연결 해제.

### DELETE /dashboard/api/sessions/{id}

세션 삭제.

---

## Rate Limiting

| 엔드포인트 | 방식 | 한도 |
|-----------|------|------|
| `/v1/events`, `/openclaw/*` | 계정별 (인메모리) | 60 req/min (계정 설정에 따름) |
| `POST /v1/sessions/create` | IP별 (Redis) | 10 req/5min |
| `GET /v1/sessions/{token}/status` | IP별 (Redis) | 30 req/min |

**Rate Limit 응답 헤더:**
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 55
X-RateLimit-Reset: 1706700060
```

초과 시 `429 Too Many Requests`.

---

## 에러 응답 형식

```json
{
  "error": "에러 메시지"
}
```
