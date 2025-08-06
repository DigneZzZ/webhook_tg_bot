# Webhook Telegram Bot

Telegram бот для обработки вебхуков из Discourse и отправки кратких резюме в Telegram группы с поддержкой топиков.

## Возможности

- 🔗 Обработка вебхуков от Discourse (тема + пост)
- 🤖 Генерация кратких резюме с помощью OpenAI GPT-4o-mini  
- 📱 Отправка **одного** объединенного сообщения на тему
- 🔐 Проверка подписи вебхуков для безопасности
- 🏷 Поддержка категорий и тегов
- � Поддержка платных разделов с уведомлениями о подписке
- �🐳 Готовая конфигурация Docker
- 🗄️ Временное хранение данных для объединения вебхуков

## Особенности работы

- **Один пост = одно сообщение**: Бот ожидает получения двух вебхуков (создание темы + создание поста) и отправляет одно объединенное сообщение
- **Умные резюме**: ИИ анализирует содержание поста и создает краткое резюме без раскрытия деталей
- **Платные разделы**: Автоматическое определение платных категорий с информацией о подписке

## Формат сообщения

```
👤 Автор создал новый пост в теме: Название темы

📋 Тема о [краткое резюме от ИИ]

🔗 Ссылка на тему

🏷 Теги: #тег1 #тег2

💎 Данный раздел доступен только по подписке.
Оформить можно в тг-боте: @gig_combot
```
(последние строки только для платных разделов)

## AI модель

Бот использует **OpenAI GPT-4o-mini** для генерации кратких и емких резюме постов.

## Установка и настройка

### 1. Клонирование репозитория

```bash
git clone <repository-url>
cd webhook_tg_bot
```

### 2. Настройка окружения

Скопируйте файл с примером и отредактируйте его:

```bash
cp .env.example .env
```

Отредактируйте `.env` файл:

```env
# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here
TELEGRAM_CHAT_ID=-1001234567890
TELEGRAM_THREAD_ID=0

# Webhook Configuration
WEBHOOK_SECRET=your_secret_key_here
WEBHOOK_PORT=8080
WEBHOOK_PATH=/webhook

# AI Configuration
OPENAI_API_KEY=your_openai_api_key_here
OPENAI_MODEL=gpt-4o-mini

# Base URL for your forum
BASE_URL=https://your-forum.com

# Premium categories (comma-separated category IDs)
PREMIUM_CATEGORIES=4,5,6

# Categories mapping
CATEGORY_MAPPING={"1":"Общие вопросы","2":"Техническая поддержка","3":"Личные вопросы"}
```

### 3. Получение Telegram Bot Token

1. Напишите [@BotFather](https://t.me/botfather) в Telegram
2. Создайте нового бота командой `/newbot`
3. Скопируйте полученный токен в `TELEGRAM_BOT_TOKEN`

### 4. Настройка Telegram группы

1. Добавьте бота в группу
2. Дайте боту права администратора
3. Получите Chat ID группы (можно использовать [@userinfobot](https://t.me/userinfobot))
4. Если нужно отправлять в топик, укажите Thread ID

### 5. Запуск

#### С помощью Docker (рекомендуется)

**Ubuntu/Linux:**
```bash
# Быстрая установка
wget https://raw.githubusercontent.com/your-repo/webhook_tg_bot/main/quick_install.sh
chmod +x quick_install.sh
./quick_install.sh

# Или с автоматической настройкой
wget https://raw.githubusercontent.com/your-repo/webhook_tg_bot/main/deploy_ubuntu.sh
chmod +x deploy_ubuntu.sh
./deploy_ubuntu.sh
```

**Windows:**
```bash
# Сборка и запуск
docker-compose up -d

# Просмотр логов
docker-compose logs -f webhook-bot
```

#### Локальный запуск

```bash
# Установка зависимостей
go mod download

# Запуск
go run main.go
```

## Настройка вебхука в Discourse

1. Перейдите в Admin → API → Webhooks
2. Создайте новый вебхук:
   - **Payload URL**: `http://your-domain:8080/webhook`
   - **Content Type**: `application/json`
   - **Secret**: укажите тот же секрет, что в `WEBHOOK_SECRET`
   - **Events**: выберите `Topic Event` и `Post Event`

## Структура проекта

```
webhook_tg_bot/
├── main.go                 # Точка входа
├── internal/
│   ├── config/            # Конфигурация
│   ├── server/            # HTTP сервер для вебхуков
│   ├── bot/               # Telegram бот
│   ├── ai/                # AI провайдеры
│   └── models/            # Модели данных
├── docker-compose.yml     # Docker конфигурация
├── Dockerfile            # Docker образ
└── .env.example          # Пример настроек
```

## Формат вебхуков

Бот обрабатывает два типа вебхуков от Discourse:

### 1. Создание темы
```json
{
  "topic": {
    "id": 158,
    "title": "Название темы",
    "category_id": 3,
    "tags": ["tag1", "tag2"],
    "created_by": {
      "username": "author"
    }
  }
}
```

### 2. Создание поста
```json
{
  "post": {
    "id": 238,
    "topic_id": 158,
    "topic_title": "Название темы",
    "raw": "Содержимое поста...",
    "username": "author",
    "post_number": 1
  }
}
```

## Troubleshooting

### Проверка работы бота
```bash
curl http://localhost:8080/health
```

### Просмотр логов
```bash
docker-compose logs -f webhook-bot
```

### Тест вебхука
```bash
curl -X POST http://localhost:8080/webhook?secret=your_secret \
  -H "Content-Type: application/json" \
  -d '{"topic":{"id":1,"title":"Test","category_id":1,"created_by":{"username":"test"}}}'
```

## Безопасность

- Используйте сложный секретный ключ для вебхука
- Не забудьте настроить HTTPS в продакшене
- Ограничьте доступ к порту бота через файрвол

## Лицензия

MIT License
