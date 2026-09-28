# wells-risk-backend

Веб-сервис справочника критериев шкалы Уэллса для оценки риска тромбоза глубоких вен.
Лабораторная работа №3: `Go` + `gin-gonic` + `gORM` + `PostgreSQL` + `MinIO`.

## Запуск

```bash
docker compose up -d          # PostgreSQL, MinIO, Adminer
go run ./cmd/migrate          # миграции и врач по умолчанию
go run ./cmd/wells            # веб-сервис на http://localhost:8080
```

Переменные окружения в `.env`:

| Переменная | Назначение |
|---|---|
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`, `DB_NAME` | подключение к PostgreSQL |
| `MINIO_ENDPOINT` | адрес MinIO для загрузки файлов |
| `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY` | ключи доступа MinIO |
| `MINIO_BUCKET_NAME` | бакет с файлами критериев |
| `MINIO_PUBLIC_URL` | публичный адрес MinIO для ссылок в ответах |

### Домен критериев Уэллса

| Метод | URL | Параметры | Ответ | Коды |
|---|---|---|---|---|
| GET | `/api/criteria` | `min_points` — минимальный вес критерия в баллах | массив опубликованных критериев, поле `is_mine` = 1 если критерий создан текущим врачом | 200, 400, 500 |
| GET | `/api/criteria/feed` | — | первый критерий ленты, ид не указывается | 200, 404 |
| GET | `/api/criteria/feed/:id` | `next=true` — следующий критерий ленты | критерий ленты | 200, 400, 404 |
| GET | `/api/criteria/draft` | — | черновик текущего врача, ид не указывается | 200, 404 |
| POST | `/api/criteria` | форма: `criterion_name`, `short_description`, `criterion_group`, `wells_points`, файлы `image`, `video` | созданный критерий в статусе черновик | 201, 400, 409, 500 |
| PUT | `/api/criteria/:id/publish` | — | опубликованный критерий | 200, 400, 404 |
| DELETE | `/api/criteria/:id` | — | пустое тело, логическое удаление | 204, 400, 404 |
| POST | `/api/criteria/:id/like` | JSON: `{"value": 1}` ставит отметку, `{"value": 0}` снимает | критерий с пересчитанными отметками | 200, 400, 404, 500 |

### Домен врачей

| Метод | URL | Параметры | Ответ | Коды |
|---|---|---|---|---|
| POST | `/api/physicians/register` | JSON: `login`, `password`, `full_name` | зарегистрированный врач | 201, 400, 409, 500 |
| POST | `/api/physicians/login` | JSON: `login`, `password` | врач; заглушка, JWT появится в ЛР4 | 200, 400, 404 |
| POST | `/api/physicians/logout` | — | пустое тело; заглушка, blacklist появится в ЛР4 | 204 |

### Правила

* Записи в статусе `удалён` на клиент не передаются.
* Статусы меняются только вперёд: `черновик` → `опубликован` → `удалён`, вернуть в черновик нельзя.
* Публиковать и удалять можно только свои критерии, у врача не более одного черновика.
* Системные поля (ид, статус, создатель, даты создания и формирования) с клиента не принимаются,
  они вычисляются на бэкенде.
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
| `image_key` | varchar(255) | ключ изображения в MinIO |
| `video_key` | varchar(255) | ключ короткого видео в MinIO |
| `wells_points` | numeric(3,1) | вес критерия в баллах Уэллса |
| `criterion_group` | varchar(50) | группа критерия |
| `created_at` | timestamp, not null | дата создания черновика |
| `formed_at` | timestamp | дата публикации |
| `creator_id` | integer, FK → `physicians` | врач-создатель критерия |

### `physicians` — врачи

| Столбец | Тип | Описание |
|---|---|---|
| `physician_id` | serial, PK | идентификатор врача |
| `login` | varchar(50), not null, unique | логин |
| `password` | varchar(100), not null | sha256-хеш пароля |
| `full_name` | varchar(150) | ФИО врача |
| `is_moderator` | boolean, not null, default false | признак модератора |

### `criterion_likes` — отметки врачей о полезности критерия

| Столбец | Тип | Описание |
|---|---|---|
| `like_id` | serial, PK | идентификатор отметки |
| `criterion_id` | integer, not null, FK → `wells_criteria` | критерий |
| `physician_id` | integer, not null, FK → `physicians` | врач |

Уникальный индекс `idx_like_criterion_physician` по паре `criterion_id` + `physician_id`:
один врач не может отметить критерий дважды.

## Структура проекта

```
cmd/wells        точка входа веб-сервиса
cmd/migrate      миграции схемы и врач по умолчанию
internal/api     регистрация эндпоинтов
internal/app/auth        функция-singleton текущего врача
internal/app/ds          модели и сериализаторы
internal/app/dsn         строка подключения к PostgreSQL
internal/app/handler     обработчики методов веб-сервиса
internal/app/repository  работа с PostgreSQL и MinIO
```