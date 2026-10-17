# TEST REPORT — A7-02 (TripService + TripRepo + TripHandler), повторная приёмка

Задача: A7-02, epic CHAT-07, ветка ai/task/A7-02, HEAD d1f9dce.

## Статус: FAILED (блок: сборка пакета internal/api/handlers)

## Блокирующий дефект (багрепорт BUG-A7-02-01)

`make backend-test` не проходит: пакет internal/api/handlers не компилируется.

```
# mytrip/internal/api/handlers
internal/api/handlers/errors.go:10:6: toAppError redeclared in this block
        internal/api/handlers/auth_handler.go:164:6: other declaration of toAppError
FAIL	mytrip/cmd/server [build failed]
FAIL	mytrip/internal/api/handlers [build failed]
```

- Выход: make backend-test → Error 1 (exit 2).
- Причина: коммит d1f9dce («wip: раунд 6», trip_handler.go + errors.go) добавил
  internal/api/handlers/errors.go с функцией toAppError, которая уже объявлена
  в internal/api/handlers/auth_handler.go:164-170 (идентичная сигнатура/логика).
- go build ./... — тоже падает на том же редекларировании; make backend-lint
  (gofmt/go vet) не запускается до зелёной сборки.
- Из-за падения сборки unit-тесты TripHandler (trip_handler_test.go) не
  выполняются вообще — приёмка по критериям 4 и 5 невозможна.

## Что проверено (до блока)

1. Сверка контрактов: TripService/TripPayload (service/interfaces.go), TripRepo
   (repo/interfaces.go), маршруты TripHandler (handlers.go) — сигнатуры в
   trip_service.go / trip_repo.go / trip_handler.go 1:1 с контрактами.
   Валидация title non-blank и end_date >= start_date → domain.NewValidation;
   чужие/нет → domain.NewNotFound (404); DELETE → 204; пустой список → [].
   Расхождений с контрактами не найдено.
2. Unit-тесты internal/service: зелёные (make backend-test, пакет ok,
   coverage 82.1%).
3. Runtime: Postgres в sandbox не поднят; интеграция через make e2e в CI
   (недоступно в песочнице — degradable шаг).

## Как снять блок

Удалить дублирующий internal/api/handlers/errors.go (функция уже есть в
auth_handler.go:164), затем: make backend-test && make backend-lint.

Задача возвращена в in_progress; багрепорт: BUG-A7-02-01.
