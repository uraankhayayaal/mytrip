# Документация проекта MyTrip

Добро пожаловать в документацию проекта **MyTrip** — приложения для организации и планирования путешествий.

## Навигация

| Документ | Описание |
| --- | --- |
| [Архитектура](./architecture.md) | Общая архитектура, layered approach, потоки данных |
| [Бэкенд](./backend.md) | Go-сервер: структура, API, middleware, DI |
| [Фронтенд](./frontend.md) | React SPA: компоненты, роутинг, клиенты, стейт |
| [База данных](./database.md) | Схема, миграции, индексы |
| [API](./api.md) | REST API: все эндпоинты, авторизация, контракты |
| [Развёртывание](./deployment.md) | Docker Compose, окружения, переменные, nginx |
| [Разработка](./development.md) | Быстрый старт, Makefile, тесты, линтинг |
| [Безопасность](./security.md) | JWT, bcrypt, CORS, CSRF, политики токенов |

## Технологии

- **Бэкенд**: Go 1.24, chi router, pgx (PostgreSQL), JWT (HS256), S3 (aws-sdk-go-v2)
- **Фронтенд**: React 18, TypeScript, Vite 5, Tailwind CSS, React Router v6
- **Инфраструктура**: Docker Compose, PostgreSQL 16, MinIO (S3-compatible), Nginx
