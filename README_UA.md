<div align="center">

# Cherkasyoblenergo API

[![EN](https://img.shields.io/badge/English-c9c9c9.svg)](README.md) [![UA](https://img.shields.io/badge/Ukrainian-0e7837.svg)](README_UA.md)

[![Deploy](https://github.com/jeffjunior470/cherkasyoblenergo-api/actions/workflows/deploy.yml/badge.svg)](https://github.com/jeffjunior470/cherkasyoblenergo-api/actions/workflows/deploy.yml)
[![Tests](https://github.com/jeffjunior470/cherkasyoblenergo-api/actions/workflows/tests.yml/badge.svg)](https://github.com/jeffjunior470/cherkasyoblenergo-api/actions/workflows/tests.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/jeffjunior470/cherkasyoblenergo-api)](https://go.dev/)
[![License](https://img.shields.io/github/license/jeffjunior470/cherkasyoblenergo-api)](LICENSE)
[![Releases](https://img.shields.io/github/release/jeffjunior470/cherkasyoblenergo-api.svg)](https://github.com/jeffjunior470/cherkasyoblenergo-api/releases)
[![Changelog](https://img.shields.io/badge/changelog-md-blue)](CHANGELOG.md)

**Неофіційне API** для отримання графіків відключень електроенергії з [cherkasyoblenergo.com](https://cherkasyoblenergo.com/)

</div>

## Публічний інстанс

Публічний інстанс цього API доступний за адресою:

```
https://hue.pp.ua/cherkasyoblenergo/api/
```

Ви можете вільно використовувати його для своїх проєктів. Однак зверніть увагу, що цей інстанс надається "як є" без будь-яких гарантій доступності чи безперебійної роботи. Для продакшн використання або якщо вам потрібна надійність, рекомендуємо розгорнути власний інстанс.

> [!NOTE]
> У цьому публічному інстансі діє обмеження на 30 запитів на хвилину з однієї IP-адреси.

## Можливості

- Автоматичне оновлення графіків (за замовчуванням кожні 5 хвилин)
- Доступ до історичних даних
- RESTful API інтерфейс
- Rate limiting по IP адресі
- Опціональна автентифікація за API ключем

## Встановлення

### Вимоги

- Go 1.25 або вище для збірки з вихідного коду

SQLite вбудовано через драйвер на Go. Компілятор C, окреме встановлення SQLite
та CGO не потрібні. Для запуску готового виконуваного файлу Go не потрібний.
Використовується `github.com/glebarez/sqlite` v1.11.0 разом із GORM.
Для Windows і Linux потрібні окремі виконувані файли відповідної архітектури.

### Налаштування

1. **Клонуйте репозиторій**

   ```bash
   git clone https://github.com/jeffjunior470/cherkasyoblenergo-api.git
   cd cherkasyoblenergo-api
   ```

2. **Налаштуйте змінні середовища**

   Створіть файл `.env` у робочій директорії: поточний завантажувач конфігурації вимагає його навіть за наявності змінних середовища. За основу можна взяти `.env.example`. Для локального HTTP встановіть `FORCE_HTTPS=false`; залиште `API_KEY=` порожнім для роботи без автентифікації або замініть приклад власним ключем. Доступні налаштування:

   | Змінна | Обов'язкова | За замовчуванням | Опис |
   |--------|-------------|------------------|------|
   | `DB_NAME` | Ні | `cherkasyoblenergo.db` | Шлях до файлу бази даних SQLite |
   | `SERVER_PORT` | Ні | `8080` | Порт для API сервера |
   | `NEWS_URL` | Ні | `https://www.cherkasyoblenergo.com/api/v1/posts/category/news?lang=uk&page=0&size=20` | URL для парсингу графіків |
   | `PARSING_INTERVAL_MINUTES` | Ні | `5` | Як часто перевіряти нові графіки (хвилини) |
   | `RATE_LIMIT_PER_MINUTE` | Ні | `60` | Макс. запитів на хвилину на IP |
   | `LOG_LEVEL` | Ні | `info` | Рівень логування (`debug`, `info`, `warn`, `error`) |
   | `FORCE_HTTPS` | Ні | `false` | Перенаправляти HTTP на HTTPS |
   | `API_KEY` | Ні | - | Якщо встановлено, вмикає автентифікацію за API ключем |
   | `PROXY_MODE` | Ні | `none` | Режим проксі для визначення реального IP клієнта: `cloudflare`, `standard` або `none` |

3. **Запуск**

   ```bash
   go run ./cmd/server/main.go
   ```

### Збірка

**Windows (cmd.exe):**

```bat
set CGO_ENABLED=0
go mod download
go build -o cherkasyoblenergo_api.exe ./cmd/server
cherkasyoblenergo_api.exe
```

**Ubuntu / Linux:**

```bash
export CGO_ENABLED=0
go mod download
go build -o cherkasyoblenergo_api ./cmd/server
./cherkasyoblenergo_api
```

Виконуйте команди з директорії проєкту, де лежить `.env`. Каталог бази даних
має бути доступним для запису. Після запуску перевірте
`http://localhost:8080/cherkasyoblenergo/api/`. Початковий парсинг графіків
виконується у фоні, після чого працює періодичне оновлення. Значення `today`
і `tomorrow` залежать від локального часового поясу комп'ютера; для українських
дат налаштуйте `Europe/Kyiv`.

Для розгортання готової програми скопіюйте виконуваний файл та свій `.env`.
Зберігайте SQLite-базу за шляхом `DB_NAME`. Для запуску готового файлу Go та
Docker не потрібні.

### Перехід на SQLite-драйвер без CGO

Замість `gorm.io/driver/sqlite` / `github.com/mattn/go-sqlite3` використовується
драйвер без CGO. База залишається файлом SQLite; для наявної схеми експорт та
імпорт даних не очікуються. Перед оновленням чинного розгортання зупиніть
програму й створіть резервну копію бази, придатну для відновлення, разом із
файлами `-wal` та `-shm`, якщо вони залишилися.

Кожне з'єднання отримує WAL, очікування блокувань 5000 мс та перевірку зовнішніх
ключів. Дати записуються в текстовому форматі, сумісному із SQLite. Під час
читання зберігається записане часове зміщення замість приведення через стару
опцію `_loc=Local`; сам момент часу не змінюється. Маршрути API та назви полів
JSON залишилися тими самими. Повторні спроби при блокуванні визначаються за
кодами SQLite, а не за текстом помилки.

### Docker

Dockerfile збирає програму з `CGO_ENABLED=0` без встановлення GCC та musl-dev.
Docker є опціональним. Наявний `docker-compose.app.yml` очікує зовнішню мережу
`think-root-network` і каталог бази, доступний для запису UID/GID `1000:1000`;
для окремого розгортання адаптуйте ці налаштування.

## Документація API

### Базовий URL

```
/cherkasyoblenergo/api
```

### Отримати графік відключень

```
GET /blackout-schedule
```

Отримати записи розкладу на основі параметрів фільтрації.

#### Параметри запиту

| Параметр | Обов'язковий | Опис |
|----------|--------------|------|
| `option` | Так | `all`, `latest_n`, `by_date` або `by_schedule_date` |
| `date` | Для `by_date`, `by_schedule_date` | `YYYY-MM-DD`, `today` або `tomorrow` |
| `limit` | Для `latest_n` | Ціле число > 0 |
| `queue` | Ні | Значення черг через кому (наприклад, `3_2` або `4_1,3_1`) |

#### Опції фільтрації

- `all` - Отримує всі записи розкладу
- `latest_n` - Отримує обмежену кількість останніх записів (потрібен `limit`)
- `by_date` - Отримує записи за датою публікації (потрібна `date`)
- `by_schedule_date` - Отримує записи за датою графіка (потрібна `date`, опціонально `limit`)

#### Приклади запитів

**Отримати всі графіки:**
```bash
curl "http://localhost:8080/cherkasyoblenergo/api/blackout-schedule?option=all"
```

**Отримати останні 5 графіків:**
```bash
curl "http://localhost:8080/cherkasyoblenergo/api/blackout-schedule?option=latest_n&limit=5"
```

**Отримати сьогоднішній графік для черги 3_2:**
```bash
curl "http://localhost:8080/cherkasyoblenergo/api/blackout-schedule?option=by_schedule_date&date=today&queue=3_2"
```

**Отримати графік з кількома чергами:**
```bash
curl "http://localhost:8080/cherkasyoblenergo/api/blackout-schedule?option=latest_n&limit=1&queue=4_1,3_1,2_2"
```

#### Приклади відповідей

**Повна відповідь (без фільтра черги):**
```json
[
  {
    "id": 1234,
    "news_id": 100,
    "title": "Графік на 14 листопада",
    "date": "2024-03-20T10:30:00Z",
    "schedule_date": "2024-11-14",
    "1_1": "08:00-10:00",
    "1_2": "10:00-12:00",
    "2_1": "12:00-14:00",
    "2_2": "14:00-16:00",
    "3_1": "09:00-11:00",
    "3_2": "11:00-13:00",
    "4_1": "13:00-15:00",
    "4_2": "15:00-17:00",
    "5_1": "07:00-09:00",
    "5_2": "09:00-11:00",
    "6_1": "11:00-13:00",
    "6_2": "13:00-15:00"
  }
]
```

**Відфільтрована відповідь (з фільтром черги):**
```json
[
  {
    "id": 1234,
    "news_id": 100,
    "title": "Графік на 14 листопада",
    "date": "2024-03-20T10:30:00Z",
    "schedule_date": "2024-11-14",
    "3_2": "00:30 - 02:30, 06:00 - 09:00"
  }
]
```

### Rate Limiting

- За замовчуванням: 60 запитів на хвилину на IP
- Налаштовується через змінну середовища `RATE_LIMIT_PER_MINUTE`
- Заголовки відповіді:
  - `X-RateLimit-Limit` - Максимальна кількість запитів
  - `X-RateLimit-Remaining` - Залишок запитів у поточному вікні

### Автентифікація (опціонально)

За замовчуванням API є публічним і не вимагає автентифікації. Для приватних інстансів можна увімкнути автентифікацію за API ключем:

1. Встановіть змінну середовища `API_KEY`:
   ```properties
   API_KEY=your-secret-key
   ```

2. Додавайте ключ до всіх запитів:
   ```bash
   curl -H "X-API-Key: your-secret-key" "http://localhost:8080/cherkasyoblenergo/api/blackout-schedule?option=latest_n&limit=1"
   ```

Якщо `API_KEY` не встановлено або порожнє, API залишається публічним.

### Коди помилок

| Код статусу | Опис |
|-------------|------|
| 200 | Успіх |
| 400 | Невірний запит (невалідні параметри) |
| 401 | Неавторизовано (невірний або відсутній API ключ, якщо автентифікацію увімкнено) |
| 429 | Забагато запитів (перевищено rate limit) |
| 500 | Внутрішня помилка сервера |

## Запуск тестів

Windows (cmd.exe):

```bat
set CGO_ENABLED=0
go test ./...
```

Ubuntu / Linux:

```bash
CGO_ENABLED=0 go test ./...
```

CI також виконує тести без CGO. Перевірки бази охоплюють PRAGMA для з'єднань,
старі UTC-дати, нові дати з часовим зміщенням, повторне відкриття файлової бази,
міграцію схеми та розпізнавання блокувань. Тести завантаження новин використовують
локальний HTTP-сервер.

## Ліцензія

Цей проєкт ліцензовано BSD 2-Clause. Деталі дивіться у файлі [LICENSE](LICENSE).
