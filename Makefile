.PHONY: help test test-v test-integration test-e2e coverage fmt fmt-check vet lint lint-install check build run clean deps mod-tidy mocks ci pre-commit docker-up docker-down docker-restart docker-logs docker-ps docker-build docker-clean

BINARY_NAME = pr-reviewer-service

.DEFAULT_GOAL := help

help:
	@echo "Доступные команды:"
	@echo "  make test            - Запустить все unit тесты"
	@echo "  make test-v          - Тесты с подробным выводом"
	@echo "  make test-integration - Запустить интеграционные тесты"
	@echo "  make test-e2e        - Запустить E2E тесты"
	@echo "  make coverage        - Показать покрытие тестами"
	@echo "  make fmt             - Форматировать код"
	@echo "  make fmt-check       - Проверить форматирование"
	@echo "  make vet             - Запустить go vet"
	@echo "  make lint            - Запустить golangci-lint"
	@echo "  make lint-install    - Установить golangci-lint"
	@echo "  make check           - Все проверки (fmt, vet, lint, test)"
	@echo "  make build           - Собрать бинарный файл"
	@echo "  make run             - Запустить приложение"
	@echo "  make clean           - Удалить сгенерированные файлы"
	@echo "  make deps            - Установить зависимости"
	@echo "  make mod-tidy        - Очистить go.mod"
	@echo "  make mocks           - Генерация моков"
	@echo "  make ci              - CI pipeline (fmt, vet, lint, test, coverage)"
	@echo "  make pre-commit      - Проверки перед коммитом (fmt, vet, test)"
	@echo "  make load-test       - Запустить нагрузочное тестирование с K6"
	@echo "  make k6-install      - Установить K6 (если не установлен)"
	@echo ""
	@echo "Docker команды:"
	@echo "  make docker-up       - Запустить Docker Compose (приложение + БД)"
	@echo "  make docker-down     - Остановить Docker Compose"
	@echo "  make docker-restart  - Перезапустить Docker Compose"
	@echo "  make docker-build    - Пересобрать образы"
	@echo "  make docker-logs     - Показать логи контейнеров"
	@echo "  make docker-ps       - Показать статус контейнеров"
	@echo "  make docker-clean    - Остановить и удалить контейнеры, volumes, сети"

test:
	go test -timeout 30s ./internal/...

test-v:
	go test -v -timeout 30s ./internal/...

test-integration:
	@echo "Запуск интеграционных тестов..."
	@cd deployments && docker-compose -f docker-compose.test.yml up -d
	@sleep 3
	@go test -v -timeout 60s ./tests/integration/... || (cd deployments && docker-compose -f docker-compose.test.yml down -v && exit 1)
	@cd deployments && docker-compose -f docker-compose.test.yml down -v

test-e2e:
	@echo "Запуск E2E тестов..."
	@cd deployments && docker-compose -f docker-compose.e2e.yml up -d
	@sleep 5
	@go test -v -timeout 120s ./tests/e2e/... || (cd deployments && docker-compose -f docker-compose.e2e.yml down -v && exit 1)
	@cd deployments && docker-compose -f docker-compose.e2e.yml down -v

coverage:
	@go test -coverprofile=coverage.out -covermode=atomic \
		./internal/delivery/http/handler/... \
		./internal/delivery/http/middleware/... \
		./internal/delivery/http/presenter/... \
		./internal/delivery/http/validator/... \
		./internal/domain/entity/... \
		./internal/usecase/... 2>&1 | grep -v "/dto" | grep "coverage:" || true
	@echo ""
	@go tool cover -func=coverage.out 2>&1 | grep "total:" || go tool cover -func=coverage.out 2>&1 | tail -1
	@rm -f coverage.out

mocks:
	@go install go.uber.org/mock/mockgen@latest
	@mockgen -package=mocks -destination=internal/domain/repository/mocks/user_repository_mock.go github.com/exPriceD/pr-reviewer-service/internal/domain/repository UserRepository
	@mockgen -package=mocks -destination=internal/domain/repository/mocks/team_repository_mock.go github.com/exPriceD/pr-reviewer-service/internal/domain/repository TeamRepository
	@mockgen -package=mocks -destination=internal/domain/repository/mocks/pull_request_repository_mock.go github.com/exPriceD/pr-reviewer-service/internal/domain/repository PullRequestRepository
	@mockgen -package=mocks -destination=internal/domain/transaction/mocks/manager_mock.go github.com/exPriceD/pr-reviewer-service/internal/domain/transaction Manager
	@mockgen -package=mocks -destination=internal/domain/logger/mocks/logger_mock.go github.com/exPriceD/pr-reviewer-service/internal/domain/logger Logger

fmt:
	gofmt -w -s .

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Код не отформатирован. Запустите: make fmt" && gofmt -l . && exit 1)

vet:
	go vet ./...

lint-install:
	@which golangci-lint > /dev/null || (echo "Установка golangci-lint..." && \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v2.6.2)

lint: lint-install
	@golangci-lint run --config golangci.yml ./...

check: fmt-check vet lint test

build:
	go build -o bin/$(BINARY_NAME) ./cmd/api

run:
	go run ./cmd/api/main.go

clean:
	rm -rf bin/

deps:
	go mod download
	go mod verify

mod-tidy:
	go mod tidy

ci: fmt-check vet lint test coverage

pre-commit: fmt-check vet test

k6-install:
	@which k6 > /dev/null || (echo "Установите K6 вручную с https://github.com/grafana/k6/releases"; exit 1)

load-test: k6-install
	@echo "Запуск нагрузочного тестирования..."
	@echo "Убедитесь, что сервер запущен на http://localhost:8080"
	@echo "Или установите BASE_URL через переменную окружения: BASE_URL=http://localhost:8080 make load-test"
	@k6 run loadtest/k6/k6_load_test.js

docker-up:
	@echo "Запуск Docker Compose..."
	@docker-compose up -d
	@echo "Ожидание готовности сервисов..."
	@echo "Проверьте статус: make docker-ps"
	@echo "Проверьте логи: make docker-logs"
	@docker-compose ps

docker-down:
	@echo "Остановка Docker Compose..."
	@docker-compose down

docker-restart:
	@echo "Перезапуск Docker Compose..."
	@docker-compose restart

docker-build:
	@echo "Пересборка Docker образов..."
	@docker-compose build --no-cache

docker-logs:
	@docker-compose logs -f

docker-ps:
	@docker-compose ps

docker-clean:
	@echo "Остановка и удаление контейнеров, volumes и сетей..."
	@docker-compose down -v --remove-orphans
