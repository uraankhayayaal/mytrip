# mytrip — корневой Makefile (BEL-01)

Единая точка входа команд для всех субагентов и приёмки. Запуск: `make help`.

## Цели

- `help` — автогенерация списка целей из комментариев `# help: ...` (подхватывает и будущие цели DOL-03: up/down/logs/ps/infra.*).
- `deps` — установка зависимостей: `go mod download && npm --prefix frontend ci`.
- `build` — сборка всего: `backend-build` + `frontend-build`.
- `backend-build` — `go build -o bin/server ./cmd/server`.
- `frontend-build` — `npm --prefix frontend run build`.
- `test` — `backend-test` + `frontend-test`.
- `backend-test` — `go test ./... -race -cover`.
- `frontend-test` — `npm --prefix frontend test -- --run`.
- `lint` — `backend-lint` + `frontend-lint`.
- `backend-lint` — `gofmt -l . && go vet ./... && golangci-lint run`.
- `frontend-lint` — `npm --prefix frontend run lint`.
- `run` — параллельный dev-запуск: `backend-run` + `frontend-run` (Ctrl+C останавливает оба).
- `backend-run` — `go run ./cmd/server`.
- `frontend-run` — `npm --prefix frontend run dev`.
- `e2e` — самозавершающийся прогон: `up` → healthcheck `http://localhost:8080/healthz` (до 60s) → `npm --prefix e2e test` → `down` → exit code прогона. Цели `up`/`down` добавит DOL-03 (ARCH-02).

## План работ

- [x] BEL-01 — корневой Makefile: help/deps/build/test/lint/run/e2e + backend-*/frontend-* (завершено).
- [x] BEL-03 — SQL-миграции golang-migrate в db/migrations/: 0001_init (users, refresh_tokens, pgcrypto), 0002_trips, 0003_stops, 0004_photos (up/down). Проверено: все 5 таблиц создаются без ошибок, down-миграции откатывают схему.
- [ ] DOL-03 (ARCH-02) — цели up/down/logs/ps/infra.* (будут ДОБАВЛЕНЫ в этот же Makefile, help подхватит автоматически).
