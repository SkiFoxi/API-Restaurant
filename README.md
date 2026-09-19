# Кухня — MVP сервиса доставки еды

<div align="center">

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-4169E1?style=flat&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-green?style=flat)
![Test Coverage](https://img.shields.io/badge/coverage-86.8%25-brightgreen?style=flat)
![Build](https://img.shields.io/badge/build-passing-success?style=flat)

</div>

---

## Описание проекта

**Авито.Кухня** — это MVP платформы для заказа еды из ресторанов. Проект разработан в рамках тестового задания для стажёрской позиции **Backend** (осенняя волна 2026) в Авито.

Сервис состоит из двух микросервисов:

- **Core API** — основной сервер на Go, предоставляющий REST API для клиентов и ресторанов.
- **Restaurant Demo** — демонстрационный сервис, имитирующий работу ресторана: регистрация, управление меню, опрос новых заказов и обновление статусов.

Архитектура построена на чистой трёхслойной модели: `Handler → Service → Repository` с внедрением зависимостей (DI).

---

## Технологии

| Компонент | Технология | Назначение |
|---|---|---|
| Язык | Go 1.26 | Основной язык разработки |
| База данных | PostgreSQL 15 | Хранение данных |
| Драйвер БД | pgx / pgxpool | Пул соединений и безопасные запросы |
| HTTP-роутинг | net/http (стандартный) | Вместо chi для чистоты кода |
| Логирование | log/slog + lumberjack | Структурированное логирование с ротацией |
| Миграции | golang-migrate | Управление схемой БД |
| Контейнеризация | Docker / Docker Compose | Запуск всех сервисов |
| Тестирование | testing + testify/assert | Юнит-тесты с моками |
| Линтер | golangci-lint | Статический анализ кода |

---

## Архитектура

Проект построен по принципу трёхслойной архитектуры:

1. **Handler** (HTTP-слой) — принимает запросы, валидирует DTO, вызывает сервисы.
2. **Service** (бизнес-логика) — содержит Use Cases, проверки, транзакции.
3. **Repository** (слой данных) — CRUD-операции с БД через pgx.

C4-диаграммы (Level 2 и Level 3) находятся в папке `docs/`:

**Level 2 — Containers**

![C4 Level 2](http://www.plantuml.com/plantuml/svg/ZLHTQnf157tVNt7hqourXaBwLafQbIGb93Ohv77PnC4iqLtP7Mkb5E8keGMXIPyAeGH2wMD19TQFgFaBC_-ezywgk0c1YguESyyzvvuvusR0RP_NgnNjcUCUL-eb1g-o6-lP3IlhkTnsNEQ_rhG2ymFFrHlioAe7p4z3Ibo9Ep4KVJ6LOOhMFqH7ZF4pa6tHm---306tRsK4yWkkpd0n0DpgY-wQR6752976ehNF0curFFyqxlwJEXDS_2Ki45osP2Xc-Ak3Mnr5-alSgtLFtCcBc7Amq86VYHW2lITAvpTaqyncm5dLUklLtPBjDtNxWyCzwzJXnoVCfPQNs6n8H1bn9IDvZi2Bb06VDC3NqXFThZbMsUQiOJSHbVLy1cyAEoHgouFdu3A-Vi9vJQcynd6dOeQDwJa2N98VFNfcu3-6I2z7I8DoeID3rDQATpDWSbHIOK604oPavexqcXVmChfBGhx7BG5j9G-diYtFIM4XhZ7VxyABbqhoxxdDmuGQsKueUN0ciet3tA3XYBl5el528F7XQ22wI8W0UKR4c0MCIbUXPsByKBfUObNFebYGdfkvOqFPMbDHwwbmK4WmXA3diDY8tLVw_uY9Z2YXo1KfJvL6OWifaO_YXDAlutEGNx7b8i7BWpMqpvfsm2ehTo0PWOEyMQHAqD-P-tjAotQnM30HIR14lfz2nfT8sHEYDoCoRgOSYyJzevQ9BYl7N2j_1BsOBIW0dkcyVCNFf_XtNCxAliqTprsoeIhCBQOjL5TugXmV6UekRUyVMhlvhVnUJZUqJUQMw7_a7m00)

`docs/c4_level2_containers.puml` — взаимодействие сервисов.

**Level 3 — Components (Core API)**

![C4 Level 3](http://www.plantuml.com/plantuml/svg/dPFVQnD14CVVxwzOyvHWIb_wL4Igbz0efOr9Ye_7jRUq1ybUsDj8HOJym5Q881N1Hz6_8AQD2JLD_eMT_yZlx0LDYqy6bfkPtPtvV6SdEwa9jUcqcjw3M1qtEv4KZ_ojJNyhy9DMEr5IcSUUvoVAXB6IEdYQT5GKwjVvuqJB85P6QXKsYoBdmnTFgizo1V764_ZzkPy_3LNKb3h5gKgZKHMLbPz3G4cpUMfCcstxbRfsPc-nvdPa9zGNTWfZYV22UZGi2Zk6_mi7hksCkZHmXvxN3_PvYuPsYZLUtQKUFbsuCxwRAzoZIgL-5HzBfgglJ94_JANmmrIc1GRQJHpMDxj0FSxr4xa0zGC0Kt1EgBS1ceMzGk261jjSSu6jQo2mZ4acu0OBtoAsgMz7rAS1tuH8Uky4yrONNeykLkYGDhVpVs1hsKxIs2GwTkYrDJT35Wz5--ICRRzsM1BWMJhs4PsZrWovXRsqdxEAaB9W98u2g-ku7JAVfe-0cA6rIqIkCj8i3gotQsDHEWhAPoOV7N6xBZErOjM5MnhIEyQf9gavqRB-OizrzRkhZhUaSnfkYqH7RaHQKdM2sCWMR2qndHqTAfEYe5UJpRLPMdi_-ucV7ah70i1ee7imkAxG5nHoG-3cYWUfJn-GZzFzJNAlh_-VycuYrn98thWxoPobuLTPLRroy3BOA--MzqlvWhSZLSJ_njy0)

`docs/c4_level3_components.puml` — внутреннее устройство Core API.

ER-диаграмма: `docs/erd.puml`.

---

## Структура проекта

```
avito-kitchen/
├── backend/                          # Core API (основной сервис)
│   ├── cmd/app/main.go               # Точка входа
│   ├── internal/
│   │   ├── database/pool.go          # Подключение к PostgreSQL
│   │   ├── logger/logger.go          # Логирование (с ротацией)
│   │   ├── handlers/                 # HTTP-обработчики
│   │   │   └── dto/                  # DTO для запросов/ответов
│   │   ├── server/server.go          # HTTP-сервер + graceful shutdown
│   │   ├── services/                 # Бизнес-логика
│   │   ├── repositories/             # Слой доступа к БД
│   │   └── models/                   # Структуры для репозиториев и сервисов
│   ├── migrations/                   # SQL-миграции
│   ├── go.mod, go.sum
│   └── Dockerfile
│
├── restaurant-demo/                  # Демо-сервис ресторана
│   ├── cmd/app/main.go
│   ├── internal/
│   │   ├── client/client.go          # HTTP-клиент для Core API
│   │   ├── worker/worker.go          # Цикл опроса и обработка заказов
│   │   └── models/models.go          # Структуры данных
│   ├── go.mod, go.sum
│   └── Dockerfile
│
├── docs/                             # Документация
│   ├── cjm_user.puml
│   ├── cjm_restaurant.puml
│   ├── c4_level2_containers.puml
│   ├── c4_level3_components.puml
│   ├── erd.puml
│   └── openapi.yaml
│
├── docker-compose.yml
├── Makefile
├── .env
├── .golangci.yml
└── .gitignore
```

---

## Запуск проекта

### Требования

- Docker & Docker Compose
- Make (не обязательно)

### Шаги

**1. Клонируйте репозиторий:**

```bash
git clone git@github.com:talense-tasks/backend-trainee-assignment-autumn-2026-skifoxi-c967418a.git
cd backend-trainee-assignment-autumn-2026-skifoxi-c967418a
```

**2. Запустите контейнеры:**

```bash
docker compose up -d
```

Будут подняты:

- PostgreSQL на порту `5432`
- Миграции (автоматически)
- Core API на порту `8080`
- Демо-сервис ресторана

**3. Проверьте работоспособность:**

```bash
curl http://localhost:8080/health
```

Ответ: `OK`.

---

## API

### Публичные эндпоинты (для клиентов)

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/api/v1/restaurants` | Список ресторанов (пагинация) |
| `GET` | `/api/v1/restaurants/{id}` | Детали ресторана |
| `GET` | `/api/v1/restaurants/{id}/menu` | Меню ресторана |
| `POST` | `/api/v1/orders` | Создание заказа |
| `GET` | `/api/v1/orders/{id}` | Получение заказа |
| `DELETE` | `/api/v1/orders/{id}` | Отмена заказа (если статус позволяет) |

### Внутренние эндпоинты (для ресторанов)

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/internal/orders` | Список заказов ресторана (фильтр по статусу) |
| `PATCH` | `/internal/orders/{id}/status` | Обновление статуса заказа |
| `POST` | `/internal/menu` | Добавление позиции в меню |
| `PATCH` | `/internal/menu/{id}/availability` | Включение/отключение позиции |
| `POST` | `/internal/restaurants` | Регистрация ресторана (для демо) |

> Полная спецификация доступна в [`docs/openapi.yaml`](docs/openapi.yaml).

### Примеры команд
 
**Публичные эндпоинты (для клиентов)**
 
Проверка работоспособности:
 
```bash
curl http://localhost:8080/health
```
 
Получить список ресторанов (пагинация: `limit`, `cursor`):
 
```bash
curl http://localhost:8080/api/v1/restaurants?limit=20&cursor=0
```
 
Получить меню конкретного ресторана:
 
```bash
curl http://localhost:8080/api/v1/restaurants/1/menu
```
 
Создать заказ:
 
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "restaurant_id": 1,
    "address": "ул. Пушкина, 10",
    "items": [
      {"menu_item_id": 2, "quantity": 2}
    ]
  }'
```
Если пишет, что ресторан закрыт, значит вы проверяете мою работу ночью.
Измените время работы ресторана вот так:
```bash
docker compose exec postgres psql -U avito -d avito_kitchen -c "UPDATE restaurants SET opening_time = '00:00:00', closing_time = '23:59:59' WHERE id = 1;"
```
 
Проверить статус заказа:
 
```bash
curl http://localhost:8080/api/v1/orders/1
```
 
Отменить заказ (если статус позволяет):
 
```bash
curl -X DELETE http://localhost:8080/api/v1/orders/1
```
 
**Внутренние эндпоинты (для ресторана)**
 
Получить заказы ресторана (с фильтром по статусу):
 
```bash
curl "http://localhost:8080/internal/orders?restaurant_id=1&status=new"
```
 
Обновить статус заказа:
 
```bash
curl -X PATCH http://localhost:8080/internal/orders/1/status \
  -H "Content-Type: application/json" \
  -d '{"status": "cooking"}'
```
 
Добавить позицию в меню (для демо):
 
```bash
curl -X POST http://localhost:8080/internal/menu \
  -H "Content-Type: application/json" \
  -d '{
    "restaurant_id": 1,
    "name": "Новое блюдо",
    "description": "Описание",
    "price": 500
  }'
```
 
Включить/отключить позицию меню:
 
```bash
curl -X PATCH http://localhost:8080/internal/menu/1/availability \
  -H "Content-Type: application/json" \
  -d '{"is_available": false}'
```

---

## Тестирование

Для запуска юнит-тестов (Core API):

```bash
cd backend
go test ./internal/services/...
```

Пример вывода:

```
ok  	avito-kitchen/backend/internal/services	0.006s	coverage: 86.8% of statements
```

Покрытие тестами основного сервиса — около **87%**.

---

## Документация

Все схемы и диаграммы находятся в папке `docs/`:

| Файл | Описание |
|---|---|
| `cjm_user.puml` | Карта пути клиента |
| `cjm_restaurant.puml` | Карта пути ресторана |
| `c4_level2_containers.puml` | Архитектура контейнеров (C4 Level 2) |
| `c4_level3_components.puml` | Внутренние компоненты Core API (C4 Level 3) |
| `erd.puml` | Схема базы данных |
| `openapi.yaml` | OpenAPI спецификация (Swagger) |

---

## Архитектурные решения и условности

### Что я использовал
- **`net/http` вместо `chi`** — стандартный пакет легче и прозрачнее для MVP.
- **`pgx` для PostgreSQL** — безопасные параметризованные запросы и пул соединений.
- **`.env` + `.gitignore`** — для сокрытия секретов. `.env` создан до `.gitignore`, чтобы проверяющему не пришлось вручную его создавать.
- **`.env`** специально не сокрыт, чтобы при проверки решения не пришлось создавать файл **`.env`** и заполнять его.
- **Структурированное логирование** — `slog` с ротацией через `lumberjack`, стиль из видео Авито.
- **Ограничение отмены заказа** — нельзя отменить заказ в статусах `cooking`, `delivering`, `delivered`.

### Что я не делал (и почему)

- **Ветки в Git** — разрабатывал один, продукт ещё не на сервере, ветки замедлили бы разработку.
- **Шардирование БД** — избыточно для MVP. При реальной нагрузке — обязательно.
- **Репликация (синхронная/асинхронная)** — сложно и не нужно для MVP.
- **Хранение цен в копейках** — использовал `float64` для простоты. В коммерческом проекте — `int` (копейки).
- **`go-playground/validator`** — не подключена, но теги в DTO оставлены как задел на будущее.

### Условности работы приложения

- **Авторизация** — заглушка (`userID = 1`) в хендлерах, т.к. по ТЗ авторизация не нужна.
- **Количество блюд** — проверки на остатки нет, т.к. готовые блюда не предполагаются (это не доставка продуктов).
- **Лимит заказов** — при `GET /internal/orders` лимит по умолчанию 20, максимум 100.
- **ID ресторана в query** — т.к. авторизации нет, при запросе списка заказов ресторана передаётся `restaurant_id` в query. В продакшене — из JWT.
- **Глобальный slog** - slog был инициализирован в main.go, чтобы вспомогательные утилиты пользовались логером без явной передачи. При переходе с MVP стоит явно передавать логгер. 
---

## Демо-сервис ресторана (restaurant-demo)

### Условности работы

- Взаимодействие с Core API через **HTTP/JSON** (не gRPC) — достаточно для MVP.
- Нет собственной БД — все данные хранятся в Core API.
- **Опрос заказов** — каждые 10 секунд, первые 20 заказов (можно увеличить).
- Логирование — через тот же пакет `logger`, что и Core API.

### Как проверить

1. Запустите проект через `docker compose up -d`.
2. Создайте заказ через `POST /api/v1/orders`.
3. Через 10–15 секунд заказ автоматически перейдёт в статус `cooking`, затем через 10 секунд — в `ready`.
4. Логи демо-сервиса можно посмотреть:

```bash
docker compose logs -f restaurant-demo
```
