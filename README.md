# Webhook Telegram Bot

![Production Ready](https://img.shields.io/badge/production-ready-green)
![Docker](https://img.shields.io/badge/docker-supported-blue)
![AI Powered](https://img.shields.io/badge/AI-GPT--5--mini-orange)

Профессиональный Telegram бот для автоматической отправки уведомлений о новых постах с форума Discourse в Telegram группы с ИИ-генерируемыми резюме.

## ✨ Возможности

### 🔗 Интеграция с Discourse
- Обработка webhook'ов при создании тем и постов
- Поддержка всех типов контента Discourse
- Объединение данных темы и поста в одно сообщение
- Проверка подписи webhook'ов для безопасности

### 🤖 ИИ-анализ контента
- **OpenAI GPT-5-mini** для генерации умных резюме
- Анализ содержания постов с учетом контекста
- Очистка HTML-тегов для лучшего анализа
- Специальные шаблоны для разных типов контента

### 📱 Гибкая доставка в Telegram
- Отправка в основной чат или конкретные топики
- **Маппинг категорий** на разные Telegram топики
- Поддержка эмодзи-префиксов для ролей пользователей
- Умное форматирование сообщений с HTML

### 🎯 Фильтрация и контроль
- **Мониторинг конкретных категорий** или всех сразу
- **Игнорирование нежелательных категорий**
- **Блокировка ботов** (discobot, chatbot и других)
- **Платные разделы** с уведомлениями о подписке

### 🚀 Production Ready
- Docker контейнеризация с healthcheck
- GitHub Actions для автоматической сборки
- Systemd интеграция для автозапуска
- Логирование и мониторинг
- Graceful shutdown и error handling

## 📋 Формат сообщения

```
👤 👑 Администратор создал новый пост: Название темы

📋 Автор делится решением проблемы с настройкой Docker контейнера и объясняет пошаговый процесс устранения ошибок подключения к базе данных.

🔗 Ссылка на тему (https://forum.example.com/t/topic/123)

🏷 Теги: #docker, #database, #troubleshooting

💎 Данный раздел доступен только по подписке.
```

### Префиксы ролей
- 👑 Администратор
- 🛡️ Модератор  
- ⭐ Персонал
- 🔥 Лидер
- 👤 Обычный пользователь

## ⚙️ Конфигурация (.env)

### 📡 Telegram настройки
```bash
# Основные настройки бота
TELEGRAM_BOT_TOKEN=1234567890:ABC-DEF1234567890          # Токен от @BotFather
TELEGRAM_CHAT_ID=-1001234567890                          # ID группы/канала  
TELEGRAM_THREAD_ID=0                                     # ID топика (0 = основной чат)
```

### 🔗 Webhook настройки
```bash
WEBHOOK_SECRET=your_super_secret_key_here                # Секретный ключ для проверки подписи
WEBHOOK_PORT=8080                                        # Порт для webhook сервера
WEBHOOK_PATH=/webhook                                     # Путь endpoint'а
WEBHOOK_DOMAIN=https://your-server.com                   # Домен сервера (опционально)
```

### 🤖 ИИ настройки
```bash
OPENAI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx  # Ключ OpenAI API
OPENAI_MODEL=gpt-4.1-nano                                   # Модель GPT (рекомендуется gpt-4.1-nano)
```

### 🏷 Категории и фильтрация
```bash
BASE_URL=https://your-forum.com                          # Адрес вашего форума

# Мониторинг категорий (пусто = все категории)
MONITORED_CATEGORIES=1,2,3,4,5                           # ID категорий для отслеживания
IGNORED_CATEGORIES=10,11,12                               # ID категорий для игнорирования (приоритет)

# Игнорирование пользователей
IGNORED_USERS=-2,-1,1234                                  # ID пользователей для игнорирования
                                                          # -2 = discobot, -1 = system, etc.

# Платные разделы
PREMIUM_CATEGORIES=4,5,6                                  # ID категорий с подпиской
```

### 📍 Маппинг категорий на Telegram топики
```bash
# Программирование в топик 1
TELEGRAM_THREAD_ID_1=123456                              # ID сообщения-топика
THREAD_CATEGORIES_1=1,2,3                                # Категории: Go, Python, JavaScript

# Дизайн в топик 2  
TELEGRAM_THREAD_ID_2=234567
THREAD_CATEGORIES_2=4,5                                  # Категории: UI/UX, Graphics

# Бизнес в топик 3
TELEGRAM_THREAD_ID_3=345678
THREAD_CATEGORIES_3=6,7,8                                # Категории: Marketing, Sales, Strategy

# Железо в топик 4
TELEGRAM_THREAD_ID_4=456789  
THREAD_CATEGORIES_4=9,10                                 # Категории: Hardware, Reviews

# Офтоп в топик 5
TELEGRAM_THREAD_ID_5=567890
THREAD_CATEGORIES_5=11,12,13                             # Категории: General, Random, Fun
```

## 🚀 Установка

### Быстрая установка (рекомендуется)
```bash
# Скачать и запустить установщик
curl -fsSL https://raw.githubusercontent.com/dignezzz/webhook_tg_bot/main/install.sh -o install.sh
chmod +x install.sh
./install.sh install
```

### Управление ботом
```bash
whtg                    # Интерактивное меню
whtg install            # Установить бота  
whtg update             # Обновить до последней версии
whtg start              # Запустить бота
whtg stop               # Остановить бота
whtg restart            # Перезапустить бота
whtg status             # Показать статус
whtg logs               # Показать логи
whtg uninstall          # Удалить бота
```

### Docker Compose (ручная установка)
```yaml
version: '3.8'
services:
  webhook-bot:
    image: ghcr.io/dignezzz/webhook_tg_bot:latest
    container_name: webhook_tg_bot
    restart: unless-stopped
    ports:
      - "8080:8080"
    env_file:
      - .env
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

## 🔧 Настройка Discourse

### 1. Создание webhook'а
1. Перейдите в **Admin → API → Webhooks**
2. Нажмите **New Webhook**
3. Заполните параметры:

```
Payload URL: http://your-server.com:8080/webhook
Content Type: application/json  
Secret: your_webhook_secret_from_env
Which events: Topic Event + Post Event
```

### 2. Получение ID категорий
Перейдите в админку: `https://your-forum.com/admin/customize/site_texts`
Или посмотрите URL категории: `https://your-forum.com/c/category-name/5` (где 5 - это ID)

### 3. Получение ID пользователей  
Перейдите в админку: `https://your-forum.com/admin/users/USERNAME`
URL покажет ID: `https://your-forum.com/admin/users/123/username`

## 📱 Настройка Telegram

### 1. Создание бота
1. Напишите [@BotFather](https://t.me/botfather)
2. Используйте команду `/newbot`
3. Скопируйте токен в `TELEGRAM_BOT_TOKEN`

### 2. Настройка группы
1. Создайте группу или канал
2. Добавьте бота как администратора
3. Получите Chat ID:
   - Добавьте [@userinfobot](https://t.me/userinfobot) в группу
   - Или используйте API: `https://api.telegram.org/bot<TOKEN>/getUpdates`

### 3. Настройка топиков (опционально)
1. Включите топики в настройках группы
2. Создайте топики для разных категорий
3. Получите Thread ID каждого топика:
   - Перешлите сообщение из топика боту
   - Посмотрите логи для получения Thread ID

## 🔍 Мониторинг и отладка

### Проверка состояния
```bash
# Healthcheck endpoint
curl http://localhost:8080/health

# Статус через управляющий скрипт
whtg status
```

### Логи
```bash
# Через управляющий скрипт  
whtg logs

# Напрямую через Docker
docker-compose logs -f webhook-bot

# Системные логи
journalctl -u webhook-tg-bot -f
```

### Тестирование webhook'а
```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -H "X-Discourse-Event-Signature: sha256=..." \
  -d '{
    "topic": {
      "id": 1,
      "title": "Test Topic", 
      "category_id": 1,
      "created_by": {"username": "test"}
    }
  }'
```

## 📊 Архитектура

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Discourse     │───▶│  Webhook Server  │───▶│   Telegram Bot  │
│    Forum        │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │                          │
                              ▼                          ▼
                       ┌──────────────┐         ┌─────────────────┐
                       │ OpenAI GPT   │         │ Telegram Groups │
                       │  (Summary)   │         │   & Topics      │
                       └──────────────┘         └─────────────────┘
```

### Компоненты
- **HTTP Server**: Обработка webhook'ов от Discourse
- **Storage**: Временное хранение для объединения данных
- **AI Provider**: Генерация резюме через OpenAI
- **Telegram Bot**: Отправка сообщений в группы/топики
- **Config**: Гибкая система конфигурации

## 🔐 Безопасность

### Рекомендации
- ✅ Используйте сложный секретный ключ (`WEBHOOK_SECRET`)
- ✅ Настройте HTTPS в продакшене  
- ✅ Ограничьте доступ к порту webhook'а через firewall
- ✅ Регулярно обновляйте Docker образ (`whtg update`)
- ✅ Мониторьте логи на предмет подозрительной активности

### Проверка подписи
Бот автоматически проверяет HMAC-SHA256 подпись каждого webhook'а для защиты от поддельных запросов.

## 🔄 Автообновления

### CI/CD Pipeline
- GitHub Actions автоматически собирает новый Docker образ при каждом commit
- Образы публикуются в GitHub Container Registry
- Используйте `whtg update` для обновления до последней версии

### Rollback
```bash
# Откат к предыдущей версии  
docker-compose down
docker pull ghcr.io/dignezzz/webhook_tg_bot:previous-tag
docker-compose up -d
```

## 🤝 Поддержка

- **GitHub**: [dignezzz/webhook_tg_bot](https://github.com/dignezzz/webhook_tg_bot)
- **Issues**: [Сообщить о проблеме](https://github.com/dignezzz/webhook_tg_bot/issues)
- **Discussions**: [Обсуждения и вопросы](https://github.com/dignezzz/webhook_tg_bot/discussions)

## 📄 Лицензия

MIT License - см. [LICENSE](LICENSE) файл.

---

**💡 Совет**: Начните с базовой конфигурации, а затем постепенно добавляйте маппинг категорий и фильтры по мере необходимости.
