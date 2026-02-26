# 시스템 아키텍처

## 개요

kakao-relay는 카카오톡 채널과 OpenClaw 인스턴스 사이의 양방향 메시지 릴레이 서버입니다. 카카오 웹훅을 수신하고, SSE로 실시간 전달하며, OpenClaw의 응답을 카카오 콜백 URL로 프록시합니다.

```
┌──────────────┐     ┌──────────────────────┐     ┌──────────────┐
│  카카오톡     │     │     kakao-relay      │     │   OpenClaw   │
│  채널 사용자  │◀───▶│                      │◀───▶│   인스턴스   │
└──────────────┘     │  ┌────────────────┐  │     └──────────────┘
                     │  │  PostgreSQL    │  │
      Webhook POST ──▶  │  Redis Pub/Sub │  ◀── SSE /v1/events
      Callback POST ◀── └────────────────┘  ◀── POST /openclaw/reply
                     └──────────────────────┘
```

---

## 메시지 흐름

### 인바운드 (카카오 → OpenClaw)

```
1. 카카오 → POST /kakao/webhook
   ├─ 서명 검증 (HMAC-SHA256, 선택)
   ├─ conversationKey 생성: ${channelId}:${plusfriendUserKey}
   ├─ conversation_mappings 조회/업데이트
   └─ 명령어 파싱 (/pair, /unpair, /status, /help)

2. 페어링된 사용자 → 메시지 큐잉
   ├─ inbound_messages INSERT (status: queued)
   ├─ SSE 브로커로 발행 (Redis Pub/Sub)
   └─ 카카오에 즉시 응답: { "version": "2.0", "useCallback": true }

3. OpenClaw ← SSE /v1/events
   ├─ 연결 시 대기 메시지 즉시 전달 (queued → delivered)
   ├─ 새 메시지 실시간 수신
   └─ 30초 하트비트로 연결 유지
```

### 아웃바운드 (OpenClaw → 카카오)

```
1. OpenClaw → POST /openclaw/reply
   ├─ 토큰 인증 → accountId 확인
   ├─ messageId로 인바운드 메시지 조회
   ├─ 테넌트 격리 검증 (message.accountId == requester.accountId)
   └─ 콜백 URL 만료 확인

2. 카카오 콜백 전송
   ├─ URL 검증: HTTPS + 카카오 도메인만 허용
   ├─ POST callbackUrl (5초 타임아웃)
   ├─ 성공 → outbound_messages status: sent
   └─ 실패 → outbound_messages status: failed + error_message
```

---

## 세션 기반 페어링

기존 `pairing_codes` 테이블 기반에서 `sessions` 테이블 기반으로 변경되었습니다.

### 흐름

```
1. 대시보드 또는 API에서 세션 생성
   POST /v1/sessions/create
   └─ 발급: sessionToken (64자 hex) + pairingCode (XXXX-XXXX)
   └─ 유효기간: 5분

2. OpenClaw가 SSE 연결 (세션 토큰으로)
   GET /v1/events?token=<sessionToken>
   └─ status: pending_pairing 상태로 대기

3. 카카오 사용자가 "/pair ABCD-1234" 입력
   └─ 웹훅에서 pairingCode 매칭
   └─ 트랜잭션:
      ├─ Account 생성 (relay_token_hash)
      ├─ Session 업데이트 (status: paired, account_id)
      └─ ConversationMapping 업데이트 (state: paired, account_id)

4. SSE로 pairing_complete 이벤트 전달
   └─ OpenClaw가 relayToken 수신, 이후 인증에 사용
```

### 페어링 코드 보안

- **문자셋**: `ABCDEFGHJKLMNPQRSTUVWXYZ23456789` (혼동 문자 I, L, O, 0, 1 제외)
- **엔트로피**: 32^8 ≈ 1.1조 조합
- **TTL**: 5분
- **일회용**: 사용 후 세션이 paired로 전환
- **Rate Limit**: IP당 10회/5분

### 사용자 명령어

카카오 채팅에서 사용 가능:

| 명령어 | 설명 |
|--------|------|
| `/pair <코드>` | OpenClaw에 연결 |
| `/unpair` | 연결 해제 |
| `/status` | 현재 연결 상태 확인 |
| `/help` | 도움말 |

---

## 데이터베이스 스키마

서버 시작 시 `internal/database/schema.sql`이 자동 실행됩니다 (`CREATE TABLE IF NOT EXISTS`).

### accounts

OpenClaw 인스턴스(봇 오너)를 나타냄.

| 컬럼 | 타입 | 설명 |
|------|------|------|
| id | uuid PK | |
| relay_token_hash | text | SHA256 해시 (평문 미저장) |
| rate_limit_per_minute | int (기본 60) | 분당 요청 한도 |
| created_at | timestamptz | |
| updated_at | timestamptz | |

### inbound_messages

카카오 → OpenClaw 메시지 큐.

