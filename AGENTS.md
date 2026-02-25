# Repository Guidelines

## Project Structure & Module Organization
The Go API entrypoint is `cmd/server/main.go`; most server code lives under `internal/` (handlers, middleware, services, jobs, config). The admin and portal UIs are in `admin/` and `portal/` and compile into `public/admin` and `public/portal`. Database schema and migrations are managed with Drizzle under `drizzle/` and `drizzle/migrations/` (files named like `0001_feature.sql`). Static assets are in `static/`, while compiled assets and bundles live in `public/`. Project docs and handoff notes are in `docs/` and root-level `HANDOFF*.md` files.

## Build, Test, and Development Commands
- `make docker-up` / `make docker-down`: start or stop PostgreSQL (see `docker-compose.yml`).
- `make db-migrate`, `make db-generate`, `make db-studio`: run or create Drizzle migrations.
- `make dev`: run the Bun dev server with hot reload (per `package.json`).
- `bun run build:admin`, `bun run build:portal`, `bun run build:all`: build UI bundles and server output.
- `go run ./cmd/server` or `go build ./cmd/server`: run or build the Go server directly.
- `make check`, `make lint`, `make format`: run Biome checks on the TypeScript code.

## Coding Style & Naming Conventions
Go code should remain gofmt-compliant (tabs, standard layout). TypeScript/React code is formatted and linted by Biome (`biome.json`), so run `bun run format` before pushing. Migration files use zero-padded numeric prefixes with snake-case names (e.g., `0002_portal_users.sql`). Tests follow `_test.go` for Go and `.test.ts` for TS.

## Testing Guidelines
Run Go tests with `go test ./...`. Run Bun tests with `bun test` or target UIs with `bun test admin/` and `bun test portal/`. Place new tests alongside the package they cover and keep fixtures local to the module.

## Commit & Pull Request Guidelines
Commit messages follow release-please: `type(scope): description` in lowercase with no trailing period. Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `ci`, `chore` (use `feat!`/`fix!` for breaking). Scopes: `api`, `kakao`, `admin`, `portal`, `db`, `sse`, `auth`, `deps`. PRs should include a short summary, test commands run, migration notes if applicable, and screenshots for UI changes.

## Security & Configuration Tips
Use `.env.example` as a template, keep secrets out of Git, and document any new env vars in the example file. Local Postgres defaults are set in `docker-compose.yml`; avoid hardcoding credentials in code or configs.
