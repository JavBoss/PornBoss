# Repository Guidelines

## Project Structure & Module Organization
- `cmd/server`: Go entrypoint wiring config, DB, thumbnail worker, and HTTP router.
- `internal/`: domain packages; key folders: `config`, `db` (GORM models/store), `scanner` (directory sync + fingerprinting), `server` (Gin handlers/router), `thumbnail` (ffmpeg-backed jobs), `logging` (shared logger).
- `web/`: React + Tailwind frontend (Vite). Components in `web/src/components`, state in `web/src/store.js`, API helpers in `web/src/api.js`.
- `scripts/`: `cli/` stores CLI sources; `scripts/cli.sh` launches the bundled CLI from `scripts/cli/build/` and auto-builds if missing.
- `data/`: runtime DB/thumbnails; keep generated files out of commits.

## Build, Test, and Development Commands
- Backend: `go run ./cmd/server -addr :17654 -static web/dist` to serve API (and built frontend when desired).
- Dev helper: `scripts/cli.sh dev backend|frontend` (flags: `WITH_STATIC=1`, `SKIP_NPM_INSTALL=1`, etc.).
- Tests: `GOCACHE=$(pwd)/.gocache go test ./...` (no Go tests yet—keep it green).
- Frontend (in `web/`): `npm install`; `npm run dev` for Vite, `npm run lint`, `npm run build` for prod bundle.
- CLI build: `cd scripts/cli && npm install && npm run build` (outputs `scripts/cli/build/pornboss-cli.cjs`).
- Release: `scripts/cli.sh release linux-x86_64 v0.1.0` builds backend + `web/dist` and archives to `release/`.

## Coding Style & Naming Conventions
- Go: run `gofmt -w` and keep imports/go mod tidy; use context as first arg, return wrapped errors with lower-case messages, and log via `internal/common/logging`. Keep package names lowercase and files scoped to their package.
- Frontend: functional React components in PascalCase (`VideoGrid.jsx`), hooks/helpers camelCase. Keep styles in Tailwind/`index.css`; prefer colocated component styles. Format with `npm run format` / `npm run format:check`; lint with `npm run lint`.
- Naming: API routes are RESTful (`/videos`, `/tags`, `/directories`); keep new endpoints consistent and document query params.

## Testing Guidelines
- Go: add table-driven `_test.go` files near the code under test; prefer integration tests around `internal/db` and handler tests via `httptest`. Use the repo-local `GOCACHE` path.
- Frontend: no unit tests today; at minimum run `npm run lint` and `npm run build` before PRs. When adding tests, colocate under `web/src` using Jest/Vitest conventions and match component names.

## Commit & Pull Request Guidelines
- Commits: short, imperative subjects (`add thumbnail retry logging`), optional scope; group related changes. Reference issues in the body when relevant.
- PRs: include intent, main changes, and tests run (`go test ./...`, `npm run lint`, `npm run build`). For UI updates, attach before/after notes or screenshots and call out config or schema impacts. Keep diffs focused; avoid committing generated assets (DB files, `.gocache`, `web/dist` unless part of release output).
