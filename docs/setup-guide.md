# 설정 가이드

> 이 문서는 릴레이 서버를 카카오톡 채널과 연동하기 위한 전체 설정 절차를 안내합니다.

## 목차

1. [사전 준비](#1-사전-준비)
2. [카카오톡 채널 만들기](#2-카카오톡-채널-만들기)
3. [카카오 i 오픈빌더 챗봇 만들기](#3-카카오-i-오픈빌더-챗봇-만들기)
4. [스킬 등록](#4-스킬-등록)
5. [폴백 블록에 스킬 연결](#5-폴백-블록에-스킬-연결)
6. [Callback 기능 신청](#6-callback-기능-신청)
7. [채널 연결 및 배포](#7-채널-연결-및-배포)
8. [릴레이 서버 설정](#8-릴레이-서버-설정)
9. [동작 확인](#9-동작-확인)

---

## 1. 사전 준비

| 항목 | 설명 |
|------|------|
| 카카오 계정 | [accounts.kakao.com](https://accounts.kakao.com) |
| 카카오비즈니스 가입 | [business.kakao.com](https://business.kakao.com) |
| 릴레이 서버 배포 완료 | 공인 IP 또는 도메인 필요 (HTTPS 권장) |

> 오픈빌더는 OBT(Open Beta Test) 신청이 필요할 수 있습니다.
> 신청: [i.kakao.com](https://i.kakao.com/) → 승인까지 약 1~6일 소요

---

## 2. 카카오톡 채널 만들기

이미 카카오톡 채널이 있으면 [3단계](#3-카카오-i-오픈빌더-챗봇-만들기)로 건너뛰세요.

1. [카카오비즈니스](https://business.kakao.com)에 로그인
2. **[+ 시작하기]** 클릭
3. 매장 보유 여부 선택 → 사업자 등록번호 입력 (없으면 일반 채널)
4. 채널 프로필 설정:
   - 프로필 이미지 (640x640px 권장)
   - 채널 이름
   - 카테고리 선택
   - 검색용 ID 생성
5. 생성 완료 후:
   - **채널 공개하기** 활성화
   - **검색 허용하기** 활성화

---

## 3. 카카오 i 오픈빌더 챗봇 만들기

1. [카카오 i 오픈빌더](https://i.kakao.com/)에 접속 → 로그인
2. **[서비스/도구] → [챗봇] → [내 챗봇]**
3. **[+ 봇 만들기] → [카카오톡 챗봇]** 선택
4. 봇 이름 입력 → **[만들기]**

생성된 챗봇에는 기본 블록이 자동으로 포함됩니다:
- **웰컴 블록**: 사용자가 처음 채팅방에 들어올 때
- **폴백 블록**: 어떤 시나리오에도 매칭되지 않는 입력을 처리
- **탈출 블록**: 대화 흐름 이탈 시

---

## 4. 스킬 등록

릴레이 서버의 웹훅 엔드포인트를 오픈빌더 스킬로 등록합니다.

1. 챗봇 선택 → 좌측 메뉴 **[스킬]**
2. **[+ 생성]** 클릭
3. 아래 정보 입력:

| 항목 | 값 |
|------|-----|
| 스킬명 | `릴레이 서버 웹훅` (자유 입력) |
| URL | `https://{YOUR_RELAY_SERVER}/kakao-talkchannel/webhook` |

4. **[저장]**

> 헤더 설정은 필요 없습니다. 카카오 서명 검증을 사용하려면 서버 측 `KAKAO_SIGNATURE_SECRET` 환경변수를 설정하세요.

---

## 5. 폴백 블록에 스킬 연결

모든 사용자 메시지가 릴레이 서버를 거치도록 폴백 블록에 스킬을 연결합니다.

1. 좌측 메뉴 **[시나리오]**
2. **[폴백 블록]** 클릭
3. **파라미터 설정** 섹션:
   - 스킬 선택 드롭다운에서 **[릴레이 서버 웹훅]** 선택
4. **봇 응답** 섹션:
   - **[응답 추가]** → **[스킬 데이터]** 선택
5. **[저장]**

```
┌────────────────────────────────────────┐
│           폴백 블록 설정 화면            │
│                                        │
│  파라미터 설정                           │
│  ┌──────────────────────────────────┐  │
│  │ 스킬 선택: [릴레이 서버 웹훅 ▾]   │  │
│  └──────────────────────────────────┘  │
│                                        │
│  봇 응답                                │
│  ┌──────────────────────────────────┐  │
│  │ 응답 유형: [스킬 데이터 ▾]        │  │
│  └──────────────────────────────────┘  │
│                                        │
│               [저장]                    │
└────────────────────────────────────────┘
```

> **중요**: 폴백 블록에 연결해야 사용자가 입력하는 **모든 메시지**가 릴레이 서버로 전달됩니다.
> `/pair`, `/unpair`, `/status` 등의 명령어도 폴백 블록을 통해 릴레이 서버에서 처리합니다.

---

## 6. Callback 기능 신청

릴레이 서버는 `useCallback: true` 응답을 사용하여 비동기로 응답합니다.
이를 위해 **AI 챗봇 Callback 기능**을 신청해야 합니다.

### 6-1. Callback 권한 신청

1. **[설정] → [AI 챗봇 관리]**
2. Callback 기능 사용 신청
3. 사용 목적 작성 (예: "외부 AI 서비스 연동을 위한 비동기 응답 처리")
4. 승인까지 1~2 영업일 소요

### 6-2. 블록에 Callback 활성화

권한 승인 후:

1. **[시나리오] → [폴백 블록]** 다시 열기
2. 블록 상세 설정에서 **Callback 사용** 활성화
3. **[저장]**

### Callback 동작 흐름

```
사용자 메시지
    │
    ▼
카카오 → 릴레이 서버: POST /kakao-talkchannel/webhook
    │                  (callbackUrl 포함)
    ▼
릴레이 서버 → 카카오: { "version": "2.0", "useCallback": true }
    │                  (5초 내 즉시 응답)
    ▼
카카오: "처리 중" 안내 표시 (블록에 설정한 기본 응답)
    │
    ▼
OpenClaw: AI 처리 완료 후 릴레이 서버에 응답 전달
    │
    ▼
릴레이 서버 → 카카오: POST callbackUrl (최종 응답)
    │                  (callbackUrl 유효시간: 1분)
    ▼
카카오 → 사용자: AI 응답 표시
```

> **callbackUrl 유효시간은 1분**입니다. 1분 내에 응답하지 못하면 사용자에게 전달되지 않습니다.

---

## 7. 채널 연결 및 배포

### 7-1. 카카오톡 채널 연결

1. 챗봇 관리자센터에서 **[설정] → [카카오톡 채널 연결]**
2. 2단계에서 만든 채널 선택 → **[연결]**

### 7-2. 배포

1. 좌측 메뉴 **[배포]**
2. 변경 내역 확인 후 **[배포]** 클릭

> 배포하지 않으면 변경사항이 실제 채널에 반영되지 않습니다.
> 스킬 URL 변경, 블록 수정 등 모든 변경 후 반드시 재배포하세요.

---

## 8. 릴레이 서버 설정

### 8-1. 환경변수

```bash
# 필수
DATABASE_URL=postgresql://...
REDIS_URL=redis://...
ADMIN_PASSWORD=<강력한_비밀번호>
ADMIN_SESSION_SECRET=<openssl rand -base64 32>
PORTAL_SESSION_SECRET=<openssl rand -base64 32>

# 선택
KAKAO_SIGNATURE_SECRET=<카카오_서명_검증_키>
PORTAL_BASE_URL=https://{YOUR_RELAY_SERVER}
```

### 8-2. Account 생성

Admin UI(`https://{YOUR_RELAY_SERVER}/admin/`)에서 OpenClaw 인스턴스용 계정을 생성합니다.

**Admin UI 사용:**
1. Admin UI 접속 → 비밀번호 입력
2. Accounts → Create Account
3. `relayToken` 안전하게 저장 (1회만 표시)

**API 사용:**
```bash
curl -X POST https://{YOUR_RELAY_SERVER}/admin/api/accounts \
  -H "Content-Type: application/json" \
  -d '{"openclawUserId": "my-openclaw-instance"}'
```

### 8-3. OpenClaw 연동

OpenClaw 측에서 아래 환경변수를 설정합니다:

```env
KAKAO_RELAY_URL=https://{YOUR_RELAY_SERVER}
KAKAO_RELAY_TOKEN=<발급받은_relay_token>
```

연동 API에 대한 자세한 내용은 [연동 가이드](integration-guide.md)와 [API 스펙](api-spec.md)을 참고하세요.

---

## 9. 동작 확인

### 9-1. 페어링 테스트

1. OpenClaw에서 페어링 코드 생성:
   ```bash
   curl -X POST https://{YOUR_RELAY_SERVER}/openclaw/pairing/generate \
     -H "Authorization: Bearer <relay_token>" \
     -H "Content-Type: application/json" \
     -d '{"expirySeconds": 600}'
   ```
2. 카카오톡에서 채널 채팅창 열기
3. `/pair ABCD-1234` 입력 (발급받은 코드)
4. "OpenClaw에 연결되었습니다!" 확인

### 9-2. 메시지 흐름 테스트

1. 카카오톡 채팅창에서 아무 메시지 입력
2. OpenClaw에서 메시지 수신 확인 (SSE 또는 polling)
3. OpenClaw에서 응답 전송
4. 카카오톡에서 응답 수신 확인

### 9-3. 명령어 테스트

| 명령어 | 예상 결과 |
|--------|----------|
| `/pair <코드>` | "OpenClaw에 연결되었습니다!" |
| `/status` | 현재 연결 상태 표시 |
| `/unpair` | "연결이 해제되었습니다" |
| `/code` | 포털 접속 코드 발급 |
| `/help` | 도움말 표시 |

---

## 문제 해결

### 스킬 서버 연결 오류

- 릴레이 서버가 **HTTPS**로 접근 가능한지 확인
- 오픈빌더 스킬 URL이 정확한지 확인: `/kakao-talkchannel/webhook`
- 서버 로그에서 웹훅 수신 여부 확인

### Callback 응답이 전달되지 않음

- **Callback 기능 승인** 여부 확인
- 폴백 블록에서 **Callback 사용** 활성화 여부 확인
- callbackUrl 유효시간(1분) 내에 응답하는지 확인
- 릴레이 서버 `CALLBACK_TTL_SECONDS` 설정 확인 (기본 55초)

### 페어링이 안 됨

- `/pair` 명령이 폴백 블록을 통해 릴레이 서버에 전달되는지 확인
- 페어링 코드 유효시간(기본 10분) 확인
- 대소문자 구분 없음

### Health Check

```bash
curl https://{YOUR_RELAY_SERVER}/health
```

정상 응답:
```json
{"status":"ok","timestamp":1706700000000}
```

---

## 참고 문서

- [카카오 비즈니스 가이드 - 채널 만들기](https://kakaobusiness.gitbook.io/main/channel/start)
- [카카오 비즈니스 가이드 - 스킬 개발 가이드](https://kakaobusiness.gitbook.io/main/tool/chatbot/skill_guide/make_skill)
- [카카오 비즈니스 가이드 - AI 챗봇 콜백 가이드](https://kakaobusiness.gitbook.io/main/tool/chatbot/skill_guide/ai_chatbot_callback_guide)
- [카카오 비즈니스 가이드 - 응답 타입별 JSON 포맷](https://kakaobusiness.gitbook.io/main/tool/chatbot/skill_guide/answer_json_format)
