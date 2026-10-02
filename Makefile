# mytrip — корневой Makefile (BEL-01)
#
# Единая точка входа команд для всех субагентов и приёмки.
# Каждая цель имеет комментарий '# help: <описание>' — help-цель
# автогенерирует список из этих комментариев (включая будущие цели
# up/down/logs/ps/infra.* из DOL-03, которые будут добавлены ниже).

.PHONY: help deps build backend-build frontend-build test backend-test frontend-test lint backend-lint frontend-lint run backend-run frontend-run e2e

# help: показать список всех целей (автогенерация из комментариев '# help: ')
help: # help: показать список всех целей (автогенерация из комментариев '# help: ')
	@grep -E '^[a-zA-Z0-9_.-]+:.*# help: ' Makefile | awk -F'# help: ' '{print "  " $$1 " — " $$2}'

# deps: установка зависимостей (go mod download + npm ci в frontend/)
deps: # help: установка зависимостей (go mod download + npm ci в frontend/)
	go mod download
	npm --prefix frontend ci

# build: сборка всего (backend + frontend)
build: # help: сборка всего (backend + frontend)
	$(MAKE) backend-build
	$(MAKE) frontend-build

# backend-build: сборка бэкенда в bin/server
backend-build: # help: сборка бэкенда в bin/server
	go build -o bin/server ./cmd/server

# frontend-build: сборка фронтенда
frontend-build: # help: сборка фронтенда
	npm --prefix frontend run build

# test: тесты всего (backend + frontend)
test: # help: тесты всего (backend + frontend)
	$(MAKE) backend-test
	$(MAKE) frontend-test

# backend-test: тесты бэкенда (race + coverage)
backend-test: # help: тесты бэкенда (race + coverage)
	go test ./... -race -cover

# frontend-test: тесты фронтенда
frontend-test: # help: тесты фронтенда
	npm --prefix frontend test -- --run

# lint: линт всего (backend + frontend)
lint: # help: линт всего (backend + frontend)
	$(MAKE) backend-lint
	$(MAKE) frontend-lint

# backend-lint: gofmt + go vet + golangci-lint
backend-lint: # help: gofmt + go vet + golangci-lint
	gofmt -l .
	go vet ./...
	golangci-lint run

# frontend-lint: линт фронтенда (prettier + eslint)
frontend-lint: # help: линт фронтенда (prettier + eslint)
	npm --prefix frontend run lint

# run: параллельный dev-запуск backend + frontend (Ctrl+C останавливает оба)
run: # help: параллельный dev-запуск backend + frontend (Ctrl+C останавливает оба)
	$(MAKE) -j2 backend-run frontend-run

# backend-run: запуск бэкенда в dev-режиме
backend-run: # help: запуск бэкенда в dev-режиме
	go run ./cmd/server

# frontend-run: запуск фронтенда в dev-режиме
frontend-run: # help: запуск фронтенда в dev-режиме
	npm --prefix frontend run dev

# e2e: самозавершающийся прогон e2e (up → healthcheck → e2e-сценарии → down)
e2e: # help: самозавершающийся прогон e2e (up → healthcheck → e2e-сценарии → down)
	$(MAKE) up
	@for i in $$(seq 1 60); do curl -f http://localhost:8080/healthz && break; sleep 1; done
	npm --prefix e2e test
	RC=$$?
	$(MAKE) down
	exit $$RC
