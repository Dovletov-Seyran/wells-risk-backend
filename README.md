# wells-risk-backend

Веб-сервис справочника критериев шкалы Уэллса для оценки риска тромбоза глубоких вен.
Лабораторная работа №4: `Go` + `gin-gonic` + `gORM` + `PostgreSQL` + `MinIO` + `Redis` + `JWT` + `Swagger`.

## Запуск

```bash
docker compose up -d          # PostgreSQL, MinIO, Redis, Adminer
go run ./cmd/migrate          # миграции и врач по умолчанию
go run ./cmd/wells            # веб-сервис на http://localhost:8080
```

Swagger: http://localhost:8080/swagger/index.html. Документация генерируется из комментариев к методам:

```bash
swag init -g cmd/wells/main.go --parseInternal
```

Настройки сервиса и таймауты Redis — в `config/config.toml`, переменные окружения — в `.env`:

| Переменная | Назначение |
|---|---|
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`, `DB_NAME` | подключение к PostgreSQL |
| `MINIO_ENDPOINT` | адрес MinIO для загрузки файлов |
| `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY` | ключи доступа MinIO |
| `MINIO_BUCKET_NAME` | бакет с файлами критериев |
| `MINIO_PUBLIC_URL` | публичный адрес MinIO для ссылок в ответах |
| `REDIS_HOST`, `REDIS_PORT`, `REDIS_USER`, `REDIS_PASSWORD` | подключение к Redis |
| `JWT_TOKEN` | секрет для подписи JWT |

## Авторизация

* Аутентификация `POST /api/physicians/login` возвращает JWT (`access_token`), подписанный HS256, срок действия 1 час.
* Токен передаётся в заголовке `Authorization: Bearer <access_token>`.
* Деавторизация `POST /api/physicians/logout` записывает токен в blacklist Redis
  (ключ `wells_service.jwt.<токен>`) до истечения его срока, после этого токен не принимается.
* Роли: `0` — врач (создатель критериев), `1` — модератор. Роль передаётся в JWT.
* Гостю доступно только чтение: список критериев и лента. Методы создателя без токена возвращают 403.
* Создатель критерия при добавлении берётся из JWT.

## Методы веб-сервиса

Авторизация: «—» — доступно гостю, «JWT» — только с токеном врача или модератора.

### Домен критериев Уэллса

| Метод | URL | Авторизация | Параметры | Ответ | Коды |
|---|---|---|---|---|---|
| GET | `/api/criteria` | — | `min_points` — минимальный вес критерия в баллах | массив опубликованных критериев, `is_mine` = 1 если критерий создан текущим врачом (для гостя 0) | 200, 400, 500 |
| GET | `/api/criteria/feed` | — | `criterion_id` — ид критерия, `next=true` — следующий после него; без параметров — первый критерий ленты | критерий ленты, `is_liked` = 1 если текущий врач поставил лайк (для гостя 0) | 200, 400, 404 |
| GET | `/api/criteria/draft` | JWT | — | черновик текущего врача, ид не указывается | 200, 403, 404 |
| POST | `/api/criteria` | JWT | форма: `criterion_name`, `short_description`, `criterion_group`, `wells_points`, файлы `image`, `video` | созданный критерий в статусе черновик | 201, 400, 403, 409, 500 |
| PUT | `/api/criteria/:id/publish` | JWT | — | опубликованный критерий | 200, 400, 403, 404 |
| DELETE | `/api/criteria/:id` | JWT | — | пустое тело, логическое удаление | 204, 400, 403, 404 |
| POST | `/api/criteria/:id/like` | JWT | JSON: `{"value": 1}` ставит лайк, `{"value": 0}` отменяет | критерий с пересчитанными лайками | 200, 400, 403, 404, 500 |

### Домен врачей

| Метод | URL | Авторизация | Параметры | Ответ | Коды |
|---|---|---|---|---|---|
| POST | `/api/physicians/register` | — | JSON: `login`, `password`, `full_name` | `{"ok": true}` | 200, 400, 409, 500 |
| POST | `/api/physicians/login` | — | JSON: `login`, `password` | `{"expires_in", "access_token", "token_type": "Bearer"}` | 200, 400, 403, 500 |
| POST | `/api/physicians/logout` | JWT | заголовок `Authorization` | пустое тело, токен в blacklist Redis | 200, 400, 403, 500 |

### Правила

* Записи в статусе `удалён` на клиент не передаются.
* Статусы меняются только вперёд: `черновик` → `опубликован` → `удалён`, вернуть в черновик нельзя.
* Публиковать и удалять можно только свои критерии, у врача не более одного черновика.
* Системные поля (ид, статус, создатель, даты создания и формирования) с клиента не принимаются,
  они вычисляются на бэкенде через авторизацию и бизнес-логику.
* Изображение и видео передаются файлами, имена объектов генерируются на латинице
  (`criterion_<ид>_image_<время>.jpg`), в БД хранятся только ключи объектов.

## Таблицы базы данных

### `wells_criteria` — критерии шкалы Уэллса (услуги)

| Столбец | Тип | Описание |
|---|---|---|
| `criterion_id` | serial, PK | идентификатор критерия |
| `criterion_name` | varchar(150), not null | название критерия |
| `short_description` | varchar(500) | клиническое описание критерия |
| `criterion_status` | varchar(20), not null | статус: черновик, опубликован, удалён |
| `image_key` | varchar(255), not null | ключ изображения в MinIO |
| `video_key` | varchar(255), not null | ключ короткого видео в MinIO |
| `wells_points` | numeric(3,1) | вес критерия в баллах Уэллса |
| `criterion_group` | varchar(50) | группа критерия |
| `created_at` | timestamp, not null | дата создания черновика |
| `formed_at` | timestamp | дата публикации |
| `creator_id` | integer, not null, FK → `physicians` | врач-создатель критерия |

### `physicians` — врачи

| Столбец | Тип | Описание |
|---|---|---|
| `physician_id` | serial, PK | идентификатор врача |
| `physician_uuid` | uuid, unique | uuid врача, передаётся в JWT |
| `login` | varchar(50), not null, unique | логин |
| `password` | varchar(100), not null | sha1-хеш пароля |
| `full_name` | varchar(150) | ФИО врача |
| `role` | bigint, not null, default 0 | роль: 0 — врач, 1 — модератор |

### `criterion_likes` — лайки врачей

| Столбец | Тип | Описание |
|---|---|---|
| `like_id` | serial, PK | идентификатор лайка |
| `criterion_id` | integer, not null, FK → `wells_criteria` | критерий |
| `physician_id` | integer, not null, FK → `physicians` | врач |

Уникальный индекс `idx_like_criterion_physician` по паре `criterion_id` + `physician_id`:
один врач не может поставить лайк критерию дважды.

## Структура проекта

```
cmd/wells                 точка входа веб-сервиса, общее описание Swagger
cmd/migrate               миграции схемы и врач по умолчанию
config/config.toml        настройки сервиса, JWT и Redis
docs                      документация Swagger, генерируется swag
internal/pkg/app          приложение: запуск сервера, эндпоинты, аутентификация, middleware
internal/app/config       чтение конфигурации
internal/app/ds           модели, сериализаторы, JWT claims
internal/app/dsn          строка подключения к PostgreSQL
internal/app/handler      обработчики методов критериев
internal/app/redis        клиент Redis и blacklist JWT
internal/app/repository   работа с PostgreSQL и MinIO
internal/app/role         роли врачей
```
