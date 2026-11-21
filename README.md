# Notes API

REST API для личных заметок на Go.

Проект демонстрирует:

- Layered architecture (handler → service → repository)
- JWT-аутентификацию
- PostgreSQL + миграции (golang-migrate)
- Middleware (logging, recovery, auth)
- Валидацию входных данных
- Structured logging (`log/slog`)
- Docker & Docker Compose
- Unit + Integration тесты
- GitHub Actions CI
- Graceful shutdown

## API

| Method | Path                  | Description              | Auth |
|--------|-----------------------|--------------------------|------|
| GET    | `/health`             | Health check             | No   |
| POST   | `/api/v1/auth/register` | Регистрация            | No   |
| POST   | `/api/v1/auth/login`    | Логин → JWT            | No   |
| GET    | `/api/v1/notes`         | Список заметок         | Yes  |
| POST   | `/api/v1/notes`         | Создать заметку        | Yes  |
| GET    | `/api/v1/notes/{id}`    | Получить заметку       | Yes  |
| PUT    | `/api/v1/notes/{id}`    | Обновить заметку       | Yes  |
| DELETE | `/api/v1/notes/{id}`    | Удалить заметку        | Yes  |

## Быстрый старт (Docker)

```bash
# Скопировать пример env
cp .env.example .env

# Поднять всё (Postgres + миграции + API)
make docker-up

# API будет доступен на http://localhost:8080
curl http://localhost:8080/health
```

Остановить:

```bash
make docker-down
```

## Локальная разработка

### Требования

- Go 1.22+
- PostgreSQL 14+
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI

### Настройка

```bash
cp .env.example .env
# Отредактируйте .env при необходимости

# Применить миграции
make migrate-up

# Запустить
make run
```

### Примеры запросов

**Регистрация**

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

**Логин**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

**Создать заметку** (подставьте токен)

```bash
curl -X POST http://localhost:8080/api/v1/notes \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Моя первая заметка","content":"Текст заметки"}'
```

**Список заметок**

```bash
curl http://localhost:8080/api/v1/notes?page=1&limit=10 \
  -H "Authorization: Bearer <TOKEN>"
```

## Тесты

```bash
# Unit-тесты
make test-unit

# Integration-тесты (нужна запущенная БД)
make test-integration

# Все тесты
make test
```

## Структура проекта

```
notes-api/
├── cmd/api/              # точка входа
├── internal/
│   ├── auth/             # JWT
│   ├── config/           # конфигурация
│   ├── domain/           # сущности + интерфейсы + ошибки
│   ├── handler/          # HTTP handlers
│   ├── middleware/       # middleware
│   ├── repository/       # PostgreSQL
│   └── service/          # бизнес-логика
├── migrations/           # SQL-миграции
├── tests/                # integration tests
├── .github/workflows/    # CI
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── README.md
```

## Переменные окружения

| Variable               | Default     | Description                    |
|------------------------|-------------|--------------------------------|
| `APP_ENV`              | development | Окружение                      |
| `APP_HOST`             | 0.0.0.0     | Хост                           |
| `APP_PORT`             | 8080        | Порт                           |
| `DB_HOST`              | localhost   | PostgreSQL host                |
| `DB_PORT`              | 5432        | PostgreSQL port                |
| `DB_USER`              | notes       | Пользователь БД                |
| `DB_PASSWORD`          | notes       | Пароль БД                      |
| `DB_NAME`              | notes       | Имя БД                         |
| `DB_SSLMODE`           | disable     | SSL mode                       |
| `JWT_SECRET`           | —           | Секрет для JWT (минимум 32 символа) |
| `JWT_EXPIRATION_HOURS` | 24          | Время жизни токена             |
| `LOG_LEVEL`            | info        | Уровень логов (debug/info/warn/error) |

## Дальнейшее развитие

Идеи, которые можно добавить в портфолио:

- [ ] Refresh tokens
- [ ] Rate limiting
- [ ] OpenAPI / Swagger документация
- [ ] Soft delete заметок
- [ ] Поиск по заметкам (full-text search)
- [ ] Metrics (Prometheus)
- [ ] Structured error codes
- [ ] Request ID в ответах

---

Сделано с ❤️ на Go.
