# Changelog

## [0.3.0](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/compare/kakao-talkchannel-relay-openclaw-v0.2.0...kakao-talkchannel-relay-openclaw-v0.3.0) (2026-02-10)


### ⚠ BREAKING CHANGES

* **security:** ADMIN_PASSWORD 환경변수가 ADMIN_PASSWORD_HASH로 변경됨. bcrypt 해시를 생성하려면: go run scripts/hash-password.go <password>

### Bug Fixes

* **security:** address medium+ security vulnerabilities ([#7](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/issues/7)) ([0043154](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/004315487cee258d0f67de5082cf993e9741e6ad))

## [0.2.0](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/compare/kakao-talkchannel-relay-openclaw-v0.1.0...kakao-talkchannel-relay-openclaw-v0.2.0) (2026-02-08)


### Features

* **admin:** add Admin/Portal APIs and SPA serving ([f2580e3](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/f2580e3c2e509c3fa0f3acdc1f16a00cce6423d4))
* **admin:** add mappings, messages, and users management ([5fbd6fc](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/5fbd6fc259808d4954ad6917b73b1dfea10337b3))
* **admin:** add plugin sessions management page ([20ed3bc](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/20ed3bce13e1e8750c6b71f66377643fba500d02))
* **api:** add cleanup jobs and Fly.io config ([6aefbe3](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/6aefbe3cab2a01dcb85aeaf7a468ae136c32bc00))
* **api:** add Go server scaffolding ([9fddc6e](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/9fddc6e538d6da1439a10e64db2427eb11922bd8))
* **api:** add OpenClaw API handlers ([62d2ce6](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/62d2ce649662409557b0652b36e616c687e2e2b3))
* **api:** migrate server from TypeScript/Bun to Go + SSE ([fce0935](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/fce0935a533dcbe1d5bd8a45d2bca0280222814a))
* **auth:** add middleware layer ([5fc5d2a](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/5fc5d2a87c88b226481853b3303c6b4a9c362a3b))
* **db:** add models and repository layer ([4196022](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/4196022c1c5a8114b3fcae6599e8d1ad66a677e2))
* **docker:** add app service to docker-compose ([42e6a8a](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/42e6a8a0d3bf97d73f7b762abc6dd8b2b3c44b77))
* **kakao:** add webhook handler and services ([f208526](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/f20852693f5360e69afc1f2feb3b0bd336adac3d))
* **oauth:** add social login with Google and Twitter ([5aa3809](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/5aa38090f307cfd2b9602575b01cd920670e2d2c))
* **portal:** add user stats dashboard with message and connection metrics ([64adb66](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/64adb6604f0228fd85cbd868dc09c35fcebaa2bb))
* **portal:** implement missing portal API endpoints ([ff6a616](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/ff6a61668163374c3bfc5da8942546c1b56b90dc))
* **portal:** implement missing portal API endpoints ([1a419a3](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/1a419a38676ddd2b2db5c492efa434f3e9c7104b))
* **security:** add Redis rate limiting, CSP headers, and audit logging ([4dd15b8](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/4dd15b8f8709eeb7d1e5d3adf272ff53aec23bbb))
* **session:** add auto-pairing session API for simplified relay mode ([0dcf1e1](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/0dcf1e1b36ae5d6b1a6e61bbc05bf5cceb1fb1a7))
* **sse:** add SSE broker with Redis Pub/Sub ([0882d3e](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/0882d3eb1c5fdebba7c0356cb444524001dbabae))


### Bug Fixes

* add Dockerfile HEALTHCHECK, remove redundant fly.toml env ([f8beb82](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/f8beb82ecfe4bf63c4a17d4a3d1de9f867680a23))
* add trailing slash redirect for /admin and /portal ([e15b677](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/e15b67749b601ad202b31cf08781b4c0c35005a0))
* add trailing slash redirect in SPA handler ([7f59bba](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/7f59bba88e0b45c5bdc3a8bddb319f8a6f1ed1b6))
* **auth:** add session token support and change SSE endpoint to /v1/events ([583fe21](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/583fe212c38d8bead0f3a21f03cb9dbd3933b5a9))
* copy public/ to static/ in Dockerfile for SPA serving ([08c196d](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/08c196d9fcba33ae2111f4e26bcee55483b88cf9))
* **deploy:** add VPC connector for Memorystore Redis ([e585b6c](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/e585b6c86d7006573307c3291a2bb058811becea))
* **deploy:** match existing Cloud Run config, add Redis secret ([cf13105](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/cf1310569a63bc29fc8e8791d8d2a25e8f185004))
* include public/ in gcloud upload, exclude src/ and drizzle/ ([2b5bda4](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/2b5bda408c4a9396157ecc48f53ebab24a62a037))
* **portal:** suppress 401 console error on unauthenticated /api/me calls ([cf1f6a2](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/cf1f6a2b8f046796b7d884bce1190b6738220657))
* **security:** implement comprehensive security hardening ([c911a72](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/c911a72ce9efaf4b169c1510699f9cd77e5d0985)), closes [#4](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/issues/4)
* use Chi URLParam for static file path resolution ([36fdfc8](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/36fdfc8c4fbb72a537ec2663d6c16f3dec08852c))
* use NotFound handler for SPA to prevent API route conflicts ([7941897](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/7941897a39cd13ff67186d6a99e24869c05fe9b6))
* use root-level patterns in gcloudignore to include public/ ([65b6c84](https://github.com/kakao-bart-lee/kakao-talkchannel-relay-openclaw/commit/65b6c844cecd52bf95861b25be601aec51c20795))
