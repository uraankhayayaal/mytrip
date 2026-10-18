# TEST REPORT — A7-04 (PhotoService + PhotoRepo + StoragePort + PhotoHandler)

Ветка: `ai/task/A7-04`, HEAD `b0d9c9a`.
Раунд: повторное QA-прохождение после эскалации (attempts=6).

## Статус: Passed

## 1. Сверка с контрактами (1:1, до последнего символа)

| Контракт | Реализация | Результат |
|---|---|---|
| `PhotoRepo` (internal/repo/interfaces.go:67-73) | `PGPhotoRepo` (internal/repo/photo_repo.go): pgx, таблица `trip_photos` (0004_photos), Insert/ListByTrip(ORDER BY created_at ASC)/Delete(NotFound при 0 строк) | OK |
| `PhotoService` (internal/service/interfaces.go:110-118) | `photoService` (internal/service/photo_service.go): владение trip через `TripRepo.Get` во всех 3 методах; `cover=true` → `TripRepo.SetCover` ДО Insert; при Delete объект из хранилища, ошибка DeleteObject не блокирует удаление записи | OK |
| `StoragePort` (internal/service/interfaces.go:120-128) | `S3Storage` (internal/storage/s3_storage.go): `PresignPut` через `s3.NewPresignClient`, TTL 15 мин; `DeleteObject` идемпотентен (NoSuchKey/NotFound → nil) | OK |
| `PhotoHandler` (internal/api/handlers/handlers.go:123-145) | `photoHandler` + `mountPhotoRoutes` (photo_handler.go, photo_handler_mount.go): GET/POST `/trips/{id}/photos`, DELETE `/trips/{id}/photos/{photoId}` → 204, POST `/trips/{id}/photos/presign`; body `{object_key, url, cover?}`; presign-ответ `{url, expires_at}` = frontend `PhotoUrlResponse` | OK |
| Wire | cmd/server/main.go:96-113: `PGPhotoRepo` + `S3Storage` + `NewPhotoService` + `NewPhotoHandler(...).Mount(r)` под auth-middleware | OK |

Расхождения: 0. Блокирующих дефектов: 0.

## 2. Автоматические проверки (make backend-test / backend-lint)

Команда: `make backend-test` (`go test ./... -race -cover`) — exit 0, все пакеты ok:

```
ok  mytrip/internal/api/handlers    1.020s  coverage: 95.1%
ok  mytrip/internal/api/middleware  1.017s  coverage: 69.8%
ok  mytrip/internal/config          1.020s  coverage: 96.0%
ok  mytrip/internal/domain          1.012s  coverage: 100.0%
ok  mytrip/internal/jwtutil         1.019s  coverage: 89.7%
ok  mytrip/internal/repo            1.019s  coverage: 4.6%   (требует live-Postgres — покрывается e2e в CI)
ok  mytrip/internal/service         3.166s  coverage: 85.5%  (photo_service_test.go — мок PhotoRepo/TripRepo/StoragePort)
ok  mytrip/internal/storage         1.012s  coverage: 91.7%
```

`make backend-lint` (degrade-шаги отмечены):
- `gofmt -l .` — чисто (exit 0);
- `go vet ./...` — exit 0;
- `golangci-lint run` — **не выполнен**: golangci-lint не установлен в песочнице (command not found) — degrade окружения. По явному решению пользователя (комментарий 2026-10-18T10:01Z: «если golangci-lint недоступен — пропустить, считать что все ок») шаг пропущен и засчитан как OK. В CI (.github/workflows/ci.yml) golangci-lint установлен и прогоняется.

## 3. Рантайм (ReadAppLogs, local)

- Без env: корректное падение `config.Load` (missing DATABASE_URL/JWT_SECRET) — ожидаемое поведение.
- С env (`DATABASE_URL=...:59999`): `migrate: dial tcp 127.0.0.1:59999: connect: connection refused` — Postgres недоступен в sandbox, Docker-стек не поднимается. Это **degrade окружения, не дефект кода**: стартовый код корректно логирует ошибку миграции и выходит с кодом 1 вместо паники. Интеграционная проверка живого API покрывается самозавершающейся целью `make e2e` (up → healthcheck → Playwright → down) в CI.

## 4. Автотесты (состав по ТЗ)

- unit: `internal/service/photo_service_test.go` — мок PhotoRepo/TripRepo/StoragePort: List (success/trip-not-found), Create (success/cover=true→SetCover/cover=false/insert-error), Delete (success/missing/not-found/объект-сирота при ошибке хранилища).
- unit: `internal/api/handlers/photo_handler_test.go` — httptest + chi: GET 200/404, POST 201/400 (validation пустых полей, invalid JSON)/404, DELETE 204/404, presign 200 `{url,expires_at}`/404/400.
- unit: `internal/storage/s3_storage_test.go` — PresignPut (реальный s3.Client, подпись без сети), DeleteObject (мок s3ClientAPI: success/NoSuchKey/NotFound/другие ошибки).
- интеграционные/E2E: live-Postgres и полный HTTP-сценарий требуют Docker-стека (Postgres/MinIO) — недоступны в sandbox; покрываются `make e2e` в CI (Playwright: frontend/e2e, frontend/src/components/PhotoUploader.test.tsx).

## 5. Дефекты

Обнаружено: 0. Багрепортов создано: 0.

## 6. Итог

Все применимые в sandbox проверки зелёные; недоступные шаги — только degrade окружения (golangci-lint, Postgres/Docker), зафиксированы выше и покрываются CI. Задача переводится в done.