| 컬럼 | 타입 | 설명 |
|------|------|------|
| id | uuid PK | |
| account_id | uuid FK | accounts(id) CASCADE |
| conversation_key | text | `${channelId}:${userKey}` |
| kakao_payload | jsonb | 카카오 원본 페이로드 |
| normalized_message | jsonb | `{ userId, text, channelId }` |
| callback_url | text | 카카오 콜백 URL |
| callback_expires_at | timestamptz | 카카오 제한: 60초 |
| status | enum | queued → delivered → acked / expired |
| source_event_id | text UNIQUE | 멱등성 키 |
| created_at | timestamptz | |
| delivered_at | timestamptz | |
| acked_at | timestamptz | |

### outbound_messages

OpenClaw → 카카오 응답 기록.

| 컬럼 | 타입 | 설명 |
|------|------|------|
| id | uuid PK | |
| account_id | uuid FK | accounts(id) CASCADE |
| inbound_message_id | uuid FK | inbound_messages(id) SET NULL |
| conversation_key | text | |
| kakao_target | jsonb | |
| response_payload | jsonb | 카카오 응답 포맷 |
| status | enum | pending → sent / failed |
| error_message | text | 실패 시 에러 메시지 |
| created_at | timestamptz | |
| sent_at | timestamptz | |

### conversation_mappings

카카오 사용자 ↔ 계정 매핑.

| 컬럼 | 타입 | 설명 |
|------|------|------|
| id | uuid PK | |
| conversation_key | text UNIQUE | `${channelId}:${userKey}` |
| kakao_channel_id | text | |
| plusfriend_user_key | text | |
| account_id | uuid FK | accounts(id) SET NULL |
| state | enum | unpaired / pending / paired / blocked |
| last_callback_url | text | 최근 웹훅의 콜백 URL |
| last_callback_expires_at | timestamptz | |
| first_seen_at | timestamptz | |
| last_seen_at | timestamptz | |
| paired_at | timestamptz | |

### sessions

페어링 세션.

| 컬럼 | 타입 | 설명 |
|------|------|------|
| id | uuid PK | |
| session_token_hash | text UNIQUE | SHA256 해시 |
| pairing_code | text UNIQUE | `XXXX-XXXX` 형식 |
| status | enum | pending_pairing / paired / expired / disconnected |
| account_id | uuid FK | 페어링 완료 시 설정 |
| paired_conversation_key | text | |
| metadata | jsonb | relay token 등 |
| expires_at | timestamptz | 기본 5분 |
| paired_at | timestamptz | |
| created_at | timestamptz | |
| updated_at | timestamptz | |

---

## 미들웨어

| 미들웨어 | 적용 범위 | 설명 |
|---------|----------|------|
| RequestID, RealIP | 전체 | Chi 내장 |
| RequestLogger | 전체 | Zerolog 구조화 로깅 |
| Recoverer | 전체 | 패닉 복구 |
| Timeout (60초) | 전체 | 요청 타임아웃 |
| BodyLimit | 전체 | 요청 본문 크기 제한 |
| KakaoSignature | `/kakao/*` | HMAC-SHA256 서명 검증 |
| Auth | `/v1/*`, `/openclaw/*` | Bearer 토큰 → 계정/세션 인증 |
| RateLimit (계정) | `/v1/*`, `/openclaw/*` | 인메모리, 분당 한도 |
| IPRateLimit (Redis) | `/v1/sessions/*` | Redis Sorted Set, 슬라이딩 윈도우 |

---

## 백그라운드 작업

### CleanupJob (5분 간격)

1. **만료 메시지 처리**: `callback_expires_at < NOW()` → status: expired
2. **오래된 메시지 삭제**: `created_at < NOW() - 7일` → 하드 삭제
3. **만료 세션 삭제**: `expires_at < NOW()` → 삭제

---

## 보안

### 토큰 관리
- SHA256 해시만 DB 저장. 평문은 생성 시 1회만 노출
- `util.GenerateToken()`: 32바이트 랜덤 → hex 인코딩 (64자)
- 재발급 시 기존 해시 즉시 교체

### 서명 검증
- `X-Kakao-Signature` 헤더로 HMAC-SHA256 검증
- `crypto/subtle.ConstantTimeCompare`로 타이밍 공격 방지
- `KAKAO_SIGNATURE_SECRET` 미설정 시 경고 로그 후 통과

### 콜백 URL 검증
- HTTPS 프로토콜 필수
- 허용 도메인: `*.kakao.com`, `*.kakaocdn.net`, `*.kakaoenterprise.com`
- 5초 타임아웃

### 테넌트 격리
- 모든 핸들러에서 `accountId` 기반 접근 제어
- 다른 계정의 메시지 접근 시 403 Forbidden

---

## SSE 브로커

Redis Pub/Sub 기반의 실시간 이벤트 분배.

- **채널 단위**: accountId 또는 `session:{sessionId}`
- **구독**: SSE 클라이언트 연결 시 Redis 채널 구독
- **발행**: 웹훅 수신/페어링 완료 시 Redis로 발행
- **하트비트**: 30초 간격 `: ping\n\n` 전송
- **클라이언트 관리**: 연결/해제 시 자동 구독/구독해제
