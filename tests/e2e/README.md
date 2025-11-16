# E2E (End-to-End) тесты

E2E тесты проверяют полные бизнес-сценарии работы приложения с реальным HTTP сервером и базой данных.

## Отличие от интеграционных тестов

- **Интеграционные тесты** (`tests/integration/`):
  - Используют `httptest.Server` (in-memory HTTP сервер)
  - Тестируют отдельные endpoints
  - Быстрее выполняются

- **E2E тесты** (`tests/e2e/`):
  - Используют реальный HTTP сервер (через Docker)
  - Тестируют полные бизнес-сценарии от начала до конца
  - Проверяют взаимодействие всех компонентов

## Требования

- Docker и Docker Compose
- Go 1.24+

## Запуск тестов

### 1. Запустить E2E окружение

```bash
cd deployments
docker-compose -f docker-compose.e2e.yml up -d
```

Это поднимет:
- PostgreSQL на порту 5434
- Автоматическое применение миграций
- Приложение на порту 8081

### 2. Запустить тесты

```bash
# Из корня проекта
make test-e2e

# Или напрямую
go test -v ./tests/e2e/...
```

### 3. Остановить E2E окружение

```bash
cd deployments
docker-compose -f docker-compose.e2e.yml down -v
```

## Переменные окружения

Можно переопределить параметры через переменные окружения:

```bash
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5434
export TEST_DB_USER=test_user
export TEST_DB_PASSWORD=test_password
export TEST_DB_NAME=pr_reviewer_e2e
export TEST_SERVER_PORT=8081
export E2E_BASE_URL=http://localhost:8081
```

## Что тестируется

- **TestCompletePRWorkflow**: Полный цикл работы с PR:
  - Создание команды
  - Создание PR с автоматическим назначением ревьюверов
  - Переназначение ревьювера
  - Merge PR (включая идемпотентность)
  - Попытка переназначения после merge (должна вернуть ошибку)

- **TestDeactivateUserWorkflow**: Работа с деактивацией пользователей:
  - Создание команды и PR
  - Деактивация пользователя
  - Создание нового PR (деактивированный пользователь не должен назначаться)

- **TestStatisticsWorkflow**: Работа со статистикой:
  - Создание команды и нескольких PR
  - Получение статистики по команде
  - Проверка корректности подсчета

## Структура

- `e2e_test.go` - настройка тестового окружения (TestMain)
- `setup.go` - вспомогательные функции (waitForServer, getBaseURL)
- `scenarios_test.go` - полные бизнес-сценарии

